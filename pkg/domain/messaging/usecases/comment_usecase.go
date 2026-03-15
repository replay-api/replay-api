package messaging_usecases

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	messaging_entities "github.com/replay-api/replay-api/pkg/domain/messaging/entities"
	messaging_in "github.com/replay-api/replay-api/pkg/domain/messaging/ports/in"
	messaging_out "github.com/replay-api/replay-api/pkg/domain/messaging/ports/out"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// CommentCommandUseCase implements the CommentCommand inbound port
type CommentCommandUseCase struct {
	commentRepo    messaging_out.CommentRepository
	eventPublisher messaging_out.MessagingEventPublisher
}

// NewCommentCommandUseCase creates a new CommentCommandUseCase
func NewCommentCommandUseCase(
	commentRepo messaging_out.CommentRepository,
	eventPublisher messaging_out.MessagingEventPublisher,
) messaging_in.CommentCommand {
	return &CommentCommandUseCase{
		commentRepo:    commentRepo,
		eventPublisher: eventPublisher,
	}
}

func (uc *CommentCommandUseCase) CreateComment(ctx context.Context, cmd messaging_in.CreateCommentCommand) (*messaging_entities.Comment, error) {
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, shared.NewErrUnauthorized()
	}

	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	// Extract author info from resource owner context
	rxn := shared.GetResourceOwner(ctx)
	userID := rxn.UserID
	displayName := userID.String()
	avatarURL := ""

	author := messaging_entities.AuthorSummary{
		ID:          userID,
		DisplayName: displayName,
		AvatarURL:   avatarURL,
	}

	comment, err := messaging_entities.NewComment(rxn, cmd.MatchID, author, cmd.Content, cmd.Mentions, cmd.ParentID)
	if err != nil {
		return nil, err
	}

	if err := uc.commentRepo.Save(ctx, comment); err != nil {
		slog.ErrorContext(ctx, "Failed to save comment", "error", err, "match_id", cmd.MatchID)
		return nil, err
	}

	// If this is a reply, increment the parent's reply count
	if cmd.ParentID != nil {
		if err := uc.commentRepo.IncrementReplyCount(ctx, *cmd.ParentID, 1); err != nil {
			slog.WarnContext(ctx, "Failed to increment reply count", "error", err, "parent_id", cmd.ParentID)
		}
	}

	// Publish event for real-time WebSocket broadcast
	if err := uc.eventPublisher.PublishCommentCreated(ctx, comment); err != nil {
		slog.WarnContext(ctx, "Failed to publish comment created event", "error", err)
	}

	// Publish mention notifications
	for _, mention := range cmd.Mentions {
		if err := uc.eventPublisher.PublishMentionNotification(ctx, mention, "comment", comment.ID, author.DisplayName); err != nil {
			slog.WarnContext(ctx, "Failed to publish mention notification", "error", err, "player_id", mention.PlayerID)
		}
	}

	slog.InfoContext(ctx, "Comment created", "comment_id", comment.ID, "match_id", cmd.MatchID, "author_id", userID)
	return comment, nil
}

func (uc *CommentCommandUseCase) EditComment(ctx context.Context, cmd messaging_in.EditCommentCommand) (*messaging_entities.Comment, error) {
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, shared.NewErrUnauthorized()
	}

	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	comment, err := uc.commentRepo.FindByID(ctx, cmd.CommentID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	rxn := shared.GetResourceOwner(ctx)
	if comment.Author.ID != rxn.UserID {
		return nil, shared.NewErrForbidden()
	}

	if err := comment.Edit(cmd.Content, cmd.Mentions); err != nil {
		return nil, err
	}

	if err := uc.commentRepo.Update(ctx, comment); err != nil {
		return nil, err
	}

	if err := uc.eventPublisher.PublishCommentEdited(ctx, comment); err != nil {
		slog.WarnContext(ctx, "Failed to publish comment edited event", "error", err)
	}

	// Publish notifications for any NEW mentions
	for _, mention := range cmd.Mentions {
		if err := uc.eventPublisher.PublishMentionNotification(ctx, mention, "comment", comment.ID, comment.Author.DisplayName); err != nil {
			slog.WarnContext(ctx, "Failed to publish mention notification", "error", err)
		}
	}

	slog.InfoContext(ctx, "Comment edited", "comment_id", cmd.CommentID)
	return comment, nil
}

func (uc *CommentCommandUseCase) DeleteComment(ctx context.Context, cmd messaging_in.DeleteCommentCommand) error {
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return shared.NewErrUnauthorized()
	}

	if err := cmd.Validate(); err != nil {
		return err
	}

	comment, err := uc.commentRepo.FindByID(ctx, cmd.CommentID)
	if err != nil {
		return err
	}

	// Verify ownership
	rxn := shared.GetResourceOwner(ctx)
	if comment.Author.ID != rxn.UserID {
		return shared.NewErrForbidden()
	}

	comment.SoftDelete()

	if err := uc.commentRepo.Update(ctx, comment); err != nil {
		return err
	}

	// Decrement parent reply count if this is a reply
	if comment.ParentID != nil {
		if err := uc.commentRepo.IncrementReplyCount(ctx, *comment.ParentID, -1); err != nil {
			slog.WarnContext(ctx, "Failed to decrement reply count", "error", err)
		}
	}

	if err := uc.eventPublisher.PublishCommentDeleted(ctx, cmd.CommentID, comment.MatchID); err != nil {
		slog.WarnContext(ctx, "Failed to publish comment deleted event", "error", err)
	}

	slog.InfoContext(ctx, "Comment deleted", "comment_id", cmd.CommentID)
	return nil
}

func (uc *CommentCommandUseCase) ReactToComment(ctx context.Context, cmd messaging_in.ReactToCommentCommand) error {
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return shared.NewErrUnauthorized()
	}

	if err := cmd.Validate(); err != nil {
		return err
	}

	comment, err := uc.commentRepo.FindByID(ctx, cmd.CommentID)
	if err != nil {
		return err
	}

	rxn := shared.GetResourceOwner(ctx)
	userID := rxn.UserID.String()

	if cmd.Remove {
		comment.RemoveReaction(cmd.Emoji, userID)
	} else {
		comment.AddReaction(cmd.Emoji, userID)
	}

	if err := uc.commentRepo.Update(ctx, comment); err != nil {
		return err
	}

	if err := uc.eventPublisher.PublishCommentReaction(ctx, cmd.CommentID, cmd.Emoji, rxn.UserID, cmd.Remove); err != nil {
		slog.WarnContext(ctx, "Failed to publish reaction event", "error", err)
	}

	return nil
}

// CommentQueryUseCase implements the CommentQuery inbound port
type CommentQueryUseCase struct {
	commentRepo messaging_out.CommentRepository
}

// NewCommentQueryUseCase creates a new CommentQueryUseCase
func NewCommentQueryUseCase(
	commentRepo messaging_out.CommentRepository,
) messaging_in.CommentQuery {
	return &CommentQueryUseCase{
		commentRepo: commentRepo,
	}
}

func (uc *CommentQueryUseCase) ListMatchComments(ctx context.Context, query messaging_in.ListMatchCommentsQuery) (*messaging_in.CommentListResult, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}

	comments, total, err := uc.commentRepo.FindByMatchID(ctx, query.MatchID, query.Limit, query.Offset, query.Sort)
	if err != nil {
		return nil, err
	}

	return &messaging_in.CommentListResult{
		Comments:   comments,
		TotalCount: total,
		Limit:      query.Limit,
		Offset:     query.Offset,
	}, nil
}

func (uc *CommentQueryUseCase) GetComment(ctx context.Context, commentID uuid.UUID) (*messaging_entities.Comment, error) {
	return uc.commentRepo.FindByID(ctx, commentID)
}

func (uc *CommentQueryUseCase) GetCommentReplies(ctx context.Context, parentID uuid.UUID, limit, offset int) (*messaging_in.CommentListResult, error) {
	comments, total, err := uc.commentRepo.FindReplies(ctx, parentID, limit, offset)
	if err != nil {
		return nil, err
	}

	return &messaging_in.CommentListResult{
		Comments:   comments,
		TotalCount: total,
		Limit:      limit,
		Offset:     offset,
	}, nil
}
