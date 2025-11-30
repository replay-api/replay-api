package db

import (
	"reflect"

	email_entities "github.com/replay-api/replay-api/pkg/domain/email/entities"
	"go.mongodb.org/mongo-driver/mongo"
)

type EmailUserRepository struct {
	MongoDBRepository[email_entities.EmailUser]
}

func NewEmailUserMongoDBRepository(client *mongo.Client, dbName string, entityType email_entities.EmailUser, collectionName string) *EmailUserRepository {
	repo := MongoDBRepository[email_entities.EmailUser]{
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
		repo,
	}
}
