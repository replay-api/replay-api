package scores_in

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	scores_entities "github.com/replay-api/replay-api/pkg/domain/scores/entities"
	scores_vo "github.com/replay-api/replay-api/pkg/domain/scores/value-objects"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

// --- Query Handlers ---

// MatchResultQueryHandler handles all read operations for match results
type MatchResultQueryHandler interface {
	GetMatchResult(ctx context.Context, query GetMatchResultQuery) (*MatchResultDTO, error)
	GetMatchResultByMatchID(ctx context.Context, query GetMatchResultByMatchIDQuery) (*MatchResultDTO, error)
	ListMatchResults(ctx context.Context, query ListMatchResultsQuery) (*MatchResultListDTO, error)
	GetMatchResultsByTournament(ctx context.Context, query GetTournamentResultsQuery) (*MatchResultListDTO, error)
}

// --- Query DTOs ---

type GetMatchResultQuery struct {
	MatchResultID uuid.UUID `json:"match_result_id"`
}

func (q GetMatchResultQuery) Validate() error {
	if q.MatchResultID == uuid.Nil {
		return fmt.Errorf("match_result_id is required")
	}
	return nil
}

type GetMatchResultByMatchIDQuery struct {
	MatchID uuid.UUID `json:"match_id"`
}

func (q GetMatchResultByMatchIDQuery) Validate() error {
	if q.MatchID == uuid.Nil {
		return fmt.Errorf("match_id is required")
	}
	return nil
}

type ListMatchResultsQuery struct {
	GameID               *replay_common.GameIDKey   `json:"game_id,omitempty"`
	TournamentID         *uuid.UUID                 `json:"tournament_id,omitempty"`
	MatchmakingSessionID *uuid.UUID                 `json:"matchmaking_session_id,omitempty"`
	Status               *scores_vo.ResultStatus    `json:"status,omitempty"`
	PlayerID             *uuid.UUID                 `json:"player_id,omitempty"`
	TeamID               *uuid.UUID                 `json:"team_id,omitempty"`
	FromDate             *time.Time                 `json:"from_date,omitempty"`
	ToDate               *time.Time                 `json:"to_date,omitempty"`
	Page                 int                        `json:"page"`
	PageSize             int                        `json:"page_size"`
}

func (q ListMatchResultsQuery) Validate() error {
	if q.Page < 0 {
		return fmt.Errorf("page must be >= 0")
	}
	if q.PageSize <= 0 || q.PageSize > 100 {
		return fmt.Errorf("page_size must be between 1 and 100")
	}
	return nil
}

type GetTournamentResultsQuery struct {
	TournamentID uuid.UUID `json:"tournament_id"`
}

func (q GetTournamentResultsQuery) Validate() error {
	if q.TournamentID == uuid.Nil {
		return fmt.Errorf("tournament_id is required")
	}
	return nil
}

// --- Response DTOs ---

// MatchResultDTO is the external representation of a match result
type MatchResultDTO struct {
	ID                   uuid.UUID                     `json:"id"`
	MatchID              uuid.UUID                     `json:"match_id"`
	TournamentID         *uuid.UUID                    `json:"tournament_id,omitempty"`
	MatchmakingSessionID *uuid.UUID                    `json:"matchmaking_session_id,omitempty"`
	GameID               replay_common.GameIDKey        `json:"game_id"`
	MapName              string                         `json:"map_name"`
	Mode                 string                         `json:"mode"`
	Source               scores_vo.ScoreSource          `json:"source"`
	SourceReplayID       *uuid.UUID                     `json:"source_replay_id,omitempty"`
	TeamResults          []TeamResultDTO                `json:"team_results"`
	PlayerResults        []PlayerResultDTO              `json:"player_results"`
	WinnerTeamID         *uuid.UUID                     `json:"winner_team_id,omitempty"`
	IsDraw               bool                           `json:"is_draw"`
	RoundsPlayed         int                            `json:"rounds_played"`
	Status               scores_vo.ResultStatus         `json:"status"`
	VerificationMethod   *scores_vo.VerificationMethod  `json:"verification_method,omitempty"`
	VerifiedAt           *time.Time                     `json:"verified_at,omitempty"`
	DisputeReason        string                         `json:"dispute_reason,omitempty"`
	DisputeCount         int                            `json:"dispute_count"`
	DisputedAt           *time.Time                     `json:"disputed_at,omitempty"`
	ConciliationNotes    string                         `json:"conciliation_notes,omitempty"`
	ConciliatedAt        *time.Time                     `json:"conciliated_at,omitempty"`
	WasAdjusted          bool                           `json:"was_adjusted"`
	FinalizedAt          *time.Time                     `json:"finalized_at,omitempty"`
	PrizeDistributionID  *uuid.UUID                     `json:"prize_distribution_id,omitempty"`
	PlayedAt             time.Time                      `json:"played_at"`
	Duration             time.Duration                  `json:"duration"`
	CreatedAt            time.Time                      `json:"created_at"`
	UpdatedAt            time.Time                      `json:"updated_at"`
}

type TeamResultDTO struct {
	TeamID   uuid.UUID   `json:"team_id"`
	TeamName string      `json:"team_name"`
	Score    int         `json:"score"`
	Position int         `json:"position"`
	Players  []uuid.UUID `json:"players"`
}

type PlayerResultDTO struct {
	PlayerID uuid.UUID              `json:"player_id"`
	TeamID   uuid.UUID              `json:"team_id"`
	Score    int                    `json:"score"`
	Kills    int                    `json:"kills"`
	Deaths   int                    `json:"deaths"`
	Assists  int                    `json:"assists"`
	Rating   float64                `json:"rating"`
	IsMVP    bool                   `json:"is_mvp"`
	Stats    map[string]interface{} `json:"stats,omitempty"`
}

type MatchResultListDTO struct {
	Results    []MatchResultDTO `json:"results"`
	TotalCount int64            `json:"total_count"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
}

// --- Conversion Helpers ---

// MatchResultToDTO converts a domain entity to its DTO representation
func MatchResultToDTO(m *scores_entities.MatchResult) MatchResultDTO {
	dto := MatchResultDTO{
		ID:                   m.ID,
		MatchID:              m.MatchID,
		TournamentID:         m.TournamentID,
		MatchmakingSessionID: m.MatchmakingSessionID,
		GameID:               m.GameID,
		MapName:              m.MapName,
		Mode:                 m.Mode,
		Source:               m.Source,
		SourceReplayID:       m.SourceReplayID,
		WinnerTeamID:         m.WinnerTeamID,
		IsDraw:               m.IsDraw,
		RoundsPlayed:         m.RoundsPlayed,
		Status:               m.Status,
		VerificationMethod:   m.VerificationMethod,
		VerifiedAt:           m.VerifiedAt,
		DisputeReason:        m.DisputeReason,
		DisputeCount:         m.DisputeCount,
		DisputedAt:           m.DisputedAt,
		ConciliationNotes:    m.ConciliationNotes,
		ConciliatedAt:        m.ConciliatedAt,
		WasAdjusted:          m.WasAdjusted(),
		FinalizedAt:          m.FinalizedAt,
		PrizeDistributionID:  m.PrizeDistributionID,
		PlayedAt:             m.PlayedAt,
		Duration:             m.Duration,
		CreatedAt:            m.CreatedAt,
		UpdatedAt:            m.UpdatedAt,
	}

	dto.TeamResults = make([]TeamResultDTO, len(m.TeamResults))
	for i, tr := range m.TeamResults {
		dto.TeamResults[i] = TeamResultDTO{
			TeamID:   tr.TeamID,
			TeamName: tr.TeamName,
			Score:    tr.Score,
			Position: tr.Position,
			Players:  tr.Players,
		}
	}

	dto.PlayerResults = make([]PlayerResultDTO, len(m.PlayerResults))
	for i, pr := range m.PlayerResults {
		dto.PlayerResults[i] = PlayerResultDTO{
			PlayerID: pr.PlayerID,
			TeamID:   pr.TeamID,
			Score:    pr.Score,
			Kills:    pr.Kills,
			Deaths:   pr.Deaths,
			Assists:  pr.Assists,
			Rating:   pr.Rating,
			IsMVP:    pr.IsMVP,
			Stats:    pr.Stats,
		}
	}

	return dto
}
