package matchmaking_usecases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	matchmaking_out "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/out"
)

// CreateCustomLobbyUseCase handles creating custom matchmaking lobbies
type CreateCustomLobbyUseCase struct {
	billableOperationHandler billing_in.BillableOperationCommandHandler
	lobbyRepository          matchmaking_out.LobbyRepository
}

// NewCreateCustomLobbyUseCase creates a new create lobby usecase
func NewCreateCustomLobbyUseCase(
	billableOperationHandler billing_in.BillableOperationCommandHandler,
	lobbyRepository matchmaking_out.LobbyRepository,
) matchmaking_in.LobbyCommand {
	return &CreateCustomLobbyUseCase{
		billableOperationHandler: billableOperationHandler,
		lobbyRepository:          lobbyRepository,
	}
}

// CreateLobby creates a new custom matchmaking lobby
func (uc *CreateCustomLobbyUseCase) CreateLobby(ctx context.Context, cmd matchmaking_in.CreateLobbyCommand) (*matchmaking_entities.MatchmakingLobby, error) {
	// auth check
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, shared.NewErrUnauthorized()
	}

	// validate max players
	if cmd.MaxPlayers < 2 || cmd.MaxPlayers > 10 {
		return nil, fmt.Errorf("max players must be between 2 and 10")
	}

	// billing validation BEFORE creating lobby
	billingCmd := billing_in.BillableOperationCommand{
		OperationID: billing_entities.OperationTypeCreateCustomLobby,
		UserID:      shared.GetResourceOwner(ctx).UserID,
		Amount:      1,
		Args: map[string]interface{}{
			"game_id":     cmd.GameID,
			"max_players": cmd.MaxPlayers,
			"invite_only": cmd.InviteOnly,
		},
	}

	err := uc.billableOperationHandler.Validate(ctx, billingCmd)
	if err != nil {
		slog.ErrorContext(ctx, "billing validation failed for create custom lobby", "error", err, "creator_id", cmd.CreatorID)
		return nil, err
	}

	// create lobby entity
	lobby, err := matchmaking_entities.NewMatchmakingLobby(
		shared.GetResourceOwner(ctx),
		cmd.CreatorID,
		cmd.GameID,
		cmd.Region,
		cmd.Tier,
		cmd.DistributionRule,
		cmd.MaxPlayers,
		cmd.AutoFill,
		cmd.InviteOnly,
	)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create lobby entity", "error", err)
		return nil, err
	}

	// save lobby
	err = uc.lobbyRepository.Save(ctx, lobby)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save lobby", "error", err)
		return nil, fmt.Errorf("failed to create lobby")
	}

	// billing execution AFTER successful operation
	_, _, err = uc.billableOperationHandler.Exec(ctx, billingCmd)
	if err != nil {
		slog.WarnContext(ctx, "failed to execute billing for create custom lobby", "error", err, "lobby_id", lobby.ID)
	}

	slog.InfoContext(ctx, "custom lobby created",
		"lobby_id", lobby.ID,
		"creator_id", cmd.CreatorID,
		"game_id", cmd.GameID,
		"max_players", cmd.MaxPlayers,
	)

	return lobby, nil
}

// JoinLobby - placeholder, will implement in join_lobby.go
func (uc *CreateCustomLobbyUseCase) JoinLobby(ctx context.Context, cmd matchmaking_in.JoinLobbyCommand) error {
	return fmt.Errorf("not implemented in this usecase")
}

// LeaveLobby - placeholder, will implement in leave_lobby.go
func (uc *CreateCustomLobbyUseCase) LeaveLobby(ctx context.Context, cmd matchmaking_in.LeaveLobbyCommand) error {
	return fmt.Errorf("not implemented in this usecase")
}

// SetPlayerReady - placeholder, will implement in set_player_ready.go
func (uc *CreateCustomLobbyUseCase) SetPlayerReady(ctx context.Context, cmd matchmaking_in.SetPlayerReadyCommand) error {
	return fmt.Errorf("not implemented in this usecase")
}

// StartReadyCheck - placeholder
func (uc *CreateCustomLobbyUseCase) StartReadyCheck(ctx context.Context, cmd matchmaking_in.StartReadyCheckCommand) error {
	return fmt.Errorf("not implemented in this usecase")
}

// StartMatch - placeholder
func (uc *CreateCustomLobbyUseCase) StartMatch(ctx context.Context, lobbyID uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, fmt.Errorf("not implemented in this usecase")
}

// CancelLobby - placeholder
func (uc *CreateCustomLobbyUseCase) CancelLobby(ctx context.Context, lobbyID uuid.UUID, reason string) error {
	return fmt.Errorf("not implemented in this usecase")
}
