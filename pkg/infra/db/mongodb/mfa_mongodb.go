package db

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	auth_entities "github.com/replay-api/replay-api/pkg/domain/auth/entities"
	auth_out "github.com/replay-api/replay-api/pkg/domain/auth/ports/out"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const mfaCollectionName = "user_mfa"

type MFAMongoDBRepository struct {
	mongodb.MongoDBRepository[auth_entities.UserMFA]
}

func NewMFAMongoDBRepository(client *mongo.Client, dbName string) auth_out.MFARepository {
	repo := mongodb.NewMongoDBRepository[auth_entities.UserMFA](client, dbName, auth_entities.UserMFA{}, mfaCollectionName, "UserMFA")

	repo.InitQueryableFields(map[string]bool{
		"ID":              true,
		"UserID":          true,
		"Method":          true,
		"Status":          true,
		"BackupCodesLeft": true,
		"VerifiedAt":      true,
		"LastUsedAt":      true,
		"CreatedAt":       true,
		"UpdatedAt":       true,
	}, map[string]string{
		"ID":              "_id",
		"UserID":          "user_id",
		"Method":          "method",
		"Status":          "status",
		"BackupCodesLeft": "backup_codes_left",
		"VerifiedAt":      "verified_at",
		"LastUsedAt":      "last_used_at",
		"CreatedAt":       "created_at",
		"UpdatedAt":       "updated_at",
	})

	mfaRepo := &MFAMongoDBRepository{
		MongoDBRepository: *repo,
	}

	// Create indexes
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := mfaRepo.MongoDBRepository.Collection().Indexes().CreateMany(ctx, indexes)
	if err != nil {
		slog.Warn("Failed to create MFA indexes", "error", err)
	}

	return mfaRepo
}

// Create creates a new MFA configuration
func (r *MFAMongoDBRepository) Create(ctx context.Context, mfa *auth_entities.UserMFA) (*auth_entities.UserMFA, error) {
	if mfa.ID == uuid.Nil {
		mfa.ID = uuid.New()
	}
	mfa.CreatedAt = time.Now()
	mfa.UpdatedAt = time.Now()

	_, err := r.MongoDBRepository.Create(ctx, mfa)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create MFA", "error", err, "user_id", mfa.UserID)
		return nil, err
	}

	slog.InfoContext(ctx, "MFA created successfully", "id", mfa.ID, "user_id", mfa.UserID)
	return mfa, nil
}

// GetByUserID gets the MFA configuration for a user
func (r *MFAMongoDBRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*auth_entities.UserMFA, error) {
	var mfa auth_entities.UserMFA

	err := r.MongoDBRepository.FindOneWithRLS(ctx, bson.M{"user_id": userID}).Decode(&mfa)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, err
		}
		slog.ErrorContext(ctx, "Failed to get MFA by user ID", "error", err, "user_id", userID)
		return nil, err
	}

	return &mfa, nil
}

// Update updates an existing MFA configuration
func (r *MFAMongoDBRepository) Update(ctx context.Context, mfa *auth_entities.UserMFA) (*auth_entities.UserMFA, error) {
	mfa.UpdatedAt = time.Now()

	_, err := r.MongoDBRepository.Update(ctx, mfa)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update MFA", "error", err, "id", mfa.ID)
		return nil, err
	}

	slog.InfoContext(ctx, "MFA updated successfully", "id", mfa.ID, "user_id", mfa.UserID)
	return mfa, nil
}

// Delete deletes an MFA configuration
func (r *MFAMongoDBRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.MongoDBRepository.DeleteOneWithRLS(ctx, bson.M{"_id": id})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to delete MFA", "error", err, "id", id)
		return err
	}

	if result.DeletedCount == 0 {
		slog.WarnContext(ctx, "MFA not found for deletion", "id", id)
		return mongo.ErrNoDocuments
	}

	slog.InfoContext(ctx, "MFA deleted successfully", "id", id)
	return nil
}

