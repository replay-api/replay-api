package db

import (
	"context"
	"log/slog"
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"

	shared "github.com/resource-ownership/go-common/pkg/common"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
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
		BsonFieldMappings: make(map[string]string),
		collectionName:    collectionName,
		entityName:        reflect.TypeOf(entityType).Name(),
		QueryableFields:   make(map[string]bool),
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
		"ID":                     "_id",
		"RIDSource":              "rid_source",
		"Type":                   "rid_source",
		"SourceKey":              "source_key",
		"Details":                "details",
		"Details.ID":             "details._id",
		"ResourceOwner":          "resource_owner", // TODO: principalmente resource ownership, que é padronizado.
		"ResourceOwner.TenantID": "resource_owner.tenant_id",
		"ResourceOwner.UserID":   "resource_owner.user_id",
		"ResourceOwner.GroupID":  "resource_owner.group_id",
		"ResourceOwner.ClientID": "resource_owner.client_id",
		"CreatedAt":              "created_at",
		"UpdatedAt":              "updated_at",
	})

	return &ProfileRepository{
		repo,
	}
}

func (r *ProfileRepository) Search(ctx context.Context, s shared.Search) ([]iam_entities.Profile, error) {
	slog.InfoContext(ctx, "searching profile entity", "search", s)
	cursor, err := r.Query(ctx, s)
	if cursor != nil {
		defer cursor.Close(ctx)
	}

	if err != nil {
		slog.ErrorContext(ctx, "error querying profile entity", "err", err)
		return nil, err
	}

	profiles := make([]iam_entities.Profile, 0)

	for cursor.Next(ctx) {
		var p iam_entities.Profile
		err := cursor.Decode(&p)

		if err != nil {
			slog.ErrorContext(ctx, "error decoding profile entity", "err", err)
			return nil, err
		}

		profiles = append(profiles, p)
	}

	return profiles, nil
}
