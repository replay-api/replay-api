package messaging_in

import (
	"context"
	"errors"

	"github.com/google/uuid"
	messaging_entities "github.com/replay-api/replay-api/pkg/domain/messaging/entities"
)

// CommentCommand defines write operations for match comments
type CommentCommand interface {
	CreateComment(ctx context.Context, cmd CreateCommentCommand) (*messaging_entities.Comment, error)
	EditComment(ctx context.Context, cmd EditCommentCommand) (*messaging_entities.Comment, error)
	DeleteComment(ctx context.Context, cmd DeleteCommentCommand) error
	ReactToComment(ctx context.Context, cmd ReactToCommentCommand) error
}

// CreateCommentCommand is the command to create a new comment on a match
type CreateCommentCommand struct {
	MatchID  uuid.UUID
	Content  string
	Mentions []messaging_entities.Mention
	ParentID *uuid.UUID
}

func (c *CreateCommentCommand) Validate() error {
	if c.MatchID == uuid.Nil {
		return errors.New("match_id is required")
	}
	if len(c.Content) < messaging_entities.MinCommentLength {
		return errors.New("comment content cannot be empty")
	}
	if len(c.Content) > messaging_entities.MaxCommentLength {
		return errors.New("comment content is too long")
	}
	return nil
}

// EditCommentCommand is the command to edit an existing comment
type EditCommentCommand struct {
	CommentID uuid.UUID
	Content   string
	Mentions  []messaging_entities.Mention
}

func (c *EditCommentCommand) Validate() error {
	if c.CommentID == uuid.Nil {
		return errors.New("comment_id is required")
	}
	if len(c.Content) < messaging_entities.MinCommentLength {
		return errors.New("comment content cannot be empty")
	}
	if len(c.Content) > messaging_entities.MaxCommentLength {
		return errors.New("comment content is too long")
	}
	return nil
}

// DeleteCommentCommand is the command to delete a comment
type DeleteCommentCommand struct {
	CommentID uuid.UUID
}

func (c *DeleteCommentCommand) Validate() error {
	if c.CommentID == uuid.Nil {
		return errors.New("comment_id is required")
	}
	return nil
}

// ReactToCommentCommand is the command to add/remove a reaction on a comment
type ReactToCommentCommand struct {
	CommentID uuid.UUID
	Emoji     string
	Remove    bool // true to remove the reaction
}

func (c *ReactToCommentCommand) Validate() error {
	if c.CommentID == uuid.Nil {
		return errors.New("comment_id is required")
	}
	if c.Emoji == "" {
		return errors.New("emoji is required")
	}
	return nil
}
