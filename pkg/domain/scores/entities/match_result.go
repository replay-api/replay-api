package scores_entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	scores_vo "github.com/replay-api/replay-api/pkg/domain/scores/value-objects"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// MatchResult is the aggregate root representing the verified result of a competitive match.
// It bridges between score submission (from replays, admins, or external sources) and
// prize distribution, providing a full audit trail with dispute/conciliation support.
type MatchResult struct {
	shared.BaseEntity

	// Context Links
	MatchID              uuid.UUID                `json:"match_id" bson:"match_id"`
	TournamentID         *uuid.UUID               `json:"tournament_id,omitempty" bson:"tournament_id,omitempty"`
	MatchmakingSessionID *uuid.UUID               `json:"matchmaking_session_id,omitempty" bson:"matchmaking_session_id,omitempty"`
	GameID               replay_common.GameIDKey   `json:"game_id" bson:"game_id"`
	MapName              string                    `json:"map_name" bson:"map_name"`
	Mode                 string                    `json:"mode" bson:"mode"`

	// Score Source
	Source         scores_vo.ScoreSource `json:"source" bson:"source"`
	SourceReplayID *uuid.UUID            `json:"source_replay_id,omitempty" bson:"source_replay_id,omitempty"`
	SubmittedBy    uuid.UUID             `json:"submitted_by" bson:"submitted_by"`

	// Results
	TeamResults  []TeamResult   `json:"team_results" bson:"team_results"`
	PlayerResults []PlayerResult `json:"player_results" bson:"player_results"`
	WinnerTeamID *uuid.UUID     `json:"winner_team_id,omitempty" bson:"winner_team_id,omitempty"`
	IsDraw       bool           `json:"is_draw" bson:"is_draw"`
	RoundsPlayed int            `json:"rounds_played" bson:"rounds_played"`

	// Status & Verification
	Status             scores_vo.ResultStatus       `json:"status" bson:"status"`
	VerificationMethod *scores_vo.VerificationMethod `json:"verification_method,omitempty" bson:"verification_method,omitempty"`
	VerifiedAt         *time.Time                    `json:"verified_at,omitempty" bson:"verified_at,omitempty"`
	VerifiedBy         *uuid.UUID                    `json:"verified_by,omitempty" bson:"verified_by,omitempty"`

	// Dispute
	DisputeReason string     `json:"dispute_reason,omitempty" bson:"dispute_reason,omitempty"`
	DisputedAt    *time.Time `json:"disputed_at,omitempty" bson:"disputed_at,omitempty"`
	DisputedBy    *uuid.UUID `json:"disputed_by,omitempty" bson:"disputed_by,omitempty"`
	DisputeCount  int        `json:"dispute_count" bson:"dispute_count"`

	// Conciliation
	ConciliatedAt       *time.Time    `json:"conciliated_at,omitempty" bson:"conciliated_at,omitempty"`
	ConciliatedBy       *uuid.UUID    `json:"conciliated_by,omitempty" bson:"conciliated_by,omitempty"`
	ConciliationNotes   string        `json:"conciliation_notes,omitempty" bson:"conciliation_notes,omitempty"`
	OriginalTeamResults []TeamResult  `json:"original_team_results,omitempty" bson:"original_team_results,omitempty"`

	// Finalization
	FinalizedAt         *time.Time `json:"finalized_at,omitempty" bson:"finalized_at,omitempty"`
	PrizeDistributionID *uuid.UUID `json:"prize_distribution_id,omitempty" bson:"prize_distribution_id,omitempty"`

	// Match Metadata
	PlayedAt time.Time     `json:"played_at" bson:"played_at"`
	Duration time.Duration `json:"duration" bson:"duration"`
}

// TeamResult represents the result of a single team in a match
type TeamResult struct {
	TeamID   uuid.UUID   `json:"team_id" bson:"team_id"`
	TeamName string      `json:"team_name" bson:"team_name"`
	Score    int         `json:"score" bson:"score"`
	Position int         `json:"position" bson:"position"` // 1 = winner, 2 = runner-up, etc.
	Players  []uuid.UUID `json:"players" bson:"players"`
}

// PlayerResult represents an individual player's performance in a match
type PlayerResult struct {
	PlayerID uuid.UUID              `json:"player_id" bson:"player_id"`
	TeamID   uuid.UUID              `json:"team_id" bson:"team_id"`
	Score    int                    `json:"score" bson:"score"`
	Kills    int                    `json:"kills" bson:"kills"`
	Deaths   int                    `json:"deaths" bson:"deaths"`
	Assists  int                    `json:"assists" bson:"assists"`
	Rating   float64                `json:"rating" bson:"rating"` // e.g., HLTV 2.0 rating
	IsMVP    bool                   `json:"is_mvp" bson:"is_mvp"`
	Stats    map[string]interface{} `json:"stats,omitempty" bson:"stats,omitempty"` // Flexible game-specific stats
}

// NewMatchResult creates a new MatchResult in the submitted state
func NewMatchResult(
	resourceOwner shared.ResourceOwner,
	matchID uuid.UUID,
	gameID replay_common.GameIDKey,
	mapName string,
	mode string,
	source scores_vo.ScoreSource,
	submittedBy uuid.UUID,
	teamResults []TeamResult,
	playerResults []PlayerResult,
	playedAt time.Time,
	duration time.Duration,
) (*MatchResult, error) {
	result := &MatchResult{
		BaseEntity:    shared.NewEntity(resourceOwner),
		MatchID:       matchID,
		GameID:        gameID,
		MapName:       mapName,
		Mode:          mode,
		Source:        source,
		SubmittedBy:   submittedBy,
		TeamResults:   teamResults,
		PlayerResults: playerResults,
		Status:        scores_vo.ResultStatusSubmitted,
		PlayedAt:      playedAt,
		Duration:      duration,
	}

	// Determine winner
	result.determineWinner()

	if err := result.Validate(); err != nil {
		return nil, fmt.Errorf("invalid match result: %w", err)
	}

	return result, nil
}

// NewMatchResultFromReplay creates a MatchResult sourced from replay file processing
func NewMatchResultFromReplay(
	resourceOwner shared.ResourceOwner,
	matchID uuid.UUID,
	replayID uuid.UUID,
	gameID replay_common.GameIDKey,
	mapName string,
	mode string,
	teamResults []TeamResult,
	playerResults []PlayerResult,
	playedAt time.Time,
	duration time.Duration,
) (*MatchResult, error) {
	result, err := NewMatchResult(
		resourceOwner, matchID, gameID, mapName, mode,
		scores_vo.ScoreSourceReplayFile, resourceOwner.UserID,
		teamResults, playerResults, playedAt, duration,
	)
	if err != nil {
		return nil, err
	}

	result.SourceReplayID = &replayID
	return result, nil
}

// NewMatchResultFromAdmin creates a MatchResult submitted by a tournament administrator
func NewMatchResultFromAdmin(
	resourceOwner shared.ResourceOwner,
	matchID uuid.UUID,
	tournamentID uuid.UUID,
	gameID replay_common.GameIDKey,
	mapName string,
	mode string,
	adminUserID uuid.UUID,
	teamResults []TeamResult,
	playerResults []PlayerResult,
	playedAt time.Time,
	duration time.Duration,
) (*MatchResult, error) {
	result, err := NewMatchResult(
		resourceOwner, matchID, gameID, mapName, mode,
		scores_vo.ScoreSourceTournamentAdmin, adminUserID,
		teamResults, playerResults, playedAt, duration,
	)
	if err != nil {
		return nil, err
	}

	result.TournamentID = &tournamentID
	return result, nil
}

// SetMatchmakingContext sets the matchmaking session context for this result
func (m *MatchResult) SetMatchmakingContext(sessionID uuid.UUID) {
	m.MatchmakingSessionID = &sessionID
}

// --- State Machine Methods ---

// Review moves the result from submitted to under_review
func (m *MatchResult) Review() error {
	if err := m.Status.ValidateTransition(scores_vo.ResultStatusUnderReview); err != nil {
		return err
	}

	m.Status = scores_vo.ResultStatusUnderReview
	m.UpdatedAt = time.Now().UTC()
	return nil
}

// Verify marks the result as verified with the given verification method
func (m *MatchResult) Verify(method scores_vo.VerificationMethod, verifiedBy *uuid.UUID) error {
	if err := m.Status.ValidateTransition(scores_vo.ResultStatusVerified); err != nil {
		return err
	}

	if err := method.Validate(); err != nil {
		return err
	}

	now := time.Now().UTC()
	m.Status = scores_vo.ResultStatusVerified
	m.VerificationMethod = &method
	m.VerifiedAt = &now
	m.VerifiedBy = verifiedBy
	m.UpdatedAt = now
	return nil
}

// AutoVerify is a convenience method for automatic verification (e.g., from replay parsing)
func (m *MatchResult) AutoVerify() error {
	return m.Verify(scores_vo.VerificationMethodAutomatic, nil)
}

// Dispute marks the result as disputed
func (m *MatchResult) Dispute(reason string, disputedBy uuid.UUID) error {
	if !m.Status.IsDisputable() {
		return fmt.Errorf("cannot dispute result in status %s", m.Status)
	}

	if reason == "" {
		return fmt.Errorf("dispute reason is required")
	}

	now := time.Now().UTC()
	m.Status = scores_vo.ResultStatusDisputed
	m.DisputeReason = reason
	m.DisputedAt = &now
	m.DisputedBy = &disputedBy
	m.DisputeCount++
	m.UpdatedAt = now
	return nil
}

// Conciliate resolves a dispute with optionally adjusted scores
func (m *MatchResult) Conciliate(
	conciliatedBy uuid.UUID,
	notes string,
	adjustedTeamResults []TeamResult,
) error {
	if err := m.Status.ValidateTransition(scores_vo.ResultStatusConciliated); err != nil {
		return err
	}

	now := time.Now().UTC()
	m.Status = scores_vo.ResultStatusConciliated
	m.ConciliatedAt = &now
	m.ConciliatedBy = &conciliatedBy
	m.ConciliationNotes = notes

	// If adjusted results are provided, preserve originals for audit and apply adjustments
	if len(adjustedTeamResults) > 0 {
		m.OriginalTeamResults = make([]TeamResult, len(m.TeamResults))
		copy(m.OriginalTeamResults, m.TeamResults)
		m.TeamResults = adjustedTeamResults
		m.determineWinner()
	}

	m.UpdatedAt = now
	return nil
}

// Finalize marks the result as finalized, making it eligible for prize distribution
func (m *MatchResult) Finalize() error {
	if err := m.Status.ValidateTransition(scores_vo.ResultStatusFinalized); err != nil {
		return err
	}

	now := time.Now().UTC()
	m.Status = scores_vo.ResultStatusFinalized
	m.FinalizedAt = &now
	m.UpdatedAt = now
	return nil
}

// Cancel voids the result entirely
func (m *MatchResult) Cancel(reason string) error {
	if err := m.Status.ValidateTransition(scores_vo.ResultStatusCancelled); err != nil {
		return err
	}

	m.Status = scores_vo.ResultStatusCancelled
	m.ConciliationNotes = reason
	m.UpdatedAt = time.Now().UTC()
	return nil
}

// SetPrizeDistribution links this result to a prize distribution
func (m *MatchResult) SetPrizeDistribution(distributionID uuid.UUID) {
	m.PrizeDistributionID = &distributionID
	m.UpdatedAt = time.Now().UTC()
}

// --- Query Methods ---

// GetRankedResults returns team results sorted by position for prize distribution.
// This produces the format expected by tournament.PrizeDistributionService.CalculateAndSetPayouts()
func (m *MatchResult) GetRankedResults() []RankedResult {
	results := make([]RankedResult, 0, len(m.TeamResults))
	for _, tr := range m.TeamResults {
		for _, playerID := range tr.Players {
			results = append(results, RankedResult{
				Position: tr.Position,
				UserID:   playerID,
				TeamID:   &tr.TeamID,
				Score:    tr.Score,
			})
		}
	}
	return results
}

// GetRankedPlayerIDs returns player IDs sorted by team position (for matchmaking prize distribution)
func (m *MatchResult) GetRankedPlayerIDs() []uuid.UUID {
	type posPlayer struct {
		position int
		playerID uuid.UUID
	}

	var sorted []posPlayer
	for _, tr := range m.TeamResults {
		for _, pid := range tr.Players {
			sorted = append(sorted, posPlayer{position: tr.Position, playerID: pid})
		}
	}

	// Sort by position
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].position < sorted[i].position {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	ids := make([]uuid.UUID, len(sorted))
	for i, s := range sorted {
		ids[i] = s.playerID
	}
	return ids
}

// GetMVPPlayerID returns the MVP player ID if one exists
func (m *MatchResult) GetMVPPlayerID() *uuid.UUID {
	for _, pr := range m.PlayerResults {
		if pr.IsMVP {
			id := pr.PlayerID
			return &id
		}
	}
	return nil
}

// GetTeamScore returns the score for a specific team
func (m *MatchResult) GetTeamScore(teamID uuid.UUID) (int, bool) {
	for _, tr := range m.TeamResults {
		if tr.TeamID == teamID {
			return tr.Score, true
		}
	}
	return 0, false
}

// GetPlayerResult returns the result for a specific player
func (m *MatchResult) GetPlayerResult(playerID uuid.UUID) (*PlayerResult, bool) {
	for i := range m.PlayerResults {
		if m.PlayerResults[i].PlayerID == playerID {
			return &m.PlayerResults[i], true
		}
	}
	return nil, false
}

// WasAdjusted returns true if the team results were modified during conciliation
func (m *MatchResult) WasAdjusted() bool {
	return len(m.OriginalTeamResults) > 0
}

// --- Internal Methods ---

func (m *MatchResult) determineWinner() {
	if len(m.TeamResults) < 2 {
		return
	}

	// Check for draw
	allSame := true
	for i := 1; i < len(m.TeamResults); i++ {
		if m.TeamResults[i].Score != m.TeamResults[0].Score {
			allSame = false
			break
		}
	}

	if allSame {
		m.IsDraw = true
		m.WinnerTeamID = nil
		// All teams get position 1 in a draw
		for i := range m.TeamResults {
			m.TeamResults[i].Position = 1
		}
		return
	}

	m.IsDraw = false

	// Sort by score descending and assign positions
	for i := 0; i < len(m.TeamResults); i++ {
		for j := i + 1; j < len(m.TeamResults); j++ {
			if m.TeamResults[j].Score > m.TeamResults[i].Score {
				m.TeamResults[i], m.TeamResults[j] = m.TeamResults[j], m.TeamResults[i]
			}
		}
	}

	for i := range m.TeamResults {
		m.TeamResults[i].Position = i + 1
	}

	winnerID := m.TeamResults[0].TeamID
	m.WinnerTeamID = &winnerID
}

// Validate validates the match result entity
func (m *MatchResult) Validate() error {
	if m.MatchID == uuid.Nil {
		return fmt.Errorf("match_id is required")
	}

	if m.GameID == "" {
		return fmt.Errorf("game_id is required")
	}

	if err := m.Source.Validate(); err != nil {
		return err
	}

	if err := m.Status.Validate(); err != nil {
		return err
	}

	if m.SubmittedBy == uuid.Nil {
		return fmt.Errorf("submitted_by is required")
	}

	if len(m.TeamResults) < 2 {
		return fmt.Errorf("at least 2 team results are required")
	}

	// Validate no duplicate team IDs
	teamIDs := make(map[uuid.UUID]bool)
	for _, tr := range m.TeamResults {
		if tr.TeamID == uuid.Nil {
			return fmt.Errorf("team_id is required for all team results")
		}
		if teamIDs[tr.TeamID] {
			return fmt.Errorf("duplicate team_id: %s", tr.TeamID)
		}
		teamIDs[tr.TeamID] = true
	}

	// Validate player results reference valid teams
	for _, pr := range m.PlayerResults {
		if pr.PlayerID == uuid.Nil {
			return fmt.Errorf("player_id is required for all player results")
		}
		if !teamIDs[pr.TeamID] {
			return fmt.Errorf("player %s references unknown team %s", pr.PlayerID, pr.TeamID)
		}
	}

	if m.PlayedAt.IsZero() {
		return fmt.Errorf("played_at is required")
	}

	return nil
}

// RankedResult is the output format for prize distribution integration.
// Matches the pattern expected by tournament.PrizeDistributionService.CalculateAndSetPayouts()
type RankedResult struct {
	Position int        `json:"position" bson:"position"`
	UserID   uuid.UUID  `json:"user_id" bson:"user_id"`
	TeamID   *uuid.UUID `json:"team_id,omitempty" bson:"team_id,omitempty"`
	Score    int        `json:"score" bson:"score"`
}
