package db

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	analytics_entities "github.com/replay-api/replay-api/pkg/domain/analytics/entities"
)

type EntityViewRepository struct {
	collection *mongo.Collection
}

func NewEntityViewRepository(client *mongo.Client, dbName string) *EntityViewRepository {
	collection := client.Database(dbName).Collection("entity_views")

	ctx := context.Background()

	// Compound index for entity lookups by date
	_, _ = collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "entity_id", Value: 1},
				{Key: "entity_type", Value: 1},
				{Key: "viewed_at", Value: -1},
			},
		},
		// Deduplication index: viewer + entity + time window
		{
			Keys: bson.D{
				{Key: "viewer_id", Value: 1},
				{Key: "entity_id", Value: 1},
				{Key: "viewed_at", Value: -1},
			},
		},
		// TTL index: auto-delete raw views after 90 days
		{
			Keys:    bson.D{{Key: "viewed_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(90 * 24 * 60 * 60),
		},
	})

	return &EntityViewRepository{collection: collection}
}

func (r *EntityViewRepository) Create(ctx context.Context, view *analytics_entities.EntityView) (*analytics_entities.EntityView, error) {
	_, err := r.collection.InsertOne(ctx, view)
	if err != nil {
		slog.ErrorContext(ctx, "error creating entity view", "err", err, "entity_id", view.EntityID)
		return nil, err
	}
	return view, nil
}

func (r *EntityViewRepository) GetLastViewByViewer(ctx context.Context, entityID uuid.UUID, viewerID uuid.UUID, since time.Time) (*analytics_entities.EntityView, error) {
	filter := bson.M{
		"entity_id": entityID,
		"viewer_id": viewerID,
		"viewed_at": bson.M{"$gte": since},
	}

	opts := options.FindOne().SetSort(bson.D{{Key: "viewed_at", Value: -1}})

	var view analytics_entities.EntityView
	err := r.collection.FindOne(ctx, filter, opts).Decode(&view)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &view, nil
}

func (r *EntityViewRepository) CountByEntity(ctx context.Context, entityID uuid.UUID, entityType analytics_entities.EntityTypeKey) (int64, error) {
	filter := bson.M{
		"entity_id":   entityID,
		"entity_type": entityType,
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	return count, nil
}
