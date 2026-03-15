package oracle_out

import (
	"context"

	oracle_entities "github.com/replay-api/replay-api/pkg/domain/oracle/entities"
)

// OracleEventPublisher defines the contract for publishing oracle-related domain events
type OracleEventPublisher interface {
	// PublishConsensusReached publishes an event when consensus is computed
	PublishConsensusReached(ctx context.Context, result *oracle_entities.OracleResult) error

	// PublishScorePublished publishes an event when score is published on-chain
	PublishScorePublished(ctx context.Context, result *oracle_entities.OracleResult) error

	// PublishScoreFinalized publishes an event when score is finalized (dispute window closed)
	PublishScoreFinalized(ctx context.Context, result *oracle_entities.OracleResult) error

	// PublishScoreDisputed publishes an event when a published score is disputed
	PublishScoreDisputed(ctx context.Context, result *oracle_entities.OracleResult) error

	// PublishExternalIngested publishes an event when a provider submission is ingested
	PublishExternalIngested(ctx context.Context, result *oracle_entities.OracleResult, sub oracle_entities.ScoreSubmission) error
}
