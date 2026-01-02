package kafka

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_out "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/out"
	"github.com/segmentio/kafka-go"
)

// WebSocketBroadcaster interface for broadcasting messages to WebSocket clients
type WebSocketBroadcaster interface {
	BroadcastFromKafka(eventType string, lobbyID *uuid.UUID, payload json.RawMessage)
}

// WebSocketBridge connects Kafka events to WebSocket broadcasts
// Enables multi-instance WebSocket server coordination via Kafka
type WebSocketBridge struct {
	client           *Client
	consumer         *Consumer
	broadcaster      WebSocketBroadcaster
	publisher        *EventPublisher
	sessionRepo      matchmaking_out.MatchmakingSessionRepository
}

// NewWebSocketBridge creates a new bridge between Kafka and WebSocket
func NewWebSocketBridge(client *Client, broadcaster WebSocketBroadcaster, instanceID string, sessionRepo matchmaking_out.MatchmakingSessionRepository) *WebSocketBridge {
	groupID := "websocket-service-" + instanceID
	config := DefaultConsumerConfig(groupID, []string{TopicWebSocketBroadcast, TopicLobbyEvents, TopicMatchesCreated})
	consumer := NewConsumer(client, config)

	bridge := &WebSocketBridge{
		client:      client,
		consumer:    consumer,
		broadcaster: broadcaster,
		publisher:   NewEventPublisher(client),
		sessionRepo: sessionRepo,
	}

	consumer.RegisterHandler(TopicWebSocketBroadcast, bridge.handleWebSocketBroadcast)
	consumer.RegisterHandler(TopicLobbyEvents, bridge.handleLobbyEvent)
	consumer.RegisterHandler(TopicMatchesCreated, bridge.handleMatchCreated)

	return bridge
}

func (b *WebSocketBridge) handleWebSocketBroadcast(ctx context.Context, msg *kafka.Message) error {
	var event WebSocketBroadcastEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.Error("Failed to unmarshal broadcast event", "error", err)
		return err
	}

	// Marshal the payload to JSON raw message
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		slog.Error("Failed to marshal payload", "error", err)
		return err
	}

	b.broadcaster.BroadcastFromKafka(event.Type, event.LobbyID, payload)
	return nil
}

func (b *WebSocketBridge) handleLobbyEvent(ctx context.Context, msg *kafka.Message) error {
	var event LobbyEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.Error("Failed to unmarshal lobby event", "error", err)
		return err
	}

	// Update matchmaking sessions to "searching" status when lobby is created
	if event.EventType == EventTypeLobbyCreated {
		for _, playerID := range event.PlayerIDs {
			if err := b.updateSessionStatus(ctx, playerID, matchmaking_entities.StatusSearching); err != nil {
				slog.Warn("Failed to update session status for player", "player_id", playerID, "error", err)
			}
		}
	}

	// Convert lobby event to WebSocket broadcast payload
	payload, err := json.Marshal(event)
	if err != nil {
		slog.Error("Failed to marshal lobby event", "error", err)
		return err
	}

	b.broadcaster.BroadcastFromKafka(event.EventType, &event.LobbyID, payload)
	return nil
}

func (b *WebSocketBridge) handleMatchCreated(ctx context.Context, msg *kafka.Message) error {
	var event MatchEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.Error("Failed to unmarshal match created event", "error", err)
		return err
	}

	// Update matchmaking sessions to "matched" status when match is created
	for _, playerID := range event.PlayerIDs {
		if err := b.updateSessionStatus(ctx, playerID, matchmaking_entities.StatusMatched); err != nil {
			slog.Warn("Failed to update session status for player", "player_id", playerID, "error", err)
		}
	}

	// Convert match event to WebSocket broadcast payload
	payload, err := json.Marshal(event)
	if err != nil {
		slog.Error("Failed to marshal match created event", "error", err)
		return err
	}

	b.broadcaster.BroadcastFromKafka(event.EventType, &event.LobbyID, payload)
	return nil
}

// Start begins consuming Kafka events and broadcasting to WebSocket clients
func (b *WebSocketBridge) Start(ctx context.Context) error {
	slog.Info("Starting WebSocket-Kafka bridge")
	return b.consumer.Start(ctx)
}

// Close shuts down the bridge
func (b *WebSocketBridge) Close() error {
	return b.consumer.Close()
}

// Publisher returns the event publisher for sending events to Kafka
func (b *WebSocketBridge) Publisher() *EventPublisher {
	return b.publisher
}

// PublishLobbyEvent is a convenience method to publish lobby events via the bridge
func (b *WebSocketBridge) PublishLobbyEvent(ctx context.Context, event *LobbyEvent) error {
	return b.publisher.PublishLobbyEvent(ctx, event)
}

// PublishMatchEvent is a convenience method to publish match events via the bridge
func (b *WebSocketBridge) PublishMatchEvent(ctx context.Context, event *MatchEvent) error {
	return b.publisher.PublishMatchCreated(ctx, event)
}

// updateSessionStatus updates the matchmaking session status for a player
func (b *WebSocketBridge) updateSessionStatus(ctx context.Context, playerID uuid.UUID, status matchmaking_entities.SessionStatus) error {
	sessions, err := b.sessionRepo.GetByPlayerID(ctx, playerID)
	if err != nil {
		return err
	}

	// Update the most recent active session
	for _, session := range sessions {
		if session.CanMatch() {
			session.Status = status
			return b.sessionRepo.Save(ctx, session)
		}
	}

	return nil
}
