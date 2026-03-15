package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	email_entities "github.com/replay-api/replay-api/pkg/domain/email/entities"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type EmailUserRepository struct {
	mongodb.MongoDBRepository[email_entities.EmailUser]
}

func NewEmailUserMongoDBRepository(client *mongo.Client, dbName string, entityType email_entities.EmailUser, collectionName string) *EmailUserRepository {
	repo := mongodb.NewMongoDBRepository(client, dbName, entityType, collectionName, "EmailUser")

	repo.InitQueryableFields(map[string]bool{
		"ID":            true,
		"VHash":         true,
		"Email":         true,
		"PasswordHash":  true,
		"EmailVerified": true,
		"DisplayName":   true,
		"ResourceOwner": true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}, map[string]string{
		"ID":            "_id",
		"VHash":         "v_hash",
		"Email":         "email",
		"PasswordHash":  "password_hash",
		"EmailVerified": "email_verified",
		"DisplayName":   "display_name",
		"ResourceOwner": "resource_owner",
		"CreatedAt":     "created_at",
		"UpdatedAt":     "updated_at",
	})

	return &EmailUserRepository{
		MongoDBRepository: *repo,
	}
}

// MarkEmailVerified sets email_verified to true for the user with the given ID
func (r *EmailUserRepository) MarkEmailVerified(ctx context.Context, userID uuid.UUID) error {
	filter := bson.M{"resource_owner.user_id": userID}
	update := bson.M{
		"$set": bson.M{
			"email_verified": true,
			"updated_at":     time.Now(),
		},
	}

	result, err := r.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return nil // User not found — not necessarily an error (could be Google user)
	}

	return nil
}
