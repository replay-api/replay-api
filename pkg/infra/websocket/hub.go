package websocket

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
)

// Message types for WebSocket protocol
const (
MessageTypeLobbyUpdate        = "lobby_update"
MessageTypePlayerJoined       = "player_joined"
MessageTypePlayerLeft         = "player_left"
MessageTypeReadyStatusChanged = "ready_status_changed"
MessageTypePrizePoolUpdate    = "prize_pool_update"
MessageTypeMatchStarting      = "match_starting"
)

// WebSocketMessage is the wire protocol format
type WebSocketMessage struct {
	Type      string          `json:"type"`
	LobbyID   *uuid.UUID      `json:"lobby_id,omitempty"`
	PoolID    *uuid.UUID      `json:"pool_id,omitempty"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp int64           `json:"timestamp"`
}

// Client represents a connected WebSocket client
type Client struct {
	ID         uuid.UUID
	Conn       *websocket.Conn
	Send       chan *WebSocketMessage
	LobbyID    *uuid.UUID
	Disconnect chan struct{}
}

// WebSocketHub manages all WebSocket connections and broadcasts
type WebSocketHub struct {
	clients    map[uuid.UUID]*Client
	lobbyRooms map[uuid.UUID]map[uuid.UUID]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan *WebSocketMessage
	mu         sync.RWMutex
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[uuid.UUID]*Client),
		lobbyRooms: make(map[uuid.UUID]map[uuid.UUID]*Client),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
		broadcast:  make(chan *WebSocketMessage, 1024),
	}
}

// RegisterClient adds a client to the hub
func (h *WebSocketHub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient removes a client from the hub
func (h *WebSocketHub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// Run starts the hub's main event loop
func (h *WebSocketHub) Run(ctx context.Context) {
for {
select {
case <-ctx.Done():
h.shutdown()
return
case client := <-h.register:
h.registerClient(client)
case client := <-h.unregister:
h.unregisterClient(client)
case message := <-h.broadcast:
h.broadcastMessage(message)
}
}
}

func (h *WebSocketHub) registerClient(client *Client) {
h.mu.Lock()
defer h.mu.Unlock()

h.clients[client.ID] = client
if client.LobbyID != nil {
if _, exists := h.lobbyRooms[*client.LobbyID]; !exists {
h.lobbyRooms[*client.LobbyID] = make(map[uuid.UUID]*Client)
}
h.lobbyRooms[*client.LobbyID][client.ID] = client
}

slog.Info("WebSocket client connected", "client_id", client.ID, "lobby_id", client.LobbyID)
}

func (h *WebSocketHub) unregisterClient(client *Client) {
h.mu.Lock()
defer h.mu.Unlock()

if _, exists := h.clients[client.ID]; exists {
delete(h.clients, client.ID)
if client.LobbyID != nil {
delete(h.lobbyRooms[*client.LobbyID], client.ID)
if len(h.lobbyRooms[*client.LobbyID]) == 0 {
delete(h.lobbyRooms, *client.LobbyID)
}
}
close(client.Send)
slog.Info("WebSocket client disconnected", "client_id", client.ID)
}
}

func (h *WebSocketHub) broadcastMessage(message *WebSocketMessage) {
h.mu.RLock()
defer h.mu.RUnlock()

if message.LobbyID != nil {
if clients, exists := h.lobbyRooms[*message.LobbyID]; exists {
for _, client := range clients {
select {
case client.Send <- message:
default:
slog.Warn("Client send buffer full", "client_id", client.ID)
}
}
}
} else {
for _, client := range h.clients {
select {
case client.Send <- message:
default:
slog.Warn("Client send buffer full", "client_id", client.ID)
}
}
}
}

// BroadcastLobbyUpdate sends lobby state to all subscribed clients
func (h *WebSocketHub) BroadcastLobbyUpdate(lobbyID uuid.UUID, lobby *matchmaking_entities.MatchmakingLobby) {
payload, err := json.Marshal(lobby)
if err != nil {
slog.Error("Failed to marshal lobby", "error", err)
return
}

message := &WebSocketMessage{
Type:      MessageTypeLobbyUpdate,
LobbyID:   &lobbyID,
Payload:   payload,
Timestamp: time.Now().Unix(),
}

h.broadcast <- message
}

// BroadcastPrizePoolUpdate sends prize pool state to subscribed clients
func (h *WebSocketHub) BroadcastPrizePoolUpdate(lobbyID uuid.UUID, pool *matchmaking_entities.PrizePool) {
payload, err := json.Marshal(pool)
if err != nil {
slog.Error("Failed to marshal prize pool", "error", err)
return
}

message := &WebSocketMessage{
Type:      MessageTypePrizePoolUpdate,
LobbyID:   &lobbyID,
Payload:   payload,
Timestamp: time.Now().Unix(),
}

h.broadcast <- message
}

func (h *WebSocketHub) shutdown() {
h.mu.Lock()
defer h.mu.Unlock()

for _, client := range h.clients {
close(client.Send)
}
slog.Info("WebSocket hub shut down")
}

// BroadcastFromKafka handles events received from Kafka and broadcasts to WebSocket clients
// This method is designed to be called by the Kafka consumer
func (h *WebSocketHub) BroadcastFromKafka(eventType string, lobbyID *uuid.UUID, payload json.RawMessage) {
message := &WebSocketMessage{
Type:      eventType,
LobbyID:   lobbyID,
Payload:   payload,
Timestamp: time.Now().Unix(),
}

h.broadcast <- message
}

// BroadcastRaw sends a raw WebSocketMessage to the broadcast channel
func (h *WebSocketHub) BroadcastRaw(message *WebSocketMessage) {
h.broadcast <- message
}

// GetConnectedClientsCount returns the total number of connected clients
func (h *WebSocketHub) GetConnectedClientsCount() int {
h.mu.RLock()
defer h.mu.RUnlock()
return len(h.clients)
}

// GetLobbyClientsCount returns the number of clients subscribed to a specific lobby
func (h *WebSocketHub) GetLobbyClientsCount(lobbyID uuid.UUID) int {
h.mu.RLock()
defer h.mu.RUnlock()
if clients, exists := h.lobbyRooms[lobbyID]; exists {
return len(clients)
}
return 0
}

// WritePump sends messages from the hub to the websocket connection
func (c *Client) WritePump() {
defer c.Conn.Close()

for {
select {
case message, ok := <-c.Send:
if !ok {
c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
return
}

if err := c.Conn.WriteJSON(message); err != nil {
slog.Error("Write error", "client_id", c.ID, "error", err)
return
}
case <-c.Disconnect:
return
}
}
}

// ReadPump reads messages from the websocket connection
func (c *Client) ReadPump(hub *WebSocketHub) {
defer func() {
hub.unregister <- c
c.Conn.Close()
}()

c.Conn.SetReadLimit(512)

for {
var msg map[string]interface{}
if err := c.Conn.ReadJSON(&msg); err != nil {
if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
slog.Error("WebSocket read error", "error", err)
}
break
}

if msgType, ok := msg["type"].(string); ok && msgType == "subscribe_lobby" {
if lobbyIDStr, ok := msg["lobby_id"].(string); ok {
lobbyID, err := uuid.Parse(lobbyIDStr)
if err == nil {
c.LobbyID = &lobbyID
hub.register <- c
}
}
}
}
}
