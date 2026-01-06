package db

import (
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReplayFileMetadataRepository struct {
	mongodb.MongoDBRepository[replay_entity.ReplayFile]
}

func NewReplayFileMetadataRepository(client *mongo.Client, dbName string, entityType replay_entity.ReplayFile, collectionName string) *ReplayFileMetadataRepository {
	repo := mongodb.NewMongoDBRepository[replay_entity.ReplayFile](client, dbName, entityType, collectionName, "ReplayFile")

	repo.InitQueryableFields(map[string]bool{
		"ID":            true,
		"GameID":        true,
		"NetworkID":     true,
		"Size":          true,
		"InternalURI":   true,
		"Status":        true,
		"Error":         true,
		"Header":        true,
		"ResourceOwner": true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
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
		MongoDBRepository: *repo,
	}
}
