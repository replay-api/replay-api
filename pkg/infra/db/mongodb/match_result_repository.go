package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	scores_entities "github.com/replay-api/replay-api/pkg/domain/scores/entities"
	scores_out "github.com/replay-api/replay-api/pkg/domain/scores/ports/out"
	scores_vo "github.com/replay-api/replay-api/pkg/domain/scores/value-objects"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoMatchResultRepository implements the MatchResultRepository port using MongoDB
type MongoMatchResultRepository struct {
	mongodb.MongoDBRepository[scores_entities.MatchResult]
}

// Compile-time interface satisfaction check
var _ scores_out.MatchResultRepository = (*MongoMatchResultRepository)(nil)

// NewMongoMatchResultRepository creates a new MongoDB-backed match result repository
func NewMongoMatchResultRepository(mongoClient *mongo.Client, dbName string) scores_out.MatchResultRepository {
	entityType := scores_entities.MatchResult{}
	repo := mongodb.NewMongoDBRepository[scores_entities.MatchResult](mongoClient, dbName, entityType, "match_results", "MatchResult")

	repo.InitQueryableFields(map[string]bool{
		"ID":                   true,
		"BaseEntityID":         true,
		"VisibilityLevel":      true,
		"VisibilityType":       true,
		"MatchID":              true,
		"TournamentID":         true,
		"MatchmakingSessionID": true,
		"GameID":               true,
		"MapName":              true,
		"Source":               true,
		"Status":               true,
		"WinnerTeamID":         true,
		"IsDraw":               true,
		"PlayedAt":             true,
		"FinalizedAt":          true,
		"CreatedAt":            true,
		"UpdatedAt":            true,
		"UserID":               true,
		"TenantID":             true,
		"GroupID":              true,
		"ClientID":             true,
	}, map[string]string{
		"ID":                   "baseentity._id",
		"BaseEntityID":         "baseentity._id",
		"VisibilityLevel":      "baseentity.visibility_level",
		"VisibilityType":       "baseentity.visibility_type",
		"MatchID":              "match_id",
		"TournamentID":         "tournament_id",
		"MatchmakingSessionID": "matchmaking_session_id",
		"GameID":               "game_id",
		"MapName":              "map_name",
		"Source":               "source",
		"Status":               "status",
		"WinnerTeamID":         "winner_team_id",
		"IsDraw":               "is_draw",
		"PlayedAt":             "played_at",
		"FinalizedAt":          "finalized_at",
		"CreatedAt":            "baseentity.created_at",
		"UpdatedAt":            "baseentity.updated_at",
		"UserID":               "baseentity.resource_owner.user_id",
		"TenantID":             "baseentity.resource_owner.tenant_id",
		"GroupID":              "baseentity.resource_owner.group_id",
		"ClientID":             "baseentity.resource_owner.client_id",
	})

	return &MongoMatchResultRepository{
		MongoDBRepository: *repo,
	}
}

func (r *MongoMatchResultRepository) Save(ctx context.Context, result *scores_entities.MatchResult) error {
	if result.GetID() == uuid.Nil {
		return fmt.Errorf("match result ID cannot be nil")
	}

	result.UpdatedAt = time.Now().UTC()

	_, err := r.MongoDBRepository.Create(ctx, result)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save match result",
			slog.String("match_result_id", result.ID.String()),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to save match result: %w", err)
	}

	slog.InfoContext(ctx, "match result saved",
		slog.String("match_result_id", result.ID.String()),
		slog.String("match_id", result.MatchID.String()),
	)
	return nil
}

func (r *MongoMatchResultRepository) FindByID(ctx context.Context, id uuid.UUID) (*scores_entities.MatchResult, error) {
	return r.MongoDBRepository.GetByID(ctx, id)
}

func (r *MongoMatchResultRepository) FindByMatchID(ctx context.Context, matchID uuid.UUID) (*scores_entities.MatchResult, error) {
	var result scores_entities.MatchResult
	filter := bson.M{"match_id": matchID}

	err := r.MongoDBRepository.FindOneWithRLS(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("match result not found for match: %s", matchID)
		}
		return nil, fmt.Errorf("failed to find match result: %w", err)
	}

	return &result, nil
}

func (r *MongoMatchResultRepository) FindByTournamentID(ctx context.Context, tournamentID uuid.UUID) ([]*scores_entities.MatchResult, error) {
	filter := bson.M{"tournament_id": tournamentID}
	return r.findMany(ctx, filter, 0, 0)
}

func (r *MongoMatchResultRepository) FindByMatchmakingSessionID(ctx context.Context, sessionID uuid.UUID) (*scores_entities.MatchResult, error) {
	var result scores_entities.MatchResult
	filter := bson.M{"matchmaking_session_id": sessionID}

	err := r.MongoDBRepository.FindOneWithRLS(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("match result not found for session: %s", sessionID)
		}
		return nil, fmt.Errorf("failed to find match result: %w", err)
	}

	return &result, nil
}

func (r *MongoMatchResultRepository) FindByStatus(ctx context.Context, status scores_vo.ResultStatus, limit int, offset int) ([]*scores_entities.MatchResult, int64, error) {
	filter := bson.M{"status": string(status)}
	results, err := r.findMany(ctx, filter, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	count, err := r.countDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return results, count, nil
}

func (r *MongoMatchResultRepository) FindByPlayerID(ctx context.Context, playerID uuid.UUID, limit int, offset int) ([]*scores_entities.MatchResult, int64, error) {
	filter := bson.M{"player_results.player_id": playerID}
	results, err := r.findMany(ctx, filter, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	count, err := r.countDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return results, count, nil
}

func (r *MongoMatchResultRepository) FindByTeamID(ctx context.Context, teamID uuid.UUID, limit int, offset int) ([]*scores_entities.MatchResult, int64, error) {
	filter := bson.M{"team_results.team_id": teamID}
	results, err := r.findMany(ctx, filter, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	count, err := r.countDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return results, count, nil
}

func (r *MongoMatchResultRepository) Update(ctx context.Context, result *scores_entities.MatchResult) error {
	if result.GetID() == uuid.Nil {
		return fmt.Errorf("match result ID cannot be nil")
	}

	result.UpdatedAt = time.Now().UTC()

	_, err := r.MongoDBRepository.Update(ctx, result)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update match result",
			slog.String("match_result_id", result.ID.String()),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to update match result: %w", err)
	}

	return nil
}

func (r *MongoMatchResultRepository) Count(ctx context.Context, filter scores_out.MatchResultFilter) (int64, error) {
	bsonFilter := r.buildFilter(filter)
	return r.countDocuments(ctx, bsonFilter)
}

func (r *MongoMatchResultRepository) Search(ctx context.Context, filter scores_out.MatchResultFilter, limit int, offset int) ([]*scores_entities.MatchResult, int64, error) {
	bsonFilter := r.buildFilter(filter)
	results, err := r.findMany(ctx, bsonFilter, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	count, err := r.countDocuments(ctx, bsonFilter)
	if err != nil {
		return nil, 0, err
	}

	return results, count, nil
}

// --- Internal helpers ---

func (r *MongoMatchResultRepository) buildFilter(filter scores_out.MatchResultFilter) bson.M {
	bsonFilter := bson.M{}

	if filter.GameID != nil {
		bsonFilter["game_id"] = *filter.GameID
	}
	if filter.TournamentID != nil {
		bsonFilter["tournament_id"] = *filter.TournamentID
	}
	if filter.MatchmakingSessionID != nil {
		bsonFilter["matchmaking_session_id"] = *filter.MatchmakingSessionID
	}
	if filter.Status != nil {
		bsonFilter["status"] = *filter.Status
	}
	if filter.PlayerID != nil {
		bsonFilter["player_results.player_id"] = *filter.PlayerID
	}
	if filter.TeamID != nil {
		bsonFilter["team_results.team_id"] = *filter.TeamID
	}
	if filter.Source != nil {
		bsonFilter["source"] = *filter.Source
	}

	return bsonFilter
}

func (r *MongoMatchResultRepository) findMany(ctx context.Context, filter bson.M, limit int, offset int) ([]*scores_entities.MatchResult, error) {
	collection := r.MongoDBRepository.Collection()

	opts := options.Find().SetSort(bson.M{"baseentity.created_at": -1})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	if offset > 0 {
		opts.SetSkip(int64(offset))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find match results: %w", err)
	}
	defer cursor.Close(ctx)

	var results []*scores_entities.MatchResult
	for cursor.Next(ctx) {
		var result scores_entities.MatchResult
		if err := cursor.Decode(&result); err != nil {
			slog.WarnContext(ctx, "failed to decode match result", slog.String("error", err.Error()))
			continue
		}
		results = append(results, &result)
	}

	return results, nil
}

func (r *MongoMatchResultRepository) countDocuments(ctx context.Context, filter bson.M) (int64, error) {
	collection := r.MongoDBRepository.Collection()
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count match results: %w", err)
	}
	return count, nil
}
