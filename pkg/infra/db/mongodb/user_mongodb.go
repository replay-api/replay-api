package db

import (
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"

	iam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
)

type UserRepository struct {
	MongoDBRepository[iam_entities.User]
}

func NewUserRepository(client *mongo.Client, dbName string, entityType *iam_entities.User, collectionName string) *UserRepository {
	repo := MongoDBRepository[iam_entities.User]{
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
		"Name":          true,
		"Type":          true,
		"ParentUserID":  true,
		"ResourceOwner": true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}, map[string]string{
		"ID":            "_id",
		"Name":          "name",
		"Type":          "type",
		"ParentUserID":  "parent_group_id",
		"ResourceOwner": "resource_owner",
		"CreatedAt":     "created_at",
		"UpdatedAt":     "updated_at",
	})

	return &UserRepository{
		repo,
	}
}
