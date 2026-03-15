package oracle_entities

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// OracleResult is the aggregate root representing a consensus-verified score
// suitable for on-chain publication. It aggregates submissions from multiple
// external providers and computes a weighted consensus outcome.
type OracleResult struct {
	shared.BaseEntity

	// Match Context
	MatchID         *uuid.UUID              `json:"match_id,omitempty" bson:"match_id,omitempty"`
	ExternalMatchID *string                 `json:"external_match_id,omitempty" bson:"external_match_id,omitempty"`
	GameID          replay_common.GameIDKey  `json:"game_id" bson:"game_id"`

	// Lifecycle
	Status          oracle_vo.OracleStatus  `json:"status" bson:"status"`
	ConfidenceLevel int                     `json:"confidence_level" bson:"confidence_level"` // 0=none, 1=low, 2=medium, 3=high

	// Aggregated Data
	Submissions     []ScoreSubmission       `json:"submissions" bson:"submissions"`
	ConsensusResult *ConsensusOutcome        `json:"consensus_result,omitempty" bson:"consensus_result,omitempty"`
	Publications    []ChainPublication       `json:"publications,omitempty" bson:"publications,omitempty"`

	// Dispute
	DisputeReason   *string                 `json:"dispute_reason,omitempty" bson:"dispute_reason,omitempty"`
	DisputedBy      *uuid.UUID              `json:"disputed_by,omitempty" bson:"disputed_by,omitempty"`
	DisputedAt      *time.Time              `json:"disputed_at,omitempty" bson:"disputed_at,omitempty"`

	// Finalization
	FinalizedAt     *time.Time              `json:"finalized_at,omitempty" bson:"finalized_at,omitempty"`
}

// ScoreSubmission represents a single score input from an external provider or OCR source
type ScoreSubmission struct {
	ID              uuid.UUID                    `json:"id" bson:"id"`
	OracleResultID  uuid.UUID                    `json:"oracle_result_id" bson:"oracle_result_id"`
	SourceType      oracle_vo.OracleSourceType   `json:"source_type" bson:"source_type"`
	ProviderMatchID string                        `json:"provider_match_id" bson:"provider_match_id"`

	// Match-level results
	WinnerTeamID    *uuid.UUID                   `json:"winner_team_id,omitempty" bson:"winner_team_id,omitempty"`
	IsDraw          bool                          `json:"is_draw" bson:"is_draw"`
	TeamAID         uuid.UUID                     `json:"team_a_id" bson:"team_a_id"`
	TeamBID         uuid.UUID                     `json:"team_b_id" bson:"team_b_id"`
	TeamAScore      int                           `json:"team_a_score" bson:"team_a_score"`
	TeamBScore      int                           `json:"team_b_score" bson:"team_b_score"`
	RoundsPlayed    int                           `json:"rounds_played" bson:"rounds_played"`

	// Detailed results
	MVPPlayerID     *uuid.UUID                   `json:"mvp_player_id,omitempty" bson:"mvp_player_id,omitempty"`
	GameDetails     []SubmissionGameDetail        `json:"game_details,omitempty" bson:"game_details,omitempty"`
	PlayerScores    []SubmissionPlayerScore       `json:"player_scores,omitempty" bson:"player_scores,omitempty"`

	// Provenance
	RawResponse     json.RawMessage              `json:"raw_response,omitempty" bson:"raw_response,omitempty"`
	SourceHash      string                        `json:"source_hash" bson:"source_hash"`
	SubmittedAt     time.Time                     `json:"submitted_at" bson:"submitted_at"`
}

// SubmissionGameDetail represents per-game (per-map) scores within a series
type SubmissionGameDetail struct {
	Position    int        `json:"position" bson:"position"`       // 1-indexed game in series
	MapName     string     `json:"map_name" bson:"map_name"`
	TeamAScore  int        `json:"team_a_score" bson:"team_a_score"`
	TeamBScore  int        `json:"team_b_score" bson:"team_b_score"`
	TeamAWon    bool       `json:"team_a_won" bson:"team_a_won"`
}

// SubmissionPlayerScore represents individual player stats from a provider
type SubmissionPlayerScore struct {
	PlayerID   uuid.UUID `json:"player_id" bson:"player_id"`
	TeamID     uuid.UUID `json:"team_id" bson:"team_id"`
	Kills      int       `json:"kills" bson:"kills"`
	Deaths     int       `json:"deaths" bson:"deaths"`
	Assists    int       `json:"assists" bson:"assists"`
	Rating     float64   `json:"rating" bson:"rating"`
}

// ConsensusOutcome represents the computed consensus from multiple submissions
type ConsensusOutcome struct {
	WinnerTeamID      *uuid.UUID              `json:"winner_team_id,omitempty" bson:"winner_team_id,omitempty"`
	IsDraw            bool                     `json:"is_draw" bson:"is_draw"`
	ConfidenceLevel   int                      `json:"confidence_level" bson:"confidence_level"` // 0-3
	AgreementRatio    float64                  `json:"agreement_ratio" bson:"agreement_ratio"`   // 0.0-1.0
	SourceCount       int                      `json:"source_count" bson:"source_count"`
	SeriesFormat      string                   `json:"series_format" bson:"series_format"`       // "bo1", "bo3", "bo5"
	GamesPlayed       int                      `json:"games_played" bson:"games_played"`
	TeamScores        []ConsensusTeamScore     `json:"team_scores" bson:"team_scores"`
	GameOutcomes      []GameConsensusOutcome   `json:"game_outcomes,omitempty" bson:"game_outcomes,omitempty"`
	MVPPlayerID       *uuid.UUID               `json:"mvp_player_id,omitempty" bson:"mvp_player_id,omitempty"`
	CrossValidation   []CrossValidationEntry   `json:"cross_validation,omitempty" bson:"cross_validation,omitempty"`
	DisagreementNotes []string                 `json:"disagreement_notes,omitempty" bson:"disagreement_notes,omitempty"`
	SourceHash        string                   `json:"source_hash" bson:"source_hash"`
	ComputedAt        time.Time                `json:"computed_at" bson:"computed_at"`
}

// ConsensusTeamScore represents team-level score from consensus
type ConsensusTeamScore struct {
	TeamID uuid.UUID `json:"team_id" bson:"team_id"`
	Score  int       `json:"score" bson:"score"`
}

// GameConsensusOutcome represents per-game consensus outcome
type GameConsensusOutcome struct {
	Position   int    `json:"position" bson:"position"`
	MapName    string `json:"map_name" bson:"map_name"`
	TeamAScore int    `json:"team_a_score" bson:"team_a_score"`
	TeamBScore int    `json:"team_b_score" bson:"team_b_score"`
	TeamAWon   bool   `json:"team_a_won" bson:"team_a_won"`
}

// CrossValidationEntry represents a pairwise validation between two sources
type CrossValidationEntry struct {
	SourceA          oracle_vo.OracleSourceType `json:"source_a" bson:"source_a"`
	SourceB          oracle_vo.OracleSourceType `json:"source_b" bson:"source_b"`
	WinnerAgree      bool                        `json:"winner_agree" bson:"winner_agree"`
	ScoreAgree       bool                        `json:"score_agree" bson:"score_agree"`
	GamesAgree       bool                        `json:"games_agree" bson:"games_agree"`
	MVPAgree         bool                        `json:"mvp_agree" bson:"mvp_agree"`
	DisagreementNote string                      `json:"disagreement_note,omitempty" bson:"disagreement_note,omitempty"`
}

// ChainPublication records a score publication to a specific blockchain
type ChainPublication struct {
	ChainID         oracle_vo.ChainID `json:"chain_id" bson:"chain_id"`
	CAIP2           string            `json:"caip2" bson:"caip2"`                       // e.g., "eip155:137" or "solana:mainnet"
	ContractAddress string            `json:"contract_address" bson:"contract_address"`
	TxHash          string            `json:"tx_hash" bson:"tx_hash"`
	BlockNumber     uint64            `json:"block_number" bson:"block_number"`
	GasUsed         int64             `json:"gas_used" bson:"gas_used"`
	Status          string            `json:"status" bson:"status"`                     // "pending", "confirmed", "failed"
	PublishedAt     time.Time         `json:"published_at" bson:"published_at"`
	ConfirmedAt     *time.Time        `json:"confirmed_at,omitempty" bson:"confirmed_at,omitempty"`
}

// --- Factory Methods ---

// NewOracleResult creates a new OracleResult from an internal match finalization
func NewOracleResult(
	resourceOwner shared.ResourceOwner,
	matchID uuid.UUID,
	gameID replay_common.GameIDKey,
) *OracleResult {
	return &OracleResult{
		BaseEntity:      shared.NewEntity(resourceOwner),
		MatchID:         &matchID,
		GameID:          gameID,
		Status:          oracle_vo.OracleStatusPending,
		ConfidenceLevel: 0,
		Submissions:     make([]ScoreSubmission, 0),
		Publications:    make([]ChainPublication, 0),
	}
}

// NewExternalOracleResult creates a new OracleResult from an external match (no internal MatchID)
func NewExternalOracleResult(
	resourceOwner shared.ResourceOwner,
	externalMatchID string,
	gameID replay_common.GameIDKey,
) *OracleResult {
	return &OracleResult{
		BaseEntity:      shared.NewEntity(resourceOwner),
		ExternalMatchID: &externalMatchID,
		GameID:          gameID,
		Status:          oracle_vo.OracleStatusPending,
		ConfidenceLevel: 0,
		Submissions:     make([]ScoreSubmission, 0),
		Publications:    make([]ChainPublication, 0),
	}
}

// --- State Machine Methods ---

// AddSubmission adds a provider submission to this oracle result
func (o *OracleResult) AddSubmission(sub ScoreSubmission) error {
	if o.Status.IsTerminal() {
		return fmt.Errorf("cannot add submission to oracle result in terminal state: %s", o.Status)
	}

	// Check for duplicate source
	for _, existing := range o.Submissions {
		if existing.SourceType == sub.SourceType && existing.ProviderMatchID == sub.ProviderMatchID {
			return fmt.Errorf("duplicate submission from provider %s for match %s", sub.SourceType, sub.ProviderMatchID)
		}
	}

	sub.ID = uuid.New()
	sub.OracleResultID = o.ID
	sub.SubmittedAt = time.Now().UTC()
	o.Submissions = append(o.Submissions, sub)
	o.UpdatedAt = time.Now().UTC()
	return nil
}

// IsReadyForConsensus checks if enough sources have submitted for consensus
func (o *OracleResult) IsReadyForConsensus(minSources int) bool {
	return len(o.Submissions) >= minSources
}

// SetConsensusResult records the computed consensus outcome and transitions state
func (o *OracleResult) SetConsensusResult(outcome ConsensusOutcome) error {
	if err := o.Status.ValidateTransition(oracle_vo.OracleStatusConsensusReached); err != nil {
		return err
	}

	o.ConsensusResult = &outcome
	o.ConfidenceLevel = outcome.ConfidenceLevel
	o.Status = oracle_vo.OracleStatusConsensusReached
	o.UpdatedAt = time.Now().UTC()
	return nil
}

// MarkPublishing transitions to publishing state before on-chain submission
func (o *OracleResult) MarkPublishing() error {
	if err := o.Status.ValidateTransition(oracle_vo.OracleStatusPublishing); err != nil {
		return err
	}

	o.Status = oracle_vo.OracleStatusPublishing
	o.UpdatedAt = time.Now().UTC()
	return nil
}

// AddPublication records a successful on-chain publication
func (o *OracleResult) AddPublication(pub ChainPublication) error {
	if o.Status != oracle_vo.OracleStatusPublishing && o.Status != oracle_vo.OracleStatusPublished {
		return fmt.Errorf("cannot add publication in state: %s", o.Status)
	}

	o.Publications = append(o.Publications, pub)
	o.Status = oracle_vo.OracleStatusPublished
	o.UpdatedAt = time.Now().UTC()
	return nil
}

// Finalize marks the oracle result as final after the dispute window
func (o *OracleResult) Finalize() error {
	if err := o.Status.ValidateTransition(oracle_vo.OracleStatusFinalized); err != nil {
		return err
	}

	now := time.Now().UTC()
	o.Status = oracle_vo.OracleStatusFinalized
	o.FinalizedAt = &now
	o.UpdatedAt = now
	return nil
}

// Dispute marks the oracle result as disputed
func (o *OracleResult) Dispute(reason string, disputedBy uuid.UUID) error {
	if err := o.Status.ValidateTransition(oracle_vo.OracleStatusDisputed); err != nil {
		return err
	}

	now := time.Now().UTC()
	o.Status = oracle_vo.OracleStatusDisputed
	o.DisputeReason = &reason
	o.DisputedBy = &disputedBy
	o.DisputedAt = &now
	o.UpdatedAt = now
	return nil
}

// ResetForReconsensus resets a disputed result back to pending for re-evaluation
func (o *OracleResult) ResetForReconsensus() error {
	if o.Status != oracle_vo.OracleStatusDisputed {
		return fmt.Errorf("can only reset disputed results, current state: %s", o.Status)
	}

	o.Status = oracle_vo.OracleStatusPending
	o.ConsensusResult = nil
	o.ConfidenceLevel = 0
	o.Submissions = make([]ScoreSubmission, 0) // Clear submissions for re-ingestion
	o.UpdatedAt = time.Now().UTC()
	return nil
}

// MarkFailed transitions to failed state
func (o *OracleResult) MarkFailed() error {
	if err := o.Status.ValidateTransition(oracle_vo.OracleStatusFailed); err != nil {
		return err
	}

	o.Status = oracle_vo.OracleStatusFailed
	o.UpdatedAt = time.Now().UTC()
	return nil
}

// --- Query Methods ---

// HasSubmissionFromSource checks if a submission from a given source already exists
func (o *OracleResult) HasSubmissionFromSource(source oracle_vo.OracleSourceType) bool {
	for _, sub := range o.Submissions {
		if sub.SourceType == source {
			return true
		}
	}
	return false
}

// GetSubmissionCount returns the number of submissions
func (o *OracleResult) GetSubmissionCount() int {
	return len(o.Submissions)
}

// IsPublishedOnChain checks if the result has been published to a specific chain
func (o *OracleResult) IsPublishedOnChain(chainID oracle_vo.ChainID) bool {
	for _, pub := range o.Publications {
		if pub.ChainID == chainID && pub.Status == "confirmed" {
			return true
		}
	}
	return false
}

// GetPublicationForChain returns the publication for a specific chain, if it exists
func (o *OracleResult) GetPublicationForChain(chainID oracle_vo.ChainID) *ChainPublication {
	for i := range o.Publications {
		if o.Publications[i].ChainID == chainID {
			return &o.Publications[i]
		}
	}
	return nil
}

// Validate checks the integrity of the oracle result
func (o *OracleResult) Validate() error {
	if o.GameID == "" {
		return fmt.Errorf("game_id is required")
	}

	if o.MatchID == nil && o.ExternalMatchID == nil {
		return fmt.Errorf("either match_id or external_match_id must be set")
	}

	if err := o.Status.Validate(); err != nil {
		return err
	}

	return nil
}
