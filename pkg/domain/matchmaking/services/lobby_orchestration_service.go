package matchmaking_services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	matchmaking_out "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/out"
	wallet_in "github.com/replay-api/replay-api/pkg/domain/wallet/ports/in"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	ws "github.com/replay-api/replay-api/pkg/infra/websocket"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// LobbyOrchestrationService coordinates lobby → prize pool → wallet operations with Saga pattern
type LobbyOrchestrationService struct {
	lobbyRepo     matchmaking_out.LobbyRepository
	prizePoolRepo matchmaking_out.PrizePoolRepository
	sessionRepo   matchmaking_out.MatchmakingSessionRepository
	walletCommand wallet_in.WalletCommand
	wsHub         *ws.WebSocketHub
}

func NewLobbyOrchestrationService(
lobbyRepo matchmaking_out.LobbyRepository,
prizePoolRepo matchmaking_out.PrizePoolRepository,
sessionRepo matchmaking_out.MatchmakingSessionRepository,
walletCommand wallet_in.WalletCommand,
wsHub *ws.WebSocketHub,
) matchmaking_in.LobbyCommand {
	return &LobbyOrchestrationService{
		lobbyRepo:     lobbyRepo,
		prizePoolRepo: prizePoolRepo,
		sessionRepo:   sessionRepo,
		walletCommand: walletCommand,
		wsHub:         wsHub,
	}
}

func (s *LobbyOrchestrationService) CreateLobby(ctx context.Context, cmd matchmaking_in.CreateLobbyCommand) (*matchmaking_entities.MatchmakingLobby, error) {
	resourceOwner := shared.GetResourceOwner(ctx)

	lobby, err := matchmaking_entities.NewMatchmakingLobby(
resourceOwner,
cmd.CreatorID,
cmd.GameID,
cmd.Region,
cmd.Tier,
cmd.DistributionRule,
cmd.MaxPlayers,
cmd.AutoFill,
cmd.InviteOnly,
)
	if err != nil {
		return nil, fmt.Errorf("failed to create lobby: %w", err)
	}

	if err := s.lobbyRepo.Save(ctx, lobby); err != nil {
		return nil, fmt.Errorf("failed to save lobby: %w", err)
	}

	// Create associated prize pool
	prizePoolCmd := matchmaking_in.CreatePrizePoolCommand{
		MatchID:              lobby.ID,
		GameID:               cmd.GameID,
		Region:               cmd.Region,
		Currency:             "USD",
		PlatformContribution: 0.50,
		DistributionRule:     cmd.DistributionRule,
	}

	_, err = s.createPrizePool(ctx, prizePoolCmd)
	if err != nil {
		// Rollback: delete lobby
		_ = s.lobbyRepo.Delete(ctx, lobby.ID)
		return nil, fmt.Errorf("failed to create prize pool: %w", err)
	}

	slog.InfoContext(ctx, "Lobby created", "lobby_id", lobby.ID, "creator", cmd.CreatorID)
	s.wsHub.BroadcastLobbyUpdate(lobby.ID, lobby)

	return lobby, nil
}

// JoinLobby implements Saga pattern: Deduct entry fee → Add to lobby → Add to prize pool (rollback on failure)
func (s *LobbyOrchestrationService) JoinLobby(ctx context.Context, cmd matchmaking_in.JoinLobbyCommand) error {
	// Step 1: Load lobby
	lobby, err := s.lobbyRepo.FindByID(ctx, cmd.LobbyID)
	if err != nil {
		return fmt.Errorf("lobby not found: %w", err)
	}

	// Step 2: Load prize pool
	prizePool, err := s.prizePoolRepo.FindByMatchID(ctx, cmd.LobbyID)
	if err != nil {
		return fmt.Errorf("prize pool not found: %w", err)
	}

	// Step 3: Calculate entry fee by tier
	entryFee := getEntryFeeByTier(lobby.Tier)

	// Step 4: Deduct entry fee from wallet (with rollback support) — skip for free tier
	if entryFee > 0 {
		walletCmd := wallet_in.DeductEntryFeeCommand{
			UserID:   cmd.PlayerID,
			Currency: string(prizePool.Currency),
			Amount:   entryFee,
		}

		if err := s.walletCommand.DeductEntryFee(ctx, walletCmd); err != nil {
			return fmt.Errorf("insufficient balance: %w", err)
		}
	}

	// Step 5: Add player to lobby
	if err := lobby.AddPlayer(cmd.PlayerID, cmd.MMR); err != nil {
		// Rollback: refund entry fee
		if entryFee > 0 {
			refundCmd := wallet_in.RefundCommand{
				UserID:   cmd.PlayerID,
				Currency: string(prizePool.Currency),
				Amount:   entryFee,
				Reason:   "failed to join lobby",
			}
			_ = s.walletCommand.Refund(ctx, refundCmd)
		}
		return fmt.Errorf("failed to add player: %w", err)
	}

	// Step 6: Add contribution to prize pool (skip for free tier)
	if entryFee > 0 {
		prizePoolAmount := wallet_vo.NewAmount(entryFee)

		if err := prizePool.AddPlayerContribution(cmd.PlayerID, prizePoolAmount); err != nil {
			// Rollback: remove from lobby + refund
			_ = lobby.RemovePlayer(cmd.PlayerID)
			_ = s.lobbyRepo.Update(ctx, lobby)
			refundCmd := wallet_in.RefundCommand{
				UserID:   cmd.PlayerID,
				Currency: string(prizePool.Currency),
				Amount:   entryFee,
				Reason:   "failed to add prize contribution",
			}
			_ = s.walletCommand.Refund(ctx, refundCmd)
			return fmt.Errorf("failed to add prize contribution: %w", err)
		}
	}

	// Step 7: Persist changes
	if err := s.lobbyRepo.Update(ctx, lobby); err != nil {
		return fmt.Errorf("failed to update lobby: %w", err)
	}
	if entryFee > 0 {
		if err := s.prizePoolRepo.Update(ctx, prizePool); err != nil {
			return fmt.Errorf("failed to update prize pool: %w", err)
		}
	}

	slog.InfoContext(ctx, "Player joined lobby", "lobby_id", cmd.LobbyID, "player_id", cmd.PlayerID)

	// Step 8: Broadcast updates via WebSocket
	s.wsHub.BroadcastLobbyUpdate(cmd.LobbyID, lobby)
	s.wsHub.BroadcastPrizePoolUpdate(cmd.LobbyID, prizePool)

	return nil
}

func (s *LobbyOrchestrationService) LeaveLobby(ctx context.Context, cmd matchmaking_in.LeaveLobbyCommand) error {
	lobby, err := s.lobbyRepo.FindByID(ctx, cmd.LobbyID)
	if err != nil {
		return fmt.Errorf("lobby not found: %w", err)
	}

	prizePool, err := s.prizePoolRepo.FindByMatchID(ctx, cmd.LobbyID)
	if err != nil {
		return fmt.Errorf("prize pool not found: %w", err)
	}

	// Get player contribution before removal
	var playerContribution float64
	if contribution, exists := prizePool.PlayerContributions[cmd.PlayerID]; exists {
		playerContribution = contribution.ToFloat()
	}

	if err := lobby.RemovePlayer(cmd.PlayerID); err != nil {
		return fmt.Errorf("failed to remove player: %w", err)
	}

	// Refund entry fee
	if playerContribution > 0 {
		refundCmd := wallet_in.RefundCommand{
			UserID:   cmd.PlayerID,
			Currency: string(prizePool.Currency),
			Amount:   playerContribution,
			Reason:   "left lobby",
		}
		if err := s.walletCommand.Refund(ctx, refundCmd); err != nil {
			slog.ErrorContext(ctx, "Failed to refund player", "error", err)
		}
	}

	if err := s.lobbyRepo.Update(ctx, lobby); err != nil {
		return fmt.Errorf("failed to update lobby: %w", err)
	}

	slog.InfoContext(ctx, "Player left lobby", "lobby_id", cmd.LobbyID, "player_id", cmd.PlayerID)
	s.wsHub.BroadcastLobbyUpdate(cmd.LobbyID, lobby)

	return nil
}

func (s *LobbyOrchestrationService) SetPlayerReady(ctx context.Context, cmd matchmaking_in.SetPlayerReadyCommand) error {
	// Atomically set the player's ready flag — avoids lost-update race
	lobby, err := s.lobbyRepo.SetPlayerReadyAtomic(ctx, cmd.LobbyID, cmd.PlayerID, cmd.IsReady)
	if err != nil {
		return fmt.Errorf("failed to set ready: %w", err)
	}

	readyCount := 0
	totalCount := 0
	for _, slot := range lobby.PlayerSlots {
		if slot.PlayerID == nil {
			continue
		}
		totalCount++
		if slot.IsReady {
			readyCount++
		}
	}

	slog.InfoContext(ctx, "Player ready status changed", "lobby_id", cmd.LobbyID, "player_id", cmd.PlayerID, "is_ready", cmd.IsReady)
	s.wsHub.BroadcastLobbyUpdate(cmd.LobbyID, lobby)
	s.broadcastReadinessUpdate(cmd.LobbyID, cmd.PlayerID, cmd.IsReady, readyCount, totalCount)

	// Auto-transition: if lobby is open and full, move to ready_check
	if cmd.IsReady && lobby.Status == matchmaking_entities.LobbyStatusOpen && totalCount >= lobby.MaxPlayers {
		if err := lobby.StartReadyCheck(); err != nil {
			slog.WarnContext(ctx, "Auto ready-check transition failed (non-fatal)", "lobby_id", cmd.LobbyID, "error", err)
		} else {
			if err := s.lobbyRepo.Update(ctx, lobby); err != nil {
				slog.ErrorContext(ctx, "Failed to persist ready-check transition", "lobby_id", cmd.LobbyID, "error", err)
			} else {
				slog.InfoContext(ctx, "Auto-transitioned to ready_check", "lobby_id", cmd.LobbyID)
				s.wsHub.BroadcastLobbyUpdate(cmd.LobbyID, lobby)
			}
		}
	}

	// Auto-transition: if all players are ready and lobby is in ready_check, start the match
	if cmd.IsReady && lobby.Status == matchmaking_entities.LobbyStatusReadyCheck {
		allReady, _ := lobby.CheckReadyStatus()
		if allReady {
			slog.InfoContext(ctx, "All players ready \u2014 auto-starting match", "lobby_id", cmd.LobbyID)
			matchID, err := s.startMatchForLobby(ctx, lobby)
			if err != nil {
				slog.ErrorContext(ctx, "Auto-start match failed (non-fatal)", "lobby_id", cmd.LobbyID, "error", err)
			} else {
				slog.InfoContext(ctx, "Match auto-started", "lobby_id", cmd.LobbyID, "match_id", matchID)
			}
		}
	}

	return nil
}

// startMatchForLobby transitions a lobby to started state using atomic CAS transitions
func (s *LobbyOrchestrationService) startMatchForLobby(ctx context.Context, lobby *matchmaking_entities.MatchmakingLobby) (uuid.UUID, error) {
	matchID := uuid.New()

	// Atomic CAS: ready_check → starting (prevents timeout worker from overwriting)
	ok, err := s.lobbyRepo.TransitionStatus(ctx, lobby.ID,
		matchmaking_entities.LobbyStatusReadyCheck,
		matchmaking_entities.LobbyStatusStarting,
		map[string]interface{}{"match_id": matchID},
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to transition to starting: %w", err)
	}
	if !ok {
		return uuid.Nil, fmt.Errorf("lobby %s is no longer in ready_check (concurrent transition)", lobby.ID)
	}

	// Prize pool lock is best-effort: matchmaking-created lobbies may not have one
	prizePool, ppErr := s.prizePoolRepo.FindByMatchID(ctx, lobby.ID)
	if ppErr == nil && prizePool != nil {
		if lockErr := prizePool.Lock(); lockErr != nil {
			slog.WarnContext(ctx, "Failed to lock prize pool (non-fatal)", "lobby_id", lobby.ID, "error", lockErr)
		} else {
			_ = s.prizePoolRepo.Update(ctx, prizePool)
		}
	} else {
		slog.InfoContext(ctx, "No prize pool for lobby (matchmaking lobby)", "lobby_id", lobby.ID)
	}

	// Atomic CAS: starting → started
	ok, err = s.lobbyRepo.TransitionStatus(ctx, lobby.ID,
		matchmaking_entities.LobbyStatusStarting,
		matchmaking_entities.LobbyStatusStarted,
		nil,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to transition to started: %w", err)
	}
	if !ok {
		slog.WarnContext(ctx, "Lobby no longer in starting state", "lobby_id", lobby.ID)
	}

	// Update matchmaking sessions to StatusMatched
	if s.sessionRepo != nil {
		for _, playerID := range lobby.GetPlayerIDs() {
			sessions, err := s.sessionRepo.GetByPlayerID(ctx, playerID)
			if err != nil {
				slog.WarnContext(ctx, "Failed to get sessions for player", "player_id", playerID, "error", err)
				continue
			}
			for _, sess := range sessions {
				if sess.MatchID != nil && *sess.MatchID == lobby.ID && sess.Status == matchmaking_entities.StatusReadyCheck {
					sess.Status = matchmaking_entities.StatusMatched
					if sess.Metadata == nil {
						sess.Metadata = make(map[string]any)
					}
					sess.Metadata["match_id"] = matchID.String()
					sess.Metadata["match_started"] = true
					if err := s.sessionRepo.Save(ctx, sess); err != nil {
						slog.WarnContext(ctx, "Failed to update session to matched", "session_id", sess.ID, "error", err)
					}
				}
			}
		}
	}

	// Re-read lobby for accurate broadcast (in-memory object has stale status)
	updatedLobby, _ := s.lobbyRepo.FindByID(ctx, lobby.ID)
	if updatedLobby == nil {
		updatedLobby = lobby
	}

	// Broadcast all_players_ready + match started to each player
	playerIDs := lobby.GetPlayerIDs()
	connectionInfo := buildGameConnectionInfoPayload(updatedLobby, matchID)
	if updatedLobby != nil {
		if err := s.lobbyRepo.Update(ctx, updatedLobby); err != nil {
			slog.WarnContext(ctx, "Failed to persist generated connection info", "lobby_id", updatedLobby.ID, "error", err)
		}
	}
	payload := map[string]interface{}{
		"lobby_id": lobby.ID.String(),
		"match_id": matchID.String(),
		"status":   "started",
		"players":  len(playerIDs),
	}
	payloadBytes, _ := json.Marshal(payload)
	lobbyPayload := &ws.WebSocketMessage{
		Type:      ws.MessageTypeAllPlayersReady,
		LobbyID:   &lobby.ID,
		Payload:   payloadBytes,
		Timestamp: updatedLobby.UpdatedAt.Unix(),
	}
	s.wsHub.BroadcastRaw(lobbyPayload)
	connectionPayloadBytes, _ := json.Marshal(connectionInfo)
	connectionPayload := &ws.WebSocketMessage{
		Type:      ws.MessageTypeGameConnectionInfo,
		LobbyID:   &lobby.ID,
		Payload:   connectionPayloadBytes,
		Timestamp: updatedLobby.UpdatedAt.Unix(),
	}
	s.wsHub.BroadcastRaw(connectionPayload)
	for _, pid := range playerIDs {
		s.wsHub.BroadcastToUser(pid, "all_players_ready", payloadBytes)
		s.wsHub.BroadcastToUser(pid, ws.MessageTypeGameConnectionInfo, connectionPayloadBytes)
	}
	s.wsHub.BroadcastLobbyUpdate(lobby.ID, updatedLobby)

	return matchID, nil
}

func (s *LobbyOrchestrationService) StartReadyCheck(ctx context.Context, cmd matchmaking_in.StartReadyCheckCommand) error {
	lobby, err := s.lobbyRepo.FindByID(ctx, cmd.LobbyID)
	if err != nil {
		return fmt.Errorf("lobby not found: %w", err)
	}

	if err := lobby.StartReadyCheck(); err != nil {
		return fmt.Errorf("failed to start ready check: %w", err)
	}

	if err := s.lobbyRepo.Update(ctx, lobby); err != nil {
		return fmt.Errorf("failed to update lobby: %w", err)
	}

	slog.InfoContext(ctx, "Ready check started", "lobby_id", cmd.LobbyID)
	s.wsHub.BroadcastLobbyUpdate(cmd.LobbyID, lobby)

	return nil
}

func (s *LobbyOrchestrationService) StartMatch(ctx context.Context, lobbyID uuid.UUID) (uuid.UUID, error) {
	lobby, err := s.lobbyRepo.FindByID(ctx, lobbyID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("lobby not found: %w", err)
	}

	prizePool, err := s.prizePoolRepo.FindByMatchID(ctx, lobbyID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("prize pool not found: %w", err)
	}

	matchID := uuid.New()

	if err := lobby.StartMatch(matchID); err != nil {
		return uuid.Nil, fmt.Errorf("failed to start match: %w", err)
	}

	if err := prizePool.Lock(); err != nil {
		return uuid.Nil, fmt.Errorf("failed to lock prize pool: %w", err)
	}

	if err := s.lobbyRepo.Update(ctx, lobby); err != nil {
		return uuid.Nil, fmt.Errorf("failed to update lobby: %w", err)
	}

	if err := s.prizePoolRepo.Update(ctx, prizePool); err != nil {
		return uuid.Nil, fmt.Errorf("failed to update prize pool: %w", err)
	}

	if err := lobby.MarkMatchStarted(); err != nil {
		return uuid.Nil, fmt.Errorf("failed to mark match started: %w", err)
	}

	if err := s.lobbyRepo.Update(ctx, lobby); err != nil {
		return uuid.Nil, fmt.Errorf("failed to update lobby: %w", err)
	}

	slog.InfoContext(ctx, "Match started", "lobby_id", lobbyID, "match_id", matchID)
	s.wsHub.BroadcastLobbyUpdate(lobbyID, lobby)

	return matchID, nil
}

func (s *LobbyOrchestrationService) CancelLobby(ctx context.Context, lobbyID uuid.UUID, reason string) error {
	lobby, err := s.lobbyRepo.FindByID(ctx, lobbyID)
	if err != nil {
		return fmt.Errorf("lobby not found: %w", err)
	}

	prizePool, err := s.prizePoolRepo.FindByMatchID(ctx, lobbyID)
	hasPrizePool := err == nil && prizePool != nil
	if err != nil {
		slog.InfoContext(ctx, "No prize pool for lobby cancellation", "lobby_id", lobbyID, "reason", reason)
	}

	if err := lobby.Cancel(reason); err != nil {
		return fmt.Errorf("failed to cancel lobby: %w", err)
	}

	if hasPrizePool {
		if err := prizePool.Cancel(reason); err != nil {
			return fmt.Errorf("failed to cancel prize pool: %w", err)
		}
	}

	// Refund all players
	if hasPrizePool {
		for playerID, contribution := range prizePool.PlayerContributions {
			refundCmd := wallet_in.RefundCommand{
				UserID:   playerID,
				Currency: string(prizePool.Currency),
				Amount:   contribution.ToFloat(),
				Reason:   fmt.Sprintf("lobby cancelled: %s", reason),
			}
			if err := s.walletCommand.Refund(ctx, refundCmd); err != nil {
				slog.ErrorContext(ctx, "Failed to refund player", "player_id", playerID, "error", err)
			}
		}
	}

	s.handleMatchmakingSessionCancellation(ctx, lobby, reason)

	if err := s.lobbyRepo.Update(ctx, lobby); err != nil {
		return fmt.Errorf("failed to update lobby: %w", err)
	}

	if hasPrizePool {
		if err := s.prizePoolRepo.Update(ctx, prizePool); err != nil {
			return fmt.Errorf("failed to update prize pool: %w", err)
		}
	}

	slog.InfoContext(ctx, "Lobby cancelled", "lobby_id", lobbyID, "reason", reason)
	s.wsHub.BroadcastLobbyUpdate(lobbyID, lobby)

	return nil
}

func (s *LobbyOrchestrationService) InviteToLobby(ctx context.Context, cmd matchmaking_in.InviteToLobbyCommand) error {
	lobby, err := s.lobbyRepo.FindByID(ctx, cmd.LobbyID)
	if err != nil {
		return fmt.Errorf("lobby not found: %w", err)
	}

	// Validate inviter is actually in the lobby
	inviterInLobby := false
	for _, slot := range lobby.PlayerSlots {
		if slot.PlayerID != nil && *slot.PlayerID == cmd.InviterID {
			inviterInLobby = true
			break
		}
	}
	if !inviterInLobby {
		return fmt.Errorf("inviter is not in the lobby")
	}

	// Validate lobby is not full
	if lobby.IsFull() {
		return fmt.Errorf("lobby is full")
	}

	// Validate lobby is in a state that allows invites
	if lobby.Status != matchmaking_entities.LobbyStatusOpen {
		return fmt.Errorf("lobby is not accepting invites (status: %s)", lobby.Status)
	}

	// Validate invitee is not already in the lobby
	for _, slot := range lobby.PlayerSlots {
		if slot.PlayerID != nil && *slot.PlayerID == cmd.InviteeID {
			return fmt.Errorf("player is already in the lobby")
		}
	}

	// Broadcast invite notification to the invitee via WebSocket
	invitePayload := map[string]interface{}{
		"lobby_id":    lobby.ID.String(),
		"inviter_id":  cmd.InviterID.String(),
		"game_id":     lobby.GameID,
		"region":      lobby.Region,
		"tier":        lobby.Tier,
		"max_players": lobby.MaxPlayers,
		"players":     lobby.GetPlayerCount(),
	}
	payloadBytes, _ := json.Marshal(invitePayload)
	s.wsHub.BroadcastToUser(cmd.InviteeID, "lobby_invite", payloadBytes)

	slog.InfoContext(ctx, "Lobby invite sent",
		"lobby_id", cmd.LobbyID,
		"inviter", cmd.InviterID,
		"invitee", cmd.InviteeID)

	return nil
}

func (s *LobbyOrchestrationService) createPrizePool(ctx context.Context, cmd matchmaking_in.CreatePrizePoolCommand) (*matchmaking_entities.PrizePool, error) {
	resourceOwner := shared.GetResourceOwner(ctx)

	currency, err := wallet_vo.ParseCurrency(cmd.Currency)
	if err != nil {
		return nil, fmt.Errorf("invalid currency: %w", err)
	}

	platformAmount := wallet_vo.NewAmount(cmd.PlatformContribution)

	pool := matchmaking_entities.NewPrizePool(
		resourceOwner,
		cmd.MatchID,
		replay_common.GameIDKey(cmd.GameID),
		cmd.Region,
		currency,
		cmd.DistributionRule,
		platformAmount,
	)

	if err := s.prizePoolRepo.Save(ctx, pool); err != nil {
		return nil, fmt.Errorf("failed to save prize pool: %w", err)
	}

	return pool, nil
}

func getEntryFeeByTier(tier string) float64 {
	fees := map[string]float64{
		"free":    0.00,
		"premium": 1.00,
		"pro":     2.00,
		"elite":   5.00,
	}
	if fee, exists := fees[tier]; exists {
		return fee
	}
	return 0.00
}

func (s *LobbyOrchestrationService) broadcastReadinessUpdate(lobbyID uuid.UUID, playerID uuid.UUID, isReady bool, readyCount, totalCount int) {
	if s.wsHub == nil {
		return
	}

	status := "declined"
	messageType := ws.MessageTypeReadinessDeclined
	if isReady {
		status = "confirmed"
		messageType = ws.MessageTypeReadinessConfirmed
	}

	payload := map[string]interface{}{
		"lobby_id":        lobbyID.String(),
		"player_id":       playerID.String(),
		"status":          status,
		"confirmed_count": readyCount,
		"total_count":     totalCount,
	}
	payloadBytes, _ := json.Marshal(payload)
	s.wsHub.BroadcastRaw(&ws.WebSocketMessage{
		Type:      messageType,
		LobbyID:   &lobbyID,
		Payload:   payloadBytes,
		Timestamp: time.Now().Unix(),
	})
}

func (s *LobbyOrchestrationService) handleMatchmakingSessionCancellation(ctx context.Context, lobby *matchmaking_entities.MatchmakingLobby, reason string) {
	if s.sessionRepo == nil || lobby == nil {
		return
	}

	declinedPlayerID := uuid.Nil
	if strings.HasPrefix(reason, "player_declined:") {
		declinedRaw := strings.TrimPrefix(reason, "player_declined:")
		parsed, err := uuid.Parse(declinedRaw)
		if err == nil {
			declinedPlayerID = parsed
		}
	}

	for _, playerID := range lobby.GetPlayerIDs() {
		sessions, err := s.sessionRepo.GetByPlayerID(ctx, playerID)
		if err != nil {
			slog.WarnContext(ctx, "Failed to get sessions for cancelled lobby", "player_id", playerID, "error", err)
			continue
		}

		for _, sess := range sessions {
			if sess.MatchID == nil || *sess.MatchID != lobby.ID || sess.Status != matchmaking_entities.StatusReadyCheck {
				continue
			}

			if sess.Metadata == nil {
				sess.Metadata = make(map[string]any)
			}
			sess.Metadata["lobby_id"] = nil
			sess.Metadata["cancel_reason"] = reason

			if declinedPlayerID != uuid.Nil && playerID == declinedPlayerID {
				sess.Status = matchmaking_entities.StatusCancelled
				sess.Metadata["declined_ready_check"] = true
			} else {
				sess.Status = matchmaking_entities.StatusQueued
				sess.MatchID = nil
				sess.MatchedAt = nil
				sess.Metadata["requeued_reason"] = reason
			}

			if err := s.sessionRepo.Save(ctx, sess); err != nil {
				slog.WarnContext(ctx, "Failed to update session during cancellation", "session_id", sess.ID, "error", err)
			}
		}
	}
}

func buildGameConnectionInfoPayload(lobby *matchmaking_entities.MatchmakingLobby, matchID uuid.UUID) map[string]interface{} {
	matchCode := strings.ToUpper(matchID.String()[:8])
	deepLink := fmt.Sprintf("leetgaming://matches/%s/join?code=%s", matchID.String(), matchCode)

	if lobby != nil {
		if lobby.Metadata == nil {
			lobby.Metadata = make(map[string]any)
		}
		lobby.Metadata["match_code"] = matchCode
		lobby.Metadata["deep_link"] = deepLink
	}

	return map[string]interface{}{
		"lobby_id":     lobby.ID.String(),
		"match_id":     matchID.String(),
		"game_id":      lobby.GameID,
		"region":       lobby.Region,
		"server_url":   fmt.Sprintf("https://join.leetgaming.pro/matches/%s", matchID.String()),
		"port":         27015,
		"passcode":     matchCode,
		"qr_code_data": deepLink,
		"deep_link":    deepLink,
		"instructions": "Open the match link or scan the QR code to join the server.",
	}
}
