package scores

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	matchmaking_out "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/out"
	scores_entities "github.com/replay-api/replay-api/pkg/domain/scores/entities"
	scores_out "github.com/replay-api/replay-api/pkg/domain/scores/ports/out"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
	tournament_services "github.com/replay-api/replay-api/pkg/domain/tournament/services"
)

type PrizeDistributionAdapter struct {
	tournamentPrizeService  *tournament_services.PrizeDistributionService
	tournamentPrizePoolRepo tournament_services.PrizePoolRepository
	matchmakingPrizeRepo    matchmaking_out.PrizePoolRepository
}

var _ scores_out.PrizeDistributionGateway = (*PrizeDistributionAdapter)(nil)

func NewPrizeDistributionAdapter(
	tournamentPrizeService *tournament_services.PrizeDistributionService,
	tournamentPrizePoolRepo tournament_services.PrizePoolRepository,
	matchmakingPrizeRepo matchmaking_out.PrizePoolRepository,
) *PrizeDistributionAdapter {
	return &PrizeDistributionAdapter{
		tournamentPrizeService:  tournamentPrizeService,
		tournamentPrizePoolRepo: tournamentPrizePoolRepo,
		matchmakingPrizeRepo:    matchmakingPrizeRepo,
	}
}

func (a *PrizeDistributionAdapter) TriggerTournamentPrizeDistribution(
	ctx context.Context,
	tournamentID uuid.UUID,
	results []scores_entities.RankedResult,
) (*uuid.UUID, error) {
	if a.tournamentPrizeService == nil {
		slog.WarnContext(ctx, "tournament prize service not available",
			slog.String("tournament_id", tournamentID.String()),
		)
		return nil, nil
	}

	tournamentResults := make([]tournament_entities.TournamentResult, len(results))
	for i, r := range results {
		tournamentResults[i] = tournament_entities.TournamentResult{
			Position: r.Position,
			UserID:   r.UserID,
			TeamID:   r.TeamID,
			Score:    r.Score,
		}
	}

	pool, err := a.tournamentPrizePoolRepo.GetByTournamentID(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("prize pool lookup failed for tournament %s: %w", tournamentID, err)
	}
	if pool == nil {
		slog.WarnContext(ctx, "no prize pool for tournament", slog.String("tournament_id", tournamentID.String()))
		return nil, nil
	}

	if err := a.tournamentPrizeService.CalculateAndSetPayouts(ctx, pool.ID, tournamentResults); err != nil {
		return nil, fmt.Errorf("payout calculation failed: %w", err)
	}

	if err := a.tournamentPrizeService.DistributePrizes(ctx, pool.ID); err != nil {
		return nil, fmt.Errorf("prize distribution failed: %w", err)
	}

	slog.InfoContext(ctx, "tournament prizes distributed",
		slog.String("tournament_id", tournamentID.String()),
		slog.String("pool_id", pool.ID.String()),
		slog.Int("results", len(results)),
	)
	return &pool.ID, nil
}

func (a *PrizeDistributionAdapter) TriggerMatchmakingPrizeDistribution(
	ctx context.Context,
	matchID uuid.UUID,
	rankedPlayerIDs []uuid.UUID,
	mvpPlayerID *uuid.UUID,
) (*uuid.UUID, error) {
	if a.matchmakingPrizeRepo == nil {
		slog.WarnContext(ctx, "matchmaking prize repo not available", slog.String("match_id", matchID.String()))
		return nil, nil
	}

	pool, err := a.matchmakingPrizeRepo.FindByMatchID(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("prize pool lookup failed for match %s: %w", matchID, err)
	}
	if pool == nil {
		slog.WarnContext(ctx, "no prize pool for match", slog.String("match_id", matchID.String()))
		return nil, nil
	}

	distribution, err := pool.CalculateDistribution(rankedPlayerIDs, mvpPlayerID)
	if err != nil {
		return nil, fmt.Errorf("distribution calculation failed: %w", err)
	}

	if err := pool.Distribute(distribution); err != nil {
		return nil, fmt.Errorf("prize distribution failed: %w", err)
	}

	if err := a.matchmakingPrizeRepo.Update(ctx, pool); err != nil {
		return nil, fmt.Errorf("prize pool update failed: %w", err)
	}

	slog.InfoContext(ctx, "matchmaking prizes distributed",
		slog.String("match_id", matchID.String()),
		slog.String("pool_id", pool.ID.String()),
		slog.Int("players", len(rankedPlayerIDs)),
	)
	poolID := pool.ID
	return &poolID, nil
}
