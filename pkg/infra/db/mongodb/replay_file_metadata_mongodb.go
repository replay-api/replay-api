package db

import (
	"reflect"

	replay_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReplayFileMetadataRepository struct {
	MongoDBRepository[replay_entity.ReplayFile]
}

func NewReplayFileMetadataRepository(client *mongo.Client, dbName string, entityType replay_entity.ReplayFile, collectionName string) *ReplayFileMetadataRepository {
	repo := MongoDBRepository[replay_entity.ReplayFile]{
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
		"ID":               true,
		"GameID":           true,
		"NetworkID":        true,
		"Size":             true,
		"InternalURI":      true,
		"Status":           true,
		"Error":            true,
		"Header":           true,
		"Header.Filestamp": true,
		"ResourceOwner":    true,
		"CreatedAt":        true,
		"UpdatedAt":        true,
	}, map[string]string{
		"ID":                     "_id",
		"GameID":                 "game_id",
		"NetworkID":              "network_id",
		"Size":                   "size",
		"InternalURI":            "uri",
		"Status":                 "status",
		"Error":                  "error",
		"Header":                 "header",
		"ResourceOwner":          "resource_owner",
		"CreatedAt":              "created_at",
		"UpdatedAt":              "updated_at",
		"Header.Filestamp":       "header.filestamp",
		"ResourceOwner.TenantID": "resource_owner.tenant_id",
		"ResourceOwner.UserID":   "resource_owner.user_id",
		"ResourceOwner.GroupID":  "resource_owner.group_id",
		"ResourceOwner.ClientID": "resource_owner.client_id",
	})

	return &ReplayFileMetadataRepository{
		repo,
	}
}
