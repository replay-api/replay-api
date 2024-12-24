package db

import (
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"

	iam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
)

type ProfileRepository struct {
	MongoDBRepository[iam_entities.Profile]
}

func NewProfileRepository(client *mongo.Client, dbName string, entityType *iam_entities.Profile, collectionName string) *ProfileRepository {
	repo := MongoDBRepository[iam_entities.Profile]{
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
		"RIDSource":     true,
		"SourceKey":     true,
		"Details":       true,
		"ResourceOwner": true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}, map[string]string{
		"ID":            `json:"id" bson:"_id"`,
		"RIDSource":     `json:"rid_source" bson:"rid_source"`,
		"SourceKey":     `json:"source_key" bson:"source_key"`,
		"Details":       `json:"details" bson:"details"`,
		"ResourceOwner": `json:"resource_owner" bson:"resource_owner"`,
		"CreatedAt":     `json:"created_at" bson:"created_at"`,
		"UpdatedAt":     `json:"updated_at" bson:"updated_at"`,
	})

	return &ProfileRepository{
		repo,
	}
}
