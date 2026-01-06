// Package db provides MongoDB repository implementations for the domain entities.
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	auth_entities "github.com/replay-api/replay-api/pkg/domain/auth/entities"
	auth_out "github.com/replay-api/replay-api/pkg/domain/auth/ports/out"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EmailVerificationRepository implements auth_out.EmailVerificationRepository
type EmailVerificationRepository struct {
	mongodb.MongoDBRepository[auth_entities.EmailVerification]
}

// NewEmailVerificationMongoDBRepository creates a new email verification repository
func NewEmailVerificationMongoDBRepository(client *mongo.Client, dbName string) *EmailVerificationRepository {
	entityType := auth_entities.EmailVerification{}
	collectionName := "email_verifications"

	repo := mongodb.NewMongoDBRepository[auth_entities.EmailVerification](client, dbName, entityType, collectionName, "EmailVerification")

	repo.InitQueryableFields(map[string]bool{
		"ID":        true,
		"UserID":    true,
		"Email":     true,
		"Token":     true,
		"Code":      true,
		"Type":      true,
		"Status":    true,
		"ExpiresAt": true,
		"CreatedAt": true,
	}, map[string]string{
		"ID":          "_id",
		"UserID":      "user_id",
		"Email":       "email",
		"Token":       "token",
		"Code":        "code",
		"Type":        "type",
		"Status":      "status",
		"Attempts":    "attempts",
		"MaxAttempts": "max_attempts",
		"ExpiresAt":   "expires_at",
		"VerifiedAt":  "verified_at",
		"CreatedAt":   "created_at",
		"UpdatedAt":   "updated_at",
		"IPAddress":   "ip_address",
		"UserAgent":   "user_agent",
	})

	return &EmailVerificationRepository{
		MongoDBRepository: *repo,
	}
}

// collection returns the MongoDB collection for email verifications
func (r *EmailVerificationRepository) collection() *mongo.Collection {
	return r.MongoDBRepository.Collection()
}

// Save creates a new email verification record
func (r *EmailVerificationRepository) Save(ctx context.Context, verification *auth_entities.EmailVerification) error {
	_, err := r.MongoDBRepository.Create(ctx, verification)
	if err != nil {
		return fmt.Errorf("failed to save email verification: %w", err)
	}
	return nil
}

// FindByID retrieves a verification by its ID
func (r *EmailVerificationRepository) FindByID(ctx context.Context, id uuid.UUID) (*auth_entities.EmailVerification, error) {
	return r.MongoDBRepository.GetByID(ctx, id)
}

// FindByToken retrieves a verification by its token
func (r *EmailVerificationRepository) FindByToken(ctx context.Context, token string) (*auth_entities.EmailVerification, error) {
	var verification auth_entities.EmailVerification
	err := r.MongoDBRepository.FindOneWithRLS(ctx, bson.M{
		"token":  token,
		"status": auth_entities.VerificationStatusPending,
	}).Decode(&verification)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("verification not found")
		}
		return nil, fmt.Errorf("failed to find verification by token: %w", err)
	}
	return &verification, nil
}

// FindByUserID retrieves the latest verification for a user
func (r *EmailVerificationRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*auth_entities.EmailVerification, error) {
	var verification auth_entities.EmailVerification

	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})
	err := r.MongoDBRepository.FindOneWithRLS(ctx, bson.M{"user_id": userID}, opts).Decode(&verification)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no verification found for user")
		}
		return nil, fmt.Errorf("failed to find verification by user_id: %w", err)
	}
	return &verification, nil
}

// FindPendingByEmail retrieves pending verifications for an email
func (r *EmailVerificationRepository) FindPendingByEmail(ctx context.Context, email string) (*auth_entities.EmailVerification, error) {
	var verification auth_entities.EmailVerification

	filter := bson.M{
		"email":      email,
		"status":     auth_entities.VerificationStatusPending,
		"expires_at": bson.M{"$gt": time.Now().UTC()},
	}
	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})

	err := r.collection().FindOne(ctx, filter, opts).Decode(&verification)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No pending verification found
		}
		return nil, fmt.Errorf("failed to find pending verification: %w", err)
	}
	return &verification, nil
}

// Update updates an existing verification record
func (r *EmailVerificationRepository) Update(ctx context.Context, verification *auth_entities.EmailVerification) error {
	verification.UpdatedAt = time.Now().UTC()

	_, err := r.MongoDBRepository.Update(ctx, verification)
	if err != nil {
		return fmt.Errorf("failed to update verification: %w", err)
	}
	return nil
}

// InvalidatePreviousVerifications invalidates all previous verifications for a user
func (r *EmailVerificationRepository) InvalidatePreviousVerifications(ctx context.Context, userID uuid.UUID, email string) error {
	filter := bson.M{
		"$or": []bson.M{
			{"user_id": userID},
			{"email": email},
		},
		"status": auth_entities.VerificationStatusPending,
	}
	update := bson.M{
		"$set": bson.M{
			"status":     auth_entities.VerificationStatusCanceled,
			"updated_at": time.Now().UTC(),
		},
	}

	_, err := r.collection().UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to invalidate previous verifications: %w", err)
	}
	return nil
}

// CountRecentAttempts counts verification attempts in the last N minutes
func (r *EmailVerificationRepository) CountRecentAttempts(ctx context.Context, email string, minutes int) (int, error) {
	since := time.Now().UTC().Add(-time.Duration(minutes) * time.Minute)

	filter := bson.M{
		"email":      email,
		"created_at": bson.M{"$gte": since},
	}

	count, err := r.collection().CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count recent attempts: %w", err)
	}
	return int(count), nil
}

// EnsureIndexes creates the necessary indexes for the email_verifications collection
func (r *EmailVerificationRepository) EnsureIndexes(ctx context.Context) error {
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
			Keys: bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(86400 * 7), // TTL: 7 days
		},
	}

	_, err := r.collection().Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}
	return nil
}

// Ensure EmailVerificationRepository implements the interface
var _ auth_out.EmailVerificationRepository = (*EmailVerificationRepository)(nil)

