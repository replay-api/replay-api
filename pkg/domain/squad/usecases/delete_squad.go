package squad_usecases

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	squad "github.com/replay-api/replay-api/pkg/domain/squad"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_out "github.com/replay-api/replay-api/pkg/domain/squad/ports/out"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type DeleteSquadUseCase struct {
	billableOperationHandler billing_in.BillableOperationCommandHandler
	SquadReader              squad_out.SquadReader
	SquadWriter              squad_out.SquadWriter
	SquadHistoryWriter       squad_out.SquadHistoryWriter
}

func NewDeleteSquadUseCase(
	billableOperationHandler billing_in.BillableOperationCommandHandler,
	squadReader squad_out.SquadReader,
	squadWriter squad_out.SquadWriter,
	squadHistoryWriter squad_out.SquadHistoryWriter,
) *DeleteSquadUseCase {
	return &DeleteSquadUseCase{
		billableOperationHandler: billableOperationHandler,
		SquadReader:              squadReader,
		SquadWriter:              squadWriter,
		SquadHistoryWriter:       squadHistoryWriter,
	}
}

func (uc *DeleteSquadUseCase) Exec(ctx context.Context, squadID uuid.UUID) error {
	// 1. Authentication check
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return shared.NewErrUnauthorized()
	}

	// 2. Check if squad exists
	squads, err := uc.SquadReader.Search(ctx, shared.NewSearchByID(ctx, squadID, shared.ClientApplicationAudienceIDKey))
	if err != nil {
		return err
	}

	if len(squads) == 0 {
		return shared.NewErrNotFound(replay_common.ResourceTypeSquad, "ID", squadID.String())
	}

	squadEntity := squads[0]

	// 3. Authorization check - only owner or admin can delete squad
	if err := squad.MustBeSquadOwnerOrAdmin(ctx, &squadEntity); err != nil {
		slog.WarnContext(ctx, "Unauthorized squad delete attempt", "squad_id", squadID, "user_id", shared.GetResourceOwner(ctx).UserID, "error", err)
		return err
	}

	// 4. Billing validation
	billingCmd := billing_in.BillableOperationCommand{
		OperationID: billing_entities.OperationTypeDeleteSquadProfile,
		UserID:      shared.GetResourceOwner(ctx).UserID,
		Amount:      1,
	}
	err = uc.billableOperationHandler.Validate(ctx, billingCmd)
	if err != nil {
		slog.ErrorContext(ctx, "Billing validation failed for delete squad", "error", err, "squad_id", squadID)
		return err
	}

	// 5. Delete squad
	err = uc.SquadWriter.Delete(ctx, squadID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to delete squad", "error", err, "squad_id", squadID)
		return err
	}

	// 6. Execute billing
	_, _, billingErr := uc.billableOperationHandler.Exec(ctx, billingCmd)
	if billingErr != nil {
		slog.WarnContext(ctx, "Failed to execute billing for delete squad", "error", billingErr, "squad_id", squadID)
	}

	// 7. Record history
	history := squad_entities.NewSquadHistory(squadID, shared.GetResourceOwner(ctx).UserID, squad_entities.SquadDeleted, shared.GetResourceOwner(ctx))
	_, err = uc.SquadHistoryWriter.Create(ctx, history)
	if err != nil {
		slog.WarnContext(ctx, "Failed to create squad history for delete", "error", err, "squad_id", squadID)
	}

	// 8. Log success
	slog.InfoContext(ctx, "squad deleted", "squad_id", squadID, "user_id", shared.GetResourceOwner(ctx).UserID)

	return nil
}
