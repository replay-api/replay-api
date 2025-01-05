package db

import (
	"context"
	"log/slog"
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"

	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	squad_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/entities"
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
		"ID":            true,
		"GroupID":       true,
		"GameID":        true,
		"FullName":      true,
		"ShortName":     true,
		"Description":   true,
		"Profiles":      true,
		"ResourceOwner": true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}, map[string]string{
		"ID":                 "_id",
		"GroupID":            "group_id",
		"GameID":             "game_id",
		"FullName":           "full_name",
		"CurrentDisplayName": "short_name",
		"Symbol":             "symbol",
		"Description":        "description",
		"Profiles":           "profiles",
		"ResourceOwner":      "resource_owner",
		"TenantID":           "resource_owner.tenant_id",
		"UserID":             "resource_owner.user_id",
		// "GroupID":            "resource_owner.group_id",
		"ClientID":  "resource_owner.client_id",
		"CreatedAt": "create_at",
		"UpdatedAt": "updated_at",
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
		slog.ErrorContext(ctx, "error querying player entity", "err", err)
		return nil, err
	}

	players := make([]squad_entities.Squad, 0)

	for cursor.Next(ctx) {
		var p squad_entities.Squad
		err := cursor.Decode(&p)

		if err != nil {
			slog.ErrorContext(ctx, "error decoding player entity", "err", err)
			return nil, err
		}

		players = append(players, p)
	}

	return players, nil
}

// func (r *SquadRepository) CreateMany(createCtx context.Context, events []interface{}) error {
// 	_, err := r.collection.InsertMany(createCtx, events)
// 	if err != nil {
// 		slog.ErrorContext(createCtx, err.Error())
// 		return err
// 	}

// 	return nil
// }

// func (r *SquadRepository) Create(createCtx context.Context, events ...squad_entities.Squad) error {
// 	toInsert := make([]interface{}, len(events))

// 	for i := range events {
// 		toInsert[i] = events[i]
// 	}

// 	_, err := r.collection.InsertMany(createCtx, toInsert)
// 	if err != nil {
// 		slog.ErrorContext(createCtx, err.Error())
// 		return err
// 	}

// 	return nil
// }
