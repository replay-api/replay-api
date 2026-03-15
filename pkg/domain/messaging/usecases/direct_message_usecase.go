package messaging_usecases

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	messaging_entities "github.com/replay-api/replay-api/pkg/domain/messaging/entities"
	messaging_in "github.com/replay-api/replay-api/pkg/domain/messaging/ports/in"
	messaging_out "github.com/replay-api/replay-api/pkg/domain/messaging/ports/out"
)

// DirectMessageCommandUseCase implements the DirectMessageCommand inbound port
type DirectMessageCommandUseCase struct {
	dmRepo         messaging_out.DirectMessageRepository
	eventPublisher messaging_out.MessagingEventPublisher
}

// NewDirectMessageCommandUseCase creates a new DirectMessageCommandUseCase
func NewDirectMessageCommandUseCase(
	dmRepo messaging_out.DirectMessageRepository,
	eventPublisher messaging_out.MessagingEventPublisher,
) messaging_in.DirectMessageCommand {
	return &DirectMessageCommandUseCase{
		dmRepo:         dmRepo,
		eventPublisher: eventPublisher,
	}
}

func (uc *DirectMessageCommandUseCase) SendDirectMessage(ctx context.Context, cmd messaging_in.SendDirectMessageCommand) (*messaging_entities.DirectMessage, error) {
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, shared.NewErrUnauthorized()
	}

	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	rxn := shared.GetResourceOwner(ctx)
	senderID := rxn.UserID

	dm, err := messaging_entities.NewDirectMessage(rxn, senderID, cmd.RecipientID, cmd.Content, cmd.Mentions)
	if err != nil {
		return nil, err
	}

	if err := uc.dmRepo.Save(ctx, dm); err != nil {
		slog.ErrorContext(ctx, "Failed to save direct message", "error", err)
		return nil, err
	}

	// Publish event for real-time delivery + notification
	if err := uc.eventPublisher.PublishDirectMessageSent(ctx, dm); err != nil {
		slog.WarnContext(ctx, "Failed to publish DM event", "error", err)
	}

	// Publish mention notifications
	for _, mention := range cmd.Mentions {
		if err := uc.eventPublisher.PublishMentionNotification(ctx, mention, "direct_message", dm.ID, ""); err != nil {
			slog.WarnContext(ctx, "Failed to publish mention notification", "error", err)
		}
	}

	slog.InfoContext(ctx, "Direct message sent", "dm_id", dm.ID, "sender", senderID, "recipient", cmd.RecipientID)
	return dm, nil
}

func (uc *DirectMessageCommandUseCase) MarkConversationRead(ctx context.Context, cmd messaging_in.MarkConversationReadCommand) error {
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return shared.NewErrUnauthorized()
	}

	if err := cmd.Validate(); err != nil {
		return err
	}

	rxn := shared.GetResourceOwner(ctx)
	conversationID := messaging_entities.MakeConversationID(rxn.UserID, cmd.OtherUserID)

	if err := uc.dmRepo.MarkConversationRead(ctx, conversationID, rxn.UserID); err != nil {
		return err
	}

	slog.InfoContext(ctx, "Conversation marked as read", "conversation_id", conversationID)
	return nil
}

func (uc *DirectMessageCommandUseCase) DeleteMessage(ctx context.Context, messageID uuid.UUID) error {
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return shared.NewErrUnauthorized()
	}

	dm, err := uc.dmRepo.FindByID(ctx, messageID)
	if err != nil {
		return err
	}

	rxn := shared.GetResourceOwner(ctx)
	switch rxn.UserID {
	case dm.SenderID:
		dm.DeleteForSender()
	case dm.RecipientID:
		dm.DeleteForRecipient()
	default:
		return shared.NewErrForbidden()
	}

	return uc.dmRepo.Update(ctx, dm)
}

// DirectMessageQueryUseCase implements the DirectMessageQuery inbound port
type DirectMessageQueryUseCase struct {
	dmRepo messaging_out.DirectMessageRepository
}

// NewDirectMessageQueryUseCase creates a new DirectMessageQueryUseCase
func NewDirectMessageQueryUseCase(
	dmRepo messaging_out.DirectMessageRepository,
) messaging_in.DirectMessageQuery {
	return &DirectMessageQueryUseCase{
		dmRepo: dmRepo,
	}
}

func (uc *DirectMessageQueryUseCase) ListConversations(ctx context.Context, query messaging_in.ListConversationsQuery) ([]*messaging_entities.Conversation, error) {
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, shared.NewErrUnauthorized()
	}

	if err := query.Validate(); err != nil {
		return nil, err
	}

	rxn := shared.GetResourceOwner(ctx)
	return uc.dmRepo.ListConversations(ctx, rxn.UserID, query.Limit, query.Offset)
}

func (uc *DirectMessageQueryUseCase) GetConversation(ctx context.Context, query messaging_in.GetConversationQuery) (*messaging_in.DirectMessageListResult, error) {
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, shared.NewErrUnauthorized()
	}

	if err := query.Validate(); err != nil {
		return nil, err
	}

	rxn := shared.GetResourceOwner(ctx)
	conversationID := messaging_entities.MakeConversationID(rxn.UserID, query.OtherUserID)

	messages, total, err := uc.dmRepo.FindByConversation(ctx, conversationID, query.Limit, query.Offset)
	if err != nil {
		return nil, err
	}

	return &messaging_in.DirectMessageListResult{
		Messages:   messages,
		TotalCount: total,
		Limit:      query.Limit,
		Offset:     query.Offset,
	}, nil
}
