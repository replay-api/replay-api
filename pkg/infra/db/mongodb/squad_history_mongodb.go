package db

import (
	"context"
	"log/slog"
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"

	common "github.com/replay-api/replay-api/pkg/domain"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
)

type SquadHistoryRepository struct {
	MongoDBRepository[squad_entities.SquadHistory]
}

func NewSquadHistoryRepository(client *mongo.Client, dbName string, entityType squad_entities.SquadHistory, collectionName string) *SquadHistoryRepository {
	repo := MongoDBRepository[squad_entities.SquadHistory]{
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
		"ID":            true,
		"SquadID":       true,
		"UserID":        true,
		"Action":        true,
		"ResourceOwner": true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}, map[string]string{
		"ID":            "_id",
		"SquadID":       "squad_id",
		"UserID":        "user_id",
		"Action":        "action",
		"ResourceOwner": "resource_owner",
		"CreatedAt":     "created_at",
		"UpdatedAt":     "updated_at",
	})

	return &SquadHistoryRepository{
		repo,
	}
}

func (r *SquadHistoryRepository) Search(ctx context.Context, s common.Search) ([]squad_entities.SquadHistory, error) {
	cursor, err := r.Query(ctx, s)
	if cursor != nil {
		defer cursor.Close(ctx)
	}

	if err != nil {
		slog.ErrorContext(ctx, "error querying SquadHistory entity", "err", err)
		return nil, err
	}

	squadHistories := make([]squad_entities.SquadHistory, 0)

	for cursor.Next(ctx) {
		var p squad_entities.SquadHistory
		err := cursor.Decode(&p)

		if err != nil {
			slog.ErrorContext(ctx, "error decoding SquadHistory entity", "err", err)
			return nil, err
		}

		squadHistories = append(squadHistories, p)
	}

	return squadHistories, nil
}
