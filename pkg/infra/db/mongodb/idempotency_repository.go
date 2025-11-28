package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_out "github.com/replay-api/replay-api/pkg/domain/wallet/ports/out"
)

// IdempotencyRepository implements wallet_out.IdempotencyRepository for MongoDB
type IdempotencyRepository struct {
	db             *mongo.Database
	collectionName string
}

// NewIdempotencyRepository creates a new MongoDB idempotency repository
func NewIdempotencyRepository(db *mongo.Database) wallet_out.IdempotencyRepository {
	repo := &IdempotencyRepository{
		db:             db,
		collectionName: "idempotent_operations",
	}

	// Ensure indexes exist (non-fatal - indexes are for performance, not correctness)
	if err := repo.ensureIndexes(context.Background()); err != nil {
		// Log error but don't panic - the app can still function without indexes
		// This can happen when MongoDB auth is misconfigured or during initial setup
		fmt.Printf("WARNING: failed to create idempotency indexes (non-fatal): %v\n", err)
	}

	return repo
}

// ensureIndexes creates necessary indexes for performance and data integrity
func (r *IdempotencyRepository) ensureIndexes(ctx context.Context) error {
	collection := r.db.Collection(r.collectionName)

	indexes := []mongo.IndexModel{
		// TTL index for automatic cleanup of expired operations (24 hours)
		// MongoDB automatically deletes documents when expires_at < current time
		{
			Keys: bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().
				SetName("idx_expires_at_ttl").
				SetExpireAfterSeconds(0), // Expire at exact time specified in expires_at
		},

		// Index for finding stale operations (stuck in Processing state)
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "last_attempt_at", Value: 1},
			},
			Options: options.Index().SetName("idx_status_last_attempt"),
		},

		// Index for operation type analytics
		{
			Keys: bson.D{{Key: "operation_type", Value: 1}},
			Options: options.Index().SetName("idx_operation_type"),
		},

		// Index for created_at (for monitoring and analytics)
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_created_at"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create idempotency indexes: %w", err)
	}

	return nil
}

// Create creates a new idempotent operation record
// The idempotency key is used as _id for uniqueness guarantee
func (r *IdempotencyRepository) Create(ctx context.Context, op *wallet_entities.IdempotentOperation) error {
	collection := r.db.Collection(r.collectionName)

	_, err := collection.InsertOne(ctx, op)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("idempotency key already exists: %s", op.Key)
		}
		return fmt.Errorf("failed to create idempotent operation: %w", err)
	}

	return nil
}

// FindByKey retrieves an idempotent operation by its key
func (r *IdempotencyRepository) FindByKey(ctx context.Context, key string) (*wallet_entities.IdempotentOperation, error) {
	collection := r.db.Collection(r.collectionName)

	var op wallet_entities.IdempotentOperation
	err := collection.FindOne(ctx, bson.M{"_id": key}).Decode(&op)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("idempotent operation not found: %s", key)
		}
		return nil, fmt.Errorf("failed to find idempotent operation: %w", err)
	}

	return &op, nil
}

// Update updates an existing idempotent operation
// Used to mark operations as completed or failed
func (r *IdempotencyRepository) Update(ctx context.Context, op *wallet_entities.IdempotentOperation) error {
	collection := r.db.Collection(r.collectionName)

	update := bson.M{
		"$set": bson.M{
			"status":           op.Status,
			"response_payload": op.ResponsePayload,
			"result_id":        op.ResultID,
			"error_message":    op.ErrorMessage,
			"completed_at":     op.CompletedAt,
			"attempt_count":    op.AttemptCount,
			"last_attempt_at":  op.LastAttemptAt,
		},
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": op.Key}, update)
	if err != nil {
		return fmt.Errorf("failed to update idempotent operation: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("idempotent operation not found: %s", op.Key)
	}

	return nil
}

// Delete removes an idempotent operation
// Used for manual cleanup or testing
func (r *IdempotencyRepository) Delete(ctx context.Context, key string) error {
	collection := r.db.Collection(r.collectionName)

	result, err := collection.DeleteOne(ctx, bson.M{"_id": key})
	if err != nil {
		return fmt.Errorf("failed to delete idempotent operation: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("idempotent operation not found: %s", key)
	}

	return nil
}

// FindStaleOperations finds operations stuck in "Processing" state for longer than threshold
// Used for monitoring and cleanup of operations that may have failed without marking themselves as failed
func (r *IdempotencyRepository) FindStaleOperations(ctx context.Context, threshold time.Duration) ([]*wallet_entities.IdempotentOperation, error) {
	collection := r.db.Collection(r.collectionName)

	// Find operations that are:
	// 1. Still in "Processing" status
	// 2. Last attempted more than threshold ago (or created more than threshold ago if never attempted)
	staleTime := time.Now().UTC().Add(-threshold)

	filter := bson.M{
		"status": wallet_entities.OperationStatusProcessing,
		"$or": []bson.M{
			// Has last_attempt_at and it's stale
			{
				"last_attempt_at": bson.M{"$exists": true, "$lt": staleTime},
			},
			// No last_attempt_at but created_at is stale
			{
				"last_attempt_at": bson.M{"$exists": false},
				"created_at":      bson.M{"$lt": staleTime},
			},
		},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find stale operations: %w", err)
	}
	defer cursor.Close(ctx)

	var staleOps []*wallet_entities.IdempotentOperation
	if err := cursor.All(ctx, &staleOps); err != nil {
		return nil, fmt.Errorf("failed to decode stale operations: %w", err)
	}

	return staleOps, nil
}

// CleanupExpired removes expired operations
// This is a fallback - MongoDB TTL index should handle this automatically
// Useful for manual cleanup or if TTL index is disabled
func (r *IdempotencyRepository) CleanupExpired(ctx context.Context) (int64, error) {
	collection := r.db.Collection(r.collectionName)

	now := time.Now().UTC()
	filter := bson.M{
		"expires_at": bson.M{"$lt": now},
	}

	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired operations: %w", err)
	}

	return result.DeletedCount, nil
}

// GetOperationStats retrieves statistics about idempotent operations
// Useful for monitoring and alerting
func (r *IdempotencyRepository) GetOperationStats(ctx context.Context) (*OperationStats, error) {
	collection := r.db.Collection(r.collectionName)

	pipeline := mongo.Pipeline{
		// Group by status to count operations
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$status"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get operation stats: %w", err)
	}
	defer cursor.Close(ctx)

	stats := &OperationStats{
		StatusCounts: make(map[wallet_entities.OperationStatus]int64),
	}

	for cursor.Next(ctx) {
		var result struct {
			ID    wallet_entities.OperationStatus `bson:"_id"`
			Count int64                            `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode stats: %w", err)
		}
		stats.StatusCounts[result.ID] = result.Count
		stats.TotalOperations += result.Count
	}

	return stats, nil
}

// GetOperationsByType retrieves operations grouped by type
// Useful for analytics and understanding system usage patterns
func (r *IdempotencyRepository) GetOperationsByType(ctx context.Context, limit int) (map[string]int64, error) {
	collection := r.db.Collection(r.collectionName)

	pipeline := mongo.Pipeline{
		// Group by operation type
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$operation_type"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		// Sort by count descending
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
	}

	if limit > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: limit}})
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get operations by type: %w", err)
	}
	defer cursor.Close(ctx)

	operationCounts := make(map[string]int64)
	for cursor.Next(ctx) {
		var result struct {
			OperationType string `bson:"_id"`
			Count         int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode operation type stats: %w", err)
		}
		operationCounts[result.OperationType] = result.Count
	}

	return operationCounts, nil
}

// OperationStats holds statistics about idempotent operations
type OperationStats struct {
	StatusCounts    map[wallet_entities.OperationStatus]int64
	TotalOperations int64
}
