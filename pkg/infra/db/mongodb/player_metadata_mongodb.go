package db

import (
	"context"
	"log/slog"
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"

	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	replay_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
)

type PlayerMetadataRepository struct {
	MongoDBRepository[replay_entity.PlayerMetadata]
}

func NewPlayerMetadataRepository(client *mongo.Client, dbName string, entityType replay_entity.PlayerMetadata, collectionName string) *PlayerMetadataRepository {
	repo := MongoDBRepository[replay_entity.PlayerMetadata]{
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

	return &PlayerMetadataRepository{
		repo,
	}
}

func (r *PlayerMetadataRepository) Search(ctx context.Context, s common.Search) ([]replay_entity.PlayerMetadata, error) {
	cursor, err := r.Query(ctx, s)
	if cursor != nil {
		defer cursor.Close(ctx)
	}

	if err != nil {
		slog.ErrorContext(ctx, "error querying player entity", "err", err)
		return nil, err
	}

	players := make([]replay_entity.PlayerMetadata, 0)

	for cursor.Next(ctx) {
		var p replay_entity.PlayerMetadata
		err := cursor.Decode(&p)

		if err != nil {
			slog.ErrorContext(ctx, "error decoding player entity", "err", err)
			return nil, err
		}

		players = append(players, p)
	}

	return players, nil
}

func (r *PlayerMetadataRepository) CreateMany(createCtx context.Context, events []interface{}) error {
	_, err := r.collection.InsertMany(createCtx, events)
	if err != nil {
		slog.ErrorContext(createCtx, err.Error())
		return err
	}

	return nil
}

func (r *PlayerMetadataRepository) Create(createCtx context.Context, events ...replay_entity.PlayerMetadata) error {
	toInsert := make([]interface{}, len(events))

	for i := range events {
		toInsert[i] = events[i]
	}

	_, err := r.collection.InsertMany(createCtx, toInsert)
	if err != nil {
		slog.ErrorContext(createCtx, err.Error())
		return err
	}

	return nil
}
