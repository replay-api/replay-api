package scores_usecases

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	scores_entities "github.com/replay-api/replay-api/pkg/domain/scores/entities"
	scores_in "github.com/replay-api/replay-api/pkg/domain/scores/ports/in"
	scores_out "github.com/replay-api/replay-api/pkg/domain/scores/ports/out"
	scores_vo "github.com/replay-api/replay-api/pkg/domain/scores/value-objects"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// DisputeWindowDuration is the minimum time after verification (or conciliation) during which
// non-admin users cannot finalize the result, ensuring participants have time to dispute.
// This mirrors the 72-hour escrow period used by the auto-finalization worker.
const DisputeWindowDuration = 72 * time.Hour

// matchResultCommandHandler implements scores_in.MatchResultCommandHandler
type matchResultCommandHandler struct {
	repository               scores_out.MatchResultRepository
	eventPublisher           scores_out.ScoreEventPublisher
	prizeDistributionGateway scores_out.PrizeDistributionGateway
	authorization            scores_out.ScoreAuthorization
	tournamentCallback       scores_out.TournamentMatchCallback
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
	authorization scores_out.ScoreAuthorization,
	tournamentCallback scores_out.TournamentMatchCallback,
) scores_in.MatchResultCommandHandler {
	return &matchResultCommandHandler{
		repository:               repository,
		eventPublisher:           eventPublisher,
		prizeDistributionGateway: prizeDistributionGateway,
		authorization:            authorization,
		tournamentCallback:       tournamentCallback,
	}
}

// SubmitMatchResult handles manual score submission (from tournament admin or external sources)
// Permission: tournament_admin source requires organizer/admin role; other sources require admin
func (h *matchResultCommandHandler) SubmitMatchResult(ctx context.Context, cmd scores_in.SubmitMatchResultCommand) (*scores_entities.MatchResult, error) {
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid command: %w", err)
	}

	// Get resource owner from context
	resourceOwner := shared.GetResourceOwner(ctx)

	// --- RBAC: Enforce submission permissions ---
	if h.authorization != nil {
		if err := h.authorizeSubmission(ctx, resourceOwner.UserID, cmd); err != nil {
			return nil, err
		}
	}

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
// Permission: tournament organizer (for tournament scores) OR platform admin
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

	// --- RBAC: Only tournament organizer or admin can verify ---
	if h.authorization != nil {
		if err := h.authorizeAdminAction(ctx, resourceOwner.UserID, result, "verify"); err != nil {
			return err
		}
	}

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
// Permission: match participant (player in team or player_results) OR platform admin
func (h *matchResultCommandHandler) DisputeMatchResult(ctx context.Context, cmd scores_in.DisputeMatchResultCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	result, err := h.repository.FindByID(ctx, cmd.MatchResultID)
	if err != nil {
		return fmt.Errorf("match result not found: %w", err)
	}

	resourceOwner := shared.GetResourceOwner(ctx)

	// --- RBAC: Only match participants or admin can dispute ---
	if h.authorization != nil {
		if err := h.authorizeDisputeAction(ctx, resourceOwner.UserID, result); err != nil {
			return err
		}
	}

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
// Permission: tournament organizer (for tournament scores) OR platform admin
func (h *matchResultCommandHandler) ConciliateMatchResult(ctx context.Context, cmd scores_in.ConciliateMatchResultCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	result, err := h.repository.FindByID(ctx, cmd.MatchResultID)
	if err != nil {
		return fmt.Errorf("match result not found: %w", err)
	}

	resourceOwner := shared.GetResourceOwner(ctx)

	// --- RBAC: Only tournament organizer or admin can conciliate ---
	if h.authorization != nil {
		if err := h.authorizeAdminAction(ctx, resourceOwner.UserID, result, "conciliate"); err != nil {
			return err
		}
	}

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
// Permission: tournament organizer (for tournament scores) OR platform admin
func (h *matchResultCommandHandler) FinalizeMatchResult(ctx context.Context, cmd scores_in.FinalizeMatchResultCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	result, err := h.repository.FindByID(ctx, cmd.MatchResultID)
	if err != nil {
		return fmt.Errorf("match result not found: %w", err)
	}

	// --- RBAC: Only tournament organizer or admin can finalize ---
	isPlatformAdmin := false
	if h.authorization != nil {
		resourceOwner := shared.GetResourceOwner(ctx)
		if err := h.authorizeAdminAction(ctx, resourceOwner.UserID, result, "finalize"); err != nil {
			return err
		}
		isPlatformAdmin = h.authorization.IsPlatformAdmin(ctx)
	}

	// --- Dispute Window Guard (Financial-Grade) ---
	// Non-admin users must wait for the dispute window to elapse before finalizing.
	// Only platform admins can force-finalize before the window expires.
	if !isPlatformAdmin {
		var referenceTime *time.Time
		if result.ConciliatedAt != nil {
			referenceTime = result.ConciliatedAt
		} else if result.VerifiedAt != nil {
			referenceTime = result.VerifiedAt
		}

		if referenceTime != nil && time.Since(*referenceTime) < DisputeWindowDuration {
			return fmt.Errorf("cannot finalize: dispute window has not elapsed (%.0f hours remaining); only platform admins can force-finalize",
				(DisputeWindowDuration - time.Since(*referenceTime)).Hours())
		}
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

	// --- Tournament Bracket Advancement Callback ---
	// After full finalization (dispute window passed, prizes distributed, event published),
	// notify the tournament domain so it can record the result and advance the bracket.
	if h.tournamentCallback != nil && result.TournamentID != nil && result.WinnerTeamID != nil {
		if err := h.tournamentCallback.OnMatchResultFinalized(ctx, *result.TournamentID, result.MatchID, *result.WinnerTeamID); err != nil {
			slog.ErrorContext(ctx, "tournament match callback failed (non-fatal)",
				slog.String("match_result_id", result.ID.String()),
				slog.String("tournament_id", result.TournamentID.String()),
				slog.String("match_id", result.MatchID.String()),
				slog.String("error", err.Error()),
			)
			// Non-fatal: tournament state can be reconciled later via event replay
		}
	}

	slog.InfoContext(ctx, "match result finalized",
		slog.String("match_result_id", result.ID.String()),
		slog.String("match_id", result.MatchID.String()),
		slog.Bool("has_prize_distribution", result.PrizeDistributionID != nil),
	)

	return nil
}

// CancelMatchResult cancels/voids a match result
// Permission: tournament organizer (for tournament scores) OR platform admin
func (h *matchResultCommandHandler) CancelMatchResult(ctx context.Context, cmd scores_in.CancelMatchResultCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	result, err := h.repository.FindByID(ctx, cmd.MatchResultID)
	if err != nil {
		return fmt.Errorf("match result not found: %w", err)
	}

	// --- RBAC: Only tournament organizer or admin can cancel ---
	resourceOwner := shared.GetResourceOwner(ctx)
	if h.authorization != nil {
		if err := h.authorizeAdminAction(ctx, resourceOwner.UserID, result, "cancel"); err != nil {
			return err
		}
	}

	if err := result.Cancel(cmd.Reason, resourceOwner.UserID); err != nil {
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

// --- Private RBAC Helper Methods (Financial-Grade Authorization) ---

// authorizeSubmission enforces permission checks for manual score submission.
// - tournament_admin source: user must be tournament organizer or platform admin
// - external_api / consensus: requires platform admin
// - replay_file: only allowed via SubmitMatchResultFromReplay (system pipeline)
func (h *matchResultCommandHandler) authorizeSubmission(ctx context.Context, userID uuid.UUID, cmd scores_in.SubmitMatchResultCommand) error {
	// Platform admin can always submit
	if h.authorization.IsPlatformAdmin(ctx) {
		return nil
	}

	switch cmd.Source {
	case scores_vo.ScoreSourceTournamentAdmin:
		// Tournament admin source: must be the tournament organizer
		if cmd.TournamentID == nil {
			return fmt.Errorf("forbidden: tournament_admin source requires a tournament_id")
		}
		isOrganizer, err := h.authorization.IsTournamentOrganizer(ctx, userID, *cmd.TournamentID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to check tournament organizer status",
				slog.String("user_id", userID.String()),
				slog.String("error", err.Error()),
			)
			return fmt.Errorf("authorization check failed: %w", err)
		}
		if !isOrganizer {
			return fmt.Errorf("forbidden: only the tournament organizer can submit scores for this tournament")
		}
		return nil

	case scores_vo.ScoreSourceExternalAPI, scores_vo.ScoreSourceConsensus:
		// External API and consensus sources require admin
		return fmt.Errorf("forbidden: %s source requires platform admin privileges", cmd.Source)

	case scores_vo.ScoreSourceReplayFile:
		// Replay file submissions should go through SubmitMatchResultFromReplay
		return fmt.Errorf("forbidden: replay_file source must use the replay processing pipeline")

	case scores_vo.ScoreSourceMatchmaking:
		// Matchmaking source: only accepted from the matchmaking pipeline (admin/system)
		return fmt.Errorf("forbidden: matchmaking source requires platform admin privileges")

	default:
		return fmt.Errorf("forbidden: unknown score source %s", cmd.Source)
	}
}

// authorizeAdminAction enforces permission checks for admin-level score operations
// (verify, conciliate, finalize, cancel). Requires tournament organizer or platform admin.
func (h *matchResultCommandHandler) authorizeAdminAction(ctx context.Context, userID uuid.UUID, result *scores_entities.MatchResult, action string) error {
	// Platform admin can always perform admin actions
	if h.authorization.IsPlatformAdmin(ctx) {
		return nil
	}

	// If this result belongs to a tournament, check if user is the organizer
	if result.TournamentID != nil {
		isOrganizer, err := h.authorization.IsTournamentOrganizer(ctx, userID, *result.TournamentID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to check tournament organizer status for "+action,
				slog.String("user_id", userID.String()),
				slog.String("match_result_id", result.ID.String()),
				slog.String("error", err.Error()),
			)
			return fmt.Errorf("authorization check failed: %w", err)
		}
		if isOrganizer {
			return nil
		}
	}

	// For matchmaking results, only platform admin can perform admin actions
	// (they don't have a tournament organizer)
	return fmt.Errorf("forbidden: insufficient permissions to %s this match result", action)
}

// authorizeDisputeAction enforces permission checks for dispute operations.
// Only match participants (players in team_results.players or player_results) or admin can dispute.
func (h *matchResultCommandHandler) authorizeDisputeAction(ctx context.Context, userID uuid.UUID, result *scores_entities.MatchResult) error {
	// Platform admin can always dispute
	if h.authorization.IsPlatformAdmin(ctx) {
		return nil
	}

	// Check if user is a tournament organizer
	if result.TournamentID != nil {
		isOrganizer, err := h.authorization.IsTournamentOrganizer(ctx, userID, *result.TournamentID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to check tournament organizer status for dispute",
				slog.String("user_id", userID.String()),
				slog.String("error", err.Error()),
			)
		}
		if isOrganizer {
			return nil
		}
	}

	// Check if user is a match participant
	isParticipant, err := h.authorization.IsMatchParticipant(ctx, userID, result.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check match participant status",
			slog.String("user_id", userID.String()),
			slog.String("match_result_id", result.ID.String()),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("authorization check failed: %w", err)
	}
	if !isParticipant {
		return fmt.Errorf("forbidden: only match participants can dispute a match result")
	}

	return nil
}
