package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"

	common "github.com/replay-api/replay-api/pkg/domain"
	replay_out "github.com/replay-api/replay-api/pkg/domain/replay/ports/out"
)

// S3ContentAdapter implements ReplayFileContentWriter, ReplayFileContentReader,
// and ChunkedUploadManager using S3-compatible object storage (e.g., MinIO).
type S3ContentAdapter struct {
	client *s3.Client
	bucket string
}

// NewS3ContentAdapter creates a new S3-compatible storage adapter.
func NewS3ContentAdapter(cfg common.S3Config) (*S3ContentAdapter, error) {
	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = cfg.S3Endpoint // backward compat
	}

	resolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               endpoint,
				HostnameImmutable: true,
			}, nil
		},
	)

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithEndpointResolverWithOptions(resolver),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
	})

	bucket := cfg.Bucket
	if bucket == "" {
		bucket = "replays"
	}

	return &S3ContentAdapter{
		client: client,
		bucket: bucket,
	}, nil
}

// blobKey builds the S3 object key for a replay file.
func blobKey(replayFileID uuid.UUID) string {
	return replayFileID.String() + ".dem"
}

// --- ReplayFileContentWriter ---

// Put uploads replay file content to S3. Supports streaming via PutObject.
func (a *S3ContentAdapter) Put(ctx context.Context, replayFileID uuid.UUID, reader io.ReadSeeker) (string, error) {
	key := blobKey(replayFileID)

	_, err := reader.Seek(0, io.SeekStart)
	if err != nil {
		return "", fmt.Errorf("seek error: %w", err)
	}

	_, err = a.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(a.bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String("application/octet-stream"),
	})
	if err != nil {
		slog.ErrorContext(ctx, "S3ContentAdapter.Put: failed to upload", "key", key, "err", err)
		return "", fmt.Errorf("s3 put error: %w", err)
	}

	slog.InfoContext(ctx, "S3ContentAdapter.Put: uploaded successfully", "key", key)
	return key, nil
}

// PutStream uploads content from a non-seekable reader by buffering into memory.
// For large files, prefer using the chunked upload flow instead.
func (a *S3ContentAdapter) PutStream(ctx context.Context, replayFileID uuid.UUID, reader io.Reader, size int64) (string, error) {
	key := blobKey(replayFileID)

	_, err := a.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(a.bucket),
		Key:           aws.String(key),
		Body:          reader,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String("application/octet-stream"),
	})
	if err != nil {
		slog.ErrorContext(ctx, "S3ContentAdapter.PutStream: failed", "key", key, "err", err)
		return "", fmt.Errorf("s3 put stream error: %w", err)
	}

	slog.InfoContext(ctx, "S3ContentAdapter.PutStream: uploaded", "key", key)
	return key, nil
}

// --- ReplayFileContentReader ---

// GetByID downloads replay content from S3 and returns it as a ReadSeekCloser.
func (a *S3ContentAdapter) GetByID(ctx context.Context, replayFileID uuid.UUID) (io.ReadSeekCloser, error) {
	key := blobKey(replayFileID)

	output, err := a.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		slog.ErrorContext(ctx, "S3ContentAdapter.GetByID: failed to get object", "key", key, "err", err)
		return nil, fmt.Errorf("s3 get error: %w", err)
	}

	// Read into memory so we can provide ReadSeekCloser interface.
	// For very large files this could be optimized with temp file buffering,
	// but the demo parser already needs random access.
	data, err := io.ReadAll(output.Body)
	output.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("s3 read error: %w", err)
	}

	slog.InfoContext(ctx, "S3ContentAdapter.GetByID: downloaded", "key", key, "size", len(data))
	return &readSeekCloser{Reader: bytes.NewReader(data)}, nil
}

// readSeekCloser wraps bytes.Reader to implement io.ReadSeekCloser.
type readSeekCloser struct {
	*bytes.Reader
}

func (r *readSeekCloser) Close() error { return nil }

// --- ChunkedUploadManager ---

// InitiateMultipartUpload starts an S3 multipart upload and returns the upload ID.
func (a *S3ContentAdapter) InitiateMultipartUpload(ctx context.Context, replayFileID uuid.UUID) (string, error) {
	key := blobKey(replayFileID)

	output, err := a.client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket:      aws.String(a.bucket),
		Key:         aws.String(key),
		ContentType: aws.String("application/octet-stream"),
	})
	if err != nil {
		slog.ErrorContext(ctx, "S3ContentAdapter.InitiateMultipartUpload: failed", "key", key, "err", err)
		return "", fmt.Errorf("s3 initiate multipart error: %w", err)
	}

	uploadID := aws.ToString(output.UploadId)
	slog.InfoContext(ctx, "S3ContentAdapter.InitiateMultipartUpload: started", "key", key, "uploadID", uploadID)
	return uploadID, nil
}

// UploadPart uploads a single part of a multipart upload. Returns the ETag for completion.
func (a *S3ContentAdapter) UploadPart(ctx context.Context, replayFileID uuid.UUID, uploadID string, partNumber int32, data io.ReadSeeker) (string, error) {
	key := blobKey(replayFileID)

	output, err := a.client.UploadPart(ctx, &s3.UploadPartInput{
		Bucket:     aws.String(a.bucket),
		Key:        aws.String(key),
		UploadId:   aws.String(uploadID),
		PartNumber: aws.Int32(partNumber),
		Body:       data,
	})
	if err != nil {
		slog.ErrorContext(ctx, "S3ContentAdapter.UploadPart: failed",
			"key", key, "uploadID", uploadID, "partNumber", partNumber, "err", err)
		return "", fmt.Errorf("s3 upload part error: %w", err)
	}

	etag := aws.ToString(output.ETag)
	slog.InfoContext(ctx, "S3ContentAdapter.UploadPart: uploaded",
		"key", key, "partNumber", partNumber, "etag", etag)
	return etag, nil
}

// CompletePart is an alias for the port type, for backward compatibility.
type CompletePart = replay_out.UploadCompletePart

// CompleteMultipartUpload finalizes the multipart upload, assembling all parts.
func (a *S3ContentAdapter) CompleteMultipartUpload(ctx context.Context, replayFileID uuid.UUID, uploadID string, parts []replay_out.UploadCompletePart) (string, error) {
	key := blobKey(replayFileID)

	completedParts := make([]types.CompletedPart, len(parts))
	for i, p := range parts {
		completedParts[i] = types.CompletedPart{
			PartNumber: aws.Int32(p.PartNumber),
			ETag:       aws.String(p.ETag),
		}
	}

	_, err := a.client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(a.bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})
	if err != nil {
		slog.ErrorContext(ctx, "S3ContentAdapter.CompleteMultipartUpload: failed",
			"key", key, "uploadID", uploadID, "err", err)
		return "", fmt.Errorf("s3 complete multipart error: %w", err)
	}

	slog.InfoContext(ctx, "S3ContentAdapter.CompleteMultipartUpload: completed", "key", key)
	return key, nil
}

// AbortMultipartUpload cancels a multipart upload and cleans up uploaded parts.
func (a *S3ContentAdapter) AbortMultipartUpload(ctx context.Context, replayFileID uuid.UUID, uploadID string) error {
	key := blobKey(replayFileID)

	_, err := a.client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(a.bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
	})
	if err != nil {
		slog.ErrorContext(ctx, "S3ContentAdapter.AbortMultipartUpload: failed",
			"key", key, "uploadID", uploadID, "err", err)
		return fmt.Errorf("s3 abort multipart error: %w", err)
	}

	slog.InfoContext(ctx, "S3ContentAdapter.AbortMultipartUpload: aborted", "key", key)
	return nil
}

// --- Streaming Hash Utilities ---

// HashingReader wraps a reader and computes SHA256 hash as data flows through.
type HashingReader struct {
	reader io.Reader
	hash   hash.Hash
}

// NewHashingReader creates a reader that computes SHA256 while data is read.
func NewHashingReader(r io.Reader) *HashingReader {
	h := sha256.New()
	return &HashingReader{
		reader: io.TeeReader(r, h),
		hash:   h,
	}
}

// Read reads from the underlying reader while updating the hash.
func (h *HashingReader) Read(p []byte) (int, error) {
	return h.reader.Read(p)
}

// ContentHash returns the hex-encoded SHA256 hash. Must be called after all data is read.
func (h *HashingReader) ContentHash() string {
	return hex.EncodeToString(h.hash.Sum(nil))
}
