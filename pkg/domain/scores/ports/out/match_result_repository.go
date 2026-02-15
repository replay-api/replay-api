package scores_out

import (
	"context"

	"github.com/google/uuid"
	scores_entities "github.com/replay-api/replay-api/pkg/domain/scores/entities"
	scores_vo "github.com/replay-api/replay-api/pkg/domain/scores/value-objects"
)

// MatchResultRepository defines persistence operations for match results
type MatchResultRepository interface {
	Save(ctx context.Context, result *scores_entities.MatchResult) error
	FindByID(ctx context.Context, id uuid.UUID) (*scores_entities.MatchResult, error)
	FindByMatchID(ctx context.Context, matchID uuid.UUID) (*scores_entities.MatchResult, error)
	FindByTournamentID(ctx context.Context, tournamentID uuid.UUID) ([]*scores_entities.MatchResult, error)
	FindByMatchmakingSessionID(ctx context.Context, sessionID uuid.UUID) (*scores_entities.MatchResult, error)
	FindByStatus(ctx context.Context, status scores_vo.ResultStatus, limit int, offset int) ([]*scores_entities.MatchResult, int64, error)
	FindByPlayerID(ctx context.Context, playerID uuid.UUID, limit int, offset int) ([]*scores_entities.MatchResult, int64, error)
	FindByTeamID(ctx context.Context, teamID uuid.UUID, limit int, offset int) ([]*scores_entities.MatchResult, int64, error)
	Update(ctx context.Context, result *scores_entities.MatchResult) error
	Count(ctx context.Context, filter MatchResultFilter) (int64, error)
	Search(ctx context.Context, filter MatchResultFilter, limit int, offset int) ([]*scores_entities.MatchResult, int64, error)
}

// MatchResultFilter defines the filter criteria for searching match results
type MatchResultFilter struct {
	GameID               *string     `json:"game_id,omitempty"`
	TournamentID         *uuid.UUID  `json:"tournament_id,omitempty"`
	MatchmakingSessionID *uuid.UUID  `json:"matchmaking_session_id,omitempty"`
	Status               *string     `json:"status,omitempty"`
	PlayerID             *uuid.UUID  `json:"player_id,omitempty"`
	TeamID               *uuid.UUID  `json:"team_id,omitempty"`
	Source               *string     `json:"source,omitempty"`
}
