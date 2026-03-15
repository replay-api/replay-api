package messaging_usecases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	messaging_entities "github.com/replay-api/replay-api/pkg/domain/messaging/entities"
	messaging_in "github.com/replay-api/replay-api/pkg/domain/messaging/ports/in"
	messaging_out "github.com/replay-api/replay-api/pkg/domain/messaging/ports/out"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
)

// TeamMessageCommandUseCase implements the TeamMessageCommand inbound port
type TeamMessageCommandUseCase struct {
	teamMsgRepo    messaging_out.TeamMessageRepository
	squadReader    squad_in.SquadReader
	eventPublisher messaging_out.MessagingEventPublisher
}

// NewTeamMessageCommandUseCase creates a new TeamMessageCommandUseCase
func NewTeamMessageCommandUseCase(
	teamMsgRepo messaging_out.TeamMessageRepository,
	squadReader squad_in.SquadReader,
	eventPublisher messaging_out.MessagingEventPublisher,
) messaging_in.TeamMessageCommand {
	return &TeamMessageCommandUseCase{
		teamMsgRepo:    teamMsgRepo,
		squadReader:    squadReader,
		eventPublisher: eventPublisher,
	}
}

func (uc *TeamMessageCommandUseCase) SendTeamMessage(ctx context.Context, cmd messaging_in.SendTeamMessageCommand) (*messaging_entities.TeamMessage, error) {
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, shared.NewErrUnauthorized()
	}

	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	rxn := shared.GetResourceOwner(ctx)
	senderID := rxn.UserID

	// Verify the sender is a member of the team
	squad, err := uc.getSquad(ctx, cmd.TeamID)
	if err != nil {
		return nil, err
	}

	isMember := false
	for _, m := range squad.Membership {
		if m.PlayerProfileID == senderID {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, shared.NewErrForbidden()
	}

	displayName := senderID.String()
	avatarURL := ""

	sender := messaging_entities.AuthorSummary{
		ID:          senderID,
		DisplayName: displayName,
		AvatarURL:   avatarURL,
	}

	teamMsg, err := messaging_entities.NewTeamMessage(rxn, senderID, cmd.TeamID, cmd.Channel, cmd.Content, cmd.Mentions, sender)
	if err != nil {
		return nil, err
	}

	if err := uc.teamMsgRepo.Save(ctx, teamMsg); err != nil {
		slog.ErrorContext(ctx, "Failed to save team message", "error", err)
		return nil, err
	}

	// Publish event for real-time delivery
	if err := uc.eventPublisher.PublishTeamMessageSent(ctx, teamMsg); err != nil {
		slog.WarnContext(ctx, "Failed to publish team message event", "error", err)
	}

	// Publish mention notifications
	for _, mention := range cmd.Mentions {
		if err := uc.eventPublisher.PublishMentionNotification(ctx, mention, "team_message", teamMsg.ID, sender.DisplayName); err != nil {
			slog.WarnContext(ctx, "Failed to publish mention notification", "error", err)
		}
	}

	slog.InfoContext(ctx, "Team message sent", "msg_id", teamMsg.ID, "team_id", cmd.TeamID, "channel", cmd.Channel)
	return teamMsg, nil
}

// TeamMessageQueryUseCase implements the TeamMessageQuery inbound port
type TeamMessageQueryUseCase struct {
	teamMsgRepo messaging_out.TeamMessageRepository
}

// NewTeamMessageQueryUseCase creates a new TeamMessageQueryUseCase
func NewTeamMessageQueryUseCase(
	teamMsgRepo messaging_out.TeamMessageRepository,
) messaging_in.TeamMessageQuery {
	return &TeamMessageQueryUseCase{
		teamMsgRepo: teamMsgRepo,
	}
}

func (uc *TeamMessageQueryUseCase) ListTeamMessages(ctx context.Context, query messaging_in.ListTeamMessagesQuery) (*messaging_in.TeamMessageListResult, error) {
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, shared.NewErrUnauthorized()
	}

	if err := query.Validate(); err != nil {
		return nil, err
	}

	messages, total, err := uc.teamMsgRepo.FindByTeamAndChannel(ctx, query.TeamID, query.Channel, query.Limit, query.Offset)
	if err != nil {
		return nil, err
	}

	return &messaging_in.TeamMessageListResult{
		Messages:   messages,
		TotalCount: total,
		Limit:      query.Limit,
		Offset:     query.Offset,
	}, nil
}

func (uc *TeamMessageQueryUseCase) ListTeamChannels(ctx context.Context, teamID uuid.UUID) ([]*messaging_entities.TeamChannelSummary, error) {
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, shared.NewErrUnauthorized()
	}

	return uc.teamMsgRepo.ListTeamChannels(ctx, teamID)
}

// getSquad retrieves a squad by ID using the Searchable interface
func (uc *TeamMessageCommandUseCase) getSquad(ctx context.Context, squadID uuid.UUID) (*squad_entities.Squad, error) {
	searchParams := []shared.SearchAggregation{
		{
			Params: []shared.SearchParameter{
				{
					ValueParams: []shared.SearchableValue{
						{Field: "ID", Values: []interface{}{squadID.String()}, Operator: shared.EqualsOperator},
					},
				},
			},
		},
	}
	resultOpts := shared.SearchResultOptions{Limit: 1}

	compiledSearch, err := uc.squadReader.Compile(ctx, searchParams, resultOpts)
	if err != nil {
		return nil, err
	}

	squads, err := uc.squadReader.Search(ctx, *compiledSearch)
	if err != nil {
		return nil, err
	}

	if len(squads) == 0 {
		return nil, fmt.Errorf("squad not found: %s", squadID)
	}

	return &squads[0], nil
}
