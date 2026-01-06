package db

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"

	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	shared "github.com/resource-ownership/go-common/pkg/common"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
)

type BadgeRepository struct {
	mongodb.MongoDBRepository[replay_entity.Badge]
}

func NewBadgeRepository(client *mongo.Client, dbName string, entityType replay_entity.Badge, collectionName string) *BadgeRepository {
	repo := mongodb.NewMongoDBRepository[replay_entity.Badge](client, dbName, entityType, collectionName, "Badge")

	repo.InitQueryableFields(map[string]bool{
		"ID":            true,
		"GameID":        true,
		"MatchID":       true,
		"PlayerID":      true,
		"Name":          true,
		"Events":        true,
		"Description":   true,
		"ImageURL":      true,
		"ResourceOwner": true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}, map[string]string{
		"ID":                     "_id",
		"GameID":                 "game_id",
		"MatchID":                "match_id",
		"PlayerID":               "player_id",
		"Name":                   "name",
		"Events":                 "events",
		"Description":            "description",
		"ResourceOwner":          "resource_owner", // TODO: principalmente resource ownership, que é padronizado.
		"ResourceOwner.TenantID": "resource_owner.tenant_id",
		"ResourceOwner.UserID":   "resource_owner.user_id",
		"ResourceOwner.GroupID":  "resource_owner.group_id",
		"ResourceOwner.ClientID": "resource_owner.client_id",
		"CreatedAt":              "create_at",
		"UpdatedAt":              "updated_at",
	})

	return &BadgeRepository{
		MongoDBRepository: *repo,
	}
}

func (r *BadgeRepository) Search(ctx context.Context, s shared.Search) ([]replay_entity.Badge, error) {
	return r.MongoDBRepository.Search(ctx, s)
}

func (r *BadgeRepository) CreateMany(createCtx context.Context, events []replay_entity.Badge) error {
	pointers := make([]*replay_entity.Badge, len(events))
	for i := range events {
		pointers[i] = &events[i]
	}
	return r.MongoDBRepository.CreateMany(createCtx, pointers)
}
