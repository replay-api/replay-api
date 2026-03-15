package oracle_in

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	oracle_entities "github.com/replay-api/replay-api/pkg/domain/oracle/entities"
	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

// --- Command Handlers ---

// OracleCommandHandler handles all write operations for oracle results
type OracleCommandHandler interface {
	// IngestExternalScore ingests a score from an external provider
	IngestExternalScore(ctx context.Context, cmd IngestExternalScoreCommand) error

	// CreateExternalMatchOracle creates an oracle result for an external match (no internal MatchID)
	CreateExternalMatchOracle(ctx context.Context, cmd CreateExternalMatchOracleCommand) (*oracle_entities.OracleResult, error)

	// PublishToChain publishes a consensus-reached result to configured blockchains
	PublishToChain(ctx context.Context, cmd PublishToChainCommand) error

	// HandleDisputeEscalation escalates a dispute against a published score
	HandleDisputeEscalation(ctx context.Context, cmd HandleDisputeCommand) error

	// TriggerIngestionFromAllProviders triggers ingestion from all available providers
	TriggerIngestionFromAllProviders(ctx context.Context, cmd TriggerIngestionCommand) error
}

// --- Command DTOs ---

// IngestExternalScoreCommand represents a score submission from an external provider
type IngestExternalScoreCommand struct {
	OracleResultID  *uuid.UUID                    `json:"oracle_result_id,omitempty"`
	MatchID         *uuid.UUID                    `json:"match_id,omitempty"`
	ExternalMatchID *string                       `json:"external_match_id,omitempty"`
	GameID          replay_common.GameIDKey        `json:"game_id"`
	SourceType      oracle_vo.OracleSourceType     `json:"source_type"`
	ProviderMatchID string                         `json:"provider_match_id"`
	WinnerTeamID    *uuid.UUID                    `json:"winner_team_id,omitempty"`
	IsDraw          bool                           `json:"is_draw"`
	TeamAID         uuid.UUID                      `json:"team_a_id"`
	TeamBID         uuid.UUID                      `json:"team_b_id"`
	TeamAScore      int                            `json:"team_a_score"`
	TeamBScore      int                            `json:"team_b_score"`
	RoundsPlayed    int                            `json:"rounds_played"`
	MVPPlayerID     *uuid.UUID                    `json:"mvp_player_id,omitempty"`
	GameDetails     []oracle_entities.SubmissionGameDetail  `json:"game_details,omitempty"`
	PlayerScores    []oracle_entities.SubmissionPlayerScore `json:"player_scores,omitempty"`
}

func (c IngestExternalScoreCommand) Validate() error {
	if c.OracleResultID == nil && c.MatchID == nil && c.ExternalMatchID == nil {
		return fmt.Errorf("one of oracle_result_id, match_id, or external_match_id is required")
	}
	if c.GameID == "" {
		return fmt.Errorf("game_id is required")
	}
	if err := c.SourceType.Validate(); err != nil {
		return err
	}
	if c.ProviderMatchID == "" {
		return fmt.Errorf("provider_match_id is required")
	}
	if c.TeamAID == uuid.Nil || c.TeamBID == uuid.Nil {
		return fmt.Errorf("team_a_id and team_b_id are required")
	}
	return nil
}

// CreateExternalMatchOracleCommand creates an oracle result for a match that only exists externally
type CreateExternalMatchOracleCommand struct {
	ExternalMatchID string                 `json:"external_match_id"`
	GameID          replay_common.GameIDKey `json:"game_id"`
}

func (c CreateExternalMatchOracleCommand) Validate() error {
	if c.ExternalMatchID == "" {
		return fmt.Errorf("external_match_id is required")
	}
	if c.GameID == "" {
		return fmt.Errorf("game_id is required")
	}
	return nil
}

// PublishToChainCommand triggers on-chain publication of a consensus result
type PublishToChainCommand struct {
	OracleResultID uuid.UUID           `json:"oracle_result_id"`
	ChainIDs       []oracle_vo.ChainID `json:"chain_ids,omitempty"` // If empty, publish to all configured chains
}

func (c PublishToChainCommand) Validate() error {
	if c.OracleResultID == uuid.Nil {
		return fmt.Errorf("oracle_result_id is required")
	}
	return nil
}

// HandleDisputeCommand registers a dispute against a published oracle result
type HandleDisputeCommand struct {
	OracleResultID uuid.UUID `json:"oracle_result_id"`
	Reason         string    `json:"reason"`
	DisputedBy     uuid.UUID `json:"disputed_by"`
}

func (c HandleDisputeCommand) Validate() error {
	if c.OracleResultID == uuid.Nil {
		return fmt.Errorf("oracle_result_id is required")
	}
	if c.Reason == "" {
		return fmt.Errorf("dispute reason is required")
	}
	if c.DisputedBy == uuid.Nil {
		return fmt.Errorf("disputed_by is required")
	}
	return nil
}

// TriggerIngestionCommand triggers ingestion from all available providers for a match
type TriggerIngestionCommand struct {
	MatchID         *uuid.UUID              `json:"match_id,omitempty"`
	ExternalMatchID *string                 `json:"external_match_id,omitempty"`
	GameID          replay_common.GameIDKey  `json:"game_id"`
}

func (c TriggerIngestionCommand) Validate() error {
	if c.MatchID == nil && c.ExternalMatchID == nil {
		return fmt.Errorf("either match_id or external_match_id is required")
	}
	if c.GameID == "" {
		return fmt.Errorf("game_id is required")
	}
	return nil
}

// --- Response DTOs ---

// OracleResultDTO is the external representation of an oracle result
type OracleResultDTO struct {
	ID              uuid.UUID                        `json:"id"`
	MatchID         *uuid.UUID                       `json:"match_id,omitempty"`
	ExternalMatchID *string                          `json:"external_match_id,omitempty"`
	GameID          replay_common.GameIDKey           `json:"game_id"`
	Status          oracle_vo.OracleStatus            `json:"status"`
	ConfidenceLevel int                               `json:"confidence_level"`
	SubmissionCount int                               `json:"submission_count"`
	Consensus       *ConsensusResultDTO               `json:"consensus,omitempty"`
	Publications    []ChainPublicationDTO             `json:"publications,omitempty"`
	DisputeReason   *string                           `json:"dispute_reason,omitempty"`
	FinalizedAt     *time.Time                        `json:"finalized_at,omitempty"`
	CreatedAt       time.Time                         `json:"created_at"`
	UpdatedAt       time.Time                         `json:"updated_at"`
}

// ConsensusResultDTO is the DTO for consensus outcome
type ConsensusResultDTO struct {
	WinnerTeamID    *uuid.UUID                              `json:"winner_team_id,omitempty"`
	IsDraw          bool                                     `json:"is_draw"`
	ConfidenceLevel int                                      `json:"confidence_level"`
	AgreementRatio  float64                                  `json:"agreement_ratio"`
	SourceCount     int                                      `json:"source_count"`
	SeriesFormat    string                                   `json:"series_format"`
	GamesPlayed     int                                      `json:"games_played"`
	TeamScores      []oracle_entities.ConsensusTeamScore     `json:"team_scores"`
	GameOutcomes    []oracle_entities.GameConsensusOutcome    `json:"game_outcomes,omitempty"`
	MVPPlayerID     *uuid.UUID                               `json:"mvp_player_id,omitempty"`
	SourceHash      string                                   `json:"source_hash"`
	ComputedAt      time.Time                                `json:"computed_at"`
}

// ChainPublicationDTO is the DTO for chain publication
type ChainPublicationDTO struct {
	ChainID         oracle_vo.ChainID `json:"chain_id"`
	CAIP2           string            `json:"caip2"`
	ContractAddress string            `json:"contract_address"`
	TxHash          string            `json:"tx_hash"`
	BlockNumber     uint64            `json:"block_number"`
	GasUsed         int64             `json:"gas_used"`
	Status          string            `json:"status"`
	PublishedAt     time.Time         `json:"published_at"`
	ConfirmedAt     *time.Time        `json:"confirmed_at,omitempty"`
}

// ScoreSubmissionDTO is the DTO for individual submissions
type ScoreSubmissionDTO struct {
	ID              uuid.UUID                    `json:"id"`
	SourceType      oracle_vo.OracleSourceType   `json:"source_type"`
	ProviderMatchID string                        `json:"provider_match_id"`
	WinnerTeamID    *uuid.UUID                   `json:"winner_team_id,omitempty"`
	IsDraw          bool                          `json:"is_draw"`
	TeamAScore      int                           `json:"team_a_score"`
	TeamBScore      int                           `json:"team_b_score"`
	RoundsPlayed    int                           `json:"rounds_played"`
	GameDetails     []oracle_entities.SubmissionGameDetail  `json:"game_details,omitempty"`
	SourceHash      string                        `json:"source_hash"`
	SubmittedAt     time.Time                     `json:"submitted_at"`
}

// --- Mappers ---

// MapOracleResultToDTO converts an entity to its DTO representation
func MapOracleResultToDTO(result *oracle_entities.OracleResult) *OracleResultDTO {
	dto := &OracleResultDTO{
		ID:              result.ID,
		MatchID:         result.MatchID,
		ExternalMatchID: result.ExternalMatchID,
		GameID:          result.GameID,
		Status:          result.Status,
		ConfidenceLevel: result.ConfidenceLevel,
		SubmissionCount: len(result.Submissions),
		DisputeReason:   result.DisputeReason,
		FinalizedAt:     result.FinalizedAt,
		CreatedAt:       result.CreatedAt,
		UpdatedAt:       result.UpdatedAt,
	}

	if result.ConsensusResult != nil {
		dto.Consensus = &ConsensusResultDTO{
			WinnerTeamID:    result.ConsensusResult.WinnerTeamID,
			IsDraw:          result.ConsensusResult.IsDraw,
			ConfidenceLevel: result.ConsensusResult.ConfidenceLevel,
			AgreementRatio:  result.ConsensusResult.AgreementRatio,
			SourceCount:     result.ConsensusResult.SourceCount,
			SeriesFormat:    result.ConsensusResult.SeriesFormat,
			GamesPlayed:     result.ConsensusResult.GamesPlayed,
			TeamScores:      result.ConsensusResult.TeamScores,
			GameOutcomes:    result.ConsensusResult.GameOutcomes,
			MVPPlayerID:     result.ConsensusResult.MVPPlayerID,
			SourceHash:      result.ConsensusResult.SourceHash,
			ComputedAt:      result.ConsensusResult.ComputedAt,
		}
	}

	for _, pub := range result.Publications {
		dto.Publications = append(dto.Publications, ChainPublicationDTO{
			ChainID:         pub.ChainID,
			CAIP2:           pub.CAIP2,
			ContractAddress: pub.ContractAddress,
			TxHash:          pub.TxHash,
			BlockNumber:     pub.BlockNumber,
			GasUsed:         pub.GasUsed,
			Status:          pub.Status,
			PublishedAt:     pub.PublishedAt,
			ConfirmedAt:     pub.ConfirmedAt,
		})
	}

	return dto
}

// MapSubmissionToDTO converts a submission entity to its DTO representation
func MapSubmissionToDTO(sub *oracle_entities.ScoreSubmission) *ScoreSubmissionDTO {
	return &ScoreSubmissionDTO{
		ID:              sub.ID,
		SourceType:      sub.SourceType,
		ProviderMatchID: sub.ProviderMatchID,
		WinnerTeamID:    sub.WinnerTeamID,
		IsDraw:          sub.IsDraw,
		TeamAScore:      sub.TeamAScore,
		TeamBScore:      sub.TeamBScore,
		RoundsPlayed:    sub.RoundsPlayed,
		GameDetails:     sub.GameDetails,
		SourceHash:      sub.SourceHash,
		SubmittedAt:     sub.SubmittedAt,
	}
}
