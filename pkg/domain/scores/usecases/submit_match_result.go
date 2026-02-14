package scores_usecases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	scores_entities "github.com/replay-api/replay-api/pkg/domain/scores/entities"
	scores_in "github.com/replay-api/replay-api/pkg/domain/scores/ports/in"
	scores_out "github.com/replay-api/replay-api/pkg/domain/scores/ports/out"
	scores_vo "github.com/replay-api/replay-api/pkg/domain/scores/value-objects"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// matchResultCommandHandler implements scores_in.MatchResultCommandHandler
type matchResultCommandHandler struct {
	repository              scores_out.MatchResultRepository
	eventPublisher          scores_out.ScoreEventPublisher
	prizeDistributionGateway scores_out.PrizeDistributionGateway
}

// NewSubmitMatchResultUseCase creates a basic command handler for submitting match results
func NewSubmitMatchResultUseCase(
	repository scores_out.MatchResultRepository,
	eventPublisher scores_out.ScoreEventPublisher,
) scores_in.MatchResultCommandHandler {
	return &matchResultCommandHandler{
		repository:     repository,
		eventPublisher: eventPublisher,
	}
}

// NewMatchResultCommandHandler creates the full command handler with all dependencies
func NewMatchResultCommandHandler(
	repository scores_out.MatchResultRepository,
	eventPublisher scores_out.ScoreEventPublisher,
	prizeDistributionGateway scores_out.PrizeDistributionGateway,
) scores_in.MatchResultCommandHandler {
	return &matchResultCommandHandler{
		repository:              repository,
		eventPublisher:          eventPublisher,
		prizeDistributionGateway: prizeDistributionGateway,
	}
}

// SubmitMatchResult handles manual score submission (from tournament admin or external sources)
func (h *matchResultCommandHandler) SubmitMatchResult(ctx context.Context, cmd scores_in.SubmitMatchResultCommand) (*scores_entities.MatchResult, error) {
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid command: %w", err)
	}

	// Get resource owner from context
	resourceOwner := shared.GetResourceOwner(ctx)

	// Check for idempotency — don't allow duplicate results for the same match
	existing, _ := h.repository.FindByMatchID(ctx, cmd.MatchID)
	if existing != nil && !existing.Status.IsTerminal() {
		return nil, fmt.Errorf("match result already exists for match %s (status: %s)", cmd.MatchID, existing.Status)
	}

	// Create the match result entity
	result, err := scores_entities.NewMatchResult(
		resourceOwner,
		cmd.MatchID,
		cmd.GameID,
		cmd.MapName,
		cmd.Mode,
		cmd.Source,
		resourceOwner.UserID,
		cmd.TeamResults,
		cmd.PlayerResults,
		cmd.PlayedAt,
		cmd.Duration,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create match result: %w", err)
	}

	// Set optional context
	if cmd.TournamentID != nil {
		result.TournamentID = cmd.TournamentID
	}
	if cmd.MatchmakingSessionID != nil {
		result.MatchmakingSessionID = cmd.MatchmakingSessionID
	}
	result.RoundsPlayed = cmd.RoundsPlayed

	// For trusted sources, auto-verify
	if cmd.Source == scores_vo.ScoreSourceTournamentAdmin || cmd.Source == scores_vo.ScoreSourceExternalAPI {
		method := scores_vo.VerificationMethodManual
		if cmd.Source == scores_vo.ScoreSourceExternalAPI {
			method = scores_vo.VerificationMethodAutomatic
		}
		verifierID := resourceOwner.UserID
		if err := result.Verify(method, &verifierID); err != nil {
			slog.WarnContext(ctx, "auto-verification failed, proceeding as submitted",
				slog.String("match_id", cmd.MatchID.String()),
				slog.String("error", err.Error()),
			)
		}
	}

	// Persist
	if err := h.repository.Save(ctx, result); err != nil {
		return nil, fmt.Errorf("failed to save match result: %w", err)
	}

	// Publish event
	if err := h.eventPublisher.PublishMatchResultSubmitted(ctx, result); err != nil {
		slog.ErrorContext(ctx, "failed to publish match result submitted event",
			slog.String("match_result_id", result.ID.String()),
			slog.String("error", err.Error()),
		)
	}

	// If auto-verified, also publish verified event
	if result.Status == scores_vo.ResultStatusVerified {
		if err := h.eventPublisher.PublishMatchResultVerified(ctx, result); err != nil {
			slog.ErrorContext(ctx, "failed to publish match result verified event",
				slog.String("match_result_id", result.ID.String()),
				slog.String("error", err.Error()),
			)
		}
	}

	slog.InfoContext(ctx, "match result submitted",
		slog.String("match_result_id", result.ID.String()),
		slog.String("match_id", cmd.MatchID.String()),
		slog.String("source", string(cmd.Source)),
		slog.String("status", string(result.Status)),
	)

	return result, nil
}

// SubmitMatchResultFromReplay handles automatic score submission from replay file processing
func (h *matchResultCommandHandler) SubmitMatchResultFromReplay(ctx context.Context, cmd scores_in.SubmitReplayResultCommand) (*scores_entities.MatchResult, error) {
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid command: %w", err)
	}

	resourceOwner := shared.GetResourceOwner(ctx)

	// Check for idempotency
	existing, _ := h.repository.FindByMatchID(ctx, cmd.MatchID)
	if existing != nil && !existing.Status.IsTerminal() {
		return nil, fmt.Errorf("match result already exists for match %s", cmd.MatchID)
	}

	// Create from replay
	result, err := scores_entities.NewMatchResultFromReplay(
		resourceOwner,
		cmd.MatchID,
		cmd.ReplayID,
		cmd.GameID,
		cmd.MapName,
		cmd.Mode,
		cmd.TeamResults,
		cmd.PlayerResults,
		cmd.PlayedAt,
		cmd.Duration,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create match result from replay: %w", err)
	}

	// Set optional context
	if cmd.MatchmakingSessionID != nil {
		result.SetMatchmakingContext(*cmd.MatchmakingSessionID)
	}
	if cmd.TournamentID != nil {
		result.TournamentID = cmd.TournamentID
	}
	result.RoundsPlayed = cmd.RoundsPlayed

	// Replay-sourced results are automatically verified
	if err := result.AutoVerify(); err != nil {
		slog.WarnContext(ctx, "auto-verification failed for replay result",
			slog.String("replay_id", cmd.ReplayID.String()),
			slog.String("error", err.Error()),
		)
	}

	// Persist
	if err := h.repository.Save(ctx, result); err != nil {
		return nil, fmt.Errorf("failed to save match result: %w", err)
	}

	// Publish events
	if err := h.eventPublisher.PublishMatchResultSubmitted(ctx, result); err != nil {
		slog.ErrorContext(ctx, "failed to publish event", slog.String("error", err.Error()))
	}

	if result.Status == scores_vo.ResultStatusVerified {
		if err := h.eventPublisher.PublishMatchResultVerified(ctx, result); err != nil {
			slog.ErrorContext(ctx, "failed to publish verified event", slog.String("error", err.Error()))
		}
	}

	slog.InfoContext(ctx, "match result submitted from replay",
		slog.String("match_result_id", result.ID.String()),
		slog.String("replay_id", cmd.ReplayID.String()),
		slog.String("status", string(result.Status)),
	)

	return result, nil
}

// VerifyMatchResult manually verifies a submitted match result
func (h *matchResultCommandHandler) VerifyMatchResult(ctx context.Context, cmd scores_in.VerifyMatchResultCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	result, err := h.repository.FindByID(ctx, cmd.MatchResultID)
	if err != nil {
		return fmt.Errorf("match result not found: %w", err)
	}

	resourceOwner := shared.GetResourceOwner(ctx)
	verifierID := resourceOwner.UserID

	if err := result.Verify(cmd.VerificationMethod, &verifierID); err != nil {
		return fmt.Errorf("failed to verify match result: %w", err)
	}

	if err := h.repository.Update(ctx, result); err != nil {
		return fmt.Errorf("failed to update match result: %w", err)
	}

	if err := h.eventPublisher.PublishMatchResultVerified(ctx, result); err != nil {
		slog.ErrorContext(ctx, "failed to publish verified event", slog.String("error", err.Error()))
	}

	slog.InfoContext(ctx, "match result verified",
		slog.String("match_result_id", result.ID.String()),
		slog.String("method", string(cmd.VerificationMethod)),
		slog.String("verified_by", verifierID.String()),
	)

	return nil
}

// DisputeMatchResult registers a dispute against a verified match result
func (h *matchResultCommandHandler) DisputeMatchResult(ctx context.Context, cmd scores_in.DisputeMatchResultCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	result, err := h.repository.FindByID(ctx, cmd.MatchResultID)
	if err != nil {
		return fmt.Errorf("match result not found: %w", err)
	}

	resourceOwner := shared.GetResourceOwner(ctx)

	if err := result.Dispute(cmd.Reason, resourceOwner.UserID); err != nil {
		return fmt.Errorf("failed to dispute match result: %w", err)
	}

	if err := h.repository.Update(ctx, result); err != nil {
		return fmt.Errorf("failed to update match result: %w", err)
	}

	if err := h.eventPublisher.PublishMatchResultDisputed(ctx, result); err != nil {
		slog.ErrorContext(ctx, "failed to publish disputed event", slog.String("error", err.Error()))
	}

	slog.InfoContext(ctx, "match result disputed",
		slog.String("match_result_id", result.ID.String()),
		slog.String("reason", cmd.Reason),
		slog.String("disputed_by", resourceOwner.UserID.String()),
	)

	return nil
}

// ConciliateMatchResult resolves a dispute with optional score adjustments
func (h *matchResultCommandHandler) ConciliateMatchResult(ctx context.Context, cmd scores_in.ConciliateMatchResultCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	result, err := h.repository.FindByID(ctx, cmd.MatchResultID)
	if err != nil {
		return fmt.Errorf("match result not found: %w", err)
	}

	resourceOwner := shared.GetResourceOwner(ctx)

	if err := result.Conciliate(resourceOwner.UserID, cmd.Notes, cmd.AdjustedTeamResults); err != nil {
		return fmt.Errorf("failed to conciliate match result: %w", err)
	}

	if err := h.repository.Update(ctx, result); err != nil {
		return fmt.Errorf("failed to update match result: %w", err)
	}

	if err := h.eventPublisher.PublishMatchResultConciliated(ctx, result); err != nil {
		slog.ErrorContext(ctx, "failed to publish conciliated event", slog.String("error", err.Error()))
	}

	slog.InfoContext(ctx, "match result conciliated",
		slog.String("match_result_id", result.ID.String()),
		slog.String("conciliated_by", resourceOwner.UserID.String()),
		slog.Bool("scores_adjusted", result.WasAdjusted()),
	)

	return nil
}

// FinalizeMatchResult finalizes a match result and triggers prize distribution
func (h *matchResultCommandHandler) FinalizeMatchResult(ctx context.Context, cmd scores_in.FinalizeMatchResultCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	result, err := h.repository.FindByID(ctx, cmd.MatchResultID)
	if err != nil {
		return fmt.Errorf("match result not found: %w", err)
	}

	if err := result.Finalize(); err != nil {
		return fmt.Errorf("failed to finalize match result: %w", err)
	}

	// Trigger prize distribution based on context
	if h.prizeDistributionGateway != nil {
		var distributionID *uuid.UUID

		if result.TournamentID != nil {
			// Tournament context: use ranked results for prize distribution
			rankedResults := result.GetRankedResults()
			distributionID, err = h.prizeDistributionGateway.TriggerTournamentPrizeDistribution(
				ctx, *result.TournamentID, rankedResults,
			)
			if err != nil {
				slog.ErrorContext(ctx, "failed to trigger tournament prize distribution",
					slog.String("match_result_id", result.ID.String()),
					slog.String("tournament_id", result.TournamentID.String()),
					slog.String("error", err.Error()),
				)
				// Don't fail finalization due to prize distribution failure
			}
		} else if result.MatchmakingSessionID != nil {
			// Matchmaking context: use ranked player IDs
			rankedPlayerIDs := result.GetRankedPlayerIDs()
			mvpPlayerID := result.GetMVPPlayerID()
			distributionID, err = h.prizeDistributionGateway.TriggerMatchmakingPrizeDistribution(
				ctx, result.MatchID, rankedPlayerIDs, mvpPlayerID,
			)
			if err != nil {
				slog.ErrorContext(ctx, "failed to trigger matchmaking prize distribution",
					slog.String("match_result_id", result.ID.String()),
					slog.String("match_id", result.MatchID.String()),
					slog.String("error", err.Error()),
				)
			}
		}

		if distributionID != nil {
			result.SetPrizeDistribution(*distributionID)
		}
	}

	if err := h.repository.Update(ctx, result); err != nil {
		return fmt.Errorf("failed to update match result: %w", err)
	}

	if err := h.eventPublisher.PublishMatchResultFinalized(ctx, result); err != nil {
		slog.ErrorContext(ctx, "failed to publish finalized event", slog.String("error", err.Error()))
	}

	slog.InfoContext(ctx, "match result finalized",
		slog.String("match_result_id", result.ID.String()),
		slog.String("match_id", result.MatchID.String()),
		slog.Bool("has_prize_distribution", result.PrizeDistributionID != nil),
	)

	return nil
}

// CancelMatchResult cancels/voids a match result
func (h *matchResultCommandHandler) CancelMatchResult(ctx context.Context, cmd scores_in.CancelMatchResultCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	result, err := h.repository.FindByID(ctx, cmd.MatchResultID)
	if err != nil {
		return fmt.Errorf("match result not found: %w", err)
	}

	if err := result.Cancel(cmd.Reason); err != nil {
		return fmt.Errorf("failed to cancel match result: %w", err)
	}

	if err := h.repository.Update(ctx, result); err != nil {
		return fmt.Errorf("failed to update match result: %w", err)
	}

	if err := h.eventPublisher.PublishMatchResultCancelled(ctx, result); err != nil {
		slog.ErrorContext(ctx, "failed to publish cancelled event", slog.String("error", err.Error()))
	}

	slog.InfoContext(ctx, "match result cancelled",
		slog.String("match_result_id", result.ID.String()),
		slog.String("reason", cmd.Reason),
	)

	return nil
}
