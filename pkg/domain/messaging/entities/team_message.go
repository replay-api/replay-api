package messaging_entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// ChannelType represents a team channel category
type ChannelType string

const (
	ChannelGeneral  ChannelType = "general"
	ChannelStrategy ChannelType = "strategy"
	ChannelComms    ChannelType = "comms"
)

// IsValid checks if the channel type is valid
func (ct ChannelType) IsValid() bool {
	switch ct {
	case ChannelGeneral, ChannelStrategy, ChannelComms:
		return true
	}
	return false
}

// TeamMessage represents a message sent in a team/squad channel
type TeamMessage struct {
	shared.BaseEntity `bson:",inline"`

	SenderID uuid.UUID     `json:"sender_id" bson:"sender_id"`
	TeamID   uuid.UUID     `json:"team_id" bson:"team_id"`
	Channel  ChannelType   `json:"channel" bson:"channel"`
	Content  string        `json:"content" bson:"content"`
	Mentions []Mention     `json:"mentions,omitempty" bson:"mentions,omitempty"`
	Sender   AuthorSummary `json:"sender" bson:"sender"`
}

const (
	MaxTeamMessageLength = 4000
	MinTeamMessageLength = 1
)

// NewTeamMessage creates a new team message
func NewTeamMessage(
	resourceOwner shared.ResourceOwner,
	senderID uuid.UUID,
	teamID uuid.UUID,
	channel ChannelType,
	content string,
	mentions []Mention,
	sender AuthorSummary,
) (*TeamMessage, error) {
	if senderID == uuid.Nil {
		return nil, fmt.Errorf("sender_id is required")
	}
	if teamID == uuid.Nil {
		return nil, fmt.Errorf("team_id is required")
	}
	if !channel.IsValid() {
		return nil, fmt.Errorf("invalid channel type: %s", channel)
	}
	if len(content) < MinTeamMessageLength {
		return nil, fmt.Errorf("message content cannot be empty")
	}
	if len(content) > MaxTeamMessageLength {
		return nil, fmt.Errorf("message content cannot exceed %d characters", MaxTeamMessageLength)
	}

	tm := &TeamMessage{
		BaseEntity: shared.NewEntity(resourceOwner),
		SenderID:   senderID,
		TeamID:     teamID,
		Channel:    channel,
		Content:    content,
		Mentions:   mentions,
		Sender:     sender,
	}

	return tm, nil
}

// TeamChannelSummary represents a summary of a team channel
type TeamChannelSummary struct {
	TeamID      uuid.UUID   `json:"team_id" bson:"team_id"`
	TeamName    string      `json:"team_name" bson:"team_name"`
	Channel     ChannelType `json:"channel" bson:"channel"`
	LastMessage string      `json:"last_message" bson:"last_message"`
	LastAt      time.Time   `json:"last_at" bson:"last_at"`
	UnreadCount int         `json:"unread_count" bson:"unread_count"`
}
