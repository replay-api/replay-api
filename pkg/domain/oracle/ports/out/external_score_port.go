package oracle_out

import (
	"context"
	"time"

	"github.com/google/uuid"
	oracle_entities "github.com/replay-api/replay-api/pkg/domain/oracle/entities"
	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

// ExternalScorePort defines the contract for fetching scores from external providers
type ExternalScorePort interface {
	// FetchMatchScore fetches the score for a specific match from this provider
	FetchMatchScore(ctx context.Context, externalMatchID string, gameID replay_common.GameIDKey) (*oracle_entities.ScoreSubmission, error)

	// ListRecentMatches fetches recently completed matches from this provider
	ListRecentMatches(ctx context.Context, gameID replay_common.GameIDKey, since time.Time, limit int) ([]ExternalMatch, error)

	// SupportsGame returns true if this provider supports the given game
	SupportsGame(gameID replay_common.GameIDKey) bool

	// ProviderID returns the source type identifier for this provider
	ProviderID() oracle_vo.OracleSourceType

	// ConfidenceWeight returns the default confidence weight for this provider
	ConfidenceWeight() float64
}

// ExternalMatch represents a match discovered from an external provider.
type ExternalMatch struct {
	ExternalMatchID string                     `json:"external_match_id"`
	GameID          replay_common.GameIDKey     `json:"game_id"`
	Provider        oracle_vo.OracleSourceType  `json:"provider"`
	TeamAName       string                     `json:"team_a_name"`
	TeamBName       string                     `json:"team_b_name"`
	TeamAID         uuid.UUID                  `json:"team_a_id"`
	TeamBID         uuid.UUID                  `json:"team_b_id"`
	TeamAScore      int                        `json:"team_a_score"`
	TeamBScore      int                        `json:"team_b_score"`
	WinnerTeamID    *uuid.UUID                 `json:"winner_team_id,omitempty"`
	IsDraw          bool                       `json:"is_draw"`
	Status          string                     `json:"status"` // "finished", "running", "not_started"
	TournamentName  string                     `json:"tournament_name,omitempty"`
	TournamentID    string                     `json:"tournament_id,omitempty"`
	SeriesType      string                     `json:"series_type,omitempty"` // "bo1", "bo3", "bo5"
	StreamURL       string                     `json:"stream_url,omitempty"`
	VODURLs         []string                   `json:"vod_urls,omitempty"`
	PlayedAt        time.Time                  `json:"played_at"`
	MapName         string                     `json:"map_name,omitempty"`
	NumberOfGames   int                        `json:"number_of_games"`
}
