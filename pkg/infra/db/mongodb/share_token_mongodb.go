package db

import (
	"context"
	"log/slog"
	"reflect"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ShareTokenRepository struct {
	MongoDBRepository[replay_entity.ShareToken]
}

func NewShareTokenRepository(client *mongo.Client, dbName string, entityType replay_entity.ShareToken, collectionName string) *ShareTokenRepository {
	repo := MongoDBRepository[replay_entity.ShareToken]{
		mongoClient:       client,
		dbName:            dbName,
		mappingCache:      make(map[string]CacheItem),
		entityModel:       reflect.TypeOf(entityType),
		BsonFieldMappings: make(map[string]string),
		collectionName:    collectionName,
		entityName:        reflect.TypeOf(entityType).Name(),
		QueryableFields:   make(map[string]bool),
	}

	repo.InitQueryableFields(map[string]bool{
		"ID":            true,
		"ResourceID":    true,
		"ResourceType":  true,
		"Status":        true,
		"ExpiresAt":     true,
		"Uri":           true,
		"EntityType":    true,
		"ResourceOwner": true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}, map[string]string{
		"ID":            "token",
		"ResourceID":    "resource_id",
		"ResourceType":  "resource_type",
		"Status":        "status",
		"ExpiresAt":     "expires_at",
		"Uri":           "uri",
		"EntityType":    "entity_type",
		"ResourceOwner": "resource_owner",
		"CreatedAt":     "created_at",
		"UpdatedAt":     "updated_at",
	})

	return &ShareTokenRepository{
		repo,
	}
}

func (r *ShareTokenRepository) Search(ctx context.Context, s common.Search) ([]replay_entity.ShareToken, error) {
	cursor, err := r.Query(ctx, s)
	if cursor != nil {
		defer cursor.Close(ctx)
	}

	if err != nil {
		slog.ErrorContext(ctx, "error querying share_token entity", "err", err)
		return nil, err
	}

	var tokens []replay_entity.ShareToken
	if err := cursor.All(ctx, &tokens); err != nil {
		slog.ErrorContext(ctx, "error decoding share_token results", "err", err)
		return nil, err
	}

	return tokens, nil
}

func (r *ShareTokenRepository) Create(ctx context.Context, token *replay_entity.ShareToken) error {
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}

	now := time.Now()
	if token.CreatedAt.IsZero() {
		token.CreatedAt = now
	}
	if token.UpdatedAt.IsZero() {
		token.UpdatedAt = now
	}

	if token.Status == "" {
		token.Status = replay_entity.ShareTokenStatusActive
	}

	_, err := r.collection.InsertOne(ctx, token)
	if err != nil {
		slog.ErrorContext(ctx, "error creating share_token", "err", err)
		return err
	}

	return nil
}

func (r *ShareTokenRepository) FindByToken(ctx context.Context, tokenID uuid.UUID) (*replay_entity.ShareToken, error) {
	var token replay_entity.ShareToken

	filter := bson.M{"token": tokenID}
	err := r.collection.FindOne(ctx, filter).Decode(&token)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		slog.ErrorContext(ctx, "error finding share_token by token", "err", err, "token_id", tokenID)
		return nil, err
	}

	return &token, nil
}

func (r *ShareTokenRepository) Update(ctx context.Context, token *replay_entity.ShareToken) error {
	token.UpdatedAt = time.Now()

	filter := bson.M{"token": token.ID}
	update := bson.M{"$set": token}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "error updating share_token", "err", err, "token_id", token.ID)
		return err
	}

	return nil
}

func (r *ShareTokenRepository) Delete(ctx context.Context, tokenID uuid.UUID) error {
	filter := bson.M{"token": tokenID}

	_, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		slog.ErrorContext(ctx, "error deleting share_token", "err", err, "token_id", tokenID)
		return err
	}

	return nil
}

func (r *ShareTokenRepository) FindByResourceID(ctx context.Context, resourceID uuid.UUID) ([]replay_entity.ShareToken, error) {
	filter := bson.M{"resource_id": resourceID, "status": replay_entity.ShareTokenStatusActive}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		slog.ErrorContext(ctx, "error finding share_tokens by resource_id", "err", err, "resource_id", resourceID)
		return nil, err
	}
	defer cursor.Close(ctx)

	var tokens []replay_entity.ShareToken
	if err := cursor.All(ctx, &tokens); err != nil {
		slog.ErrorContext(ctx, "error decoding share_tokens", "err", err)
		return nil, err
	}

	return tokens, nil
}

// ExpireOldTokens marks tokens as expired if they're past their expiration date
func (r *ShareTokenRepository) ExpireOldTokens(ctx context.Context) (int64, error) {
	filter := bson.M{
		"expires_at": bson.M{"$lt": time.Now()},
		"status":     replay_entity.ShareTokenStatusActive,
	}

	update := bson.M{
		"$set": bson.M{
			"status":     replay_entity.ShareTokenStatusExpired,
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "error expiring old tokens", "err", err)
		return 0, err
	}

	return result.ModifiedCount, nil
}
