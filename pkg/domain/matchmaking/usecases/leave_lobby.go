package matchmaking_usecases

import (
	"context"
	"fmt"
	"log/slog"

	common "github.com/replay-api/replay-api/pkg/domain"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	matchmaking_out "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/out"
)

// LeaveLobbyUseCase handles player leaving a lobby
type LeaveLobbyUseCase struct {
	billableOperationHandler billing_in.BillableOperationCommandHandler
	lobbyRepository          matchmaking_out.LobbyRepository
}

// NewLeaveLobbyUseCase creates a new leave lobby usecase
func NewLeaveLobbyUseCase(
	billableOperationHandler billing_in.BillableOperationCommandHandler,
	lobbyRepository matchmaking_out.LobbyRepository,
) *LeaveLobbyUseCase {
	return &LeaveLobbyUseCase{
		billableOperationHandler: billableOperationHandler,
		lobbyRepository:          lobbyRepository,
	}
}

// Exec removes a player from a lobby
func (uc *LeaveLobbyUseCase) Exec(ctx context.Context, cmd matchmaking_in.LeaveLobbyCommand) error {
	// auth check
	isAuthenticated := ctx.Value(common.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return common.NewErrUnauthorized()
	}

	// get lobby
	lobby, err := uc.lobbyRepository.FindByID(ctx, cmd.LobbyID)
	if err != nil {
		slog.ErrorContext(ctx, "lobby not found", "error", err, "lobby_id", cmd.LobbyID)
		return fmt.Errorf("lobby not found")
	}

	// billing validation BEFORE leaving
	billingCmd := billing_in.BillableOperationCommand{
		OperationID: billing_entities.OperationTypeLeaveLobby,
		UserID:      common.GetResourceOwner(ctx).UserID,
		Amount:      1,
		Args: map[string]interface{}{
			"lobby_id": cmd.LobbyID.String(),
		},
	}

	err = uc.billableOperationHandler.Validate(ctx, billingCmd)
	if err != nil {
		slog.ErrorContext(ctx, "billing validation failed for leave lobby", "error", err, "player_id", cmd.PlayerID)
		return err
	}

	// remove player from lobby
	err = lobby.RemovePlayer(cmd.PlayerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to remove player from lobby", "error", err, "lobby_id", cmd.LobbyID, "player_id", cmd.PlayerID)
		return err
	}

	// save updated lobby
	err = uc.lobbyRepository.Update(ctx, lobby)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update lobby", "error", err, "lobby_id", cmd.LobbyID)
		return fmt.Errorf("failed to leave lobby")
	}

	// billing execution AFTER successful operation
	_, _, err = uc.billableOperationHandler.Exec(ctx, billingCmd)
	if err != nil {
		slog.WarnContext(ctx, "failed to execute billing for leave lobby", "error", err, "lobby_id", cmd.LobbyID)
	}

	slog.InfoContext(ctx, "player left lobby",
		"lobby_id", cmd.LobbyID,
		"player_id", cmd.PlayerID,
	)

	return nil
}
