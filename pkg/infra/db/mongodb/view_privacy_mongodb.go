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

type ViewPrivacyRepository struct {
	collection *mongo.Collection
}

func NewViewPrivacyRepository(client *mongo.Client, dbName string) *ViewPrivacyRepository {
	collection := client.Database(dbName).Collection("view_privacy_settings")

	ctx := context.Background()

	// Unique index: one privacy settings document per user
	_, _ = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	return &ViewPrivacyRepository{collection: collection}
}

func (r *ViewPrivacyRepository) Upsert(ctx context.Context, settings *analytics_entities.ViewPrivacySettings) error {
	filter := bson.M{"user_id": settings.UserID}
	opts := options.Replace().SetUpsert(true)

	_, err := r.collection.ReplaceOne(ctx, filter, settings, opts)
	if err != nil {
		slog.ErrorContext(ctx, "error upserting view privacy settings", "err", err, "user_id", settings.UserID)
		return err
	}

	return nil
}

func (r *ViewPrivacyRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*analytics_entities.ViewPrivacySettings, error) {
	filter := bson.M{"user_id": userID}

	var settings analytics_entities.ViewPrivacySettings
	err := r.collection.FindOne(ctx, filter).Decode(&settings)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &settings, nil
}
