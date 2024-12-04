package db

import (
	"context"
	"log/slog"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	replay_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
)

type PlayerRepository struct {
	MongoDBRepository[replay_entity.Player]
}

func NewPlayerRepository(client *mongo.Client, dbName string, entityType replay_entity.Player, collectionName string) *PlayerRepository {
	repo := MongoDBRepository[replay_entity.Player]{
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
		"ID":                 true,
		"GameID":             true,
		"NetworkID":          true,
		"NetworkUserID":      true,
		"CurrentDisplayName": true,
		"NameHistory":        true,
		"ResourceOwner":      true,
		"CreatedAt":          true,
		"UpdatedAt":          true,
	}, map[string]string{
		"ID":                 "_id", // TODO: review ; (opcional: aqui a exceção se tornou a regra. deixar default o que está na annotation da prop.) talvez seja melhor refletir os tipos nas anotacoes json/bson
		"GameID":             "game_id",
		"NetworkID":          "network_id",
		"NetworkUserID":      "network_user_id",
		"CurrentDisplayName": "current_display_name",
		"NameHistory":        "name_history",
		"ResourceOwner":      "resource_owner", // TODO: principalmente resource ownership, que é padronizado.
		"TenantID":           "resource_owner.tenant_id",
		"UserID":             "resource_owner.user_id",
		"GroupID":            "resource_owner.group_id",
		"ClientID":           "resource_owner.client_id",
		"CreatedAt":          "create_at",
		"UpdatedAt":          "updated_at",
	})

	return &PlayerRepository{
		repo,
	}
}

func (r *PlayerRepository) Search(ctx context.Context, s common.Search) ([]replay_entity.Player, error) {
	cursor, err := r.Query(ctx, s)
	if cursor != nil {
		defer cursor.Close(ctx)
	}

	if err != nil {
		slog.ErrorContext(ctx, "error querying player entity", "err", err)
		return nil, err
	}

	players := make([]replay_entity.Player, 0)

	for cursor.Next(ctx) {
		var p replay_entity.Player
		err := cursor.Decode(&p)

		if err != nil {
			slog.ErrorContext(ctx, "error decoding player entity", "err", err)
			return nil, err
		}

		players = append(players, p)
	}

	return players, nil
}

func (r *PlayerRepository) CreateMany(createCtx context.Context, events []interface{}) error {
	collection := r.mongoClient.Database(r.dbName).Collection(r.collectionName)

	queryCtx, cancel := context.WithTimeout(createCtx, 10*time.Second)
	defer cancel()

	_, err := collection.InsertMany(queryCtx, events)
	if err != nil {
		slog.ErrorContext(queryCtx, err.Error())
		return err
	}

	return nil
}

func (r *PlayerRepository) Create(createCtx context.Context, events ...replay_entity.Player) error {
	collection := r.mongoClient.Database(r.dbName).Collection(r.collectionName)

	queryCtx, cancel := context.WithTimeout(createCtx, 10*time.Second)
	defer cancel()

	toInsert := make([]interface{}, len(events))

	for i := range events {
		toInsert[i] = events[i]
	}

	_, err := collection.InsertMany(queryCtx, toInsert)
	if err != nil {
		slog.ErrorContext(queryCtx, err.Error())
		return err
	}

	return nil
}
