package db

import (
	"context"
	"log/slog"
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"

	common "github.com/replay-api/replay-api/pkg/domain"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
)

type SquadRepository struct {
	MongoDBRepository[squad_entities.Squad]
}

func NewSquadRepository(client *mongo.Client, dbName string, entityType squad_entities.Squad, collectionName string) *SquadRepository {
	repo := MongoDBRepository[squad_entities.Squad]{
		mongoClient:       client,
		dbName:            dbName,
		mappingCache:      make(map[string]CacheItem),
		entityModel:       reflect.TypeOf(entityType),
		bsonFieldMappings: make(map[string]string),
		collectionName:    collectionName,
		entityName:        reflect.TypeOf(entityType).Name(),
		queryableFields:   make(map[string]bool),
		collection:        client.Database(dbName).Collection(collectionName),
	}

	repo.InitQueryableFields(map[string]bool{
		"ID":              true,
		"GameID":          true,
		"Name":            true,
		"Symbol":          true,
		"SlugURI":         true,
		"Description":     true,
		"Membership.*":    true,
		"VisibilityLevel": true,
		"VisibilityType":  true,
		"ResourceOwner.*": true,
		"CreatedAt":       true,
		"UpdatedAt":       true,
	}, map[string]string{
		"ID":                 "baseentity._id",
		"GameID":             "game_id",
		"Name":               "name",
		"Symbol":             "symbol",
		"SlugURI":            "slug_uri",
		"Description":        "description",
		"Membership":         "membership",
		"Membership.Type":    "membership.type",
		"Membership.Roles":   "membership.roles",
		"Membership.Status":  "membership.status",
		"Membership.History": "membership.history",
		"VisibilityLevel":    "baseentity.visibility_level",
		"VisibilityType":     "baseentity.visibility_type",
		"ResourceOwner":      "baseentity.resource_owner",
		"TenantID":           "baseentity.resource_owner.tenant_id",
		"UserID":             "baseentity.resource_owner.user_id",
		"GroupID":            "baseentity.resource_owner.group_id",
		"ClientID":           "baseentity.resource_owner.client_id",
		"CreatedAt":          "baseentity.create_at",
		"UpdatedAt":          "baseentity.updated_at",
	})

	return &SquadRepository{
		repo,
	}
}

func (r *SquadRepository) Search(ctx context.Context, s common.Search) ([]squad_entities.Squad, error) {
	cursor, err := r.Query(ctx, s)
	if cursor != nil {
		defer cursor.Close(ctx)
	}

	if err != nil {
		slog.ErrorContext(ctx, "error querying squad entity", "err", err)
		return nil, err
	}

	squads := make([]squad_entities.Squad, 0)

	for cursor.Next(ctx) {
		var p squad_entities.Squad
		err := cursor.Decode(&p)

		if err != nil {
			slog.ErrorContext(ctx, "error decoding squad entity", "err", err)
			return nil, err
		}

		squads = append(squads, p)
	}

	return squads, nil
}
