package entities

import (
	"time"

	"github.com/google/uuid"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// ChunkedUploadStatus tracks the state of a chunked upload session.
type ChunkedUploadStatus string

const (
	ChunkedUploadStatusInitiated  ChunkedUploadStatus = "Initiated"
	ChunkedUploadStatusUploading  ChunkedUploadStatus = "Uploading"
	ChunkedUploadStatusCompleting ChunkedUploadStatus = "Completing"
	ChunkedUploadStatusCompleted  ChunkedUploadStatus = "Completed"
	ChunkedUploadStatusAborted    ChunkedUploadStatus = "Aborted"
	ChunkedUploadStatusFailed     ChunkedUploadStatus = "Failed"
)

// ChunkedUpload represents an in-progress chunked upload session.
type ChunkedUpload struct {
	ID             uuid.UUID              `json:"id" bson:"_id"`
	ReplayFileID   uuid.UUID              `json:"replay_file_id" bson:"replay_file_id"`
	GameID         replay_common.GameIDKey `json:"game_id" bson:"game_id"`
	FileName       string                 `json:"file_name" bson:"file_name"`
	FileSize       int64                  `json:"file_size" bson:"file_size"`
	ChunkSize      int64                  `json:"chunk_size" bson:"chunk_size"`
	TotalChunks    int32                  `json:"total_chunks" bson:"total_chunks"`
	S3UploadID     string                 `json:"s3_upload_id" bson:"s3_upload_id"`
	Status         ChunkedUploadStatus    `json:"status" bson:"status"`
	UploadedParts  []ChunkResult          `json:"uploaded_parts" bson:"uploaded_parts"`
	Options        *ReplayFileOptions     `json:"options,omitempty" bson:"options,omitempty"`
	ResourceOwner  shared.ResourceOwner   `json:"resource_owner" bson:"resource_owner"`
	CreatedAt      time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" bson:"updated_at"`
	ExpiresAt      time.Time              `json:"expires_at" bson:"expires_at"`
	Error          string                 `json:"error,omitempty" bson:"error,omitempty"`
}

// ChunkResult contains the result of uploading a single chunk.
type ChunkResult struct {
	PartNumber int32  `json:"part_number" bson:"part_number"`
	ETag       string `json:"etag" bson:"etag"`
	Size       int64  `json:"size" bson:"size"`
}

// DefaultChunkSize is 5MB — the minimum for S3 multipart upload parts (except the last part).
const DefaultChunkSize int64 = 5 * 1024 * 1024

// MaxFileSize is 500MB — maximum allowed replay file size.
const MaxFileSize int64 = 500 * 1024 * 1024

// ChunkedUploadTTL is the duration after which incomplete uploads are automatically cleaned up.
const ChunkedUploadTTL = 24 * time.Hour

// NewChunkedUpload creates a new chunked upload session.
func NewChunkedUpload(gameID replay_common.GameIDKey, fileName string, fileSize int64, resourceOwner shared.ResourceOwner, opts *ReplayFileOptions) *ChunkedUpload {
	chunkSize := DefaultChunkSize
	totalChunks := int32(fileSize / chunkSize)
	if fileSize%chunkSize != 0 {
		totalChunks++
	}

	now := time.Now().UTC()
	replayFileID := uuid.New()

	return &ChunkedUpload{
		ID:            uuid.New(),
		ReplayFileID:  replayFileID,
		GameID:        gameID,
		FileName:      fileName,
		FileSize:      fileSize,
		ChunkSize:     chunkSize,
		TotalChunks:   totalChunks,
		Status:        ChunkedUploadStatusInitiated,
		UploadedParts: make([]ChunkResult, 0),
		Options:       opts,
		ResourceOwner: resourceOwner,
		CreatedAt:     now,
		UpdatedAt:     now,
		ExpiresAt:     now.Add(ChunkedUploadTTL),
	}
}

// IsExpired returns true if the upload session has exceeded its TTL.
func (cu *ChunkedUpload) IsExpired() bool {
	return time.Now().UTC().After(cu.ExpiresAt)
}

// AllPartsUploaded returns true if all expected parts have been received.
func (cu *ChunkedUpload) AllPartsUploaded() bool {
	return int32(len(cu.UploadedParts)) >= cu.TotalChunks
}

// UploadedBytes returns the total bytes uploaded so far.
func (cu *ChunkedUpload) UploadedBytes() int64 {
	var total int64
	for _, p := range cu.UploadedParts {
		total += p.Size
	}
	return total
}
