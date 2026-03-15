package messaging_entities

import (
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// DirectMessage represents a private message between two users
type DirectMessage struct {
	shared.BaseEntity `bson:",inline"`

	SenderID       uuid.UUID  `json:"sender_id" bson:"sender_id"`
	RecipientID    uuid.UUID  `json:"recipient_id" bson:"recipient_id"`
	ConversationID string     `json:"conversation_id" bson:"conversation_id"` // sorted pair: "uuid1:uuid2"
	Content        string     `json:"content" bson:"content"`
	Mentions       []Mention  `json:"mentions,omitempty" bson:"mentions,omitempty"`
	ReadAt         *time.Time `json:"read_at,omitempty" bson:"read_at,omitempty"`
	DeletedBySender    bool   `json:"deleted_by_sender,omitempty" bson:"deleted_by_sender,omitempty"`
	DeletedByRecipient bool   `json:"deleted_by_recipient,omitempty" bson:"deleted_by_recipient,omitempty"`
}

const (
	MaxDirectMessageLength = 4000
	MinDirectMessageLength = 1
)

// MakeConversationID creates a deterministic conversation ID from two user IDs
func MakeConversationID(userA, userB uuid.UUID) string {
	ids := []string{userA.String(), userB.String()}
	sort.Strings(ids)
	return ids[0] + ":" + ids[1]
}

// NewDirectMessage creates a new direct message
func NewDirectMessage(
	resourceOwner shared.ResourceOwner,
	senderID uuid.UUID,
	recipientID uuid.UUID,
	content string,
	mentions []Mention,
) (*DirectMessage, error) {
	if senderID == uuid.Nil {
		return nil, fmt.Errorf("sender_id is required")
	}
	if recipientID == uuid.Nil {
		return nil, fmt.Errorf("recipient_id is required")
	}
	if senderID == recipientID {
		return nil, fmt.Errorf("cannot send a message to yourself")
	}
	if len(content) < MinDirectMessageLength {
		return nil, fmt.Errorf("message content cannot be empty")
	}
	if len(content) > MaxDirectMessageLength {
		return nil, fmt.Errorf("message content cannot exceed %d characters", MaxDirectMessageLength)
	}

	dm := &DirectMessage{
		BaseEntity:     shared.NewEntity(resourceOwner),
		SenderID:       senderID,
		RecipientID:    recipientID,
		ConversationID: MakeConversationID(senderID, recipientID),
		Content:        content,
		Mentions:       mentions,
	}

	return dm, nil
}

// MarkAsRead marks the message as read
func (dm *DirectMessage) MarkAsRead() {
	if dm.ReadAt == nil {
		now := time.Now()
		dm.ReadAt = &now
		dm.UpdatedAt = now
	}
}

// DeleteForSender soft-deletes the message for the sender
func (dm *DirectMessage) DeleteForSender() {
	dm.DeletedBySender = true
	dm.UpdatedAt = time.Now()
}

// DeleteForRecipient soft-deletes the message for the recipient
func (dm *DirectMessage) DeleteForRecipient() {
	dm.DeletedByRecipient = true
	dm.UpdatedAt = time.Now()
}

// Conversation represents a DM thread summary
type Conversation struct {
	ConversationID string        `json:"conversation_id" bson:"conversation_id"`
	Participant    AuthorSummary `json:"participant" bson:"participant"`
	LastMessage    string        `json:"last_message" bson:"last_message"`
	LastMessageAt  time.Time     `json:"last_message_at" bson:"last_message_at"`
	UnreadCount    int           `json:"unread_count" bson:"unread_count"`
}
