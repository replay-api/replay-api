package messaging_entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// CommentStatus represents the lifecycle of a comment
type CommentStatus string

const (
	CommentStatusActive  CommentStatus = "active"
	CommentStatusEdited  CommentStatus = "edited"
	CommentStatusDeleted CommentStatus = "deleted"
)

// Mention represents a player mention within a message or comment
type Mention struct {
	PlayerID    uuid.UUID `json:"player_id" bson:"player_id"`
	DisplayName string    `json:"display_name" bson:"display_name"`
	Offset      int       `json:"offset" bson:"offset"`
	Length      int       `json:"length" bson:"length"`
}

// AuthorSummary holds minimal author info embedded in the comment
type AuthorSummary struct {
	ID          uuid.UUID `json:"id" bson:"id"`
	DisplayName string    `json:"display_name" bson:"display_name"`
	AvatarURL   string    `json:"avatar_url,omitempty" bson:"avatar_url,omitempty"`
	Slug        string    `json:"slug,omitempty" bson:"slug,omitempty"`
}

// Comment represents a comment on a match
type Comment struct {
	shared.BaseEntity `bson:",inline"`

	MatchID    uuid.UUID           `json:"match_id" bson:"match_id"`
	Author     AuthorSummary       `json:"author" bson:"author"`
	Content    string              `json:"content" bson:"content"`
	Mentions   []Mention           `json:"mentions,omitempty" bson:"mentions,omitempty"`
	ParentID   *uuid.UUID          `json:"parent_id,omitempty" bson:"parent_id,omitempty"`
	Reactions  map[string][]string `json:"reactions,omitempty" bson:"reactions,omitempty"` // emoji -> list of user IDs
	Status     CommentStatus       `json:"status" bson:"status"`
	EditedAt   *time.Time          `json:"edited_at,omitempty" bson:"edited_at,omitempty"`
	ReplyCount int                 `json:"reply_count" bson:"reply_count"`
}

const (
	MaxCommentLength = 2000
	MinCommentLength = 1
)

// NewComment creates a new comment on a match
func NewComment(
	resourceOwner shared.ResourceOwner,
	matchID uuid.UUID,
	author AuthorSummary,
	content string,
	mentions []Mention,
	parentID *uuid.UUID,
) (*Comment, error) {
	if matchID == uuid.Nil {
		return nil, fmt.Errorf("match_id is required")
	}
	if author.ID == uuid.Nil {
		return nil, fmt.Errorf("author_id is required")
	}
	if len(content) < MinCommentLength {
		return nil, fmt.Errorf("comment content cannot be empty")
	}
	if len(content) > MaxCommentLength {
		return nil, fmt.Errorf("comment content cannot exceed %d characters", MaxCommentLength)
	}

	comment := &Comment{
		BaseEntity: shared.NewEntity(resourceOwner),
		MatchID:    matchID,
		Author:     author,
		Content:    content,
		Mentions:   mentions,
		ParentID:   parentID,
		Reactions:  make(map[string][]string),
		Status:     CommentStatusActive,
		ReplyCount: 0,
	}

	return comment, nil
}

// Edit updates the comment content
func (c *Comment) Edit(content string, mentions []Mention) error {
	if c.Status == CommentStatusDeleted {
		return fmt.Errorf("cannot edit a deleted comment")
	}
	if len(content) < MinCommentLength {
		return fmt.Errorf("comment content cannot be empty")
	}
	if len(content) > MaxCommentLength {
		return fmt.Errorf("comment content cannot exceed %d characters", MaxCommentLength)
	}

	now := time.Now()
	c.Content = content
	c.Mentions = mentions
	c.Status = CommentStatusEdited
	c.EditedAt = &now
	c.UpdatedAt = now

	return nil
}

// SoftDelete marks the comment as deleted
func (c *Comment) SoftDelete() {
	now := time.Now()
	c.Status = CommentStatusDeleted
	c.Content = "[deleted]"
	c.Mentions = nil
	c.UpdatedAt = now
}

// AddReaction adds a reaction emoji from a user
func (c *Comment) AddReaction(emoji string, userID string) {
	if c.Reactions == nil {
		c.Reactions = make(map[string][]string)
	}
	// Check if user already reacted with this emoji
	for _, uid := range c.Reactions[emoji] {
		if uid == userID {
			return
		}
	}
	c.Reactions[emoji] = append(c.Reactions[emoji], userID)
	c.UpdatedAt = time.Now()
}

// RemoveReaction removes a reaction emoji from a user
func (c *Comment) RemoveReaction(emoji string, userID string) {
	if c.Reactions == nil {
		return
	}
	users := c.Reactions[emoji]
	for i, uid := range users {
		if uid == userID {
			c.Reactions[emoji] = append(users[:i], users[i+1:]...)
			if len(c.Reactions[emoji]) == 0 {
				delete(c.Reactions, emoji)
			}
			c.UpdatedAt = time.Now()
			return
		}
	}
}
