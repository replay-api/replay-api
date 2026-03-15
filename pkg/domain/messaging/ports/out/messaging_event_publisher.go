package messaging_out

import (
	"context"

	"github.com/google/uuid"
	messaging_entities "github.com/replay-api/replay-api/pkg/domain/messaging/entities"
)

// MessagingEventPublisher defines the outbound port for publishing messaging events to Kafka
type MessagingEventPublisher interface {
	PublishCommentCreated(ctx context.Context, comment *messaging_entities.Comment) error
	PublishCommentEdited(ctx context.Context, comment *messaging_entities.Comment) error
	PublishCommentDeleted(ctx context.Context, commentID uuid.UUID, matchID uuid.UUID) error
	PublishCommentReaction(ctx context.Context, commentID uuid.UUID, emoji string, userID uuid.UUID, removed bool) error
	PublishDirectMessageSent(ctx context.Context, message *messaging_entities.DirectMessage) error
	PublishTeamMessageSent(ctx context.Context, message *messaging_entities.TeamMessage) error
	PublishMentionNotification(ctx context.Context, mention messaging_entities.Mention, sourceType string, sourceID uuid.UUID, authorName string) error
}
