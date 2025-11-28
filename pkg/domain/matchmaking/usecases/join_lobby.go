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

// JoinLobbyUseCase handles player joining a lobby
type JoinLobbyUseCase struct {
	billableOperationHandler billing_in.BillableOperationCommandHandler
	lobbyRepository          matchmaking_out.LobbyRepository
}

// NewJoinLobbyUseCase creates a new join lobby usecase
func NewJoinLobbyUseCase(
	billableOperationHandler billing_in.BillableOperationCommandHandler,
	lobbyRepository matchmaking_out.LobbyRepository,
) *JoinLobbyUseCase {
	return &JoinLobbyUseCase{
		billableOperationHandler: billableOperationHandler,
		lobbyRepository:          lobbyRepository,
	}
}

// Exec joins a player to a lobby
func (uc *JoinLobbyUseCase) Exec(ctx context.Context, cmd matchmaking_in.JoinLobbyCommand) error {
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

	// billing validation BEFORE joining
	billingCmd := billing_in.BillableOperationCommand{
		OperationID: billing_entities.OperationTypeJoinLobby,
		UserID:      common.GetResourceOwner(ctx).UserID,
		Amount:      1,
		Args: map[string]interface{}{
			"lobby_id": cmd.LobbyID.String(),
		},
	}

	err = uc.billableOperationHandler.Validate(ctx, billingCmd)
	if err != nil {
		slog.ErrorContext(ctx, "billing validation failed for join lobby", "error", err, "player_id", cmd.PlayerID)
		return err
	}

	// add player to lobby
	err = lobby.AddPlayer(cmd.PlayerID, cmd.MMR)
	if err != nil {
		slog.ErrorContext(ctx, "failed to add player to lobby", "error", err, "lobby_id", cmd.LobbyID, "player_id", cmd.PlayerID)
		return err
	}

	// save updated lobby
	err = uc.lobbyRepository.Update(ctx, lobby)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update lobby", "error", err, "lobby_id", cmd.LobbyID)
		return fmt.Errorf("failed to join lobby")
	}

	// billing execution AFTER successful operation
	_, _, err = uc.billableOperationHandler.Exec(ctx, billingCmd)
	if err != nil {
		slog.WarnContext(ctx, "failed to execute billing for join lobby", "error", err, "lobby_id", cmd.LobbyID)
	}

	slog.InfoContext(ctx, "player joined lobby",
		"lobby_id", cmd.LobbyID,
		"player_id", cmd.PlayerID,
		"mmr", cmd.MMR,
	)

	return nil
}
