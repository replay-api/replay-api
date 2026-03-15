package websocket_controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	ws "github.com/gorilla/websocket"
	wsHub "github.com/replay-api/replay-api/pkg/infra/websocket"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// MessagingWebSocketHandler handles WebSocket connections for the messaging
// system. It subscribes authenticated users to their personal messaging stream
// (DMs, team messages) via the userRooms mechanism.
type MessagingWebSocketHandler struct {
	container container.Container
	hub       *wsHub.WebSocketHub
	upgrader  ws.Upgrader
}

func NewMessagingWebSocketHandler(c container.Container, hub *wsHub.WebSocketHub) *MessagingWebSocketHandler {
	allowed := buildAllowedOrigins()

	return &MessagingWebSocketHandler{
		container: c,
		hub:       hub,
		upgrader: ws.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 4096,
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

// UpgradeConnection upgrades an HTTP request to a WebSocket connection for
// the messaging channel.
func (h *MessagingWebSocketHandler) UpgradeConnection(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		conn, err := h.upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.ErrorContext(ctx, "failed to upgrade messaging WebSocket", "error", err)
			return
		}

		client := &wsHub.Client{
			ID:         userID,
			Conn:       conn,
			Send:       make(chan *wsHub.WebSocketMessage, 256),
			LobbyID:    nil,
			Disconnect: make(chan struct{}),
		}

		// Register in user room for DM/team message delivery
		h.hub.RegisterUserClient(userID, client)

		go client.WritePump()
		go h.readPump(client, userID)

		slog.InfoContext(ctx, "Messaging WebSocket connected", "user_id", userID)
	}
}

// messagingClientMessage represents a message from the client to subscribe to
// team rooms or send typing indicators.
type messagingClientMessage struct {
	Type   string `json:"type"`              // "subscribe_team", "unsubscribe_team", "typing"
	TeamID string `json:"team_id,omitempty"` // team UUID
}

func (h *MessagingWebSocketHandler) readPump(client *wsHub.Client, userID uuid.UUID) {
	defer func() {
		h.hub.UnregisterUserClient(userID, client)
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(2048)

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseNormalClosure) {
				slog.Warn("Messaging WebSocket unexpected close", "user_id", userID, "error", err)
			}
			break
		}

		var msg messagingClientMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "subscribe_team":
			teamID, err := uuid.Parse(msg.TeamID)
			if err != nil {
				continue
			}
			// Subscribe client to team room via lobbyRooms mechanism
			teamClient := &wsHub.Client{
				ID:         userID,
				Conn:       client.Conn,
				Send:       client.Send,
				LobbyID:    &teamID,
				Disconnect: client.Disconnect,
			}
			h.hub.RegisterClient(teamClient)

		case "unsubscribe_team":
			teamID, err := uuid.Parse(msg.TeamID)
			if err != nil {
				continue
			}
			unsubClient := &wsHub.Client{
				ID:      userID,
				Conn:    client.Conn,
				Send:    client.Send,
				LobbyID: &teamID,
			}
			h.hub.UnregisterClient(unsubClient)

		case "typing":
			// Broadcast typing indicator to the user's team room
			if msg.TeamID != "" {
				teamID, err := uuid.Parse(msg.TeamID)
				if err != nil {
					continue
				}
				typingPayload, _ := json.Marshal(map[string]string{
					"user_id": userID.String(),
					"team_id": teamID.String(),
				})
				h.hub.BroadcastRaw(&wsHub.WebSocketMessage{
					Type:      "typing",
					LobbyID:   &teamID,
					Payload:   typingPayload,
					Timestamp: time.Now().Unix(),
				})
			}
		}
	}
}
