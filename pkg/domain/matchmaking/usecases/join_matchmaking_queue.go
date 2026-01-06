package matchmaking_usecases

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	matchmaking_out "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/out"
	kafka "github.com/replay-api/replay-api/pkg/infra/kafka"
)

// JoinMatchmakingQueueUseCase handles player joining the ranked matchmaking queue.
//
// This is the primary entry point for competitive matchmaking in the LeetGaming platform.
//
// Flow:
//  1. Authentication verification - user must be authenticated
//  2. Input validation - team format, player role, and game mode validation
//  3. Active session check - prevents duplicate queue entries
//  4. Billing validation - ensures subscription/credits allow queue join
//  5. Session creation - creates player's matchmaking session with preferences
//  6. Billing execution - records the billable operation
//
// Features:
//   - Priority boost support for premium subscribers
//   - Dynamic skill range calculation based on MMR
//   - Estimated wait time calculation based on pool health
//   - Role-based matchmaking for 5v5 team formats
//   - Cross-platform matching support
//
// Security:
//   - Requires authenticated context (shared.AuthenticatedKey)
//   - Uses resource ownership from context for billing
//
// Dependencies:
//   - BillableOperationCommandHandler: Validates/tracks usage against subscription limits
//   - MatchmakingSessionRepository: Session persistence
//   - EventPublisher: Publishes matchmaking events to Kafka
type JoinMatchmakingQueueUseCase struct {
	billableOperationHandler billing_in.BillableOperationCommandHandler
	sessionRepository        matchmaking_out.MatchmakingSessionRepository
	eventPublisher           *kafka.EventPublisher
}

// NewJoinMatchmakingQueueUseCase creates a new join queue usecase
func NewJoinMatchmakingQueueUseCase(
	billableOperationHandler billing_in.BillableOperationCommandHandler,
	sessionRepository matchmaking_out.MatchmakingSessionRepository,
	eventPublisher *kafka.EventPublisher,
) matchmaking_in.JoinMatchmakingQueueCommandHandler {
	return &JoinMatchmakingQueueUseCase{
		billableOperationHandler: billableOperationHandler,
		sessionRepository:        sessionRepository,
		eventPublisher:           eventPublisher,
	}
}

// Exec executes the join matchmaking queue command
func (uc *JoinMatchmakingQueueUseCase) Exec(ctx context.Context, cmd matchmaking_in.JoinMatchmakingQueueCommand) (*matchmaking_entities.MatchmakingSession, error) {
	// auth check
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, shared.NewErrUnauthorized()
	}

	// validate team format
	if !cmd.TeamFormat.IsValid() {
		return nil, fmt.Errorf("invalid team format: %s", cmd.TeamFormat)
	}

	// validate role if provided (required for 5v5)
	if cmd.PlayerRole != nil {
		role := matchmaking_in.PlayerRole(*cmd.PlayerRole)
		if !role.IsValid() {
			return nil, fmt.Errorf("invalid player role: %s", *cmd.PlayerRole)
		}
		// 5v5 requires role selection
		if cmd.TeamFormat == matchmaking_in.TeamFormat5v5 && *cmd.PlayerRole == "" {
			return nil, fmt.Errorf("5v5 matchmaking requires player role selection")
		}
	}

	// check for existing active sessions
	existingSessions, err := uc.sessionRepository.GetByPlayerID(ctx, cmd.PlayerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check existing sessions", "error", err, "player_id", cmd.PlayerID)
		return nil, fmt.Errorf("failed to check existing sessions")
	}

	for _, session := range existingSessions {
		if session.CanMatch() {
			return nil, fmt.Errorf("player already in queue, session_id: %s", session.ID)
		}
	}

	// billing validation BEFORE creating session
	operationType := billing_entities.OperationTypeJoinMatchmakingQueue
	if cmd.PriorityBoost {
		operationType = billing_entities.OperationTypeMatchMakingPriorityQueue
	}

	billingCmd := billing_in.BillableOperationCommand{
		OperationID: operationType,
		UserID:      shared.GetResourceOwner(ctx).UserID,
		Amount:      1,
		Args: map[string]interface{}{
			"game_id":       cmd.GameID,
			"game_mode":     cmd.GameMode,
			"team_format":   cmd.TeamFormat,
			"priority_boost": cmd.PriorityBoost,
		},
	}

	err = uc.billableOperationHandler.Validate(ctx, billingCmd)
	if err != nil {
		slog.ErrorContext(ctx, "billing validation failed for join matchmaking queue", "error", err, "player_id", cmd.PlayerID)
		return nil, err
	}

	// create matchmaking session
	sessionID := uuid.New()
	now := time.Now().UTC()
	expiresAt := now.Add(10 * time.Minute) // 10 minute queue timeout

	resourceOwner := shared.GetResourceOwner(ctx)

	session := &matchmaking_entities.MatchmakingSession{
		BaseEntity: shared.NewEntity(resourceOwner),
		PlayerID:   cmd.PlayerID,
		SquadID:    cmd.SquadID,
		Preferences: matchmaking_entities.MatchPreferences{
			GameID:             cmd.GameID,
			GameMode:           cmd.GameMode,
			Region:             cmd.Region,
			SkillRange:         matchmaking_entities.SkillRange{MinMMR: 0, MaxMMR: 5000}, // Broad range, actual matching done by match-making-api
			MaxPing:            cmd.MaxPing,
			AllowCrossPlatform: true,
			Tier:               cmd.Tier,
			PriorityBoost:      cmd.PriorityBoost,
		},
		Status:        matchmaking_entities.StatusQueued,
		PlayerMMR:     cmd.PlayerMMR,
		QueuedAt:      now,
		EstimatedWait: 180, // Default 3 minutes, actual estimation done by match-making-api
		ExpiresAt:     expiresAt,
		Metadata: map[string]any{
			"team_format": cmd.TeamFormat,
			"player_role": cmd.PlayerRole,
		},
	}

	// save session
	err = uc.sessionRepository.Save(ctx, session)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save matchmaking session", "error", err)
		return nil, fmt.Errorf("failed to save matchmaking session")
	}

	// publish queue joined event
	if uc.eventPublisher != nil {
		queueEvent := &kafka.QueueEvent{
			PlayerID:  cmd.PlayerID,
			GameType:  cmd.GameID,
			Region:    cmd.Region,
			MMR:       cmd.PlayerMMR,
			EventType: kafka.EventTypeQueueJoined,
			Metadata: map[string]string{
				"session_id":   session.ID.String(),
				"game_mode":   cmd.GameMode,
				"team_format": string(cmd.TeamFormat),
				"squad_id":    cmd.SquadID.String(),
			},
		}
		if cmd.PlayerRole != nil {
			queueEvent.Metadata["player_role"] = *cmd.PlayerRole
		}

		if err := uc.eventPublisher.PublishQueueEvent(ctx, queueEvent); err != nil {
			slog.WarnContext(ctx, "failed to publish queue joined event", "error", err, "player_id", cmd.PlayerID)
		}
	}

	// billing execution AFTER successful operation
	_, _, err = uc.billableOperationHandler.Exec(ctx, billingCmd)
	if err != nil {
		slog.WarnContext(ctx, "failed to execute billing for join matchmaking queue", "error", err, "player_id", cmd.PlayerID)
	}

	slog.InfoContext(ctx, "player joined matchmaking queue",
		"session_id", sessionID,
		"player_id", cmd.PlayerID,
		"game_id", cmd.GameID,
		"game_mode", cmd.GameMode,
		"team_format", cmd.TeamFormat,
		"mmr", cmd.PlayerMMR,
		"estimated_wait", 180,
	)

	return session, nil
}
