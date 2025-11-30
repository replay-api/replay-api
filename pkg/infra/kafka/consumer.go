package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

// ConsumerConfig holds consumer-specific configuration
type ConsumerConfig struct {
	GroupID          string
	Topics           []string
	MinBytes         int
	MaxBytes         int
	MaxWait          time.Duration
	CommitInterval   time.Duration
	StartOffset      int64
	HeartbeatInterval time.Duration
	SessionTimeout   time.Duration
}

// DefaultConsumerConfig returns sensible defaults
func DefaultConsumerConfig(groupID string, topics []string) *ConsumerConfig {
	return &ConsumerConfig{
		GroupID:          groupID,
		Topics:           topics,
		MinBytes:         1e3,    // 1KB
		MaxBytes:         10e6,   // 10MB
		MaxWait:          time.Second,
		CommitInterval:   time.Second,
		StartOffset:      kafka.LastOffset,
		HeartbeatInterval: 3 * time.Second,
		SessionTimeout:   30 * time.Second,
	}
}

// Consumer provides Kafka message consumption
type Consumer struct {
	client   *Client
	config   *ConsumerConfig
	reader   *kafka.Reader
	handlers map[string]MessageHandler
}

// MessageHandler processes a single Kafka message
type MessageHandler func(ctx context.Context, msg *kafka.Message) error

// NewConsumer creates a new Kafka consumer
func NewConsumer(client *Client, config *ConsumerConfig) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:           client.Brokers(),
		GroupID:           config.GroupID,
		GroupTopics:       config.Topics,
		MinBytes:          config.MinBytes,
		MaxBytes:          config.MaxBytes,
		MaxWait:           config.MaxWait,
		CommitInterval:    config.CommitInterval,
		StartOffset:       config.StartOffset,
		HeartbeatInterval: config.HeartbeatInterval,
		SessionTimeout:    config.SessionTimeout,
		Dialer:            client.Dialer(),
	})

	return &Consumer{
		client:   client,
		config:   config,
		reader:   reader,
		handlers: make(map[string]MessageHandler),
	}
}

// RegisterHandler registers a handler for a specific topic
func (c *Consumer) RegisterHandler(topic string, handler MessageHandler) {
	c.handlers[topic] = handler
}

// Start begins consuming messages
func (c *Consumer) Start(ctx context.Context) error {
	slog.Info("Starting Kafka consumer",
		"group_id", c.config.GroupID,
		"topics", c.config.Topics)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Consumer context cancelled, shutting down")
			return c.reader.Close()
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return nil // Context cancelled
				}
				slog.Error("Error fetching message", "error", err)
				continue
			}

			if err := c.processMessage(ctx, &msg); err != nil {
				slog.Error("Error processing message",
					"topic", msg.Topic,
					"partition", msg.Partition,
					"offset", msg.Offset,
					"error", err)
				// Don't commit failed messages - they'll be reprocessed
				continue
			}

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				slog.Error("Error committing message", "error", err)
			}
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, msg *kafka.Message) error {
	handler, exists := c.handlers[msg.Topic]
	if !exists {
		slog.Warn("No handler for topic", "topic", msg.Topic)
		return nil
	}

	return handler(ctx, msg)
}

// Close closes the consumer
func (c *Consumer) Close() error {
	return c.reader.Close()
}

// WebSocketBroadcastConsumer consumes events and broadcasts to WebSocket hub
type WebSocketBroadcastConsumer struct {
	consumer      *Consumer
	broadcastFunc func(event *WebSocketBroadcastEvent)
}

// NewWebSocketBroadcastConsumer creates a consumer for WebSocket broadcasts
func NewWebSocketBroadcastConsumer(client *Client, groupID string, broadcastFunc func(event *WebSocketBroadcastEvent)) *WebSocketBroadcastConsumer {
	config := DefaultConsumerConfig(groupID, []string{TopicWebSocketBroadcast, TopicLobbyEvents})
	consumer := NewConsumer(client, config)

	wsc := &WebSocketBroadcastConsumer{
		consumer:      consumer,
		broadcastFunc: broadcastFunc,
	}

	consumer.RegisterHandler(TopicWebSocketBroadcast, wsc.handleWebSocketBroadcast)
	consumer.RegisterHandler(TopicLobbyEvents, wsc.handleLobbyEvent)

	return wsc
}

func (wsc *WebSocketBroadcastConsumer) handleWebSocketBroadcast(ctx context.Context, msg *kafka.Message) error {
	var event WebSocketBroadcastEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal broadcast event: %w", err)
	}

	wsc.broadcastFunc(&event)
	return nil
}

func (wsc *WebSocketBroadcastConsumer) handleLobbyEvent(ctx context.Context, msg *kafka.Message) error {
	var event LobbyEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal lobby event: %w", err)
	}

	// Convert lobby event to WebSocket broadcast
	broadcastEvent := &WebSocketBroadcastEvent{
		Type:      event.EventType,
		LobbyID:   &event.LobbyID,
		Payload:   event,
		Timestamp: event.CreatedAt,
	}

	wsc.broadcastFunc(broadcastEvent)
	return nil
}

// Start begins consuming WebSocket broadcast events
func (wsc *WebSocketBroadcastConsumer) Start(ctx context.Context) error {
	return wsc.consumer.Start(ctx)
}

// Close closes the consumer
func (wsc *WebSocketBroadcastConsumer) Close() error {
	return wsc.consumer.Close()
}

// MatchResultConsumer processes match result events
type MatchResultConsumer struct {
	consumer    *Consumer
	processFunc func(ctx context.Context, event *MatchEvent) error
}

// NewMatchResultConsumer creates a consumer for match results
func NewMatchResultConsumer(client *Client, groupID string, processFunc func(ctx context.Context, event *MatchEvent) error) *MatchResultConsumer {
	config := DefaultConsumerConfig(groupID, []string{TopicMatchesResults})
	consumer := NewConsumer(client, config)

	mrc := &MatchResultConsumer{
		consumer:    consumer,
		processFunc: processFunc,
	}

	consumer.RegisterHandler(TopicMatchesResults, mrc.handleMatchResult)

	return mrc
}

func (mrc *MatchResultConsumer) handleMatchResult(ctx context.Context, msg *kafka.Message) error {
	var event MatchEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal match result: %w", err)
	}

	return mrc.processFunc(ctx, &event)
}

// Start begins consuming match results
func (mrc *MatchResultConsumer) Start(ctx context.Context) error {
	return mrc.consumer.Start(ctx)
}

// Close closes the consumer
func (mrc *MatchResultConsumer) Close() error {
	return mrc.consumer.Close()
}

// HealthCheck verifies Kafka connectivity
func (c *Client) HealthCheck(ctx context.Context) error {
	conn, err := c.dialer.DialContext(ctx, "tcp", strings.Split(c.config.BootstrapServers, ",")[0])
	if err != nil {
		return fmt.Errorf("failed to connect to Kafka: %w", err)
	}
	defer conn.Close()

	_, err = conn.ApiVersions()
	if err != nil {
		return fmt.Errorf("failed to get API versions: %w", err)
	}

	return nil
}
