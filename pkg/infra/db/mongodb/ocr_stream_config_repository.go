package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	oracle_entities "github.com/replay-api/replay-api/pkg/domain/oracle/entities"
	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoOCRStreamConfigRepository implements the OCRStreamConfigRepository port using MongoDB.
type MongoOCRStreamConfigRepository struct {
	collection *mongo.Collection
}

// Compile-time interface satisfaction check
var _ oracle_out.OCRStreamConfigRepository = (*MongoOCRStreamConfigRepository)(nil)

// NewMongoOCRStreamConfigRepository creates a new MongoDB-backed OCR stream config repository
func NewMongoOCRStreamConfigRepository(mongoClient *mongo.Client, dbName string) oracle_out.OCRStreamConfigRepository {
	collection := mongoClient.Database(dbName).Collection("ocr_stream_configs")

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "external_match_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys:    bson.D{{Key: "video_id", Value: 1}},
			Options: options.Index().SetSparse(true),
		},
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "game_id", Value: 1},
			},
		},
		{
			Keys: bson.D{{Key: "oracle_result_id", Value: 1}},
			Options: options.Index().SetSparse(true),
		},
	}

	_, err := collection.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		slog.Warn("failed to create ocr_stream_configs indexes", slog.String("error", err.Error()))
	}

	return &MongoOCRStreamConfigRepository{collection: collection}
}

func (r *MongoOCRStreamConfigRepository) Save(ctx context.Context, config *oracle_entities.OCRStreamConfig) error {
	if config.ID == uuid.Nil {
		config.ID = uuid.New()
	}

	_, err := r.collection.InsertOne(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to save OCR stream config: %w", err)
	}

	return nil
}

func (r *MongoOCRStreamConfigRepository) FindByID(ctx context.Context, id uuid.UUID) (*oracle_entities.OCRStreamConfig, error) {
	var config oracle_entities.OCRStreamConfig
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find OCR stream config: %w", err)
	}

	return &config, nil
}

func (r *MongoOCRStreamConfigRepository) FindByExternalMatchID(ctx context.Context, externalMatchID string) (*oracle_entities.OCRStreamConfig, error) {
	var config oracle_entities.OCRStreamConfig
	err := r.collection.FindOne(ctx, bson.M{"external_match_id": externalMatchID}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find OCR stream config by external match ID: %w", err)
	}

	return &config, nil
}

func (r *MongoOCRStreamConfigRepository) FindByVideoID(ctx context.Context, videoID string) (*oracle_entities.OCRStreamConfig, error) {
	var config oracle_entities.OCRStreamConfig
	err := r.collection.FindOne(ctx, bson.M{"video_id": videoID}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find OCR stream config by video ID: %w", err)
	}

	return &config, nil
}

func (r *MongoOCRStreamConfigRepository) FindByStatus(ctx context.Context, status oracle_entities.OCRStreamStatus, limit int) ([]*oracle_entities.OCRStreamConfig, error) {
	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.D{{Key: "_id", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{"status": string(status)}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find OCR stream configs by status: %w", err)
	}
	defer cursor.Close(ctx)

	var configs []*oracle_entities.OCRStreamConfig
	if err := cursor.All(ctx, &configs); err != nil {
		return nil, fmt.Errorf("failed to decode OCR stream configs: %w", err)
	}

	return configs, nil
}

func (r *MongoOCRStreamConfigRepository) FindPending(ctx context.Context, limit int) ([]*oracle_entities.OCRStreamConfig, error) {
	return r.FindByStatus(ctx, oracle_entities.OCRStreamStatusPending, limit)
}

func (r *MongoOCRStreamConfigRepository) FindByGameID(ctx context.Context, gameID replay_common.GameIDKey, limit int) ([]*oracle_entities.OCRStreamConfig, error) {
	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.D{{Key: "_id", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{"game_id": string(gameID)}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find OCR stream configs by game ID: %w", err)
	}
	defer cursor.Close(ctx)

	var configs []*oracle_entities.OCRStreamConfig
	if err := cursor.All(ctx, &configs); err != nil {
		return nil, fmt.Errorf("failed to decode OCR stream configs: %w", err)
	}

	return configs, nil
}

func (r *MongoOCRStreamConfigRepository) Update(ctx context.Context, config *oracle_entities.OCRStreamConfig) error {
	now := time.Now().UTC()
	_ = now

	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": config.ID}, config)
	if err != nil {
		return fmt.Errorf("failed to update OCR stream config: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("OCR stream config not found: %s", config.ID)
	}

	return nil
}

func (r *MongoOCRStreamConfigRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete OCR stream config: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("OCR stream config not found: %s", id)
	}

	return nil
}
