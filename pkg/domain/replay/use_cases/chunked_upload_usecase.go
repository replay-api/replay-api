package use_cases

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_out "github.com/replay-api/replay-api/pkg/domain/replay/ports/out"
)

// InitiateChunkedUploadUseCase creates a new chunked upload session.
type InitiateChunkedUploadUseCase struct {
	UploadManager  replay_out.ChunkedUploadManager
	UploadWriter   replay_out.ChunkedUploadWriter
	MetadataWriter replay_out.ReplayFileMetadataWriter
}

func (uc *InitiateChunkedUploadUseCase) Exec(ctx context.Context, gameID string, fileName string, fileSize int64, opts *replay_entity.ReplayFileOptions) (*replay_entity.ChunkedUpload, error) {
	if fileSize <= 0 || fileSize > replay_entity.MaxFileSize {
		return nil, fmt.Errorf("file size must be between 1 byte and %d bytes", replay_entity.MaxFileSize)
	}

	if gameID == "" {
		gameID = string(replay_common.CS2_GAME_ID)
	}

	resourceOwner := shared.GetResourceOwner(ctx)

	// Create chunked upload tracking record
	upload := replay_entity.NewChunkedUpload(
		replay_common.GameIDKey(gameID),
		fileName,
		fileSize,
		resourceOwner,
		opts,
	)

	// Create replay file metadata entry (status = Pending)
	entity := replay_entity.NewReplayFileWithOptions(
		replay_common.GameIDKey(gameID),
		replay_common.NetworkIDKey("steam"),
		int(fileSize),
		"",
		resourceOwner,
		opts,
	)
	entity.Status = replay_entity.ReplayFileStatusPending
	entity.ID = upload.ReplayFileID

	replayFile, err := uc.MetadataWriter.Create(ctx, entity)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create replay metadata for chunked upload", "err", err)
		return nil, err
	}

	// Initiate multipart upload in S3
	s3UploadID, err := uc.UploadManager.InitiateMultipartUpload(ctx, replayFile.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to initiate multipart upload", "replayFileID", replayFile.ID, "err", err)
		return nil, err
	}

	upload.S3UploadID = s3UploadID

	if err := uc.UploadWriter.Create(ctx, upload); err != nil {
		// Best-effort abort the S3 upload
		_ = uc.UploadManager.AbortMultipartUpload(ctx, replayFile.ID, s3UploadID)
		slog.ErrorContext(ctx, "failed to persist chunked upload state", "err", err)
		return nil, err
	}

	slog.InfoContext(ctx, "initiated chunked upload",
		"uploadID", upload.ID,
		"replayFileID", replayFile.ID,
		"fileSize", fileSize,
		"totalChunks", upload.TotalChunks,
		"chunkSize", upload.ChunkSize,
	)

	return upload, nil
}

// UploadChunkUseCase handles uploading a single chunk/part.
type UploadChunkUseCase struct {
	UploadManager replay_out.ChunkedUploadManager
	UploadReader  replay_out.ChunkedUploadReader
	UploadWriter  replay_out.ChunkedUploadWriter
}

func (uc *UploadChunkUseCase) Exec(ctx context.Context, uploadID uuid.UUID, partNumber int32, data io.ReadSeeker) (*replay_entity.ChunkResult, error) {
	upload, err := uc.UploadReader.GetByID(ctx, uploadID)
	if err != nil {
		return nil, fmt.Errorf("upload session not found: %w", err)
	}

	// Verify ownership
	currentOwner := shared.GetResourceOwner(ctx)
	if currentOwner.UserID == uuid.Nil || currentOwner.UserID != upload.ResourceOwner.UserID {
		return nil, fmt.Errorf("forbidden: not the owner of this upload session")
	}

	if upload.IsExpired() {
		return nil, fmt.Errorf("upload session expired")
	}

	if upload.Status == replay_entity.ChunkedUploadStatusCompleted ||
		upload.Status == replay_entity.ChunkedUploadStatusAborted {
		return nil, fmt.Errorf("upload session is %s", upload.Status)
	}

	if partNumber < 1 || partNumber > upload.TotalChunks {
		return nil, fmt.Errorf("part number %d out of range [1, %d]", partNumber, upload.TotalChunks)
	}

	// Upload part to S3
	etag, err := uc.UploadManager.UploadPart(ctx, upload.ReplayFileID, upload.S3UploadID, partNumber, data)
	if err != nil {
		slog.ErrorContext(ctx, "failed to upload chunk to S3", "uploadID", uploadID, "partNumber", partNumber, "err", err)
		return nil, err
	}

	// Atomically track the uploaded part (rejects duplicates)
	result := &replay_entity.ChunkResult{
		PartNumber: partNumber,
		ETag:       etag,
	}

	if err := uc.UploadWriter.AddPart(ctx, uploadID, *result); err != nil {
		slog.ErrorContext(ctx, "failed to record chunk in upload state", "uploadID", uploadID, "partNumber", partNumber, "err", err)
		return nil, err
	}

	slog.InfoContext(ctx, "uploaded chunk",
		"uploadID", uploadID,
		"partNumber", partNumber,
		"totalChunks", upload.TotalChunks,
	)

	return result, nil
}

// CompleteChunkedUploadUseCase finalizes the upload and triggers processing.
type CompleteChunkedUploadUseCase struct {
	UploadManager  replay_out.ChunkedUploadManager
	UploadReader   replay_out.ChunkedUploadReader
	UploadWriter   replay_out.ChunkedUploadWriter
	MetadataWriter replay_out.ReplayFileMetadataWriter
	MetadataReader replay_out.ReplayFileMetadataReader
	EventPublisher replay_out.ReplayEventPublisher
}

func (uc *CompleteChunkedUploadUseCase) Exec(ctx context.Context, uploadID uuid.UUID) (*replay_entity.ReplayFile, error) {
	upload, err := uc.UploadReader.GetByID(ctx, uploadID)
	if err != nil {
		return nil, fmt.Errorf("upload session not found: %w", err)
	}

	// Verify ownership
	currentOwner := shared.GetResourceOwner(ctx)
	if currentOwner.UserID == uuid.Nil || currentOwner.UserID != upload.ResourceOwner.UserID {
		return nil, fmt.Errorf("forbidden: not the owner of this upload session")
	}

	if upload.IsExpired() {
		return nil, fmt.Errorf("upload session expired")
	}

	if !upload.AllPartsUploaded() {
		return nil, fmt.Errorf("not all parts uploaded: got %d, need %d", len(upload.UploadedParts), upload.TotalChunks)
	}

	// Convert entity parts to port's UploadCompletePart for S3 API
	parts := make([]replay_out.UploadCompletePart, len(upload.UploadedParts))
	for i, p := range upload.UploadedParts {
		parts[i] = replay_out.UploadCompletePart{
			PartNumber: p.PartNumber,
			ETag:       p.ETag,
		}
	}

	// Complete multipart upload in S3
	upload.Status = replay_entity.ChunkedUploadStatusCompleting
	_ = uc.UploadWriter.Update(ctx, upload)

	internalURI, err := uc.UploadManager.CompleteMultipartUpload(ctx, upload.ReplayFileID, upload.S3UploadID, parts)
	if err != nil {
		upload.Status = replay_entity.ChunkedUploadStatusFailed
		_ = uc.UploadWriter.Update(ctx, upload)
		slog.ErrorContext(ctx, "failed to complete multipart upload", "uploadID", uploadID, "err", err)
		return nil, err
	}

	// Update upload state
	upload.Status = replay_entity.ChunkedUploadStatusCompleted
	_ = uc.UploadWriter.Update(ctx, upload)

	// Update replay file metadata with S3 URI + set to Processing
	replayFile, err := uc.findReplayFile(ctx, upload.ReplayFileID)
	if err != nil {
		return nil, err
	}

	replayFile.InternalURI = internalURI
	replayFile.Status = replay_entity.ReplayFileStatusProcessing
	replayFile, err = uc.MetadataWriter.Update(ctx, replayFile)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update replay file after chunked upload complete", "err", err)
		return nil, err
	}

	// Publish event for async processing via Kafka
	if uc.EventPublisher != nil {
		if err := uc.EventPublisher.PublishReplayUploaded(ctx, replayFile); err != nil {
			slog.WarnContext(ctx, "failed to publish replay uploaded event after chunked upload", "replayFileID", replayFile.ID, "err", err)
		}
	}

	slog.InfoContext(ctx, "completed chunked upload",
		"uploadID", uploadID,
		"replayFileID", replayFile.ID,
		"internalURI", internalURI,
	)

	return replayFile, nil
}

func (uc *CompleteChunkedUploadUseCase) findReplayFile(ctx context.Context, replayFileID uuid.UUID) (*replay_entity.ReplayFile, error) {
	valueParams := []shared.SearchableValue{
		{Field: "ID", Values: []interface{}{replayFileID}, Operator: shared.EqualsOperator},
	}
	search := shared.NewSearchByValues(ctx, valueParams, shared.SearchResultOptions{Limit: 1}, shared.ClientApplicationAudienceIDKey)
	results, err := uc.MetadataReader.Search(ctx, search)
	if err != nil || len(results) == 0 {
		return nil, fmt.Errorf("replay file not found: %v", replayFileID)
	}
	return &results[0], nil
}

// AbortChunkedUploadUseCase cancels an in-progress upload.
type AbortChunkedUploadUseCase struct {
	UploadManager  replay_out.ChunkedUploadManager
	UploadReader   replay_out.ChunkedUploadReader
	UploadWriter   replay_out.ChunkedUploadWriter
	MetadataWriter replay_out.ReplayFileMetadataWriter
	MetadataReader replay_out.ReplayFileMetadataReader
}

func (uc *AbortChunkedUploadUseCase) Exec(ctx context.Context, uploadID uuid.UUID) error {
	upload, err := uc.UploadReader.GetByID(ctx, uploadID)
	if err != nil {
		return fmt.Errorf("upload session not found: %w", err)
	}

	// Verify ownership
	currentOwner := shared.GetResourceOwner(ctx)
	if currentOwner.UserID == uuid.Nil || currentOwner.UserID != upload.ResourceOwner.UserID {
		return fmt.Errorf("forbidden: not the owner of this upload session")
	}

	// Abort S3 multipart upload
	if err := uc.UploadManager.AbortMultipartUpload(ctx, upload.ReplayFileID, upload.S3UploadID); err != nil {
		slog.WarnContext(ctx, "failed to abort S3 multipart upload (may already be completed/aborted)", "err", err)
	}

	// Update upload state
	upload.Status = replay_entity.ChunkedUploadStatusAborted
	_ = uc.UploadWriter.Update(ctx, upload)

	// Mark replay file as failed
	replayFile, err := uc.findReplayFile(ctx, upload.ReplayFileID)
	if err == nil && replayFile != nil {
		replayFile.Status = replay_entity.ReplayFileStatusFailed
		replayFile.Error = "upload aborted by user"
		_, _ = uc.MetadataWriter.Update(ctx, replayFile)
	}

	slog.InfoContext(ctx, "aborted chunked upload", "uploadID", uploadID, "replayFileID", upload.ReplayFileID)
	return nil
}

func (uc *AbortChunkedUploadUseCase) findReplayFile(ctx context.Context, replayFileID uuid.UUID) (*replay_entity.ReplayFile, error) {
	valueParams := []shared.SearchableValue{
		{Field: "ID", Values: []interface{}{replayFileID}, Operator: shared.EqualsOperator},
	}
	search := shared.NewSearchByValues(ctx, valueParams, shared.SearchResultOptions{Limit: 1}, shared.ClientApplicationAudienceIDKey)
	results, err := uc.MetadataReader.Search(ctx, search)
	if err != nil || len(results) == 0 {
		return nil, fmt.Errorf("replay file not found: %v", replayFileID)
	}
	return &results[0], nil
}
