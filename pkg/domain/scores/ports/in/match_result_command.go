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

// --- Command Handlers ---

// MatchResultCommandHandler handles all write operations for match results
type MatchResultCommandHandler interface {
	SubmitMatchResult(ctx context.Context, cmd SubmitMatchResultCommand) (*scores_entities.MatchResult, error)
	SubmitMatchResultFromReplay(ctx context.Context, cmd SubmitReplayResultCommand) (*scores_entities.MatchResult, error)
	VerifyMatchResult(ctx context.Context, cmd VerifyMatchResultCommand) error
	DisputeMatchResult(ctx context.Context, cmd DisputeMatchResultCommand) error
	ConciliateMatchResult(ctx context.Context, cmd ConciliateMatchResultCommand) error
	FinalizeMatchResult(ctx context.Context, cmd FinalizeMatchResultCommand) error
	CancelMatchResult(ctx context.Context, cmd CancelMatchResultCommand) error
}

// --- Command DTOs ---

// SubmitMatchResultCommand is used by tournament admins or external sources to submit match results
type SubmitMatchResultCommand struct {
	MatchID              uuid.UUID                         `json:"match_id"`
	TournamentID         *uuid.UUID                        `json:"tournament_id,omitempty"`
	MatchmakingSessionID *uuid.UUID                        `json:"matchmaking_session_id,omitempty"`
	GameID               replay_common.GameIDKey            `json:"game_id"`
	MapName              string                             `json:"map_name"`
	Mode                 string                             `json:"mode"`
	Source               scores_vo.ScoreSource              `json:"source"`
	TeamResults          []scores_entities.TeamResult        `json:"team_results"`
	PlayerResults        []scores_entities.PlayerResult      `json:"player_results"`
	PlayedAt             time.Time                          `json:"played_at"`
	Duration             time.Duration                      `json:"duration"`
	RoundsPlayed         int                                `json:"rounds_played"`
}

func (c SubmitMatchResultCommand) Validate() error {
	if c.MatchID == uuid.Nil {
		return fmt.Errorf("match_id is required")
	}
	if c.GameID == "" {
		return fmt.Errorf("game_id is required")
	}
	if err := c.Source.Validate(); err != nil {
		return err
	}
	if len(c.TeamResults) < 2 {
		return fmt.Errorf("at least 2 team results are required")
	}
	if c.PlayedAt.IsZero() {
		return fmt.Errorf("played_at is required")
	}
	return nil
}

// SubmitReplayResultCommand is used by the replay processing pipeline to automatically submit results
type SubmitReplayResultCommand struct {
	MatchID              uuid.UUID                         `json:"match_id"`
	ReplayID             uuid.UUID                         `json:"replay_id"`
	MatchmakingSessionID *uuid.UUID                        `json:"matchmaking_session_id,omitempty"`
	TournamentID         *uuid.UUID                        `json:"tournament_id,omitempty"`
	GameID               replay_common.GameIDKey            `json:"game_id"`
	MapName              string                             `json:"map_name"`
	Mode                 string                             `json:"mode"`
	TeamResults          []scores_entities.TeamResult        `json:"team_results"`
	PlayerResults        []scores_entities.PlayerResult      `json:"player_results"`
	PlayedAt             time.Time                          `json:"played_at"`
	Duration             time.Duration                      `json:"duration"`
	RoundsPlayed         int                                `json:"rounds_played"`
}

func (c SubmitReplayResultCommand) Validate() error {
	if c.MatchID == uuid.Nil {
		return fmt.Errorf("match_id is required")
	}
	if c.ReplayID == uuid.Nil {
		return fmt.Errorf("replay_id is required")
	}
	if c.GameID == "" {
		return fmt.Errorf("game_id is required")
	}
	if len(c.TeamResults) < 2 {
		return fmt.Errorf("at least 2 team results are required")
	}
	return nil
}

// VerifyMatchResultCommand triggers verification of a submitted match result
type VerifyMatchResultCommand struct {
	MatchResultID      uuid.UUID                    `json:"match_result_id"`
	VerificationMethod scores_vo.VerificationMethod `json:"verification_method"`
}

func (c VerifyMatchResultCommand) Validate() error {
	if c.MatchResultID == uuid.Nil {
		return fmt.Errorf("match_result_id is required")
	}
	return c.VerificationMethod.Validate()
}

// DisputeMatchResultCommand registers a dispute against a verified match result
type DisputeMatchResultCommand struct {
	MatchResultID uuid.UUID `json:"match_result_id"`
	Reason        string    `json:"reason"`
}

func (c DisputeMatchResultCommand) Validate() error {
	if c.MatchResultID == uuid.Nil {
		return fmt.Errorf("match_result_id is required")
	}
	if c.Reason == "" {
		return fmt.Errorf("dispute reason is required")
	}
	return nil
}

// ConciliateMatchResultCommand resolves a disputed match result
type ConciliateMatchResultCommand struct {
	MatchResultID       uuid.UUID                     `json:"match_result_id"`
	Notes               string                         `json:"notes"`
	AdjustedTeamResults []scores_entities.TeamResult   `json:"adjusted_team_results,omitempty"`
}

func (c ConciliateMatchResultCommand) Validate() error {
	if c.MatchResultID == uuid.Nil {
		return fmt.Errorf("match_result_id is required")
	}
	if c.Notes == "" {
		return fmt.Errorf("conciliation notes are required")
	}
	return nil
}

// FinalizeMatchResultCommand finalizes a verified/conciliated match result and triggers prize distribution
type FinalizeMatchResultCommand struct {
	MatchResultID uuid.UUID `json:"match_result_id"`
}

func (c FinalizeMatchResultCommand) Validate() error {
	if c.MatchResultID == uuid.Nil {
		return fmt.Errorf("match_result_id is required")
	}
	return nil
}

// CancelMatchResultCommand cancels/voids a match result
type CancelMatchResultCommand struct {
	MatchResultID uuid.UUID `json:"match_result_id"`
	Reason        string    `json:"reason"`
}

func (c CancelMatchResultCommand) Validate() error {
	if c.MatchResultID == uuid.Nil {
		return fmt.Errorf("match_result_id is required")
	}
	if c.Reason == "" {
		return fmt.Errorf("cancellation reason is required")
	}
	return nil
}
