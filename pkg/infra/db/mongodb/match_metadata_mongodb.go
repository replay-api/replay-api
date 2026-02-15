package db

import (
	"context"
	"log/slog"

	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	shared "github.com/resource-ownership/go-common/pkg/common"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MatchMetadataRepository struct {
	mongodb.MongoDBRepository[replay_entity.Match]
}

func NewMatchMetadataRepository(client *mongo.Client, dbName string, entityType replay_entity.Match, collectionName string) *MatchMetadataRepository {
	repo := mongodb.NewMongoDBRepository[replay_entity.Match](client, dbName, entityType, collectionName, "Match")

	repo.InitQueryableFields(map[string]bool{
		"ID":                             true,
		"ReplayFileID":                   true,
		"GameID":                         true,
		"MapName":                        true, // Added for search
		"Duration":                       true, // Match duration in seconds
		"Mode":                           true, // Game mode (competitive, casual)
		"GameMode":                       true, // Added for search (legacy)
		"Status":                         true, // Added for search
		"ServerName":                     true, // Server name from replay
		"Visibility":                     true,
		"VisibilityLevel":                true,
		"VisibilityType":                 true,
		"ResourceOwner":                  true,
		"CreatedAt":                      true,
		"UpdatedAt":                      true,
		"EventCount":                     true, // Number of events in match
		"RegionID":                       true, // Region identifier
		"Teams":                          true, // Team data
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
		"MapName":                        "map_name",
		"Duration":                       "duration",
		"Mode":                           "mode",
		"GameMode":                       "game_mode",
		"Status":                         "status",
		"ServerName":                     "server_name",
		"Visibility":                     "visibility",
		"VisibilityLevel":                "visibility_level",
		"VisibilityType":                 "visibility_type",
		"ResourceOwner":                  "resource_owner",
		"CreatedAt":                      "created_at",
		"UpdatedAt":                      "updated_at",
		"EventCount":                     "event_count",
		"RegionID":                       "region_id",
		"Teams":                          "teams",
		"Scoreboard":                     "scoreboard",
		"Events":                         "game_events",
		"ShareTokens":                    "share_tokens",
		"Scoreboard.MVP":                 "scoreboard.match_mvp",
		"Scoreboard.Teams":               "scoreboard.team_mvp",
		"Scoreboard.Teams.MVP":           "scoreboard.team_mvp",
		"Scoreboard.Teams.Players":       "scoreboard.team_scoreboards.players",
		"Scoreboard.Teams.Players.Stats": "scoreboard.team_scoreboards.player_stats",
		"Scoreboard.Teams.Rounds":        "scoreboard.team_scoreboards.rounds",
		"Scoreboard.Teams.Rounds.Stats":  "scoreboard.team_scoreboards.round_stats",
	})

	return &MatchMetadataRepository{
		MongoDBRepository: *repo,
	}
}

func (r *MatchMetadataRepository) Search(ctx context.Context, s shared.Search) ([]replay_entity.Match, error) {
	return r.MongoDBRepository.Search(ctx, s)
}

func (r *MatchMetadataRepository) CreateMany(createCtx context.Context, events []replay_entity.Match) error {
	pointers := make([]*replay_entity.Match, len(events))
	for i := range events {
		pointers[i] = &events[i]
	}
	return r.MongoDBRepository.CreateMany(createCtx, pointers)
}

func (r *MatchMetadataRepository) Create(createCtx context.Context, event replay_entity.Match) error {
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
