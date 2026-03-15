package db

import (
	"context"

	"github.com/google/uuid"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReplayFileMetadataRepository struct {
	mongodb.MongoDBRepository[replay_entity.ReplayFile]
}

func NewReplayFileMetadataRepository(client *mongo.Client, dbName string, entityType replay_entity.ReplayFile, collectionName string) *ReplayFileMetadataRepository {
	repo := mongodb.NewMongoDBRepository(client, dbName, entityType, collectionName, "ReplayFile")

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
		"ContentHash":   true,
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
		"ContentHash":            "content_hash",
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

// FindByContentHash finds a replay file by its content hash (SHA256)
// Used for deduplication - returns nil if no matching hash found
func (r *ReplayFileMetadataRepository) FindByContentHash(ctx context.Context, contentHash string) (*replay_entity.ReplayFile, error) {
	if contentHash == "" {
		return nil, nil
	}

	collection := r.MongoDBRepository.Collection()
	filter := bson.M{"content_hash": contentHash}

	var result replay_entity.ReplayFile
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &result, nil
}

// GetByID retrieves a replay file by its ID
func (r *ReplayFileMetadataRepository) GetByID(ctx context.Context, replayFileID uuid.UUID) (*replay_entity.ReplayFile, error) {
	collection := r.MongoDBRepository.Collection()
	filter := bson.M{"_id": replayFileID}

	var result replay_entity.ReplayFile
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &result, nil
}
