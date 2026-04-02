package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const chunkedUploadsCollection = "chunked_uploads"

// ChunkedUploadRepository implements ChunkedUploadWriter and ChunkedUploadReader ports.
type ChunkedUploadRepository struct {
	collection *mongo.Collection
}

func NewChunkedUploadRepository(client *mongo.Client, dbName string) *ChunkedUploadRepository {
	col := client.Database(dbName).Collection(chunkedUploadsCollection)

	// Create TTL index for automatic cleanup of expired uploads
	_, err := col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	})
	if err != nil {
		slog.Warn("failed to create TTL index on chunked_uploads", "err", err)
	}

	return &ChunkedUploadRepository{collection: col}
}

func (r *ChunkedUploadRepository) Create(ctx context.Context, upload *replay_entity.ChunkedUpload) error {
	_, err := r.collection.InsertOne(ctx, upload)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create chunked upload record", "uploadID", upload.ID, "err", err)
		return fmt.Errorf("create chunked upload: %w", err)
	}
	return nil
}

func (r *ChunkedUploadRepository) Update(ctx context.Context, upload *replay_entity.ChunkedUpload) error {
	filter := bson.M{"_id": upload.ID}
	update := bson.M{"$set": upload}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update chunked upload record", "uploadID", upload.ID, "err", err)
		return fmt.Errorf("update chunked upload: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("chunked upload not found: %v", upload.ID)
	}
	return nil
}

func (r *ChunkedUploadRepository) Delete(ctx context.Context, uploadID uuid.UUID) error {
	filter := bson.M{"_id": uploadID}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("delete chunked upload: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("chunked upload not found: %v", uploadID)
	}
	return nil
}

func (r *ChunkedUploadRepository) GetByID(ctx context.Context, uploadID uuid.UUID) (*replay_entity.ChunkedUpload, error) {
	filter := bson.M{"_id": uploadID}
	var upload replay_entity.ChunkedUpload
	if err := r.collection.FindOne(ctx, filter).Decode(&upload); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("chunked upload not found: %v", uploadID)
		}
		return nil, fmt.Errorf("get chunked upload: %w", err)
	}
	return &upload, nil
}

// AddPart atomically appends a chunk result to the upload's parts list using
// MongoDB $push. The filter includes a $ne guard on the part number so that
// duplicate uploads of the same part are rejected without a separate read.
func (r *ChunkedUploadRepository) AddPart(ctx context.Context, uploadID uuid.UUID, part replay_entity.ChunkResult) error {
	filter := bson.M{
		"_id":                          uploadID,
		"uploaded_parts.part_number":   bson.M{"$ne": part.PartNumber},
	}
	update := bson.M{
		"$push": bson.M{"uploaded_parts": part},
		"$set": bson.M{
			"status":     replay_entity.ChunkedUploadStatusUploading,
			"updated_at": time.Now().UTC(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to add part to chunked upload", "uploadID", uploadID, "partNumber", part.PartNumber, "err", err)
		return fmt.Errorf("add part to chunked upload: %w", err)
	}

	if result.MatchedCount == 0 {
		// Distinguish between "upload not found" and "duplicate part"
		exists, _ := r.GetByID(ctx, uploadID)
		if exists == nil {
			return fmt.Errorf("chunked upload not found: %v", uploadID)
		}
		return fmt.Errorf("part %d already uploaded", part.PartNumber)
	}

	return nil
}
