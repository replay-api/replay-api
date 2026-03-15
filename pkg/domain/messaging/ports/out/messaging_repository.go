package messaging_out

import (
	"context"

	"github.com/google/uuid"
	messaging_entities "github.com/replay-api/replay-api/pkg/domain/messaging/entities"
)

// CommentRepository defines persistence operations for match comments
type CommentRepository interface {
	Save(ctx context.Context, comment *messaging_entities.Comment) error
	FindByID(ctx context.Context, id uuid.UUID) (*messaging_entities.Comment, error)
	FindByMatchID(ctx context.Context, matchID uuid.UUID, limit, offset int, sort string) ([]*messaging_entities.Comment, int64, error)
	FindReplies(ctx context.Context, parentID uuid.UUID, limit, offset int) ([]*messaging_entities.Comment, int64, error)
	Update(ctx context.Context, comment *messaging_entities.Comment) error
	Delete(ctx context.Context, id uuid.UUID) error
	IncrementReplyCount(ctx context.Context, parentID uuid.UUID, delta int) error
}

// DirectMessageRepository defines persistence operations for direct messages
type DirectMessageRepository interface {
	Save(ctx context.Context, message *messaging_entities.DirectMessage) error
	FindByID(ctx context.Context, id uuid.UUID) (*messaging_entities.DirectMessage, error)
	FindByConversation(ctx context.Context, conversationID string, limit, offset int) ([]*messaging_entities.DirectMessage, int64, error)
	ListConversations(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*messaging_entities.Conversation, error)
	MarkConversationRead(ctx context.Context, conversationID string, userID uuid.UUID) error
	Update(ctx context.Context, message *messaging_entities.DirectMessage) error
	CountUnread(ctx context.Context, userID uuid.UUID) (int64, error)
}

// TeamMessageRepository defines persistence operations for team messages
type TeamMessageRepository interface {
	Save(ctx context.Context, message *messaging_entities.TeamMessage) error
	FindByID(ctx context.Context, id uuid.UUID) (*messaging_entities.TeamMessage, error)
	FindByTeamAndChannel(ctx context.Context, teamID uuid.UUID, channel messaging_entities.ChannelType, limit, offset int) ([]*messaging_entities.TeamMessage, int64, error)
	ListTeamChannels(ctx context.Context, teamID uuid.UUID) ([]*messaging_entities.TeamChannelSummary, error)
}
