package websocket_controllers

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	wsHub "github.com/replay-api/replay-api/pkg/infra/websocket"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type LobbyWebSocketHandler struct {
	container      container.Container
	hub            *wsHub.WebSocketHub
	upgrader       websocket.Upgrader
	allowedOrigins map[string]bool
}

// buildAllowedOrigins reads CORS_ALLOWED_ORIGINS / CORS_ALLOWED_ORIGIN env vars
// and returns the set of allowed origins for WebSocket upgrade requests.
func buildAllowedOrigins() map[string]bool {
	origins := make(map[string]bool)
	origins["http://localhost:3030"] = true
	origins["http://localhost:3000"] = true

	if envOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); envOrigins != "" {
		for _, o := range strings.Split(envOrigins, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				origins[o] = true
			}
		}
	}
	if single := os.Getenv("CORS_ALLOWED_ORIGIN"); single != "" {
		origins[strings.TrimSpace(single)] = true
	}
	return origins
}

func NewLobbyWebSocketHandler(container container.Container, hub *wsHub.WebSocketHub) *LobbyWebSocketHandler {
	allowed := buildAllowedOrigins()

	return &LobbyWebSocketHandler{
		container:      container,
		hub:            hub,
		allowedOrigins: allowed,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// SECURITY: Validate Origin header against allowed list
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return false // Reject requests without Origin header
				}
				return allowed[origin]
			},
		},
	}
}

func (h *LobbyWebSocketHandler) UpgradeConnection(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// SECURITY: Require authentication before WebSocket upgrade
		reqCtx := r.Context()
		authenticated, ok := reqCtx.Value(shared.AuthenticatedKey).(bool)
		if !ok || !authenticated {
			http.Error(w, `{"success":false,"error":"Authentication required"}`, http.StatusUnauthorized)
			return
		}
		userID, ok := reqCtx.Value(shared.UserIDKey).(uuid.UUID)
		if !ok || userID == uuid.Nil {
			http.Error(w, `{"success":false,"error":"User identity not found"}`, http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		lobbyIDStr := vars["lobby_id"]

		lobbyID, err := uuid.Parse(lobbyIDStr)
		if err != nil {
			slog.ErrorContext(ctx, "invalid lobby_id in WebSocket request", "lobby_id", lobbyIDStr)
			http.Error(w, "invalid lobby_id", http.StatusBadRequest)
			return
		}

		// Upgrade HTTP connection to WebSocket
		conn, err := h.upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.ErrorContext(ctx, "failed to upgrade WebSocket connection", "error", err)
			return
		}

		// Create client with authenticated user ID
		client := &wsHub.Client{
			ID:         userID, // Use authenticated user ID, not random UUID
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

		slog.InfoContext(ctx, "WebSocket client connected", "client_id", client.ID, "lobby_id", lobbyID, "user_id", userID)
	}
}
