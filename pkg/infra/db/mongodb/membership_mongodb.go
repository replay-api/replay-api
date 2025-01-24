package db

import (
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"

	iam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
)

type MembershipRepository struct {
	MongoDBRepository[iam_entities.Membership]
}

func NewMembershipRepository(client *mongo.Client, dbName string, entityType *iam_entities.Membership, collectionName string) *MembershipRepository {
	repo := MongoDBRepository[iam_entities.Membership]{
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
		"Type":          true,
		"ResourceOwner": true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}, map[string]string{
		"ID":            "_id",
		"Type":          "type",
		"ResourceOwner": "resource_owner",
		"CreatedAt":     "created_at",
		"UpdatedAt":     "updated_at",
	})

	return &MembershipRepository{
		repo,
	}
}
