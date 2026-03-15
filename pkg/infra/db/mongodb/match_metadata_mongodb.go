package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
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
	repo := mongodb.NewMongoDBRepository(client, dbName, entityType, collectionName, "Match")

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
		"Source":                         true,
		"ExternalMatchID":               true,
		"Slug":                          true,
		"PlayedAt":                      true,
		"LinkedMatchIDs":                true,
		"SourceConfirmations":           true,
		"NeedsReview":                   true,
		"ConflictDetails":               true,
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
		"Source":                         "source",
		"ExternalMatchID":               "external_match_id",
		"Slug":                          "slug",
		"PlayedAt":                      "played_at",
		"LinkedMatchIDs":                "linked_match_ids",
		"SourceConfirmations":           "source_confirmations",
		"NeedsReview":                   "needs_review",
		"ConflictDetails":               "conflict_details",
	})

	return &MatchMetadataRepository{
		MongoDBRepository: *repo,
	}
}

// EnsureIndexes creates indexes for slug, external_match_id, source, and needs_review fields.
func (r *MatchMetadataRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "slug", Value: 1}},
			Options: options.Index().SetSparse(true).SetUnique(true).SetName("idx_slug_unique"),
		},
		{
			Keys:    bson.D{{Key: "external_match_id", Value: 1}},
			Options: options.Index().SetSparse(true).SetUnique(true).SetName("idx_external_match_id"),
		},
		{
			Keys:    bson.D{{Key: "source", Value: 1}},
			Options: options.Index().SetName("idx_source"),
		},
		{
			Keys:    bson.D{{Key: "needs_review", Value: 1}},
			Options: options.Index().SetSparse(true).SetName("idx_needs_review"),
		},
	}

	_, err := r.MongoDBRepository.Collection().Indexes().CreateMany(ctx, indexes)
	if err != nil {
		slog.WarnContext(ctx, "failed to create match indexes (may already exist)",
			slog.String("error", err.Error()),
		)
	}
	return nil
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

func (r *MatchMetadataRepository) Update(ctx context.Context, match replay_entity.Match) error {
	filter := bson.M{"_id": match.ID}
	update := bson.M{"$set": match}
	result, err := r.MongoDBRepository.Collection().UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update match", slog.String("error", err.Error()))
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("match not found: %s", match.ID)
	}
	return nil
}

// FindOneAndUpsertBySlug atomically finds a match by slug or creates it.
// Uses MongoDB FindOneAndUpdate with upsert + $setOnInsert for race-condition-free atomic operation.
func (r *MatchMetadataRepository) FindOneAndUpsertBySlug(ctx context.Context, slug string, match replay_entity.Match) (*replay_entity.Match, bool, error) {
	if slug == "" {
		// No slug — cannot do atomic upsert, fall back to regular create
		if err := r.Create(ctx, match); err != nil {
			return nil, false, fmt.Errorf("failed to create match without slug: %w", err)
		}
		return &match, true, nil
	}

	filter := bson.M{"slug": slug}
	update := bson.M{
		"$setOnInsert": match,
	}
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var result replay_entity.Match
	err := r.MongoDBRepository.Collection().FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return nil, false, fmt.Errorf("FindOneAndUpsertBySlug failed: %w", err)
	}

	// If the returned document's ID matches our input, we created it
	created := result.ID == match.ID
	return &result, created, nil
}

// AppendSourceConfirmation atomically appends a source confirmation to a match.
// Uses $push for the confirmation array and $set for conflict fields.
// This avoids full-document replacement and is safe for concurrent operations.
func (r *MatchMetadataRepository) AppendSourceConfirmation(
	ctx context.Context,
	matchID uuid.UUID,
	confirmation replay_entity.SourceConfirmation,
	needsReview bool,
	conflictDetails string,
) error {
	filter := bson.M{"_id": matchID}
	update := bson.M{
		"$push": bson.M{
			"source_confirmations": confirmation,
		},
		"$set": bson.M{
			"needs_review":     needsReview,
			"conflict_details": conflictDetails,
			"updated_at":       time.Now().UTC(),
		},
	}

	result, err := r.MongoDBRepository.Collection().UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to append source confirmation: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("match not found for source confirmation: %s", matchID)
	}

	return nil
}

func (r *MatchMetadataRepository) FindBySlug(ctx context.Context, slug string) (*replay_entity.Match, error) {
	var match replay_entity.Match
	filter := bson.M{"slug": slug}
	err := r.MongoDBRepository.Collection().FindOne(ctx, filter).Decode(&match)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &match, nil
}

func (r *MatchMetadataRepository) FindByExternalMatchID(ctx context.Context, externalMatchID string) (*replay_entity.Match, error) {
	var match replay_entity.Match
	filter := bson.M{"external_match_id": externalMatchID}
	err := r.MongoDBRepository.Collection().FindOne(ctx, filter).Decode(&match)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &match, nil
}
