package kafka

import (
	"context"

	"github.com/google/uuid"
	messaging_entities "github.com/replay-api/replay-api/pkg/domain/messaging/entities"
	messaging_out "github.com/replay-api/replay-api/pkg/domain/messaging/ports/out"
)

// NoopMessagingEventPublisher is a no-op messaging publisher for local and test environments without Kafka.
type NoopMessagingEventPublisher struct{}

// NewNoopMessagingEventPublisher creates a no-op messaging event publisher.
func NewNoopMessagingEventPublisher() messaging_out.MessagingEventPublisher {
	return &NoopMessagingEventPublisher{}
}

func (p *NoopMessagingEventPublisher) PublishCommentCreated(context.Context, *messaging_entities.Comment) error {
	return nil
}

func (p *NoopMessagingEventPublisher) PublishCommentEdited(context.Context, *messaging_entities.Comment) error {
	return nil
}

func (p *NoopMessagingEventPublisher) PublishCommentDeleted(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func (p *NoopMessagingEventPublisher) PublishCommentReaction(context.Context, uuid.UUID, string, uuid.UUID, bool) error {
	return nil
}

func (p *NoopMessagingEventPublisher) PublishDirectMessageSent(context.Context, *messaging_entities.DirectMessage) error {
	return nil
}

func (p *NoopMessagingEventPublisher) PublishTeamMessageSent(context.Context, *messaging_entities.TeamMessage) error {
	return nil
}

func (p *NoopMessagingEventPublisher) PublishMentionNotification(context.Context, messaging_entities.Mention, string, uuid.UUID, string) error {
	return nil
}
