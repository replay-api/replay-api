package squad_usecases

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	squad_out "github.com/replay-api/replay-api/pkg/domain/squad/ports/out"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type DeletePlayerUseCase struct {
	billableOperationHandler billing_in.BillableOperationCommandHandler
	PlayerProfileReader      squad_in.PlayerProfileReader
	PlayerProfileWriter      squad_out.PlayerProfileWriter
	PlayerHistoryWriter      squad_out.PlayerProfileHistoryWriter
}

func NewDeletePlayerUseCase(
	billableOperationHandler billing_in.BillableOperationCommandHandler,
	playerProfileReader squad_in.PlayerProfileReader,
	playerProfileWriter squad_out.PlayerProfileWriter,
	playerHistoryWriter squad_out.PlayerProfileHistoryWriter,
) *DeletePlayerUseCase {
	return &DeletePlayerUseCase{
		billableOperationHandler: billableOperationHandler,
		PlayerProfileReader:      playerProfileReader,
		PlayerProfileWriter:      playerProfileWriter,
		PlayerHistoryWriter:      playerHistoryWriter,
	}
}

func (uc *DeletePlayerUseCase) Exec(ctx context.Context, playerID uuid.UUID) error {
	// 1. Authentication check
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return shared.NewErrUnauthorized()
	}

	// 2. Check if player exists and ownership
	getByIdSearch := squad_entities.NewSearchByID(ctx, playerID)
	playerProfile, err := uc.PlayerProfileReader.Search(ctx, getByIdSearch)
	if err != nil {
		return err
	}

	if len(playerProfile) == 0 {
		return shared.NewErrNotFound(replay_common.ResourceTypePlayerProfile, "ID", playerID.String())
	}

	if playerProfile[0].ResourceOwner.UserID != shared.GetResourceOwner(ctx).UserID {
		return shared.NewErrUnauthorized()
	}

	// 3. Billing validation
	billingCmd := billing_in.BillableOperationCommand{
		OperationID: billing_entities.OperationTypeDeletePlayerProfile,
		UserID:      shared.GetResourceOwner(ctx).UserID,
		Amount:      1,
	}
	err = uc.billableOperationHandler.Validate(ctx, billingCmd)
	if err != nil {
		slog.ErrorContext(ctx, "Billing validation failed for delete player", "error", err, "player_id", playerID)
		return err
	}

	// 4. Delete player
	err = uc.PlayerProfileWriter.Delete(ctx, playerID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to delete player profile", "error", err, "player_id", playerID)
		return err
	}

	// 5. Execute billing
	_, _, billingErr := uc.billableOperationHandler.Exec(ctx, billingCmd)
	if billingErr != nil {
		slog.WarnContext(ctx, "Failed to execute billing for delete player", "error", billingErr, "player_id", playerID)
	}

	// 6. Record history
	history := squad_entities.NewPlayerProfileHistory(playerID, squad_entities.PlayerHistoryActionDelete, shared.GetResourceOwner(ctx))
	_, err = uc.PlayerHistoryWriter.Create(ctx, history)
	if err != nil {
		slog.WarnContext(ctx, "Failed to create player history for delete", "error", err, "player_id", playerID)
	}

	// 7. Log success
	slog.InfoContext(ctx, "player profile deleted", "player_id", playerID, "user_id", shared.GetResourceOwner(ctx).UserID)

	return nil
}
