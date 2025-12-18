// Package db provides MongoDB repository implementations for the domain entities.
package db

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	auth_entities "github.com/replay-api/replay-api/pkg/domain/auth/entities"
	auth_out "github.com/replay-api/replay-api/pkg/domain/auth/ports/out"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PasswordResetRepository implements auth_out.PasswordResetRepository
type PasswordResetRepository struct {
	MongoDBRepository[auth_entities.PasswordReset]
}

// NewPasswordResetMongoDBRepository creates a new password reset repository
func NewPasswordResetMongoDBRepository(client *mongo.Client, dbName string) *PasswordResetRepository {
	entityType := auth_entities.PasswordReset{}
	collectionName := "password_resets"

	repo := MongoDBRepository[auth_entities.PasswordReset]{
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
		"ID":        true,
		"UserID":    true,
		"Email":     true,
		"Token":     true,
		"Status":    true,
		"ExpiresAt": true,
		"CreatedAt": true,
	}, map[string]string{
		"ID":        "_id",
		"UserID":    "user_id",
		"Email":     "email",
		"Token":     "token",
		"Status":    "status",
		"ExpiresAt": "expires_at",
		"UsedAt":    "used_at",
		"CreatedAt": "created_at",
		"UpdatedAt": "updated_at",
		"IPAddress": "ip_address",
		"UserAgent": "user_agent",
	})

	return &PasswordResetRepository{repo}
}

// collection returns the MongoDB collection for password resets
func (r *PasswordResetRepository) collection() *mongo.Collection {
	return r.mongoClient.Database(r.dbName).Collection(r.collectionName)
}

// Save creates a new password reset record
func (r *PasswordResetRepository) Save(ctx context.Context, reset *auth_entities.PasswordReset) error {
	_, err := r.collection().InsertOne(ctx, reset)
	if err != nil {
		return fmt.Errorf("failed to save password reset: %w", err)
	}
	return nil
}

// FindByID retrieves a reset request by its ID
func (r *PasswordResetRepository) FindByID(ctx context.Context, id uuid.UUID) (*auth_entities.PasswordReset, error) {
	var reset auth_entities.PasswordReset
	err := r.collection().FindOne(ctx, bson.M{"_id": id}).Decode(&reset)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("password reset not found")
		}
		return nil, fmt.Errorf("failed to find password reset: %w", err)
	}
	return &reset, nil
}

// FindByToken retrieves a reset request by its token
func (r *PasswordResetRepository) FindByToken(ctx context.Context, token string) (*auth_entities.PasswordReset, error) {
	var reset auth_entities.PasswordReset
	err := r.collection().FindOne(ctx, bson.M{
		"token":  token,
		"status": auth_entities.PasswordResetStatusPending,
	}).Decode(&reset)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("password reset not found")
		}
		return nil, fmt.Errorf("failed to find password reset by token: %w", err)
	}
	return &reset, nil
}

// FindByUserID retrieves the latest reset request for a user
func (r *PasswordResetRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*auth_entities.PasswordReset, error) {
	var reset auth_entities.PasswordReset

	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})
	err := r.collection().FindOne(ctx, bson.M{"user_id": userID}, opts).Decode(&reset)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no password reset found for user")
		}
		return nil, fmt.Errorf("failed to find password reset by user_id: %w", err)
	}
	return &reset, nil
}

// FindPendingByEmail retrieves pending reset requests for an email
func (r *PasswordResetRepository) FindPendingByEmail(ctx context.Context, email string) (*auth_entities.PasswordReset, error) {
	var reset auth_entities.PasswordReset

	filter := bson.M{
		"email":      email,
		"status":     auth_entities.PasswordResetStatusPending,
		"expires_at": bson.M{"$gt": time.Now().UTC()},
	}
	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})

	err := r.collection().FindOne(ctx, filter, opts).Decode(&reset)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No pending reset found
		}
		return nil, fmt.Errorf("failed to find pending password reset: %w", err)
	}
	return &reset, nil
}

// Update updates an existing reset request
func (r *PasswordResetRepository) Update(ctx context.Context, reset *auth_entities.PasswordReset) error {
	reset.UpdatedAt = time.Now().UTC()

	filter := bson.M{"_id": reset.ID}
	update := bson.M{"$set": reset}

	result, err := r.collection().UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update password reset: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("password reset not found")
	}
	return nil
}

// InvalidatePreviousResets invalidates all previous resets for a user
func (r *PasswordResetRepository) InvalidatePreviousResets(ctx context.Context, userID uuid.UUID, email string) error {
	filter := bson.M{
		"$or": []bson.M{
			{"user_id": userID},
			{"email": email},
		},
		"status": auth_entities.PasswordResetStatusPending,
	}
	update := bson.M{
		"$set": bson.M{
			"status":     auth_entities.PasswordResetStatusCanceled,
			"updated_at": time.Now().UTC(),
		},
	}

	_, err := r.collection().UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to invalidate previous password resets: %w", err)
	}
	return nil
}

// CountRecentAttempts counts reset attempts in the last N minutes
func (r *PasswordResetRepository) CountRecentAttempts(ctx context.Context, email string, minutes int) (int, error) {
	since := time.Now().UTC().Add(-time.Duration(minutes) * time.Minute)

	filter := bson.M{
		"email":      email,
		"created_at": bson.M{"$gte": since},
	}

	count, err := r.collection().CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count recent password reset attempts: %w", err)
	}
	return int(count), nil
}

// EnsureIndexes creates the necessary indexes for the password_resets collection
func (r *PasswordResetRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "token", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "email", Value: 1}},
		},
		{
			Keys: bson.D{
				{Key: "email", Value: 1},
				{Key: "status", Value: 1},
			},
		},
		{
			Keys:    bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(86400 * 7), // TTL: 7 days
		},
	}

	_, err := r.collection().Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}
	return nil
}

// Ensure PasswordResetRepository implements the interface
var _ auth_out.PasswordResetRepository = (*PasswordResetRepository)(nil)

