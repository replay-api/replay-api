package db

import (
	"context"
	"log/slog"
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"

	common "github.com/replay-api/replay-api/pkg/domain"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
)

type PlayerProfileHistoryRepository struct {
	MongoDBRepository[squad_entities.PlayerProfileHistory]
}

func NewPlayerProfileHistoryRepository(client *mongo.Client, dbName string, entityType squad_entities.PlayerProfileHistory, collectionName string) *PlayerProfileHistoryRepository {
	repo := MongoDBRepository[squad_entities.PlayerProfileHistory]{
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
		"ID":        true,
		"PlayerID":  true,
		"Changes":   true,
		"CreatedAt": true,
	}, map[string]string{
		"ID":              "baseentity._id", // TODO: review ; (opcional: aqui a exceção se tornou a regra. deixar default o que está na annotation da prop.) talvez seja melhor refletir os tipos nas anotacoes json/bson
		"PlayerID":        "player_id",
		"Changes":         "changes",
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

	return &PlayerProfileHistoryRepository{
		repo,
	}
}

func (r *PlayerProfileHistoryRepository) Search(ctx context.Context, s common.Search) ([]squad_entities.PlayerProfileHistory, error) {
	cursor, err := r.Query(ctx, s)
	if cursor != nil {
		defer cursor.Close(ctx)
	}

	if err != nil {
		slog.ErrorContext(ctx, "error querying player profile entity", "err", err)
		return nil, err
	}

	players := make([]squad_entities.PlayerProfileHistory, 0)

	for cursor.Next(ctx) {
		var p squad_entities.PlayerProfileHistory
		err := cursor.Decode(&p)

		if err != nil {
			slog.ErrorContext(ctx, "error decoding player profile entity", "err", err)
			return nil, err
		}

		players = append(players, p)
	}

	return players, nil
}
