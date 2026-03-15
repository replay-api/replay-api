package messaging_in

import (
	"context"
	"errors"

	"github.com/google/uuid"
	messaging_entities "github.com/replay-api/replay-api/pkg/domain/messaging/entities"
)

// DirectMessageCommand defines write operations for direct messages
type DirectMessageCommand interface {
	SendDirectMessage(ctx context.Context, cmd SendDirectMessageCommand) (*messaging_entities.DirectMessage, error)
	MarkConversationRead(ctx context.Context, cmd MarkConversationReadCommand) error
	DeleteMessage(ctx context.Context, messageID uuid.UUID) error
}

// SendDirectMessageCommand is the command to send a direct message
type SendDirectMessageCommand struct {
	RecipientID uuid.UUID
	Content     string
	Mentions    []messaging_entities.Mention
}

func (c *SendDirectMessageCommand) Validate() error {
	if c.RecipientID == uuid.Nil {
		return errors.New("recipient_id is required")
	}
	if len(c.Content) < messaging_entities.MinDirectMessageLength {
		return errors.New("message content cannot be empty")
	}
	if len(c.Content) > messaging_entities.MaxDirectMessageLength {
		return errors.New("message content is too long")
	}
	return nil
}

// MarkConversationReadCommand marks all messages in a conversation as read
type MarkConversationReadCommand struct {
	OtherUserID uuid.UUID
}

func (c *MarkConversationReadCommand) Validate() error {
	if c.OtherUserID == uuid.Nil {
		return errors.New("other_user_id is required")
	}
	return nil
}

// DirectMessageQuery defines read operations for direct messages
type DirectMessageQuery interface {
	ListConversations(ctx context.Context, query ListConversationsQuery) ([]*messaging_entities.Conversation, error)
	GetConversation(ctx context.Context, query GetConversationQuery) (*DirectMessageListResult, error)
}

// ListConversationsQuery defines query parameters for listing conversations
type ListConversationsQuery struct {
	Limit  int
	Offset int
}

func (q *ListConversationsQuery) Validate() error {
	if q.Limit <= 0 {
		q.Limit = 20
	}
	if q.Limit > 50 {
		q.Limit = 50
	}
	if q.Offset < 0 {
		q.Offset = 0
	}
	return nil
}

// GetConversationQuery defines query parameters for getting a DM conversation
type GetConversationQuery struct {
	OtherUserID uuid.UUID
	Limit       int
	Offset      int
}

func (q *GetConversationQuery) Validate() error {
	if q.OtherUserID == uuid.Nil {
		return errors.New("other_user_id is required")
	}
	if q.Limit <= 0 {
		q.Limit = 50
	}
	if q.Limit > 100 {
		q.Limit = 100
	}
	if q.Offset < 0 {
		q.Offset = 0
	}
	return nil
}

// DirectMessageListResult contains paginated DM results
type DirectMessageListResult struct {
	Messages   []*messaging_entities.DirectMessage `json:"messages"`
	TotalCount int64                                `json:"total_count"`
	Limit      int                                  `json:"limit"`
	Offset     int                                  `json:"offset"`
}
