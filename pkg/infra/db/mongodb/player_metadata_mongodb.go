package db

import (
	"context"
	"log/slog"

	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type PlayerMetadataRepository struct {
	mongodb.MongoDBRepository[replay_entity.PlayerMetadata]
}

func NewPlayerMetadataRepository(client *mongo.Client, dbName string, entityType replay_entity.PlayerMetadata, collectionName string) *PlayerMetadataRepository {
	repo := mongodb.NewMongoDBRepository[replay_entity.PlayerMetadata](client, dbName, entityType, collectionName, "PlayerMetadata")

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
		MongoDBRepository: *repo,
	}
}

func (r *PlayerMetadataRepository) Search(ctx context.Context, s shared.Search) ([]replay_entity.PlayerMetadata, error) {
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

func (r *PlayerMetadataRepository) CreateMany(createCtx context.Context, events []replay_entity.PlayerMetadata) error {
	toInsert := make([]interface{}, len(events))

	for i := range events {
		toInsert[i] = events[i]
	}

	_, err := r.MongoDBRepository.Collection().InsertMany(createCtx, toInsert)
	if err != nil {
		slog.ErrorContext(createCtx, err.Error())
		return err
	}

	return nil
}

func (r *PlayerMetadataRepository) Create(createCtx context.Context, event replay_entity.PlayerMetadata) error {
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"_id": event.ID}
	update := bson.M{"$set": event}
	_, err := r.MongoDBRepository.Collection().UpdateOne(createCtx, filter, update, opts)
	if err != nil {
		slog.ErrorContext(createCtx, err.Error())
		return err
	}

	return nil
}
