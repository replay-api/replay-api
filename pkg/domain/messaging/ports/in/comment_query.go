package messaging_in

import (
	"context"

	"github.com/google/uuid"
	messaging_entities "github.com/replay-api/replay-api/pkg/domain/messaging/entities"
)

// CommentQuery defines read operations for match comments
type CommentQuery interface {
	ListMatchComments(ctx context.Context, query ListMatchCommentsQuery) (*CommentListResult, error)
	GetComment(ctx context.Context, commentID uuid.UUID) (*messaging_entities.Comment, error)
	GetCommentReplies(ctx context.Context, parentID uuid.UUID, limit, offset int) (*CommentListResult, error)
}

// ListMatchCommentsQuery defines the query parameters for listing comments
type ListMatchCommentsQuery struct {
	MatchID uuid.UUID
	Limit   int
	Offset  int
	Sort    string // "newest", "oldest", "most_reactions"
}

func (q *ListMatchCommentsQuery) Validate() error {
	if q.MatchID == uuid.Nil {
		return ErrMatchIDRequired
	}
	if q.Limit <= 0 {
		q.Limit = 20
	}
	if q.Limit > 100 {
		q.Limit = 100
	}
	if q.Offset < 0 {
		q.Offset = 0
	}
	if q.Sort == "" {
		q.Sort = "newest"
	}
	return nil
}

// CommentListResult contains paginated comment results
type CommentListResult struct {
	Comments   []*messaging_entities.Comment `json:"comments"`
	TotalCount int64                          `json:"total_count"`
	Limit      int                            `json:"limit"`
	Offset     int                            `json:"offset"`
}

var ErrMatchIDRequired = errMatchIDRequired{}

type errMatchIDRequired struct{}

func (e errMatchIDRequired) Error() string { return "match_id is required" }
