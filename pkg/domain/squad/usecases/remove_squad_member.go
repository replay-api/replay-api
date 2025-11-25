package squad_usecases

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	common "github.com/replay-api/replay-api/pkg/domain"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	squad_out "github.com/replay-api/replay-api/pkg/domain/squad/ports/out"
)

type RemoveSquadMemberUseCase struct {
	billableOperationHandler billing_in.BillableOperationCommandHandler
	SquadReader              squad_out.SquadReader
	SquadWriter              squad_out.SquadWriter
	SquadHistoryWriter       squad_out.SquadHistoryWriter
}

func NewRemoveSquadMemberUseCase(
	billableOperationHandler billing_in.BillableOperationCommandHandler,
	squadReader squad_out.SquadReader,
	squadWriter squad_out.SquadWriter,
	squadHistoryWriter squad_out.SquadHistoryWriter,
) *RemoveSquadMemberUseCase {
	return &RemoveSquadMemberUseCase{
		billableOperationHandler: billableOperationHandler,
		SquadReader:              squadReader,
		SquadWriter:              squadWriter,
		SquadHistoryWriter:       squadHistoryWriter,
	}
}

func (uc *RemoveSquadMemberUseCase) Exec(ctx context.Context, cmd squad_in.RemoveSquadMemberCommand) (*squad_entities.Squad, error) {
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

	// 3. Check if member exists
	if squad.Membership == nil {
		return nil, common.NewErrNotFound(common.ResourceTypeSquad, "MemberID", cmd.PlayerID.String())
	}

	if _, exists := squad.Membership[cmd.PlayerID.String()]; !exists {
		return nil, common.NewErrNotFound(common.ResourceTypeSquad, "MemberID", cmd.PlayerID.String())
	}

	// 4. Billing validation
	billingCmd := billing_in.BillableOperationCommand{
		OperationID: billing_entities.OperationTypeRemoveSquadMember,
		UserID:      common.GetResourceOwner(ctx).UserID,
		Amount:      1,
	}
	err = uc.billableOperationHandler.Validate(ctx, billingCmd)
	if err != nil {
		slog.ErrorContext(ctx, "Billing validation failed for remove squad member", "error", err, "squad_id", cmd.SquadID)
		return nil, err
	}

	// 5. Remove member from squad
	delete(squad.Membership, cmd.PlayerID.String())

	// 6. Update squad
	updatedSquad, err := uc.SquadWriter.Update(ctx, &squad)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to remove squad member", "error", err, "squad_id", cmd.SquadID, "player_id", cmd.PlayerID)
		return nil, err
	}

	// 7. Execute billing
	err = uc.billableOperationHandler.Exec(ctx, billingCmd)
	if err != nil {
		slog.WarnContext(ctx, "Failed to execute billing for remove squad member", "error", err, "squad_id", cmd.SquadID)
	}

	// 8. Record history
	history := squad_entities.NewSquadHistory(cmd.SquadID, common.GetResourceOwner(ctx).UserID, squad_entities.SquadMemberRemoved, common.GetResourceOwner(ctx))
	_, err = uc.SquadHistoryWriter.Create(ctx, history)
	if err != nil {
		slog.WarnContext(ctx, "Failed to create squad history for remove member", "error", err, "squad_id", cmd.SquadID)
	}

	// 9. Log success
	slog.InfoContext(ctx, "squad member removed", "squad_id", cmd.SquadID, "player_id", cmd.PlayerID, "user_id", common.GetResourceOwner(ctx).UserID)

	return updatedSquad, nil
}
