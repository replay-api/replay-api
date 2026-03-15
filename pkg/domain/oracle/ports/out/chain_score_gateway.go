package oracle_out

import (
	"context"

	"github.com/google/uuid"
	oracle_entities "github.com/replay-api/replay-api/pkg/domain/oracle/entities"
	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
)

// ChainScoreGateway defines the contract for publishing scores to blockchains
type ChainScoreGateway interface {
	// PublishScore publishes a verified score to a specific blockchain
	PublishScore(ctx context.Context, chainID oracle_vo.ChainID, result *oracle_entities.OracleResult) (*oracle_entities.ChainPublication, error)

	// GetPublishedScore retrieves a published score from a blockchain
	GetPublishedScore(ctx context.Context, chainID oracle_vo.ChainID, matchID uuid.UUID) (*OnChainScore, error)

	// IsScoreFinalized checks if a score has been finalized on-chain (dispute window closed)
	IsScoreFinalized(ctx context.Context, chainID oracle_vo.ChainID, matchID uuid.UUID) (bool, error)

	// SupportedChains returns the list of chains this gateway supports
	SupportedChains() []oracle_vo.ChainID
}

// OnChainScore represents score data read from a blockchain
type OnChainScore struct {
	MatchID      uuid.UUID         `json:"match_id"`
	WinnerTeamID *uuid.UUID        `json:"winner_team_id,omitempty"`
	TeamAScore   int               `json:"team_a_score"`
	TeamBScore   int               `json:"team_b_score"`
	SourceHash   string            `json:"source_hash"`
	IsFinalized  bool              `json:"is_finalized"`
	ChainID      oracle_vo.ChainID `json:"chain_id"`
	BlockNumber  uint64            `json:"block_number"`
}
