package messaging_in

import (
	"context"
	"errors"

	"github.com/google/uuid"
	messaging_entities "github.com/replay-api/replay-api/pkg/domain/messaging/entities"
)

// TeamMessageCommand defines write operations for team messages
type TeamMessageCommand interface {
	SendTeamMessage(ctx context.Context, cmd SendTeamMessageCommand) (*messaging_entities.TeamMessage, error)
}

// SendTeamMessageCommand is the command to send a team message
type SendTeamMessageCommand struct {
	TeamID   uuid.UUID
	Channel  messaging_entities.ChannelType
	Content  string
	Mentions []messaging_entities.Mention
}

func (c *SendTeamMessageCommand) Validate() error {
	if c.TeamID == uuid.Nil {
		return errors.New("team_id is required")
	}
	if !c.Channel.IsValid() {
		return errors.New("invalid channel type")
	}
	if len(c.Content) < messaging_entities.MinTeamMessageLength {
		return errors.New("message content cannot be empty")
	}
	if len(c.Content) > messaging_entities.MaxTeamMessageLength {
		return errors.New("message content is too long")
	}
	return nil
}

// TeamMessageQuery defines read operations for team messages
type TeamMessageQuery interface {
	ListTeamMessages(ctx context.Context, query ListTeamMessagesQuery) (*TeamMessageListResult, error)
	ListTeamChannels(ctx context.Context, teamID uuid.UUID) ([]*messaging_entities.TeamChannelSummary, error)
}

// ListTeamMessagesQuery defines query parameters for listing team messages
type ListTeamMessagesQuery struct {
	TeamID  uuid.UUID
	Channel messaging_entities.ChannelType
	Limit   int
	Offset  int
}

func (q *ListTeamMessagesQuery) Validate() error {
	if q.TeamID == uuid.Nil {
		return errors.New("team_id is required")
	}
	if q.Channel != "" && !q.Channel.IsValid() {
		return errors.New("invalid channel type")
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

// TeamMessageListResult contains paginated team message results
type TeamMessageListResult struct {
	Messages   []*messaging_entities.TeamMessage `json:"messages"`
	TotalCount int64                              `json:"total_count"`
	Limit      int                                `json:"limit"`
	Offset     int                                `json:"offset"`
}
