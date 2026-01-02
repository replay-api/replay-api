package matchmaking_usecases

import (
	"context"
	"fmt"
	"log/slog"

	common "github.com/replay-api/replay-api/pkg/domain"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	matchmaking_out "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/out"
	kafka "github.com/replay-api/replay-api/pkg/infra/kafka"
)

// LeaveMatchmakingQueueUseCase handles player leaving matchmaking queue
type LeaveMatchmakingQueueUseCase struct {
	billableOperationHandler billing_in.BillableOperationCommandHandler
	sessionRepository        matchmaking_out.MatchmakingSessionRepository
	eventPublisher           *kafka.EventPublisher
}

// NewLeaveMatchmakingQueueUseCase creates a new leave queue usecase
func NewLeaveMatchmakingQueueUseCase(
	billableOperationHandler billing_in.BillableOperationCommandHandler,
	sessionRepository matchmaking_out.MatchmakingSessionRepository,
	eventPublisher *kafka.EventPublisher,
) matchmaking_in.LeaveMatchmakingQueueCommandHandler {
	return &LeaveMatchmakingQueueUseCase{
		billableOperationHandler: billableOperationHandler,
		sessionRepository:        sessionRepository,
		eventPublisher:           eventPublisher,
	}
}

// Exec executes the leave matchmaking queue command
func (uc *LeaveMatchmakingQueueUseCase) Exec(ctx context.Context, cmd matchmaking_in.LeaveMatchmakingQueueCommand) error {
	// auth check
	isAuthenticated := ctx.Value(common.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return common.NewErrUnauthorized()
	}

	// get session
	session, err := uc.sessionRepository.GetByID(ctx, cmd.SessionID)
	if err != nil {
		slog.ErrorContext(ctx, "session not found", "error", err, "session_id", cmd.SessionID)
		return fmt.Errorf("session not found")
	}

	// verify player owns session
	if session.PlayerID != cmd.PlayerID {
		return common.NewErrForbidden("player does not own this session")
	}

	// check if session can be left
	if !session.CanMatch() {
		return fmt.Errorf("cannot leave queue from status: %s", session.Status)
	}

	// billing validation BEFORE leaving queue
	billingCmd := billing_in.BillableOperationCommand{
		OperationID: billing_entities.OperationTypeLeaveMatchmakingQueue,
		UserID:      common.GetResourceOwner(ctx).UserID,
		Amount:      1,
		Args: map[string]interface{}{
			"session_id": cmd.SessionID.String(),
		},
	}

	err = uc.billableOperationHandler.Validate(ctx, billingCmd)
	if err != nil {
		slog.ErrorContext(ctx, "billing validation failed for leave matchmaking queue", "error", err, "session_id", cmd.SessionID)
		return err
	}

	// update session status to cancelled
	err = uc.sessionRepository.UpdateStatus(ctx, cmd.SessionID, matchmaking_entities.StatusCancelled)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update session status", "error", err, "session_id", cmd.SessionID)
		return fmt.Errorf("failed to leave matchmaking queue")
	}

	// publish queue left event
	if uc.eventPublisher != nil {
		queueEvent := &kafka.QueueEvent{
			PlayerID:  cmd.PlayerID,
			GameType:  session.Preferences.GameID,
			Region:    session.Preferences.Region,
			MMR:       session.PlayerMMR,
			EventType: kafka.EventTypeQueueLeft,
			Metadata: map[string]string{
				"session_id": cmd.SessionID.String(),
				"game_mode": session.Preferences.GameMode,
			},
		}

		if err := uc.eventPublisher.PublishQueueEvent(ctx, queueEvent); err != nil {
			slog.WarnContext(ctx, "failed to publish queue left event", "error", err, "player_id", cmd.PlayerID)
		}
	}

	// billing execution AFTER successful operation
	_, _, err = uc.billableOperationHandler.Exec(ctx, billingCmd)
	if err != nil {
		slog.WarnContext(ctx, "failed to execute billing for leave matchmaking queue", "error", err, "session_id", cmd.SessionID)
	}

	slog.InfoContext(ctx, "player left matchmaking queue",
		"session_id", cmd.SessionID,
		"player_id", cmd.PlayerID,
	)

	return nil
}
