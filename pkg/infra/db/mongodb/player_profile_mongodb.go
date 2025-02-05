package db

import (
	"context"
	"log/slog"
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"

	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	squad_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/entities"
)

type PlayerProfileRepository struct {
	MongoDBRepository[squad_entities.PlayerProfile]
}

func NewPlayerProfileRepository(client *mongo.Client, dbName string, entityType squad_entities.PlayerProfile, collectionName string) *PlayerProfileRepository {
	repo := MongoDBRepository[squad_entities.PlayerProfile]{
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
		"Nickname":        true,
		"Avatar":          true,
		"Roles":           true,
		"Description":     true,
		"VisibilityLevel": true,
		"VisibilityType":  true,
		"ResourceOwner":   true,
		"CreatedAt":       true,
		"UpdatedAt":       true,
	}, map[string]string{
		"ID":              "baseentity._id", // TODO: review ; (opcional: aqui a exceção se tornou a regra. deixar default o que está na annotation da prop.) talvez seja melhor refletir os tipos nas anotacoes json/bson
		"GameID":          "game_id",
		"Nickname":        "nickname",
		"Avatar":          "avatar",
		"Roles":           "roles",
		"Description":     "description",
		"VisibilityLevel": "baseentity.visibility_level",
		"VisibilityType":  "baseentity.visibility_type",
		"ResourceOwner":   "baseentity.resource_owner", // TODO: principalmente resource ownership, que é padronizado.
		"TenantID":        "baseentity.resource_owner.tenant_id",
		"UserID":          "baseentity.resource_owner.user_id",
		"GroupID":         "baseentity.resource_owner.group_id",
		"ClientID":        "baseentity.resource_owner.client_id",
		"CreatedAt":       "baseentity.create_at",
		"UpdatedAt":       "baseentity.updated_at",
	})

	return &PlayerProfileRepository{
		repo,
	}
}

func (r *PlayerProfileRepository) Search(ctx context.Context, s common.Search) ([]squad_entities.PlayerProfile, error) {
	cursor, err := r.Query(ctx, s)
	if cursor != nil {
		defer cursor.Close(ctx)
	}

	if err != nil {
		slog.ErrorContext(ctx, "error querying player profile entity", "err", err)
		return nil, err
	}

	players := make([]squad_entities.PlayerProfile, 0)

	for cursor.Next(ctx) {
		var p squad_entities.PlayerProfile
		err := cursor.Decode(&p)

		if err != nil {
			slog.ErrorContext(ctx, "error decoding player profile entity", "err", err)
			return nil, err
		}

		players = append(players, p)
	}

	return players, nil
}
