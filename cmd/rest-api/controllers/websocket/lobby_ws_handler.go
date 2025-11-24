package websocket_controllers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	wsHub "github.com/psavelis/team-pro/replay-api/pkg/infra/websocket"
)

type LobbyWebSocketHandler struct {
	container container.Container
	hub       *wsHub.WebSocketHub
	upgrader  websocket.Upgrader
}

func NewLobbyWebSocketHandler(container container.Container, hub *wsHub.WebSocketHub) *LobbyWebSocketHandler {
	return &LobbyWebSocketHandler{
		container: container,
		hub:       hub,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // TODO: Implement proper CORS check
			},
		},
	}
}

func (h *LobbyWebSocketHandler) UpgradeConnection(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		lobbyIDStr := vars["lobby_id"]
		
		lobbyID, err := uuid.Parse(lobbyIDStr)
		if err != nil {
			slog.ErrorContext(ctx, "invalid lobby_id in WebSocket request", "lobby_id", lobbyIDStr, "error", err)
			http.Error(w, "invalid lobby_id", http.StatusBadRequest)
			return
		}

		// Upgrade HTTP connection to WebSocket
		conn, err := h.upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.ErrorContext(ctx, "failed to upgrade WebSocket connection", "error", err)
			return
		}

		// Create client
		client := &wsHub.Client{
			ID:         uuid.New(),
			Conn:       conn,
			Send:       make(chan *wsHub.WebSocketMessage, 256),
			LobbyID:    &lobbyID,
			Disconnect: make(chan struct{}),
		}

		// Register client with hub
		h.hub.RegisterClient(client)

		// Start goroutines for read/write pumps
		go client.WritePump()
		go client.ReadPump(h.hub)

		slog.InfoContext(ctx, "WebSocket client connected", "client_id", client.ID, "lobby_id", lobbyID)
	}
}
