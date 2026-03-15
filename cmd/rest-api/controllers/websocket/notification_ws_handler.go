package websocket_controllers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	wsHub "github.com/replay-api/replay-api/pkg/infra/websocket"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// NotificationWebSocketHandler handles WebSocket connections for user-scoped
// real-time notification delivery. Unlike the lobby handler, this does not
// require a lobby_id — it subscribes the authenticated user to their
// personal notification stream.
type NotificationWebSocketHandler struct {
	container container.Container
	hub       *wsHub.WebSocketHub
	upgrader  websocket.Upgrader
}

func NewNotificationWebSocketHandler(c container.Container, hub *wsHub.WebSocketHub) *NotificationWebSocketHandler {
	allowed := buildAllowedOrigins() // shared helper in lobby_ws_handler.go

	return &NotificationWebSocketHandler{
		container: c,
		hub:       hub,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return false
				}
				return allowed[origin]
			},
		},
	}
}

// UpgradeConnection upgrades an HTTP request to a WebSocket connection and
// registers the client in the hub's userRooms under the authenticated user ID.
func (h *NotificationWebSocketHandler) UpgradeConnection(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Require authentication
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

		// Upgrade to WebSocket
		conn, err := h.upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.ErrorContext(ctx, "failed to upgrade notification WebSocket", "error", err)
			return
		}

		client := &wsHub.Client{
			ID:         userID,
			Conn:       conn,
			Send:       make(chan *wsHub.WebSocketMessage, 256),
			LobbyID:    nil, // not lobby-scoped
			Disconnect: make(chan struct{}),
		}

		// Register into the user room (not a lobby room)
		h.hub.RegisterUserClient(userID, client)

		// Write pump sends messages from hub → client
		go client.WritePump()

		// Read pump processes client → server messages (keep-alive, ack)
		go h.readPump(client, userID)

		slog.InfoContext(ctx, "Notification WebSocket connected", "user_id", userID)
	}
}

// readPump reads messages from the client. For the notification channel we
// mostly just keep the connection alive, but we also handle subscribe_notifications
// as a no-op ack and any future client-to-server message types.
func (h *NotificationWebSocketHandler) readPump(client *wsHub.Client, userID uuid.UUID) {
	defer func() {
		h.hub.UnregisterUserClient(userID, client)
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(512)

	for {
		var msg map[string]interface{}
		if err := client.Conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("Notification WebSocket read error", "error", err, "user_id", userID)
			}
			break
		}

		// Handle optional message types from the client
		if msgType, ok := msg["type"].(string); ok {
			switch msgType {
			case wsHub.MessageTypeSubscribeNotifications:
				// Client explicitly subscribing — already registered, this is a no-op ack
				slog.Debug("Notification subscription confirmed", "user_id", userID)
			case wsHub.MessageTypeNotificationRead:
				// Client may ACK a read — we could relay this to other tabs
				slog.Debug("Notification read ack", "user_id", userID)
			default:
				slog.Debug("Unknown notification message type", "type", msgType, "user_id", userID)
			}
		}
	}
}
