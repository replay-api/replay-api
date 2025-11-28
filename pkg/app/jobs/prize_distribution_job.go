package jobs

import (
	"context"
	"log/slog"
	"time"

	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_out "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/out"
	wallet_in "github.com/replay-api/replay-api/pkg/domain/wallet/ports/in"
)

type PrizeDistributionJob struct {
	prizePoolRepo matchmaking_out.PrizePoolRepository
	walletCommand wallet_in.WalletCommand
	ticker        *time.Ticker
	interval      time.Duration
}

func NewPrizeDistributionJob(
	poolRepo matchmaking_out.PrizePoolRepository,
	walletCmd wallet_in.WalletCommand,
	interval time.Duration,
) *PrizeDistributionJob {
	return &PrizeDistributionJob{
		prizePoolRepo: poolRepo,
		walletCommand: walletCmd,
		ticker:        time.NewTicker(interval),
		interval:      interval,
	}
}

func (j *PrizeDistributionJob) Run(ctx context.Context) {
	slog.InfoContext(ctx, "Prize distribution job started", "interval", j.interval)
	defer j.ticker.Stop()

	// Run once immediately on start
	j.processPendingDistributions(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "Prize distribution job stopped")
			return
		case <-j.ticker.C:
			j.processPendingDistributions(ctx)
		}
	}
}

func (j *PrizeDistributionJob) processPendingDistributions(ctx context.Context) {
	slog.InfoContext(ctx, "Processing pending prize distributions")

	pools, err := j.prizePoolRepo.FindPendingDistributions(ctx, 100)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find pending distributions", "error", err)
		return
	}

	if len(pools) == 0 {
		slog.InfoContext(ctx, "No pending distributions found")
		return
	}

	slog.InfoContext(ctx, "Found pending distributions", "count", len(pools))

	for _, pool := range pools {
		if err := j.distributePool(ctx, pool); err != nil {
			slog.ErrorContext(ctx, "Failed to distribute pool", "pool_id", pool.ID, "match_id", pool.MatchID, "error", err)
			continue
		}
	}
}

func (j *PrizeDistributionJob) distributePool(ctx context.Context, pool *matchmaking_entities.PrizePool) error {
	slog.InfoContext(ctx, "Distributing prize pool", "pool_id", pool.ID, "match_id", pool.MatchID, "total_amount", pool.TotalAmount.String())

	// Verify pool is in correct state
	if pool.Status != matchmaking_entities.PrizePoolStatusInEscrow {
		slog.WarnContext(ctx, "Pool not in escrow status, skipping", "pool_id", pool.ID, "status", pool.Status)
		return nil
	}

	// Verify escrow period has ended
	if pool.EscrowEndTime != nil && time.Now().UTC().Before(*pool.EscrowEndTime) {
		slog.WarnContext(ctx, "Escrow period not ended yet", "pool_id", pool.ID, "escrow_end_time", pool.EscrowEndTime)
		return nil
	}

	// Check if winners are set
	if len(pool.Winners) == 0 {
		slog.ErrorContext(ctx, "No winners set for pool", "pool_id", pool.ID, "match_id", pool.MatchID)
		return nil // Don't retry, requires manual intervention
	}

	// Distribute prizes to each winner
	successCount := 0
	for _, winner := range pool.Winners {
		addPrizeCmd := wallet_in.AddPrizeCommand{
			UserID:   winner.PlayerID,
			Currency: string(pool.Currency),
			Amount:   winner.Amount.ToFloat(),
		}

		if err := j.walletCommand.AddPrize(ctx, addPrizeCmd); err != nil {
			slog.ErrorContext(ctx, "Failed to add prize to wallet", "pool_id", pool.ID, "player_id", winner.PlayerID, "amount", winner.Amount.String(), "error", err)
			// Continue with other winners even if one fails
			continue
		}

		// Mark this winner as paid
		now := time.Now().UTC()
		winner.PaidAt = &now
		successCount++

		slog.InfoContext(ctx, "Prize distributed to winner", "pool_id", pool.ID, "player_id", winner.PlayerID, "rank", winner.Rank, "amount", winner.Amount.String())
	}

	// If all prizes distributed successfully, mark pool as distributed
	if successCount == len(pool.Winners) {
		pool.Status = matchmaking_entities.PrizePoolStatusDistributed
		now := time.Now().UTC()
		pool.DistributedAt = &now

		if err := j.prizePoolRepo.Update(ctx, pool); err != nil {
			slog.ErrorContext(ctx, "Failed to update pool status to distributed", "pool_id", pool.ID, "error", err)
			return err
		}

		slog.InfoContext(ctx, "Prize pool fully distributed", "pool_id", pool.ID, "match_id", pool.MatchID, "winner_count", successCount, "total_amount", pool.TotalAmount.String())
	} else {
		// Partial distribution - update winners but keep status as in_escrow for retry
		if err := j.prizePoolRepo.Update(ctx, pool); err != nil {
			slog.ErrorContext(ctx, "Failed to update pool with partial distribution", "pool_id", pool.ID, "error", err)
			return err
		}

		slog.WarnContext(ctx, "Prize pool partially distributed", "pool_id", pool.ID, "success_count", successCount, "total_winners", len(pool.Winners))
	}

	return nil
}
