package squad_usecases

import (
	"context"
	"log/slog"

	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	common "github.com/replay-api/replay-api/pkg/domain"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	squad_out "github.com/replay-api/replay-api/pkg/domain/squad/ports/out"
)

type UpdateSquadMemberRoleUseCase struct {
	billableOperationHandler billing_in.BillableOperationCommandHandler
	SquadReader              squad_out.SquadReader
	SquadWriter              squad_out.SquadWriter
	SquadHistoryWriter       squad_out.SquadHistoryWriter
}

func NewUpdateSquadMemberRoleUseCase(
	billableOperationHandler billing_in.BillableOperationCommandHandler,
	squadReader squad_out.SquadReader,
	squadWriter squad_out.SquadWriter,
	squadHistoryWriter squad_out.SquadHistoryWriter,
) *UpdateSquadMemberRoleUseCase {
	return &UpdateSquadMemberRoleUseCase{
		billableOperationHandler: billableOperationHandler,
		SquadReader:              squadReader,
		SquadWriter:              squadWriter,
		SquadHistoryWriter:       squadHistoryWriter,
	}
}

func (uc *UpdateSquadMemberRoleUseCase) Exec(ctx context.Context, cmd squad_in.UpdateSquadMemberRoleCommand) (*squad_entities.Squad, error) {
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

	// 3. Find member in slice
	memberIndex := -1
	for i, m := range squad.Membership {
		if m.PlayerProfileID == cmd.PlayerID {
			memberIndex = i
			break
		}
	}

	if memberIndex == -1 {
		return nil, common.NewErrNotFound(common.ResourceTypeSquad, "MemberID", cmd.PlayerID.String())
	}

	// 4. Billing validation
	billingCmd := billing_in.BillableOperationCommand{
		OperationID: billing_entities.OperationTypeUpdateSquadMemberRole,
		UserID:      common.GetResourceOwner(ctx).UserID,
		Amount:      1,
	}
	err = uc.billableOperationHandler.Validate(ctx, billingCmd)
	if err != nil {
		slog.ErrorContext(ctx, "Billing validation failed for update squad member role", "error", err, "squad_id", cmd.SquadID)
		return nil, err
	}

	// 5. Update member roles
	squad.Membership[memberIndex].Roles = cmd.Roles

	// 6. Update squad
	updatedSquad, err := uc.SquadWriter.Update(ctx, &squad)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update squad member role", "error", err, "squad_id", cmd.SquadID, "player_id", cmd.PlayerID)
		return nil, err
	}

	// 7. Execute billing
	_, _, billingErr := uc.billableOperationHandler.Exec(ctx, billingCmd)
	if billingErr != nil {
		slog.WarnContext(ctx, "Failed to execute billing for update squad member role", "error", billingErr, "squad_id", cmd.SquadID)
	}

	// 8. Record history - using SquadMemberPromoted as generic role update action
	historyAction := squad_entities.SquadMemberPromoted
	// Could add logic to determine if promoted or demoted based on role hierarchy if needed

	history := squad_entities.NewSquadHistory(cmd.SquadID, common.GetResourceOwner(ctx).UserID, historyAction, common.GetResourceOwner(ctx))
	if history.Action == "" {
		history.Action = squad_entities.SquadUpdated
	}
	_, err = uc.SquadHistoryWriter.Create(ctx, history)
	if err != nil {
		slog.WarnContext(ctx, "Failed to create squad history for update member role", "error", err, "squad_id", cmd.SquadID)
	}

	// 9. Log success
	slog.InfoContext(ctx, "squad member role updated", "squad_id", cmd.SquadID, "player_id", cmd.PlayerID, "roles", cmd.Roles, "user_id", common.GetResourceOwner(ctx).UserID)

	return updatedSquad, nil
}
