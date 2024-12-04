package db

import (
	"context"
	"log/slog"
	"reflect"
	"time"

	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	replay_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
	"go.mongodb.org/mongo-driver/mongo"
)

type MatchMetadataRepository struct {
	MongoDBRepository[replay_entity.Match]
}

func NewMatchMetadataRepository(client *mongo.Client, dbName string, entityType replay_entity.Match, collectionName string) *MatchMetadataRepository {
	repo := MongoDBRepository[replay_entity.Match]{
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
		"ID":                             true,
		"ReplayFileID":                   true,
		"GameID":                         true,
		"Visibility":                     true,
		"ResourceOwner":                  true,
		"CreatedAt":                      true,
		"UpdatedAt":                      true,
		"Scoreboard":                     true,
		"Events":                         true,
		"ShareTokens.*":                  true,
		"Scoreboard.MVP":                 true,
		"Scoreboard.Teams":               true,
		"Scoreboard.Teams.MVP":           true,
		"Scoreboard.Teams.Players":       true,
		"Scoreboard.Teams.Players.Stats": true,
		"Scoreboard.Teams.Rounds":        true,
		"Scoreboard.Teams.Rounds.Stats":  true,
	}, map[string]string{
		"ID":                             "_id",
		"ReplayFileID":                   "replay_file_id",
		"GameID":                         "game_id",
		"Visibility":                     "visibility",
		"ResourceOwner":                  "resource_owner",
		"CreatedAt":                      "created_at",
		"UpdatedAt":                      "updated_at",
		"Scoreboard":                     "scoreboard",
		"Events":                         "game_events",
		"ShareTokens":                    "share_tokens",
		"Scoreboard.MVP":                 "scoreboard.match_mvp",
		"Scoreboard.Teams":               "scoreboard.team_scoreboards",
		"Scoreboard.Teams.MVP":           "scoreboard.team_mvp",
		"Scoreboard.Teams.Players":       "scoreboard.team_scoreboards.players",
		"Scoreboard.Teams.Players.Stats": "scoreboard.team_scoreboards.player_stats",
		"Scoreboard.Teams.Rounds":        "scoreboard.team_scoreboards.rounds",
		"Scoreboard.Teams.Rounds.Stats":  "scoreboard.team_scoreboards.round_stats",
	})

	return &MatchMetadataRepository{
		repo,
	}
}

func (r *MatchMetadataRepository) Search(ctx context.Context, s common.Search) ([]replay_entity.Match, error) {
	cursor, err := r.Query(ctx, s)
	if cursor != nil {
		defer cursor.Close(ctx)
	}

	if err != nil {
		slog.ErrorContext(ctx, "error querying match entity", "err", err)
		return nil, err
	}

	players := make([]replay_entity.Match, 0)

	for cursor.Next(ctx) {
		var p replay_entity.Match
		err := cursor.Decode(&p)

		if err != nil {
			slog.ErrorContext(ctx, "error decoding match entity", "err", err)
			return nil, err
		}

		players = append(players, p)
	}

	return players, nil
}

func (r *MatchMetadataRepository) CreateMany(createCtx context.Context, events []interface{}) error {
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

func (r *MatchMetadataRepository) Create(createCtx context.Context, events ...replay_entity.Match) error {
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
