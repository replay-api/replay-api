package db

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	iam_entity "github.com/replay-api/replay-api/pkg/domain/iam/entities"
)

type RIDTokenRepository struct {
	mongodb.MongoDBRepository[iam_entity.RIDToken]
}

func NewRIDTokenRepository(client *mongo.Client, dbName string, entityType iam_entity.RIDToken, collectionName string) *RIDTokenRepository {
	// TODO: create Factory for encapsulating this and reducing bloat/repetition of some fields like queryableFields, mappingCache.. this needs a facade for a clearer instantiation
	repo := mongodb.NewMongoDBRepository[iam_entity.RIDToken](client, dbName, entityType, collectionName, "RIDToken")

	repo.InitQueryableFields(map[string]bool{
		"ID":            true,
		"Key":           true,
		"Source":        true,
		"ResourceOwner": true,
		"ExpiresAt":     true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
		"RevokedAt":     true,
	}, map[string]string{
		"ID":            "_id",
		"Key":           "key",
		"Source":        "source",
		"ResourceOwner": "resource_owner",
		"ExpiresAt":     "expires_at",
		"CreatedAt":     "created_at",
		"UpdatedAt":     "updated_at",
		"RevokedAt":     "revoked_at",
		"TenantID":      "resource_owner.tenant_id",
		"UserID":        "resource_owner.user_id",
		"GroupID":       "resource_owner.group_id",
		"ClientID":      "resource_owner.client_id",
	})

	return &RIDTokenRepository{
		MongoDBRepository: *repo,
	}
}

// Delete removes a token from the database
func (r *RIDTokenRepository) Delete(ctx context.Context, tokenID string) error {
	id, err := uuid.Parse(tokenID)
	if err != nil {
		slog.ErrorContext(ctx, "invalid token ID for delete", "tokenID", tokenID, "err", err)
		return err
	}

	collection := r.MongoDBRepository.Collection()
	result, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		slog.ErrorContext(ctx, "failed to delete token", "tokenID", tokenID, "err", err)
		return err
	}

	if result.DeletedCount == 0 {
		slog.WarnContext(ctx, "token not found for deletion", "tokenID", tokenID)
	}

	return nil
}

// Revoke marks a token as revoked by setting the revoked_at timestamp
// This is preferred over deletion for audit trail purposes
func (r *RIDTokenRepository) Revoke(ctx context.Context, tokenID string) error {
	id, err := uuid.Parse(tokenID)
	if err != nil {
		slog.ErrorContext(ctx, "invalid token ID for revoke", "tokenID", tokenID, "err", err)
		return err
	}

	collection := r.MongoDBRepository.Collection()
	
	update := bson.M{
		"$set": bson.M{
			"revoked_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to revoke token", "tokenID", tokenID, "err", err)
		return err
	}

	if result.MatchedCount == 0 {
		slog.WarnContext(ctx, "token not found for revocation", "tokenID", tokenID)
	}

	slog.InfoContext(ctx, "token revoked successfully", "tokenID", tokenID)
	return nil
}

// FindByID retrieves a token by its ID
func (r *RIDTokenRepository) FindByID(ctx context.Context, tokenID uuid.UUID) (*iam_entity.RIDToken, error) {
	return r.GetByID(ctx, tokenID)
}

// IsRevoked checks if a token has been revoked
func (r *RIDTokenRepository) IsRevoked(ctx context.Context, tokenID uuid.UUID) (bool, error) {
	collection := r.MongoDBRepository.Collection()
	
	var result struct {
		RevokedAt *time.Time `bson:"revoked_at"`
	}

	err := collection.FindOne(ctx, bson.M{"_id": tokenID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil // Token doesn't exist, treat as not revoked
		}
		return false, err
	}

	return result.RevokedAt != nil, nil
}
