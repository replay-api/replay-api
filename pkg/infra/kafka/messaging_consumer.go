package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	kafka_go "github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"

	websocket "github.com/replay-api/replay-api/pkg/infra/websocket"
)

// MessagingConsumer handles messaging-related Kafka events
// and creates notifications + WebSocket broadcasts
type MessagingConsumer struct {
	consumer             *Consumer
	notificationColl     *mongo.Collection
	hub                  *websocket.WebSocketHub
}

// NewMessagingConsumer creates a consumer for messaging events
func NewMessagingConsumer(
	kafkaClient *Client,
	notificationColl *mongo.Collection,
	hub *websocket.WebSocketHub,
) *MessagingConsumer {
	topics := []string{
		TopicMentionNotification,
		TopicDirectMessageSent,
		TopicCommentCreated,
	}

	config := DefaultConsumerConfig("messaging-consumer", topics)
	consumer := NewConsumer(kafkaClient, config)

	mc := &MessagingConsumer{
		consumer:         consumer,
		notificationColl: notificationColl,
		hub:              hub,
	}

	consumer.RegisterHandler(TopicMentionNotification, mc.handleMentionNotification)
	consumer.RegisterHandler(TopicDirectMessageSent, mc.handleDirectMessageNotification)
	consumer.RegisterHandler(TopicCommentCreated, mc.handleCommentBroadcast)

	return mc
}

// Start begins consuming messaging events
func (mc *MessagingConsumer) Start(ctx context.Context) error {
	slog.InfoContext(ctx, "Starting messaging consumer")
	return mc.consumer.Start(ctx)
}

// handleMentionNotification creates a notification when a player is @mentioned
func (mc *MessagingConsumer) handleMentionNotification(ctx context.Context, msg *kafka_go.Message) error {
	var event MessagingEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal mention event", "error", err)
		return err
	}

	payload, ok := event.Payload.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid mention notification payload")
	}

	playerIDStr, _ := payload["player_id"].(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		return fmt.Errorf("invalid player_id in mention: %w", err)
	}

	sourceType, _ := payload["source_type"].(string)
	sourceIDStr, _ := payload["source_id"].(string)
	authorName, _ := payload["author_name"].(string)
	displayName, _ := payload["display_name"].(string)

	// Build notification
	title := "You were mentioned"
	message := fmt.Sprintf("@%s mentioned you", authorName)
	actionURL := ""

	switch sourceType {
	case "comment":
		title = "Mentioned in a comment"
		message = fmt.Sprintf("%s mentioned you in a match comment", authorName)
		actionURL = fmt.Sprintf("/matches?comment=%s", sourceIDStr)
	case "direct_message":
		title = "Mentioned in a message"
		message = fmt.Sprintf("You were mentioned by %s", displayName)
		actionURL = fmt.Sprintf("/messages?user=%s", sourceIDStr)
	case "team_message":
		title = "Mentioned in team chat"
		message = fmt.Sprintf("%s mentioned you in team chat", authorName)
		actionURL = fmt.Sprintf("/messages?team=%s", sourceIDStr)
	}

	notification := map[string]interface{}{
		"_id":        uuid.New(),
		"user_id":    playerID,
		"type":       "mention",
		"title":      title,
		"message":    message,
		"timestamp":  time.Now().Format(time.RFC3339),
		"read":       false,
		"action_url": actionURL,
		"metadata": map[string]interface{}{
			"source_type": sourceType,
			"source_id":   sourceIDStr,
			"author_name": authorName,
		},
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}

	if _, err := mc.notificationColl.InsertOne(ctx, notification); err != nil {
		slog.ErrorContext(ctx, "Failed to create mention notification", "error", err, "player_id", playerID)
		return err
	}

	// Broadcast via WebSocket to the mentioned user
	notifPayload, _ := json.Marshal(notification)
	mc.hub.BroadcastToUser(playerID, websocket.MessageTypeNotification, notifPayload)

	slog.InfoContext(ctx, "Mention notification created and broadcast",
		"player_id", playerID,
		"source_type", sourceType)
	return nil
}

// handleDirectMessageNotification creates a notification for DM recipients
func (mc *MessagingConsumer) handleDirectMessageNotification(ctx context.Context, msg *kafka_go.Message) error {
	var event MessagingEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal DM event", "error", err)
		return err
	}

	// Parse the DM from the event payload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	var dm struct {
		ID          uuid.UUID `json:"id"`
		SenderID    uuid.UUID `json:"sender_id"`
		RecipientID uuid.UUID `json:"recipient_id"`
		Content     string    `json:"content"`
	}
	if err := json.Unmarshal(payloadBytes, &dm); err != nil {
		return fmt.Errorf("failed to parse DM payload: %w", err)
	}

	// Create notification for recipient
	notification := map[string]interface{}{
		"_id":        uuid.New(),
		"user_id":    dm.RecipientID,
		"type":       "message",
		"title":      "New message",
		"message":    truncateMessage(dm.Content, 100),
		"timestamp":  time.Now().Format(time.RFC3339),
		"read":       false,
		"action_url": fmt.Sprintf("/messages?user=%s", dm.SenderID),
		"metadata": map[string]interface{}{
			"sender_id":  dm.SenderID,
			"message_id": dm.ID,
		},
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}

	if _, err := mc.notificationColl.InsertOne(ctx, notification); err != nil {
		slog.ErrorContext(ctx, "Failed to create DM notification", "error", err)
		return err
	}

	// Broadcast notification + DM content to receiver via WebSocket
	notifPayload, _ := json.Marshal(notification)
	mc.hub.BroadcastToUser(dm.RecipientID, websocket.MessageTypeNotification, notifPayload)

	// Also broadcast the DM itself for real-time chat
	dmPayload, _ := json.Marshal(event.Payload)
	mc.hub.BroadcastToUser(dm.RecipientID, "dm_received", dmPayload)

	slog.InfoContext(ctx, "DM notification created", "recipient", dm.RecipientID, "sender", dm.SenderID)
	return nil
}

// handleCommentBroadcast broadcasts new comments to match room subscribers
func (mc *MessagingConsumer) handleCommentBroadcast(ctx context.Context, msg *kafka_go.Message) error {
	var event MessagingEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal comment event", "error", err)
		return err
	}

	// Broadcast the comment to all clients watching this match
	// The key is the match_id, so we broadcast to the match room
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	matchIDStr := string(msg.Key)
	matchID, err := uuid.Parse(matchIDStr)
	if err != nil {
		return fmt.Errorf("invalid match_id key: %w", err)
	}

	// Broadcast to match room (using lobby room infrastructure — match rooms)
	mc.hub.BroadcastFromKafka("comment_created", &matchID, payloadBytes)

	slog.InfoContext(ctx, "Comment broadcast to match room", "match_id", matchIDStr)
	return nil
}

func truncateMessage(msg string, maxLen int) string {
	if len(msg) <= maxLen {
		return msg
	}
	return msg[:maxLen-3] + "..."
}
