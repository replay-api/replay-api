package replay_out

import (
	"context"
	"io"

	"github.com/google/uuid"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
)

type ReplayParser interface {
	Parse(ctx context.Context, match uuid.UUID, content io.Reader, eventsChan chan *replay_entity.GameEvent) error
}

type GameEventWriter interface {
	CreateMany(createCtx context.Context, events []*replay_entity.GameEvent) error
	Create(createCtx context.Context, events *replay_entity.GameEvent) (*replay_entity.GameEvent, error)
}

type MatchMetadataWriter interface {
	Create(createCtx context.Context, match replay_entity.Match) error
	CreateMany(createCtx context.Context, matches []replay_entity.Match) error
	Update(ctx context.Context, match replay_entity.Match) error
	// FindOneAndUpsertBySlug atomically finds a match by slug or creates it if not found.
	// Returns the existing or newly created match, whether it was created, and any error.
	// Uses MongoDB FindOneAndUpdate with $setOnInsert for TOCTOU-safe atomic upsert.
	FindOneAndUpsertBySlug(ctx context.Context, slug string, match replay_entity.Match) (existing *replay_entity.Match, created bool, err error)
	// AppendSourceConfirmation atomically appends a source confirmation to a match
	// and updates conflict detection fields. Uses $push instead of full document $set.
	AppendSourceConfirmation(ctx context.Context, matchID uuid.UUID, confirmation replay_entity.SourceConfirmation, needsReview bool, conflictDetails string) error
}

type PlayerMetadataWriter interface {
	Create(createCtx context.Context, player replay_entity.PlayerMetadata) error
	CreateMany(createCtx context.Context, players []replay_entity.PlayerMetadata) error
}

type ReplayFileMetadataWriter interface {
	Create(createCtx context.Context, replayFile *replay_entity.ReplayFile) (*replay_entity.ReplayFile, error)
	Update(createCtx context.Context, replayFile *replay_entity.ReplayFile) (*replay_entity.ReplayFile, error)
}

type ReplayFileContentWriter interface {
	Put(createCtx context.Context, replayFileID uuid.UUID, reader io.ReadSeeker) (string, error)
}

// ChunkedUploadManager handles multipart/chunked uploads for large replay files.
// Implementations map to S3 multipart upload API or equivalent.
type ChunkedUploadManager interface {
	// InitiateMultipartUpload starts a new chunked upload session.
	// Returns an uploadID that must be used for subsequent part uploads.
	InitiateMultipartUpload(ctx context.Context, replayFileID uuid.UUID) (uploadID string, err error)

	// UploadPart uploads a single chunk. partNumber is 1-based.
	// Returns an ETag identifying the uploaded part (needed for completion).
	UploadPart(ctx context.Context, replayFileID uuid.UUID, uploadID string, partNumber int32, data io.ReadSeeker) (etag string, err error)

	// CompleteMultipartUpload finalizes the upload, assembling all parts into the final object.
	// parts must contain all previously uploaded part numbers and their ETags.
	CompleteMultipartUpload(ctx context.Context, replayFileID uuid.UUID, uploadID string, parts []UploadCompletePart) (internalURI string, err error)

	// AbortMultipartUpload cancels an in-progress upload and cleans up any uploaded parts.
	AbortMultipartUpload(ctx context.Context, replayFileID uuid.UUID, uploadID string) error
}

// UploadCompletePart represents a completed part for multipart upload finalization.
type UploadCompletePart struct {
	PartNumber int32  `json:"part_number"`
	ETag       string `json:"etag"`
}

// StreamingContentWriter extends ReplayFileContentWriter with non-seekable stream support.
type StreamingContentWriter interface {
	ReplayFileContentWriter
	// PutStream uploads content from a non-seekable reader with known size.
	PutStream(ctx context.Context, replayFileID uuid.UUID, reader io.Reader, size int64) (string, error)
}

// ChunkedUploadWriter persists chunked upload session state.
type ChunkedUploadWriter interface {
	Create(ctx context.Context, upload *replay_entity.ChunkedUpload) error
	Update(ctx context.Context, upload *replay_entity.ChunkedUpload) error
	Delete(ctx context.Context, uploadID uuid.UUID) error
	// AddPart atomically appends a chunk result to the upload's parts list.
	// Returns an error if the part number was already uploaded (duplicate guard).
	AddPart(ctx context.Context, uploadID uuid.UUID, part replay_entity.ChunkResult) error
}

// ChunkedUploadReader retrieves chunked upload session state.
type ChunkedUploadReader interface {
	GetByID(ctx context.Context, uploadID uuid.UUID) (*replay_entity.ChunkedUpload, error)
}

type ShareTokenWriter interface {
	Create(ctx context.Context, token *replay_entity.ShareToken) error
	Update(ctx context.Context, token *replay_entity.ShareToken) error
	Delete(ctx context.Context, tokenID uuid.UUID) error
}

// ReplayEventPublisher publishes replay-related events to message broker (Kafka)
type ReplayEventPublisher interface {
	// PublishReplayUploaded publishes an event when a replay file is uploaded
	PublishReplayUploaded(ctx context.Context, replayFile *replay_entity.ReplayFile) error
	// PublishReplayProcessing publishes progress events during replay processing
	PublishReplayProcessing(ctx context.Context, replayFileID uuid.UUID, stage string, progress int, eventCount int, playerCount int) error
	// PublishReplayCompleted publishes an event when replay processing completes successfully
	PublishReplayCompleted(ctx context.Context, replayFile *replay_entity.ReplayFile, matchID uuid.UUID, eventCount int, playerCount int, processingDurationMs int64) error
	// PublishReplayFailed publishes an event when replay processing fails
	PublishReplayFailed(ctx context.Context, replayFile *replay_entity.ReplayFile, stage string, errorType string, errorMessage string, retryable bool, retryCount int) error
}
