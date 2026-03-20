package matchmaking_usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	matchmaking_out "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/out"
	kafka "github.com/replay-api/replay-api/pkg/infra/kafka"
	matchmaking_vo "github.com/replay-api/replay-api/pkg/domain/matchmaking/value-objects"
	ws "github.com/replay-api/replay-api/pkg/infra/websocket"
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
//   - WebSocketHub: Notifies matched players in real-time
//   - LobbyRepository: Creates lobbies for matched players
type JoinMatchmakingQueueUseCase struct {
	billableOperationHandler billing_in.BillableOperationCommandHandler
	sessionRepository        matchmaking_out.MatchmakingSessionRepository
	lobbyRepository          matchmaking_out.LobbyRepository
	eventPublisher           *kafka.EventPublisher
	wsHub                    *ws.WebSocketHub
}

// NewJoinMatchmakingQueueUseCase creates a new join queue usecase
func NewJoinMatchmakingQueueUseCase(
	billableOperationHandler billing_in.BillableOperationCommandHandler,
	sessionRepository matchmaking_out.MatchmakingSessionRepository,
	lobbyRepository matchmaking_out.LobbyRepository,
	eventPublisher *kafka.EventPublisher,
	wsHub *ws.WebSocketHub,
) matchmaking_in.JoinMatchmakingQueueCommandHandler {
	return &JoinMatchmakingQueueUseCase{
		billableOperationHandler: billableOperationHandler,
		sessionRepository:        sessionRepository,
		lobbyRepository:          lobbyRepository,
		eventPublisher:           eventPublisher,
		wsHub:                    wsHub,
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

	// Dynamic queue estimation: check how many compatible players are already queued
	estimatedWait := uc.estimateWaitTime(ctx, cmd)

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
		EstimatedWait: estimatedWait,
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
				"session_id":  session.ID.String(),
				"game_mode":  cmd.GameMode,
				"team_format": string(cmd.TeamFormat),
			},
		}
		if cmd.SquadID != nil {
			queueEvent.Metadata["squad_id"] = cmd.SquadID.String()
		}
		if cmd.PlayerRole != nil {
			queueEvent.Metadata["player_role"] = *cmd.PlayerRole
		}

		if err := uc.eventPublisher.PublishQueueEvent(ctx, queueEvent); err != nil {
			slog.WarnContext(ctx, "failed to publish queue joined event", "error", err, "player_id", cmd.PlayerID)
		}
	}

	// attempt immediate match: find a compatible queued session
	// Build a background context with resource owner info (the HTTP request context gets cancelled)
	matchCtx := context.Background()
	ro := shared.GetResourceOwner(ctx)
	matchCtx = context.WithValue(matchCtx, shared.TenantIDKey, ro.TenantID)
	matchCtx = context.WithValue(matchCtx, shared.ClientIDKey, ro.ClientID)
	matchCtx = context.WithValue(matchCtx, shared.GroupIDKey, ro.GroupID)
	matchCtx = context.WithValue(matchCtx, shared.UserIDKey, ro.UserID)
	go uc.tryMatch(matchCtx, session)

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

// tryMatch attempts to find a compatible opponent for the given session.
// Runs asynchronously after queue join to avoid blocking the HTTP response.
// Matching criteria: same game, same game mode, same region, MMR within 500 range.
func (uc *JoinMatchmakingQueueUseCase) tryMatch(ctx context.Context, newSession *matchmaking_entities.MatchmakingSession) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("panic in tryMatch", "recover", r)
		}
	}()

	status := matchmaking_entities.StatusQueued
	candidates, err := uc.sessionRepository.GetActiveSessions(ctx, matchmaking_out.SessionFilters{
		GameID:   newSession.Preferences.GameID,
		GameMode: newSession.Preferences.GameMode,
		Region:   newSession.Preferences.Region,
		Status:   &status,
		Limit:    50,
	})
	if err != nil {
		slog.Error("tryMatch: failed to get active sessions", "error", err)
		return
	}

	const maxMMRDelta = 500
	teamFormat := getSessionTeamFormat(newSession)
	requiredPlayers := teamFormat.GetTotalPlayers()
	selectedSessions := []*matchmaking_entities.MatchmakingSession{newSession}
	compatibleCandidates := make([]*matchmaking_entities.MatchmakingSession, 0, len(candidates))

	for _, candidate := range candidates {
		if candidate.ID == newSession.ID {
			continue // skip self
		}
		if candidate.PlayerID == newSession.PlayerID {
			continue
		}
		if getSessionTeamFormat(candidate) != teamFormat {
			continue
		}
		delta := int(math.Abs(float64(candidate.PlayerMMR - newSession.PlayerMMR)))
		if delta <= maxMMRDelta {
			compatibleCandidates = append(compatibleCandidates, candidate)
		}
	}

	sort.SliceStable(compatibleCandidates, func(i, j int) bool {
		leftDelta := math.Abs(float64(compatibleCandidates[i].PlayerMMR - newSession.PlayerMMR))
		rightDelta := math.Abs(float64(compatibleCandidates[j].PlayerMMR - newSession.PlayerMMR))
		if leftDelta == rightDelta {
			return compatibleCandidates[i].QueuedAt.Before(compatibleCandidates[j].QueuedAt)
		}
		return leftDelta < rightDelta
	})

	for _, candidate := range compatibleCandidates {
		if len(selectedSessions) >= requiredPlayers {
			break
		}
		selectedSessions = append(selectedSessions, candidate)
	}

	if len(selectedSessions) < requiredPlayers {
		slog.Info("tryMatch: insufficient compatible players for team format",
			"session_id", newSession.ID,
			"player_id", newSession.PlayerID,
			"team_format", teamFormat,
			"required_players", requiredPlayers,
			"compatible_candidates", len(compatibleCandidates)+1)
		return
	}

	playerIDs := make([]uuid.UUID, 0, len(selectedSessions))
	readyCheckPlayers := make([]string, 0, len(selectedSessions))
	playerPayloads := make([]map[string]interface{}, 0, len(selectedSessions))
	playerSlots := make([]matchmaking_entities.PlayerSlot, 0, len(selectedSessions))
	avgMMR := 0
	now := time.Now().UTC()

	for index, session := range selectedSessions {
		displayName := fmt.Sprintf("Player %d", index+1)
		playerIDs = append(playerIDs, session.PlayerID)
		readyCheckPlayers = append(readyCheckPlayers, session.PlayerID.String())
		playerPayloads = append(playerPayloads, map[string]interface{}{
			"player_id":    session.PlayerID.String(),
			"display_name": displayName,
			"status":       "pending",
			"mmr":          session.PlayerMMR,
			"slot":         index + 1,
		})
		playerSlots = append(playerSlots, matchmaking_entities.PlayerSlot{
			SlotNumber: index + 1,
			PlayerID:   &session.PlayerID,
			IsReady:    false,
			JoinedAt:   now,
			MMR:        &session.PlayerMMR,
		})
		avgMMR += session.PlayerMMR
	}

	avgMMR = avgMMR / len(selectedSessions)

	slog.Info("tryMatch: match group assembled",
		"team_format", teamFormat,
		"required_players", requiredPlayers,
		"selected_players", len(selectedSessions),
		"lobby_creator", newSession.PlayerID)

	// Create a lobby for the matched players
	resourceOwner := shared.GetResourceOwner(ctx)

	lobby := &matchmaking_entities.MatchmakingLobby{
		BaseEntity:       shared.NewEntity(resourceOwner),
		CreatorID:        newSession.PlayerID,
		GameID:           newSession.Preferences.GameID,
		Region:           newSession.Preferences.Region,
		Tier:             string(newSession.Preferences.Tier),
		DistributionRule: matchmaking_vo.DistributionRuleWinnerTakesAll,
		MaxPlayers:       requiredPlayers,
		PlayerSlots:      playerSlots,
		Status:           matchmaking_entities.LobbyStatusReadyCheck,
		Metadata: map[string]any{
			"team_format":                 string(teamFormat),
			"ready_check_players":         playerPayloads,
			"ready_check_expected_players": requiredPlayers,
			"declined_players":            []string{},
		},
		AutoFill:         false,
		InviteOnly:       false,
		ReadyTimeout:     30 * time.Second,
	}

	endTime := now.Add(lobby.ReadyTimeout)
	lobby.ReadyCheckStart = &now
	lobby.ReadyCheckEnd = &endTime

	if err := uc.lobbyRepository.Save(ctx, lobby); err != nil {
		slog.Error("tryMatch: failed to save lobby", "error", err, "lobby_id", lobby.ID)
		return
	}

	// Atomically claim both sessions with CAS to prevent double-booking
	claimExtras := map[string]interface{}{
		"match_id":                             lobby.ID,
		"matched_at":                           now,
		"metadata.lobby_id":                    lobby.ID.String(),
		"metadata.ready_check_started_at":      now.Format(time.RFC3339),
		"metadata.ready_check_players":         playerPayloads,
		"metadata.team_format":                 string(teamFormat),
		"metadata.ready_check_expected_players": requiredPlayers,
	}

	claimedSessions := make([]*matchmaking_entities.MatchmakingSession, 0, len(selectedSessions))
	for _, session := range selectedSessions {
		ok, claimErr := uc.sessionRepository.CompareAndSetStatus(ctx, session.ID,
			matchmaking_entities.StatusQueued, matchmaking_entities.StatusReadyCheck, claimExtras)
		if claimErr != nil || !ok {
			slog.Warn("tryMatch: failed to claim session for ready check",
				"session_id", session.ID,
				"player_id", session.PlayerID,
				"error", claimErr)
			for _, claimed := range claimedSessions {
				_, _ = uc.sessionRepository.CompareAndSetStatus(ctx, claimed.ID,
					matchmaking_entities.StatusReadyCheck, matchmaking_entities.StatusQueued,
					map[string]interface{}{"match_id": nil, "matched_at": nil, "metadata": nil})
			}
			_ = uc.lobbyRepository.Delete(ctx, lobby.ID)
			return
		}
		claimedSessions = append(claimedSessions, session)
	}

	slog.Info("tryMatch: lobby created, sessions updated to ready_check",
		"lobby_id", lobby.ID,
		"team_format", teamFormat,
		"player_count", len(playerIDs))

	// Notify matched players via WebSocket
	if uc.wsHub != nil {
		matchPayload := map[string]interface{}{
			"lobby_id":              lobby.ID.String(),
			"match_type":            fmt.Sprintf("ranked_%s", teamFormat),
			"game_id":               newSession.Preferences.GameID,
			"region":                newSession.Preferences.Region,
			"team_format":           string(teamFormat),
			"players":               playerPayloads,
			"ready_timeout_seconds": 30,
		}

		payloadBytes, _ := json.Marshal(matchPayload)

		for _, playerID := range playerIDs {
			uc.wsHub.BroadcastToUser(playerID, "match_found", payloadBytes)
		}
	}

	// Publish match-found Kafka event
	if uc.eventPublisher != nil {
		lobbyEvent := &kafka.LobbyEvent{
			LobbyID:   lobby.ID,
			EventType: kafka.EventTypeLobbyCreated,
			PlayerIDs: playerIDs,
			GameType:  newSession.Preferences.GameID,
			Region:    newSession.Preferences.Region,
			AvgMMR:    avgMMR,
		}
		if err := uc.eventPublisher.PublishLobbyEvent(ctx, lobbyEvent); err != nil {
			slog.Warn("tryMatch: failed to publish lobby event", "error", err)
		}
	}
}

// estimateWaitTime computes a dynamic wait estimate based on the number of
// compatible queued players.  When at least one other player is already
// queued the estimate is very short (instant match likely); otherwise it
// falls back to a conservative default.
func (uc *JoinMatchmakingQueueUseCase) estimateWaitTime(ctx context.Context, cmd matchmaking_in.JoinMatchmakingQueueCommand) int {
	const defaultWait = 120 // 2 minutes if nobody else is queued
	const fastWait = 10     // 10 seconds when match is likely
	const moderateWait = 45

	status := matchmaking_entities.StatusQueued
	candidates, err := uc.sessionRepository.GetActiveSessions(ctx, matchmaking_out.SessionFilters{
		GameID:   cmd.GameID,
		GameMode: cmd.GameMode,
		Region:   cmd.Region,
		Status:   &status,
		Limit:    10,
	})
	if err != nil {
		slog.WarnContext(ctx, "estimateWaitTime: query failed, using default", "error", err)
		return defaultWait
	}

	// Exclude self (the session isn't saved yet but there could be stale sessions for this player)
	count := 0
	for _, c := range candidates {
		if c.PlayerID != cmd.PlayerID {
			count++
		}
	}

	requiredOpponents := cmd.TeamFormat.GetTotalPlayers() - 1
	halfway := requiredOpponents / 2

	if count >= requiredOpponents {
		return fastWait
	}
	if halfway > 0 && count >= halfway {
		return moderateWait
	}
	return defaultWait
}

func getSessionTeamFormat(session *matchmaking_entities.MatchmakingSession) matchmaking_in.TeamFormat {
	if session == nil || session.Metadata == nil {
		return matchmaking_in.TeamFormat1v1
	}

	value, ok := session.Metadata["team_format"]
	if !ok {
		return matchmaking_in.TeamFormat1v1
	}

	switch typed := value.(type) {
	case matchmaking_in.TeamFormat:
		if typed.IsValid() {
			return typed
		}
	case string:
		format := matchmaking_in.TeamFormat(typed)
		if format.IsValid() {
			return format
		}
	}

	return matchmaking_in.TeamFormat1v1
}
