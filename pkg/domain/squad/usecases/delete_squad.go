package squad_usecases

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	common "github.com/replay-api/replay-api/pkg/domain"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_out "github.com/replay-api/replay-api/pkg/domain/squad/ports/out"
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
	isAuthenticated := ctx.Value(common.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return common.NewErrUnauthorized()
	}

	// 2. Check if squad exists and ownership
	squads, err := uc.SquadReader.Search(ctx, common.NewSearchByID(ctx, squadID, common.ClientApplicationAudienceIDKey))
	if err != nil {
		return err
	}

	if len(squads) == 0 {
		return common.NewErrNotFound(common.ResourceTypeSquad, "ID", squadID.String())
	}

	squad := squads[0]

	if squad.ResourceOwner.UserID != common.GetResourceOwner(ctx).UserID {
		return common.NewErrUnauthorized()
	}

	// 3. Billing validation
	billingCmd := billing_in.BillableOperationCommand{
		OperationID: billing_entities.OperationTypeDeleteSquadProfile,
		UserID:      common.GetResourceOwner(ctx).UserID,
		Amount:      1,
	}
	err = uc.billableOperationHandler.Validate(ctx, billingCmd)
	if err != nil {
		slog.ErrorContext(ctx, "Billing validation failed for delete squad", "error", err, "squad_id", squadID)
		return err
	}

	// 4. Delete squad
	err = uc.SquadWriter.Delete(ctx, squadID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to delete squad", "error", err, "squad_id", squadID)
		return err
	}

	// 5. Execute billing
	_, _, billingErr := uc.billableOperationHandler.Exec(ctx, billingCmd)
	if billingErr != nil {
		slog.WarnContext(ctx, "Failed to execute billing for delete squad", "error", billingErr, "squad_id", squadID)
	}

	// 6. Record history
	history := squad_entities.NewSquadHistory(squadID, common.GetResourceOwner(ctx).UserID, squad_entities.SquadDeleted, common.GetResourceOwner(ctx))
	_, err = uc.SquadHistoryWriter.Create(ctx, history)
	if err != nil {
		slog.WarnContext(ctx, "Failed to create squad history for delete", "error", err, "squad_id", squadID)
	}

	// 7. Log success
	slog.InfoContext(ctx, "squad deleted", "squad_id", squadID, "user_id", common.GetResourceOwner(ctx).UserID)

	return nil
}
