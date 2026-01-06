package db

import (
	"context"
	"log/slog"

	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/mongo"

	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type SquadHistoryRepository struct {
	mongodb.MongoDBRepository[squad_entities.SquadHistory]
}

func NewSquadHistoryRepository(client *mongo.Client, dbName string, entityType squad_entities.SquadHistory, collectionName string) *SquadHistoryRepository {
	repo := mongodb.NewMongoDBRepository[squad_entities.SquadHistory](client, dbName, entityType, collectionName, "SquadHistory")

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
		MongoDBRepository: *repo,
	}
}

func (r *SquadHistoryRepository) Search(ctx context.Context, s shared.Search) ([]squad_entities.SquadHistory, error) {
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
