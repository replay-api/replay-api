package db

import (
	"reflect"

	google_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/google/entities"
	"go.mongodb.org/mongo-driver/mongo"
)

type GoogleUserRepository struct {
	MongoDBRepository[google_entities.GoogleUser]
}

func NewGoogleUserMongoDBRepository(client *mongo.Client, dbName string, entityType google_entities.GoogleUser, collectionName string) *GoogleUserRepository {
	repo := MongoDBRepository[google_entities.GoogleUser]{
		mongoClient:       client,
		dbName:            dbName,
		mappingCache:      make(map[string]CacheItem),
		entityModel:       reflect.TypeOf(entityType),
		bsonFieldMappings: make(map[string]string),
		collectionName:    collectionName,
		entityName:        reflect.TypeOf(entityType).Name(),
		queryableFields:   make(map[string]bool),
	}

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
		repo,
	}
}
