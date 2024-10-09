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

type BadgeRepository struct {
	MongoDBRepository[replay_entity.Badge]
}

func NewBadgeRepository(client *mongo.Client, dbName string, entityType replay_entity.Badge, collectionName string) *BadgeRepository {
	repo := MongoDBRepository[replay_entity.Badge]{
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
		repo,
	}
}

func (r *BadgeRepository) Search(ctx context.Context, s common.Search) ([]replay_entity.Badge, error) {
	cursor, err := r.Query(ctx, s)
	if cursor != nil {
		defer cursor.Close(ctx)
	}

	if err != nil {
		slog.ErrorContext(ctx, "error querying Badge entity", "err", err)
		return nil, err
	}

	Badges := make([]replay_entity.Badge, 0)

	for cursor.Next(ctx) {
		var p replay_entity.Badge
		err := cursor.Decode(&p)

		if err != nil {
			slog.ErrorContext(ctx, "error decoding Badge entity", "err", err)
			return nil, err
		}

		Badges = append(Badges, p)
	}

	return Badges, nil
}

func (r *BadgeRepository) CreateMany(createCtx context.Context, events []replay_entity.Badge) error {
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
