package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_out "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/out"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoPrizePoolRepository struct {
	*mongodb.MongoDBRepository[matchmaking_entities.PrizePool]
}

func NewMongoPrizePoolRepository(mongoClient *mongo.Client, dbName string) matchmaking_out.PrizePoolRepository {
	entityType := matchmaking_entities.PrizePool{}
	collectionName := "prize_pools"

	repo := mongodb.NewMongoDBRepository[matchmaking_entities.PrizePool](mongoClient, dbName, entityType, collectionName, "PrizePool")

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

	return &MongoPrizePoolRepository{
		MongoDBRepository: repo,
	}
}

func (r *MongoPrizePoolRepository) Save(ctx context.Context, pool *matchmaking_entities.PrizePool) error {
	if pool.GetID() == uuid.Nil {
		return fmt.Errorf("prize pool ID cannot be nil")
	}

	pool.UpdatedAt = time.Now().UTC()

	_, err := r.MongoDBRepository.Update(ctx, pool)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save prize pool", "pool_id", pool.ID, "error", err)
		return fmt.Errorf("failed to save prize pool: %w", err)
	}

	slog.InfoContext(ctx, "prize pool saved successfully", "pool_id", pool.ID, "total_amount", pool.TotalAmount)
	return nil
}

func (r *MongoPrizePoolRepository) FindByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.PrizePool, error) {
	return r.MongoDBRepository.GetByID(ctx, id)
}

func (r *MongoPrizePoolRepository) FindByMatchID(ctx context.Context, matchID uuid.UUID) (*matchmaking_entities.PrizePool, error) {
	var pool matchmaking_entities.PrizePool

	filter := bson.M{"match_id": matchID}
	err := r.MongoDBRepository.FindOneWithRLS(ctx, filter).Decode(&pool)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("prize pool not found for match: %s", matchID)
		}
		slog.ErrorContext(ctx, "failed to find prize pool by match ID", "match_id", matchID, "error", err)
		return nil, fmt.Errorf("failed to find prize pool: %w", err)
	}

	return &pool, nil
}

func (r *MongoPrizePoolRepository) Update(ctx context.Context, pool *matchmaking_entities.PrizePool) error {
	if pool.GetID() == uuid.Nil {
		return fmt.Errorf("prize pool ID cannot be nil")
	}

	pool.UpdatedAt = time.Now().UTC()

	_, err := r.MongoDBRepository.Update(ctx, pool)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update prize pool", "pool_id", pool.ID, "error", err)
		return fmt.Errorf("failed to update prize pool: %w", err)
	}

	slog.InfoContext(ctx, "prize pool updated successfully", "pool_id", pool.ID, "status", pool.Status)
	return nil
}

func (r *MongoPrizePoolRepository) Delete(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"_id": id}

	result, err := r.MongoDBRepository.DeleteOneWithRLS(ctx, filter)
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

func (r *MongoPrizePoolRepository) UpdateUnsafe(ctx context.Context, pool *matchmaking_entities.PrizePool) error {
	if pool.GetID() == uuid.Nil {
		return fmt.Errorf("prize pool ID cannot be nil")
	}

	pool.UpdatedAt = time.Now().UTC()

	_, err := r.MongoDBRepository.UpdateUnsafe(ctx, pool)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update prize pool", "pool_id", pool.ID, "error", err)
		return fmt.Errorf("failed to update prize pool: %w", err)
	}

	slog.InfoContext(ctx, "prize pool updated successfully", "pool_id", pool.ID)
	return nil
}

// Ensure MongoPrizePoolRepository implements PrizePoolRepository interface
var _ matchmaking_out.PrizePoolRepository = (*MongoPrizePoolRepository)(nil)
