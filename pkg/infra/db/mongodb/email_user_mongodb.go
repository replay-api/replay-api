package db

import (
	email_entities "github.com/replay-api/replay-api/pkg/domain/email/entities"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
)

type EmailUserRepository struct {
	mongodb.MongoDBRepository[email_entities.EmailUser]
}

func NewEmailUserMongoDBRepository(client *mongo.Client, dbName string, entityType email_entities.EmailUser, collectionName string) *EmailUserRepository {
	repo := mongodb.NewMongoDBRepository[email_entities.EmailUser](client, dbName, entityType, collectionName, "EmailUser")

	repo.InitQueryableFields(map[string]bool{
		"ID":            true,
		"VHash":         true,
		"Email":         true,
		"EmailVerified": true,
		"DisplayName":   true,
		"ResourceOwner": true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}, map[string]string{
		"ID":            "_id",
		"VHash":         "v_hash",
		"Email":         "email",
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
