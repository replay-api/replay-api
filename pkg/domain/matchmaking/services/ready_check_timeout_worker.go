package matchmaking_services

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_out "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/out"
	ws "github.com/replay-api/replay-api/pkg/infra/websocket"
)

// ReadyCheckTimeoutWorker periodically checks for expired ready checks and
// cancels lobbies / re-queues players when the timeout elapses.
type ReadyCheckTimeoutWorker struct {
	lobbyRepo   matchmaking_out.LobbyRepository
	sessionRepo matchmaking_out.MatchmakingSessionRepository
	wsHub       *ws.WebSocketHub
	interval    time.Duration
}

// NewReadyCheckTimeoutWorker creates a new timeout worker.
func NewReadyCheckTimeoutWorker(
	lobbyRepo matchmaking_out.LobbyRepository,
	sessionRepo matchmaking_out.MatchmakingSessionRepository,
	wsHub *ws.WebSocketHub,
) *ReadyCheckTimeoutWorker {
	return &ReadyCheckTimeoutWorker{
		lobbyRepo:   lobbyRepo,
		sessionRepo: sessionRepo,
		wsHub:       wsHub,
		interval:    5 * time.Second,
	}
}

// Run starts the ticker loop. Blocks until ctx is cancelled.
func (w *ReadyCheckTimeoutWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	slog.Info("ReadyCheckTimeoutWorker started", "interval", w.interval)

	for {
		select {
		case <-ctx.Done():
			slog.Info("ReadyCheckTimeoutWorker stopped")
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *ReadyCheckTimeoutWorker) tick(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("ReadyCheckTimeoutWorker panic recovered", "recover", r)
		}
	}()

	expired, err := w.lobbyRepo.FindExpiredReadyChecks(ctx)
	if err != nil {
		slog.Error("ReadyCheckTimeoutWorker: failed to find expired ready checks", "error", err)
		return
	}

	for _, lobby := range expired {
		w.handleExpiredLobby(ctx, lobby)
	}
}

func (w *ReadyCheckTimeoutWorker) handleExpiredLobby(ctx context.Context, lobby *matchmaking_entities.MatchmakingLobby) {
	slog.Info("ReadyCheckTimeoutWorker: handling expired lobby",
		"lobby_id", lobby.ID,
		"ready_check_end", lobby.ReadyCheckEnd)

	playerIDs := lobby.GetPlayerIDs()

	// Cancel the lobby
	if err := lobby.Cancel("ready_check_timeout"); err != nil {
		slog.Error("ReadyCheckTimeoutWorker: failed to cancel lobby", "lobby_id", lobby.ID, "error", err)
		return
	}

	if err := w.lobbyRepo.Update(ctx, lobby); err != nil {
		slog.Error("ReadyCheckTimeoutWorker: failed to update cancelled lobby", "lobby_id", lobby.ID, "error", err)
		return
	}

	// Re-queue players who were ready (they didn't cause the timeout)
	_, notReadyPlayers := lobby.CheckReadyStatus()
	notReadySet := make(map[string]bool)
	for _, pid := range notReadyPlayers {
		notReadySet[pid.String()] = true
	}

	for _, playerID := range playerIDs {
		sessions, err := w.sessionRepo.GetByPlayerID(ctx, playerID)
		if err != nil {
			slog.Warn("ReadyCheckTimeoutWorker: failed to get sessions", "player_id", playerID, "error", err)
			continue
		}
		for _, sess := range sessions {
			if sess.MatchID != nil && *sess.MatchID == lobby.ID && sess.Status == matchmaking_entities.StatusReadyCheck {
				if notReadySet[playerID.String()] {
					// Player didn't accept → expire their session
					sess.Status = matchmaking_entities.StatusExpired
					if sess.Metadata == nil {
						sess.Metadata = make(map[string]any)
					}
					sess.Metadata["timeout_reason"] = "did_not_accept_ready_check"
				} else {
					// Player was ready → re-queue them
					sess.Status = matchmaking_entities.StatusQueued
					sess.MatchID = nil
					sess.MatchedAt = nil
					if sess.Metadata == nil {
						sess.Metadata = make(map[string]any)
					}
					sess.Metadata["requeued_reason"] = "ready_check_timeout_other_player"
					sess.Metadata["lobby_id"] = nil
				}
				if err := w.sessionRepo.Save(ctx, sess); err != nil {
					slog.Warn("ReadyCheckTimeoutWorker: failed to update session", "session_id", sess.ID, "error", err)
				}
			}
		}
	}

	// Notify players via WebSocket
	payload := map[string]interface{}{
		"lobby_id": lobby.ID.String(),
		"reason":   "ready_check_timeout",
	}
	payloadBytes, _ := json.Marshal(payload)
	for _, pid := range playerIDs {
		w.wsHub.BroadcastToUser(pid, "ready_check_timeout", payloadBytes)
	}

	w.wsHub.BroadcastLobbyUpdate(lobby.ID, lobby)

	slog.Info("ReadyCheckTimeoutWorker: lobby cancelled and players notified",
		"lobby_id", lobby.ID,
		"total_players", len(playerIDs),
		"not_ready_players", len(notReadyPlayers))
}
