package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	messaging_entities "github.com/replay-api/replay-api/pkg/domain/messaging/entities"
	messaging_out "github.com/replay-api/replay-api/pkg/domain/messaging/ports/out"
	kafka_go "github.com/segmentio/kafka-go"
)

// MessagingEventPublisherAdapter implements the MessagingEventPublisher outbound port using Kafka
type MessagingEventPublisherAdapter struct {
	client *Client
}

// NewMessagingEventPublisherAdapter creates a new Kafka-based messaging event publisher
func NewMessagingEventPublisherAdapter(client *Client) messaging_out.MessagingEventPublisher {
	return &MessagingEventPublisherAdapter{
		client: client,
	}
}

// MessagingEvent is the envelope for messaging events published to Kafka
type MessagingEvent struct {
	EventType string      `json:"event_type"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

func (p *MessagingEventPublisherAdapter) publish(ctx context.Context, topic string, key string, event MessagingEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal messaging event: %w", err)
	}

	msg := kafka_go.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
	}

	if err := p.client.GetWriter(topic).WriteMessages(ctx, msg); err != nil {
		slog.ErrorContext(ctx, "Failed to publish messaging event",
			"topic", topic,
			"event_type", event.EventType,
			"error", err)
		return fmt.Errorf("failed to publish messaging event: %w", err)
	}

	slog.InfoContext(ctx, "Published messaging event",
		"topic", topic,
		"event_type", event.EventType,
		"key", key)
	return nil
}

func (p *MessagingEventPublisherAdapter) PublishCommentCreated(ctx context.Context, comment *messaging_entities.Comment) error {
	return p.publish(ctx, TopicCommentCreated, comment.MatchID.String(), MessagingEvent{
		EventType: EventTypeCommentCreated,
		Timestamp: time.Now(),
		Payload:   comment,
	})
}

func (p *MessagingEventPublisherAdapter) PublishCommentEdited(ctx context.Context, comment *messaging_entities.Comment) error {
	return p.publish(ctx, TopicCommentEdited, comment.MatchID.String(), MessagingEvent{
		EventType: EventTypeCommentEdited,
		Timestamp: time.Now(),
		Payload:   comment,
	})
}

func (p *MessagingEventPublisherAdapter) PublishCommentDeleted(ctx context.Context, commentID uuid.UUID, matchID uuid.UUID) error {
	return p.publish(ctx, TopicCommentDeleted, matchID.String(), MessagingEvent{
		EventType: EventTypeCommentDeleted,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"comment_id": commentID,
			"match_id":   matchID,
		},
	})
}

func (p *MessagingEventPublisherAdapter) PublishCommentReaction(ctx context.Context, commentID uuid.UUID, emoji string, userID uuid.UUID, removed bool) error {
	return p.publish(ctx, TopicCommentReaction, commentID.String(), MessagingEvent{
		EventType: EventTypeCommentReaction,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"comment_id": commentID,
			"emoji":      emoji,
			"user_id":    userID,
			"removed":    removed,
		},
	})
}

func (p *MessagingEventPublisherAdapter) PublishDirectMessageSent(ctx context.Context, message *messaging_entities.DirectMessage) error {
	return p.publish(ctx, TopicDirectMessageSent, message.ConversationID, MessagingEvent{
		EventType: EventTypeDirectMessageSent,
		Timestamp: time.Now(),
		Payload:   message,
	})
}

func (p *MessagingEventPublisherAdapter) PublishTeamMessageSent(ctx context.Context, message *messaging_entities.TeamMessage) error {
	return p.publish(ctx, TopicTeamMessageSent, message.TeamID.String(), MessagingEvent{
		EventType: EventTypeTeamMessageSent,
		Timestamp: time.Now(),
		Payload:   message,
	})
}

func (p *MessagingEventPublisherAdapter) PublishMentionNotification(ctx context.Context, mention messaging_entities.Mention, sourceType string, sourceID uuid.UUID, authorName string) error {
	return p.publish(ctx, TopicMentionNotification, mention.PlayerID.String(), MessagingEvent{
		EventType: EventTypeMentionNotification,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"player_id":    mention.PlayerID,
			"display_name": mention.DisplayName,
			"source_type":  sourceType,
			"source_id":    sourceID,
			"author_name":  authorName,
		},
	})
}
