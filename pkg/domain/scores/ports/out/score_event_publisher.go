package scores_out

import (
	"context"

	"github.com/google/uuid"
	scores_entities "github.com/replay-api/replay-api/pkg/domain/scores/entities"
)

// ScoreEventPublisher defines the contract for publishing score-related events
type ScoreEventPublisher interface {
	PublishMatchResultSubmitted(ctx context.Context, result *scores_entities.MatchResult) error
	PublishMatchResultVerified(ctx context.Context, result *scores_entities.MatchResult) error
	PublishMatchResultDisputed(ctx context.Context, result *scores_entities.MatchResult) error
	PublishMatchResultConciliated(ctx context.Context, result *scores_entities.MatchResult) error
	PublishMatchResultFinalized(ctx context.Context, result *scores_entities.MatchResult) error
	PublishMatchResultCancelled(ctx context.Context, result *scores_entities.MatchResult) error
}

// PrizeDistributionGateway defines the contract for triggering prize distribution
// after match results are finalized. Supports both tournament and matchmaking contexts.
type PrizeDistributionGateway interface {
	// TriggerTournamentPrizeDistribution triggers prize calculation and distribution for a tournament match
	TriggerTournamentPrizeDistribution(
		ctx context.Context,
		tournamentID uuid.UUID,
		results []scores_entities.RankedResult,
	) (*uuid.UUID, error)

	// TriggerMatchmakingPrizeDistribution triggers prize distribution for a matchmaking match
	TriggerMatchmakingPrizeDistribution(
		ctx context.Context,
		matchID uuid.UUID,
		rankedPlayerIDs []uuid.UUID,
		mvpPlayerID *uuid.UUID,
	) (*uuid.UUID, error)
}
