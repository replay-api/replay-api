package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_out "github.com/replay-api/replay-api/pkg/domain/wallet/ports/out"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoVaultActivityRepository implements VaultActivityRepository using MongoDB
type MongoVaultActivityRepository struct {
	mongodb.MongoDBRepository[wallet_entities.VaultActivity]
	coll *mongo.Collection
}

// NewMongoVaultActivityRepository creates a new VaultActivity MongoDB repository
func NewMongoVaultActivityRepository(mongoClient *mongo.Client, dbName string) wallet_out.VaultActivityRepository {
	entityType := wallet_entities.VaultActivity{}
	repo := mongodb.NewMongoDBRepository(mongoClient, dbName, entityType, "vault_activities", "VaultActivity")

	repo.InitQueryableFields(map[string]bool{
		"ID":           true,
		"BaseEntityID": true,
		"VaultID":      true,
		"ActorID":      true,
		"ActivityType": true,
		"CreatedAt":    true,
		"GroupID":      true,
	}, map[string]string{
		"ID":           "baseentity._id",
		"BaseEntityID": "baseentity._id",
		"VaultID":      "vault_id",
		"ActorID":      "actor_id",
		"ActivityType": "activity_type",
		"CreatedAt":    "baseentity.created_at",
		"GroupID":      "baseentity.resource_owner.group_id",
	})

	coll := mongoClient.Database(dbName).Collection("vault_activities")
	r := &MongoVaultActivityRepository{MongoDBRepository: *repo, coll: coll}
	r.ensureIndexes(coll)
	return r
}

func (r *MongoVaultActivityRepository) ensureIndexes(coll *mongo.Collection) {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "vault_id", Value: 1}, {Key: "baseentity.created_at", Value: -1}},
			Options: options.Index().SetName("idx_vault_activities_vault_time"),
		},
		{
			Keys:    bson.D{{Key: "actor_id", Value: 1}},
			Options: options.Index().SetName("idx_vault_activities_actor"),
		},
		{
			Keys:    bson.D{{Key: "activity_type", Value: 1}},
			Options: options.Index().SetName("idx_vault_activities_type"),
		},
	}
	if _, err := coll.Indexes().CreateMany(context.Background(), indexes); err != nil {
		slog.Warn("Failed to create vault_activities indexes", "error", err)
	}
}

func (r *MongoVaultActivityRepository) Append(ctx context.Context, activity *wallet_entities.VaultActivity) error {
	if activity.GetID() == uuid.Nil {
		return fmt.Errorf("activity ID cannot be nil")
	}
	activity.CreatedAt = time.Now().UTC()
	_, err := r.MongoDBRepository.Create(ctx, activity)
	if err != nil {
		return fmt.Errorf("failed to append vault activity: %w", err)
	}
	return nil
}

func (r *MongoVaultActivityRepository) FindByVaultID(ctx context.Context, vaultID uuid.UUID, limit, offset int) ([]wallet_entities.VaultActivity, int64, error) {
	filter := bson.M{"vault_id": vaultID}
	opts := options.Find().
		SetSort(bson.D{{Key: "baseentity.created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count activities: %w", err)
	}

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find activities: %w", err)
	}
	defer cursor.Close(ctx)

	var activities []wallet_entities.VaultActivity
	if err := cursor.All(ctx, &activities); err != nil {
		return nil, 0, fmt.Errorf("failed to decode activities: %w", err)
	}
	return activities, total, nil
}

func (r *MongoVaultActivityRepository) FindByActorID(ctx context.Context, vaultID uuid.UUID, actorID uuid.UUID, limit, offset int) ([]wallet_entities.VaultActivity, int64, error) {
	filter := bson.M{"vault_id": vaultID, "actor_id": actorID}
	opts := options.Find().
		SetSort(bson.D{{Key: "baseentity.created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count actor activities: %w", err)
	}

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find actor activities: %w", err)
	}
	defer cursor.Close(ctx)

	var activities []wallet_entities.VaultActivity
	if err := cursor.All(ctx, &activities); err != nil {
		return nil, 0, fmt.Errorf("failed to decode activities: %w", err)
	}
	return activities, total, nil
}

func (r *MongoVaultActivityRepository) FindByType(ctx context.Context, vaultID uuid.UUID, activityType wallet_entities.VaultActivityType, limit, offset int) ([]wallet_entities.VaultActivity, int64, error) {
	filter := bson.M{"vault_id": vaultID, "activity_type": activityType}
	opts := options.Find().
		SetSort(bson.D{{Key: "baseentity.created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count activities by type: %w", err)
	}

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find activities by type: %w", err)
	}
	defer cursor.Close(ctx)

	var activities []wallet_entities.VaultActivity
	if err := cursor.All(ctx, &activities); err != nil {
		return nil, 0, fmt.Errorf("failed to decode activities: %w", err)
	}
	return activities, total, nil
}

func (r *MongoVaultActivityRepository) CountByVaultID(ctx context.Context, vaultID uuid.UUID) (int64, error) {
	filter := bson.M{"vault_id": vaultID}
	count, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count vault activities: %w", err)
	}
	return count, nil
}
