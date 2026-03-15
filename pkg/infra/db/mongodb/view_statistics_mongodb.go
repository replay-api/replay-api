package db

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	analytics_entities "github.com/replay-api/replay-api/pkg/domain/analytics/entities"
)

type ViewStatisticsRepository struct {
	collection *mongo.Collection
}

func NewViewStatisticsRepository(client *mongo.Client, dbName string) *ViewStatisticsRepository {
	collection := client.Database(dbName).Collection("view_statistics")

	ctx := context.Background()

	// Unique index: one stats document per entity
	_, _ = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "entity_id", Value: 1},
			{Key: "entity_type", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})

	return &ViewStatisticsRepository{collection: collection}
}

func (r *ViewStatisticsRepository) Upsert(ctx context.Context, stats *analytics_entities.ViewStatistics) error {
	filter := bson.M{
		"entity_id":   stats.EntityID,
		"entity_type": stats.EntityType,
	}

	opts := options.Replace().SetUpsert(true)
	_, err := r.collection.ReplaceOne(ctx, filter, stats, opts)
	if err != nil {
		slog.ErrorContext(ctx, "error upserting view statistics", "err", err, "entity_id", stats.EntityID)
		return err
	}

	return nil
}

func (r *ViewStatisticsRepository) GetByEntity(ctx context.Context, entityID uuid.UUID, entityType analytics_entities.EntityTypeKey) (*analytics_entities.ViewStatistics, error) {
	filter := bson.M{
		"entity_id":   entityID,
		"entity_type": entityType,
	}

	var stats analytics_entities.ViewStatistics
	err := r.collection.FindOne(ctx, filter).Decode(&stats)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &stats, nil
}

func (r *ViewStatisticsRepository) GetByOwner(ctx context.Context, ownerID uuid.UUID, entityType *analytics_entities.EntityTypeKey) ([]analytics_entities.ViewStatistics, error) {
	filter := bson.M{
		"resource_owner.user_id": ownerID,
	}
	if entityType != nil {
		filter["entity_type"] = *entityType
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []analytics_entities.ViewStatistics
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}
