package use_cases

import (
	"bytes"
	"context"
	"io"
	"log/slog"

	shared "github.com/resource-ownership/go-common/pkg/common"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_out "github.com/replay-api/replay-api/pkg/domain/replay/ports/out"
)

type UploadReplayFileUseCase struct {
	MetadataWriter replay_out.ReplayFileMetadataWriter
	ContentWriter  replay_out.ReplayFileContentWriter
}

func NewUploadReplayFileUseCase(metadataWriter replay_out.ReplayFileMetadataWriter, dataCommand replay_out.ReplayFileContentWriter) *UploadReplayFileUseCase {
	return &UploadReplayFileUseCase{
		MetadataWriter: metadataWriter,
		ContentWriter:  dataCommand,
	}
}

// Exec uploads a replay file and creates associated metadata.
//
// This use case handles:
//  1. Authentication verification - user must be authenticated
//  2. Reading replay file content from the provided reader
//  3. Creating replay metadata entry with initial "processing" status
//  4. Uploading file content to blob storage
//  5. Updating metadata with storage URI and final status
//
// Parameters:
//   - ctx: Context containing authentication and resource ownership
//   - reader: io.Reader providing the replay file content
//
// Returns:
//   - *ReplayFile: Created replay file metadata with storage URI
//   - error: Returns ErrUnauthorized if not authenticated, or storage/DB errors
func (usecase *UploadReplayFileUseCase) Exec(ctx context.Context, reader io.Reader) (*replay_entity.ReplayFile, error) {
	// Authentication check - user must be logged in to upload replays
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		slog.WarnContext(ctx, "unauthorized replay upload attempt")
		return nil, shared.NewErrUnauthorized()
	}

	file, err := io.ReadAll(reader)
	if err != nil {
		slog.ErrorContext(ctx, "error reading replay file", "err", err)
		return nil, err
	}

	slog.InfoContext(ctx, "uploading replay file", "size", len(file))

	// create Metadata
	entity := replay_entity.NewReplayFile("cs", "steam", len(file), "", shared.GetResourceOwner(ctx))
	replayFile, err := usecase.MetadataWriter.Create(ctx, entity)

	if err != nil {
		slog.ErrorContext(ctx, "error creating new replay metadata", "err", err)
		return nil, err
	}

	slog.InfoContext(ctx, "created new replay metadata", "replayFile", replayFile)

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

	// return updated metadata
	return replayFile, nil
}
