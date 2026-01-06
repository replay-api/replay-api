package squad_usecases

import (
	"context"
	"log/slog"
	"time"

	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	squad "github.com/replay-api/replay-api/pkg/domain/squad"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	squad_out "github.com/replay-api/replay-api/pkg/domain/squad/ports/out"
	squad_value_objects "github.com/replay-api/replay-api/pkg/domain/squad/value-objects"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
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
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, shared.NewErrUnauthorized()
	}

	// 2. Check if squad exists
	squads, err := uc.SquadReader.Search(ctx, shared.NewSearchByID(ctx, cmd.SquadID, shared.ClientApplicationAudienceIDKey))
	if err != nil {
		return nil, err
	}

	if len(squads) == 0 {
		return nil, shared.NewErrNotFound(replay_common.ResourceTypeSquad, "ID", cmd.SquadID.String())
	}

	squadEntity := squads[0]

	// 3. Find member in slice
	memberIndex := -1
	for i, m := range squadEntity.Membership {
		if m.PlayerProfileID == cmd.PlayerID {
			memberIndex = i
			break
		}
	}

	if memberIndex == -1 {
		return nil, shared.NewErrNotFound(replay_common.ResourceTypeSquad, "MemberID", cmd.PlayerID.String())
	}

	// 4. Authorization check - only owner/admin can update roles
	// Determine new membership type based on role change
	newMembershipType := squadEntity.Membership[memberIndex].Type
	if cmd.MembershipType != "" {
		newMembershipType = cmd.MembershipType
	}

	if err := squad.CanUpdateMemberRole(ctx, &squadEntity, cmd.PlayerID, newMembershipType); err != nil {
		slog.WarnContext(ctx, "Unauthorized squad member role update attempt", "squad_id", cmd.SquadID, "target_player", cmd.PlayerID, "user_id", shared.GetResourceOwner(ctx).UserID, "error", err)
		return nil, err
	}

	// 5. Billing validation
	billingCmd := billing_in.BillableOperationCommand{
		OperationID: billing_entities.OperationTypeUpdateSquadMemberRole,
		UserID:      shared.GetResourceOwner(ctx).UserID,
		Amount:      1,
	}
	err = uc.billableOperationHandler.Validate(ctx, billingCmd)
	if err != nil {
		slog.ErrorContext(ctx, "Billing validation failed for update squad member role", "error", err, "squad_id", cmd.SquadID)
		return nil, err
	}

	// 6. Determine the history action based on role change
	oldType := squadEntity.Membership[memberIndex].Type
	historyAction := squad_entities.SquadUpdated
	if newMembershipType != oldType {
		if isPromotion(oldType, newMembershipType) {
			historyAction = squad_entities.SquadMemberPromoted
		} else {
			historyAction = squad_entities.SquadMemberDemoted
		}
	}

	// 7. Update member roles and type
	squadEntity.Membership[memberIndex].Roles = cmd.Roles
	if cmd.MembershipType != "" {
		squadEntity.Membership[memberIndex].Type = cmd.MembershipType
		squadEntity.Membership[memberIndex].History[time.Now()] = cmd.MembershipType
	}

	// 8. Update squad
	updatedSquad, err := uc.SquadWriter.Update(ctx, &squadEntity)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update squad member role", "error", err, "squad_id", cmd.SquadID, "player_id", cmd.PlayerID)
		return nil, err
	}

	// 9. Execute billing
	_, _, billingErr := uc.billableOperationHandler.Exec(ctx, billingCmd)
	if billingErr != nil {
		slog.WarnContext(ctx, "Failed to execute billing for update squad member role", "error", billingErr, "squad_id", cmd.SquadID)
	}

	// 10. Record history
	history := squad_entities.NewSquadHistory(cmd.SquadID, shared.GetResourceOwner(ctx).UserID, historyAction, shared.GetResourceOwner(ctx))
	_, err = uc.SquadHistoryWriter.Create(ctx, history)
	if err != nil {
		slog.WarnContext(ctx, "Failed to create squad history for update member role", "error", err, "squad_id", cmd.SquadID)
	}

	// 11. Log success
	slog.InfoContext(ctx, "squad member role updated", "squad_id", cmd.SquadID, "player_id", cmd.PlayerID, "roles", cmd.Roles, "user_id", shared.GetResourceOwner(ctx).UserID)

	return updatedSquad, nil
}

// isPromotion determines if a membership type change is a promotion
func isPromotion(from, to squad_value_objects.SquadMembershipType) bool {
	hierarchy := map[squad_value_objects.SquadMembershipType]int{
		squad_value_objects.SquadMembershipTypeInactive: 0,
		squad_value_objects.SquadMembershipTypeGuest:    1,
		squad_value_objects.SquadMembershipTypeMember:   2,
		squad_value_objects.SquadMembershipTypeAdmin:    3,
		squad_value_objects.SquadMembershipTypeOwner:    4,
	}

	return hierarchy[to] > hierarchy[from]
}
