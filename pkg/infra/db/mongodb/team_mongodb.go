package db

import (
	"context"
	"log/slog"

	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/mongo"

	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type TeamRepository struct {
	mongodb.MongoDBRepository[replay_entity.Team]
}

func NewTeamRepository(client *mongo.Client, dbName string, entityType replay_entity.Team, collectionName string) *TeamRepository {
	repo := mongodb.NewMongoDBRepository[replay_entity.Team](client, dbName, entityType, collectionName, "Team")

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
		"ID":                     "_id", // TODO: review ; (opcional: aqui a exceção se tornou a regra. deixar default o que está na annotation da prop.) talvez seja melhor refletir os tipos nas anotacoes json/bson
		"GameID":                 "game_id",
		"NetworkID":              "network_id",
		"NetworkUserID":          "network_user_id",
		"CurrentDisplayName":     "current_display_name",
		"NameHistory":            "name_history",
		"ResourceOwner":          "resource_owner", // TODO: principalmente resource ownership, que é padronizado.
		"ResourceOwner.TenantID": "resource_owner.tenant_id",
		"ResourceOwner.UserID":   "resource_owner.user_id",
		"ResourceOwner.GroupID":  "resource_owner.group_id",
		"ResourceOwner.ClientID": "resource_owner.client_id",
		"CreatedAt":              "created_at",
		"UpdatedAt":              "updated_at",
	})

	return &TeamRepository{
		MongoDBRepository: *repo,
	}
}

func (r *TeamRepository) Search(ctx context.Context, s shared.Search) ([]replay_entity.Team, error) {
	cursor, err := r.Query(ctx, s)
	if cursor != nil {
		defer cursor.Close(ctx)
	}

	if err != nil {
		slog.ErrorContext(ctx, "error querying player entity", "err", err)
		return nil, err
	}

	players := make([]replay_entity.Team, 0)

	for cursor.Next(ctx) {
		var p replay_entity.Team
		err := cursor.Decode(&p)

		if err != nil {
			slog.ErrorContext(ctx, "error decoding player entity", "err", err)
			return nil, err
		}

		players = append(players, p)
	}

	return players, nil
}

func (r *TeamRepository) CreateMany(createCtx context.Context, events []replay_entity.Team) error {
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
