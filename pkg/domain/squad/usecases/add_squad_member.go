package squad_usecases

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	common "github.com/replay-api/replay-api/pkg/domain"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	squad_out "github.com/replay-api/replay-api/pkg/domain/squad/ports/out"
	squad_value_objects "github.com/replay-api/replay-api/pkg/domain/squad/value-objects"
)

type AddSquadMemberUseCase struct {
	billableOperationHandler billing_in.BillableOperationCommandHandler
	SquadReader              squad_out.SquadReader
	SquadWriter              squad_out.SquadWriter
	SquadHistoryWriter       squad_out.SquadHistoryWriter
	PlayerProfileReader      squad_in.PlayerProfileReader
}

func NewAddSquadMemberUseCase(
	billableOperationHandler billing_in.BillableOperationCommandHandler,
	squadReader squad_out.SquadReader,
	squadWriter squad_out.SquadWriter,
	squadHistoryWriter squad_out.SquadHistoryWriter,
	playerProfileReader squad_in.PlayerProfileReader,
) *AddSquadMemberUseCase {
	return &AddSquadMemberUseCase{
		billableOperationHandler: billableOperationHandler,
		SquadReader:              squadReader,
		SquadWriter:              squadWriter,
		SquadHistoryWriter:       squadHistoryWriter,
		PlayerProfileReader:      playerProfileReader,
	}
}

func (uc *AddSquadMemberUseCase) Exec(ctx context.Context, cmd squad_in.AddSquadMemberCommand) (*squad_entities.Squad, error) {
	// 1. Authentication check
	isAuthenticated := ctx.Value(common.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, common.NewErrUnauthorized()
	}

	// 2. Check if squad exists
	squads, err := uc.SquadReader.Search(ctx, common.NewSearchByID(ctx, cmd.SquadID, common.ClientApplicationAudienceIDKey))
	if err != nil {
		return nil, err
	}

	if len(squads) == 0 {
		return nil, common.NewErrNotFound(common.ResourceTypeSquad, "ID", cmd.SquadID.String())
	}

	squad := squads[0]

	// 3. Validate player exists
	players, err := uc.PlayerProfileReader.Search(ctx, common.NewSearchByID(ctx, cmd.PlayerID, common.ClientApplicationAudienceIDKey))
	if err != nil {
		return nil, err
	}

	if len(players) == 0 {
		return nil, common.NewErrNotFound(common.ResourceTypePlayerProfile, "ID", cmd.PlayerID.String())
	}

	// 4. Check if member already exists
	if squad.Membership != nil {
		if _, exists := squad.Membership[cmd.PlayerID.String()]; exists {
			return nil, common.NewErrAlreadyExists(common.ResourceTypeSquad, "MemberID", cmd.PlayerID.String())
		}
	}

	// 5. Billing validation
	billingCmd := billing_in.BillableOperationCommand{
		OperationID: billing_entities.OperationTypeAddSquadMember,
		UserID:      common.GetResourceOwner(ctx).UserID,
		Amount:      1,
	}
	err = uc.billableOperationHandler.Validate(ctx, billingCmd)
	if err != nil {
		slog.ErrorContext(ctx, "Billing validation failed for add squad member", "error", err, "squad_id", cmd.SquadID)
		return nil, err
	}

	// 6. Add member to squad
	if squad.Membership == nil {
		squad.Membership = make(map[string]squad_value_objects.SquadMembership)
	}

	now := time.Now()
	memberType := cmd.Type
	if memberType == "" {
		memberType = squad_value_objects.SquadMembershipTypeMember
	}

	squad.Membership[cmd.PlayerID.String()] = squad_value_objects.SquadMembership{
		UserID: players[0].ResourceOwner.UserID,
		Type:   memberType,
		Roles:  cmd.Roles,
		Status: map[time.Time]squad_value_objects.SquadMembershipStatus{
			now: squad_value_objects.SquadMembershipStatusActive,
		},
		History: map[time.Time]squad_value_objects.SquadMembershipType{
			now: memberType,
		},
	}

	// 7. Update squad
	updatedSquad, err := uc.SquadWriter.Update(ctx, &squad)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to add squad member", "error", err, "squad_id", cmd.SquadID, "player_id", cmd.PlayerID)
		return nil, err
	}

	// 8. Execute billing
	err = uc.billableOperationHandler.Exec(ctx, billingCmd)
	if err != nil {
		slog.WarnContext(ctx, "Failed to execute billing for add squad member", "error", err, "squad_id", cmd.SquadID)
	}

	// 9. Record history
	history := squad_entities.NewSquadHistory(cmd.SquadID, common.GetResourceOwner(ctx).UserID, squad_entities.SquadMemberAdded, common.GetResourceOwner(ctx))
	_, err = uc.SquadHistoryWriter.Create(ctx, history)
	if err != nil {
		slog.WarnContext(ctx, "Failed to create squad history for add member", "error", err, "squad_id", cmd.SquadID)
	}

	// 10. Log success
	slog.InfoContext(ctx, "squad member added", "squad_id", cmd.SquadID, "player_id", cmd.PlayerID, "user_id", common.GetResourceOwner(ctx).UserID)

	return updatedSquad, nil
}
