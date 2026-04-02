package use_cases

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"log/slog"

	shared "github.com/resource-ownership/go-common/pkg/common"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_out "github.com/replay-api/replay-api/pkg/domain/replay/ports/out"
)

type UploadReplayFileUseCase struct {
	MetadataReader replay_out.ReplayFileMetadataReader
	MetadataWriter replay_out.ReplayFileMetadataWriter
	ContentWriter  replay_out.ReplayFileContentWriter
	EventPublisher replay_out.ReplayEventPublisher
}

func NewUploadReplayFileUseCase(metadataReader replay_out.ReplayFileMetadataReader, metadataWriter replay_out.ReplayFileMetadataWriter, dataCommand replay_out.ReplayFileContentWriter, eventPublisher replay_out.ReplayEventPublisher) *UploadReplayFileUseCase {
	return &UploadReplayFileUseCase{
		MetadataReader: metadataReader,
		MetadataWriter: metadataWriter,
		ContentWriter:  dataCommand,
		EventPublisher: eventPublisher,
	}
}

// calculateContentHash computes SHA256 hash of the file content
func calculateContentHash(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

// streamingHash reads content through a SHA256 hasher and returns the buffered bytes + hex hash.
// This avoids reading the file twice (once for hash, once for upload) while keeping memory bounded
// to a single copy of the file content.
func streamingHash(reader io.Reader) ([]byte, string, error) {
	h := sha256.New()
	var buf bytes.Buffer
	tee := io.TeeReader(reader, h)
	if _, err := io.Copy(&buf, tee); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), hex.EncodeToString(h.(hash.Hash).Sum(nil)), nil
}

// Exec uploads a replay file and creates associated metadata with default options.
func (usecase *UploadReplayFileUseCase) Exec(ctx context.Context, reader io.Reader) (*replay_entity.ReplayFile, error) {
	return usecase.ExecWithOptions(ctx, reader, nil)
}

// ExecWithOptions uploads a replay file with optional metadata (title, description, visibility, tags).
//
// This use case handles:
//  1. Authentication verification - allows both authenticated and guest uploads
//  2. Reading replay file content from the provided reader
//  3. Deduplication check using SHA256 content hash:
//     - Same user uploading same file = upsert (update existing)
//     - Different user uploading same file = create reference to original
//  4. Creating replay metadata entry with initial "processing" status
//  5. Uploading file content to blob storage
//  6. Updating metadata with storage URI and final status
//
// Parameters:
//   - ctx: Context containing authentication and resource ownership
//   - reader: io.Reader providing the replay file content
//   - opts: Optional metadata (title, description, visibility, tags)
//
// Returns:
//   - *ReplayFile: Created replay file metadata with storage URI
//   - error: Returns storage/DB errors
func (usecase *UploadReplayFileUseCase) ExecWithOptions(ctx context.Context, reader io.Reader, opts *replay_entity.ReplayFileOptions) (*replay_entity.ReplayFile, error) {
	// Check authentication - allow guest uploads
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	resourceOwner := shared.GetResourceOwner(ctx)
	
	// For guest uploads, use the resource owner set by ResourceContextMiddleware
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		slog.InfoContext(ctx, "allowing guest replay upload", "resourceOwner", resourceOwner)
	} else {
		slog.InfoContext(ctx, "authenticated user replay upload", "resourceOwner", resourceOwner)
	}

	// Stream content through SHA256 hasher — single-pass read, no double buffering
	file, contentHash, err := streamingHash(reader)
	if err != nil {
		slog.ErrorContext(ctx, "error reading replay file", "err", err)
		return nil, err
	}

	slog.InfoContext(ctx, "uploading replay file", "size", len(file), "hash", contentHash)

	// Check for existing replay with same content hash (deduplication)
	existingReplay, err := usecase.MetadataReader.FindByContentHash(ctx, contentHash)
	if err != nil {
		slog.WarnContext(ctx, "error checking for duplicate replay", "err", err)
		// Continue with upload - don't fail due to deduplication check error
	}

	// Get game_id from context
	gameIDValue := ctx.Value(shared.GameIDParamKey)
	var gameID string
	if gameIDValue != nil {
		gameID = gameIDValue.(string)
	}
	if gameID == "" {
		gameID = string(replay_common.CS2_GAME_ID) // default fallback to cs2
	}

	// Handle deduplication scenarios
	if existingReplay != nil {
		slog.InfoContext(ctx, "found existing replay with same content hash", 
			"existingReplayID", existingReplay.ID,
			"existingOwnerUserID", existingReplay.ResourceOwner.UserID,
			"currentUserID", resourceOwner.UserID)

		// Check if same user is uploading
		isSameUser := existingReplay.ResourceOwner.UserID == resourceOwner.UserID

		if isSameUser {
			// Same user uploading same file = UPSERT (update existing)
			slog.InfoContext(ctx, "same user uploading duplicate - performing upsert", "replayID", existingReplay.ID)
			
			// Update metadata if new options provided
			if opts != nil {
				if opts.Title != "" {
					existingReplay.Title = opts.Title
				}
				if opts.Description != "" {
					existingReplay.Description = opts.Description
				}
				if len(opts.Tags) > 0 {
					existingReplay.Tags = opts.Tags
				}
				if opts.Visibility != 0 {
					existingReplay.VisibilityType = opts.Visibility
				}
			}

			// Update the existing replay
			updatedReplay, err := usecase.MetadataWriter.Update(ctx, existingReplay)
			if err != nil {
				slog.ErrorContext(ctx, "error updating existing replay during upsert", "err", err)
				return nil, err
			}

			slog.InfoContext(ctx, "successfully upserted existing replay", "replayID", updatedReplay.ID)
			return updatedReplay, nil
		} else {
			// Different user uploading same file = CREATE REFERENCE
			slog.InfoContext(ctx, "different user uploading duplicate - creating reference", 
				"originalReplayID", existingReplay.ID)

			// Create a new replay that references the original
			// This allows the user to have their own copy with their own metadata
			originalID := existingReplay.ID
			replayOpts := &replay_entity.ReplayFileOptions{
				ContentHash:      contentHash,
				OriginalReplayID: &originalID,
			}

			// Merge with provided options
			if opts != nil {
				replayOpts.Title = opts.Title
				replayOpts.Description = opts.Description
				replayOpts.Tags = opts.Tags
				replayOpts.Visibility = opts.Visibility
			}

			// Create new entity referencing the original
			entity := replay_entity.NewReplayFileWithOptions(
				replay_common.GameIDKey(gameID),
				replay_common.NetworkIDKey("steam"),
				len(file),
				existingReplay.InternalURI, // Reference same storage URI
				resourceOwner,
				replayOpts,
			)
			
			// Copy status from original since content already processed
			entity.Status = existingReplay.Status
			entity.Header = existingReplay.Header

			replayFile, err := usecase.MetadataWriter.Create(ctx, entity)
			if err != nil {
				slog.ErrorContext(ctx, "error creating reference replay", "err", err)
				return nil, err
			}

			slog.InfoContext(ctx, "created reference to existing replay", 
				"newReplayID", replayFile.ID,
				"originalReplayID", originalID)

			return replayFile, nil
		}
	}

	// No duplicate found - create new replay
	if opts == nil {
		opts = &replay_entity.ReplayFileOptions{}
	}
	opts.ContentHash = contentHash

	// Create Metadata with options (visibility, title, description, tags, hash)
	entity := replay_entity.NewReplayFileWithOptions(
		replay_common.GameIDKey(gameID),
		replay_common.NetworkIDKey("steam"),
		len(file),
		"",
		resourceOwner,
		opts,
	)
	replayFile, err := usecase.MetadataWriter.Create(ctx, entity)

	if err != nil {
		slog.ErrorContext(ctx, "error creating new replay metadata", "err", err)
		return nil, err
	}

	slog.InfoContext(ctx, "created new replay metadata", "replayFile", replayFile, "visibility", replayFile.VisibilityType)

	// Put Contents into Blob Store
	uri, err := usecase.ContentWriter.Put(ctx, replayFile.ID, bytes.NewReader(file))
	if err != nil {
		replayFile.Status = replay_entity.ReplayFileStatusFailed
		replayFile.Error = err.Error()
		_, _ = usecase.MetadataWriter.Update(ctx, replayFile)
		slog.ErrorContext(ctx, "error uploading replay data", "err", err, "replayFile", replayFile)
		return nil, err
	}

	slog.InfoContext(ctx, "uploaded replay data", "replayFile", replayFile, "uri", uri)

	// Update Metadata
	replayFile.InternalURI = uri
	replayFile.Status = replay_entity.ReplayFileStatusProcessing
	replayFile, err = usecase.MetadataWriter.Update(ctx, replayFile)

	if err != nil {
		slog.ErrorContext(ctx, "error updating uploaded replay metadata", "replayFile", replayFile, "err", err)
		return nil, err
	}

	// Publish replay uploaded event for async processing via Kafka
	if usecase.EventPublisher != nil {
		if err := usecase.EventPublisher.PublishReplayUploaded(ctx, replayFile); err != nil {
			// Log error but don't fail the upload - the replay can be processed later
			slog.WarnContext(ctx, "failed to publish replay uploaded event", "replayFile", replayFile.ID, "err", err)
		} else {
			slog.InfoContext(ctx, "published replay uploaded event for async processing", "replayFile", replayFile.ID)
		}
	}

	// return updated metadata
	return replayFile, nil
}
