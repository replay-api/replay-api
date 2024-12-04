package db

import (
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"

	iam_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
)

type RIDTokenRepository struct {
	MongoDBRepository[iam_entity.RIDToken]
}

func NewRIDTokenRepository(client *mongo.Client, dbName string, entityType iam_entity.RIDToken, collectionName string) *RIDTokenRepository {
	// TODO: create Factory for encapsulating this and reducing bloat/repetition of some fields like queryableFields, mappingCache.. this needs a facade for a clearer instantiation
	repo := MongoDBRepository[iam_entity.RIDToken]{
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
		"ID":                     true,
		"Key":                    true,
		"Source":                 true,
		"ResourceOwner":          true,
		"ExpiresAt":              true,
		"CreatedAt":              true,
		"UpdatedAt":              true,
		"ResourceOwner.TenantID": true,
		"ResourceOwner.UserID":   true,
		"ResourceOwner.GroupID":  true,
		"ResourceOwner.ClientID": true,
	}, map[string]string{
		"ID":                     "_id",
		"Key":                    "key",
		"Source":                 "source",
		"ResourceOwner":          "resource_owner",
		"ExpiresAt":              "expires_at",
		"CreatedAt":              "created_at",
		"UpdatedAt":              "updated_at",
		"ResourceOwner.TenantID": "resource_owner.tenant_id",
		"ResourceOwner.UserID":   "resource_owner.user_id",
		"ResourceOwner.GroupID":  "resource_owner.group_id",
		"ResourceOwner.ClientID": "resource_owner.client_id",
	})

	return &RIDTokenRepository{
		repo,
	}
}
