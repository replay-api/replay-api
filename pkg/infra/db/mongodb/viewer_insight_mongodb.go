package db

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	analytics_entities "github.com/replay-api/replay-api/pkg/domain/analytics/entities"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type ViewerInsightRepository struct {
	collection *mongo.Collection
}

func NewViewerInsightRepository(client *mongo.Client, dbName string) *ViewerInsightRepository {
	collection := client.Database(dbName).Collection("viewer_insights")

	ctx := context.Background()

	// Unique index: one insight per viewer+entity pair
	_, _ = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "entity_id", Value: 1},
			{Key: "entity_type", Value: 1},
			{Key: "viewer_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})

	return &ViewerInsightRepository{collection: collection}
}

func (r *ViewerInsightRepository) Upsert(ctx context.Context, insight *analytics_entities.ViewerInsight) error {
	filter := bson.M{
		"entity_id":   insight.EntityID,
		"entity_type": insight.EntityType,
		"viewer_id":   insight.ViewerID,
	}

	opts := options.Replace().SetUpsert(true)
	_, err := r.collection.ReplaceOne(ctx, filter, insight, opts)
	if err != nil {
		slog.ErrorContext(ctx, "error upserting viewer insight", "err", err, "entity_id", insight.EntityID)
		return err
	}

	return nil
}

func (r *ViewerInsightRepository) GetByEntity(ctx context.Context, entityID uuid.UUID, entityType analytics_entities.EntityTypeKey, search shared.Search) ([]analytics_entities.ViewerInsight, int64, error) {
	filter := bson.M{
		"entity_id":    entityID,
		"entity_type":  entityType,
		"is_anonymous": false,
	}

	// Count total
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Build find options from search
	findOpts := options.Find()
	findOpts.SetSkip(int64(search.ResultOptions.Skip))
	if search.ResultOptions.Limit > 0 {
		findOpts.SetLimit(int64(search.ResultOptions.Limit))
	}

	if len(search.SortOptions) > 0 {
		sort := bson.D{}
		for _, s := range search.SortOptions {
			sort = append(sort, bson.E{Key: s.Field, Value: int(s.Direction)})
		}
		findOpts.SetSort(sort)
	} else {
		findOpts.SetSort(bson.D{{Key: "last_viewed_at", Value: -1}})
	}

	cursor, err := r.collection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var insights []analytics_entities.ViewerInsight
	if err := cursor.All(ctx, &insights); err != nil {
		return nil, 0, err
	}

	return insights, total, nil
}
