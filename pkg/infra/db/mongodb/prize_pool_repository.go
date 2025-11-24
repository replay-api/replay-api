package db

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	"github.com/google/uuid"
	matchmaking_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_out "github.com/psavelis/team-pro/replay-api/pkg/domain/matchmaking/ports/out"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoPrizePoolRepository struct {
	*MongoDBRepository[*matchmaking_entities.PrizePool]
}

func NewMongoPrizePoolRepository(mongoClient *mongo.Client, dbName string) matchmaking_out.PrizePoolRepository {
	mappingCache := make(map[string]CacheItem)
	entityModel := reflect.TypeOf(matchmaking_entities.PrizePool{})
	repo := &MongoPrizePoolRepository{
		MongoDBRepository: &MongoDBRepository[*matchmaking_entities.PrizePool]{
			mongoClient:       mongoClient,
			dbName:            dbName,
			mappingCache:      mappingCache,
			entityModel:       entityModel,
			collectionName:    "prize_pools",
			entityName:        "PrizePool",
			bsonFieldMappings: make(map[string]string),
			queryableFields:   make(map[string]bool),
		},
	}

	// Define BSON field mappings
	bsonFieldMappings := map[string]string{
		"ID":                   "_id",
		"MatchID":              "match_id",
		"GameID":               "game_id",
		"Region":               "region",
		"Currency":             "currency",
		"TotalAmount":          "total_amount",
		"PlatformContribution": "platform_contribution",
		"PlayerContributions":  "player_contributions",
		"DistributionRule":     "distribution_rule",
		"Status":               "status",
		"LockedAt":             "locked_at",
		"DistributedAt":        "distributed_at",
		"Winners":              "winners",
		"MVPPlayerID":          "mvp_player_id",
		"EscrowEndTime":        "escrow_end_time",
		"CreatedAt":            "created_at",
		"UpdatedAt":            "updated_at",
		"ResourceOwner":        "resource_owner",
	}

	// Define queryable fields for search operations
	queryableFields := map[string]bool{
		"MatchID":       true,
		"GameID":        true,
		"Region":        true,
		"Currency":      true,
		"TotalAmount":   true,
		"Status":        true,
		"LockedAt":      true,
		"DistributedAt": true,
		"EscrowEndTime": true,
		"MVPPlayerID":   true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}

	repo.InitQueryableFields(queryableFields, bsonFieldMappings)

	return repo
}

func (r *MongoPrizePoolRepository) Save(ctx context.Context, pool *matchmaking_entities.PrizePool) error {
	if pool.GetID() == uuid.Nil {
		return fmt.Errorf("prize pool ID cannot be nil")
	}

	pool.UpdatedAt = time.Now().UTC()

	_, err := r.collection.InsertOne(ctx, pool)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save prize pool", "pool_id", pool.ID, "error", err)
		return fmt.Errorf("failed to save prize pool: %w", err)
	}

	slog.InfoContext(ctx, "prize pool saved successfully", "pool_id", pool.ID, "total_amount", pool.TotalAmount)
	return nil
}

func (r *MongoPrizePoolRepository) FindByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.PrizePool, error) {
	var pool matchmaking_entities.PrizePool

	filter := bson.M{"_id": id}
	err := r.collection.FindOne(ctx, filter).Decode(&pool)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("prize pool not found: %s", id)
		}
		slog.ErrorContext(ctx, "failed to find prize pool by ID", "id", id, "error", err)
		return nil, fmt.Errorf("failed to find prize pool: %w", err)
	}

	return &pool, nil
}

func (r *MongoPrizePoolRepository) FindByMatchID(ctx context.Context, matchID uuid.UUID) (*matchmaking_entities.PrizePool, error) {
	var pool matchmaking_entities.PrizePool

	filter := bson.M{"match_id": matchID}
	err := r.collection.FindOne(ctx, filter).Decode(&pool)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("prize pool not found for match: %s", matchID)
		}
		slog.ErrorContext(ctx, "failed to find prize pool by match ID", "match_id", matchID, "error", err)
		return nil, fmt.Errorf("failed to find prize pool: %w", err)
	}

	return &pool, nil
}

func (r *MongoPrizePoolRepository) FindPendingDistributions(ctx context.Context, limit int) ([]*matchmaking_entities.PrizePool, error) {
	now := time.Now().UTC()

	// Find pools that:
	// 1. In escrow status (match completed)
	// 2. Escrow period has ended
	filter := bson.M{
		"status":          matchmaking_entities.PrizePoolStatusInEscrow,
		"escrow_end_time": bson.M{"$lte": now},
	}

	findOptions := options.Find()
	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find pending distributions", "error", err)
		return nil, fmt.Errorf("failed to find pending distributions: %w", err)
	}
	defer cursor.Close(ctx)

	pools := make([]*matchmaking_entities.PrizePool, 0)
	for cursor.Next(ctx) {
		var pool matchmaking_entities.PrizePool
		if err := cursor.Decode(&pool); err != nil {
			slog.ErrorContext(ctx, "failed to decode prize pool", "error", err)
			continue
		}
		pools = append(pools, &pool)
	}

	slog.InfoContext(ctx, "found pending distributions", "count", len(pools))
	return pools, nil
}

func (r *MongoPrizePoolRepository) Update(ctx context.Context, pool *matchmaking_entities.PrizePool) error {
	if pool.GetID() == uuid.Nil {
		return fmt.Errorf("prize pool ID cannot be nil")
	}

	pool.UpdatedAt = time.Now().UTC()

	filter := bson.M{"_id": pool.ID}
	update := bson.M{"$set": pool}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update prize pool", "pool_id", pool.ID, "error", err)
		return fmt.Errorf("failed to update prize pool: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("prize pool not found for update: %s", pool.ID)
	}

	slog.InfoContext(ctx, "prize pool updated successfully", "pool_id", pool.ID, "status", pool.Status)
	return nil
}

func (r *MongoPrizePoolRepository) Delete(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"_id": id}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		slog.ErrorContext(ctx, "failed to delete prize pool", "id", id, "error", err)
		return fmt.Errorf("failed to delete prize pool: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("prize pool not found for deletion: %s", id)
	}

	slog.InfoContext(ctx, "prize pool deleted successfully", "pool_id", id)
	return nil
}

// Ensure MongoPrizePoolRepository implements PrizePoolRepository interface
var _ matchmaking_out.PrizePoolRepository = (*MongoPrizePoolRepository)(nil)
