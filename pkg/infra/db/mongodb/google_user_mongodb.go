package db

import (
	google_entities "github.com/replay-api/replay-api/pkg/domain/google/entities"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
)

type GoogleUserRepository struct {
	mongodb.MongoDBRepository[google_entities.GoogleUser]
}

func NewGoogleUserMongoDBRepository(client *mongo.Client, dbName string, entityType google_entities.GoogleUser, collectionName string) *GoogleUserRepository {
	repo := mongodb.NewMongoDBRepository[google_entities.GoogleUser](client, dbName, entityType, collectionName, "GoogleUser")

	repo.InitQueryableFields(map[string]bool{
		"ID":            true,
		"VHash":         true,
		"Sub":           true,
		"Hd":            true,
		"GivenName":     true,
		"FamilyName":    true,
		"Email":         true,
		"Locale":        true,
		"EmailVerified": true,
		"ResourceOwner": true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}, map[string]string{
		"ID":            "id",
		"VHash":         "v_hash",
		"Sub":           "sub",
		"Hd":            "hd",
		"GivenName":     "given_name",
		"FamilyName":    "family_name",
		"Email":         "email",
		"Locale":        "locale",
		"EmailVerified": "email_verified",
		"ResourceOwner": "resource_owner",
		"CreatedAt":     "created_at",
		"UpdatedAt":     "updated_at",
	})

	return &GoogleUserRepository{
		MongoDBRepository: *repo,
	}
}
