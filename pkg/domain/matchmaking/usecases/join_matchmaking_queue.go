package matchmaking_usecases

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	matchmaking_out "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/out"
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
//  5. Pool management - creates/retrieves matchmaking pool for region/mode
//  6. Session creation - creates player's matchmaking session with preferences
//  7. Billing execution - records the billable operation
//
// Features:
//   - Priority boost support for premium subscribers
//   - Dynamic skill range calculation based on MMR
//   - Estimated wait time calculation based on pool health
//   - Role-based matchmaking for 5v5 team formats
//   - Cross-platform matching support
//
// Security:
//   - Requires authenticated context (common.AuthenticatedKey)
//   - Uses resource ownership from context for billing
//
// Dependencies:
//   - BillableOperationCommandHandler: Validates/tracks usage against subscription limits
//   - MatchmakingSessionRepository: Session persistence
//   - MatchmakingPoolRepository: Pool management for game mode/region
type JoinMatchmakingQueueUseCase struct {
	billableOperationHandler billing_in.BillableOperationCommandHandler
	sessionRepository        matchmaking_out.MatchmakingSessionRepository
	poolRepository           matchmaking_out.MatchmakingPoolRepository
}

// NewJoinMatchmakingQueueUseCase creates a new join queue usecase
func NewJoinMatchmakingQueueUseCase(
	billableOperationHandler billing_in.BillableOperationCommandHandler,
	sessionRepository matchmaking_out.MatchmakingSessionRepository,
	poolRepository matchmaking_out.MatchmakingPoolRepository,
) matchmaking_in.JoinMatchmakingQueueCommandHandler {
	return &JoinMatchmakingQueueUseCase{
		billableOperationHandler: billableOperationHandler,
		sessionRepository:        sessionRepository,
		poolRepository:           poolRepository,
	}
}

// Exec executes the join matchmaking queue command
func (uc *JoinMatchmakingQueueUseCase) Exec(ctx context.Context, cmd matchmaking_in.JoinMatchmakingQueueCommand) (*matchmaking_entities.MatchmakingSession, error) {
	// auth check
	isAuthenticated := ctx.Value(common.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, common.NewErrUnauthorized()
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
		UserID:      common.GetResourceOwner(ctx).UserID,
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

	// get or create pool
	pool, err := uc.poolRepository.GetByGameModeRegion(ctx, cmd.GameID, cmd.GameMode, cmd.Region)
	if err != nil {
		// create new pool if doesn't exist
		pool = &matchmaking_entities.MatchmakingPool{
			ID:             uuid.New(),
			GameID:         cmd.GameID,
			GameMode:       cmd.GameMode,
			Region:         cmd.Region,
			ActiveSessions: []uuid.UUID{},
			PoolStats: matchmaking_entities.PoolStatistics{
				TotalPlayers:    0,
				AverageWaitTime: 0,
				PlayersByTier:   make(map[matchmaking_entities.MatchmakingTier]int),
				PlayersBySkill:  make(map[string]int),
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		err = uc.poolRepository.Save(ctx, pool)
		if err != nil {
			slog.ErrorContext(ctx, "failed to create pool", "error", err)
			return nil, fmt.Errorf("failed to create matchmaking pool")
		}
	}

	// create matchmaking session
	sessionID := uuid.New()
	now := time.Now().UTC()
	expiresAt := now.Add(10 * time.Minute) // 10 minute queue timeout

	// calculate estimated wait based on pool health
	estimatedWait := uc.calculateEstimatedWait(pool, cmd.TeamFormat)

	session := &matchmaking_entities.MatchmakingSession{
		ID:       sessionID,
		PlayerID: cmd.PlayerID,
		SquadID:  cmd.SquadID,
		Preferences: matchmaking_entities.MatchPreferences{
			GameID:             cmd.GameID,
			GameMode:           cmd.GameMode,
			Region:             cmd.Region,
			SkillRange:         uc.calculateSkillRange(cmd.PlayerMMR),
			MaxPing:            cmd.MaxPing,
			AllowCrossPlatform: true,
			Tier:               cmd.Tier,
			PriorityBoost:      cmd.PriorityBoost,
		},
		Status:        matchmaking_entities.StatusQueued,
		PlayerMMR:     cmd.PlayerMMR,
		QueuedAt:      now,
		EstimatedWait: estimatedWait,
		ExpiresAt:     expiresAt,
		Metadata: map[string]any{
			"team_format": cmd.TeamFormat,
			"player_role": cmd.PlayerRole,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	// save session
	err = uc.sessionRepository.Save(ctx, session)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save matchmaking session", "error", err)
		return nil, fmt.Errorf("failed to save matchmaking session")
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
		"estimated_wait", estimatedWait,
	)

	return session, nil
}

// calculateSkillRange determines acceptable MMR range based on player rating
func (uc *JoinMatchmakingQueueUseCase) calculateSkillRange(playerMMR int) matchmaking_entities.SkillRange {
	// tighter range for high MMR, wider for low MMR
	rangeFactor := 200
	if playerMMR > 2000 {
		rangeFactor = 150
	} else if playerMMR < 1000 {
		rangeFactor = 300
	}

	return matchmaking_entities.SkillRange{
		MinMMR: playerMMR - rangeFactor,
		MaxMMR: playerMMR + rangeFactor,
	}
}

// calculateEstimatedWait estimates queue time based on pool stats
func (uc *JoinMatchmakingQueueUseCase) calculateEstimatedWait(pool *matchmaking_entities.MatchmakingPool, format matchmaking_in.TeamFormat) int {
	if pool.PoolStats.AverageWaitTime > 0 {
		return pool.PoolStats.AverageWaitTime
	}

	// base estimate on pool size and format
	playersNeeded := format.GetTotalPlayers()
	if pool.PoolStats.TotalPlayers >= playersNeeded {
		return 30 // 30 seconds if enough players
	} else if pool.PoolStats.TotalPlayers >= playersNeeded/2 {
		return 90 // 90 seconds if halfway there
	}

	return 180 // 3 minutes default
}
