package db

import (
	"context"
	"log/slog"
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"

	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
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

func (r *ProfileRepository) Search(ctx context.Context, s common.Search) ([]iam_entities.Profile, error) {
	cursor, err := r.Query(ctx, s)
	if cursor != nil {
		defer cursor.Close(ctx)
	}

	if err != nil {
		slog.ErrorContext(ctx, "error querying user entity", "err", err)
		return nil, err
	}

	profiles := make([]iam_entities.Profile, 0)

	for cursor.Next(ctx) {
		var p iam_entities.Profile
		err := cursor.Decode(&p)

		if err != nil {
			slog.ErrorContext(ctx, "error decoding user entity", "err", err)
			return nil, err
		}

		profiles = append(profiles, p)
	}

	return profiles, nil
}
