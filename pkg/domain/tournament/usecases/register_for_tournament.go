package tournament_usecases

import (
	"context"
	"fmt"
	"log/slog"

	common "github.com/replay-api/replay-api/pkg/domain"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	tournament_in "github.com/psavelis/team-pro/replay-api/pkg/domain/tournament/ports/in"
	tournament_out "github.com/psavelis/team-pro/replay-api/pkg/domain/tournament/ports/out"
)

// RegisterForTournamentUseCase handles player registration for tournaments
type RegisterForTournamentUseCase struct {
	billableOperationHandler billing_in.BillableOperationCommandHandler
	tournamentRepository     tournament_out.TournamentRepository
}

// NewRegisterForTournamentUseCase creates a new register for tournament usecase
func NewRegisterForTournamentUseCase(
	billableOperationHandler billing_in.BillableOperationCommandHandler,
	tournamentRepository tournament_out.TournamentRepository,
) *RegisterForTournamentUseCase {
	return &RegisterForTournamentUseCase{
		billableOperationHandler: billableOperationHandler,
		tournamentRepository:     tournamentRepository,
	}
}

// Exec registers a player for a tournament
func (uc *RegisterForTournamentUseCase) Exec(ctx context.Context, cmd tournament_in.RegisterPlayerCommand) error {
	// auth check
	isAuthenticated := ctx.Value(common.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return common.NewErrUnauthorized()
	}

	// get tournament
	tournament, err := uc.tournamentRepository.FindByID(ctx, cmd.TournamentID)
	if err != nil {
		slog.ErrorContext(ctx, "tournament not found", "error", err, "tournament_id", cmd.TournamentID)
		return fmt.Errorf("tournament not found")
	}

	// billing validation BEFORE registering (includes entry fee)
	amount := 1.0
	if !tournament.EntryFee.IsZero() {
		amount = tournament.EntryFee.ToFloat64()
	}

	billingCmd := billing_in.BillableOperationCommand{
		OperationID: billing_entities.OperationTypeRegisterForTournament,
		UserID:      common.GetResourceOwner(ctx).UserID,
		Amount:      amount,
		Args: map[string]interface{}{
			"tournament_id": cmd.TournamentID.String(),
			"entry_fee":     tournament.EntryFee.ToFloat64(),
		},
	}

	err = uc.billableOperationHandler.Validate(ctx, billingCmd)
	if err != nil {
		slog.ErrorContext(ctx, "billing validation failed for register for tournament", "error", err, "player_id", cmd.PlayerID)
		return err
	}

	// register player
	err = tournament.RegisterPlayer(cmd.PlayerID, cmd.DisplayName)
	if err != nil {
		slog.ErrorContext(ctx, "failed to register player", "error", err, "tournament_id", cmd.TournamentID, "player_id", cmd.PlayerID)
		return err
	}

	// save updated tournament
	err = uc.tournamentRepository.Update(ctx, tournament)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update tournament", "error", err, "tournament_id", cmd.TournamentID)
		return fmt.Errorf("failed to register for tournament")
	}

	// billing execution AFTER successful operation
	_, _, err = uc.billableOperationHandler.Exec(ctx, billingCmd)
	if err != nil {
		slog.WarnContext(ctx, "failed to execute billing for register for tournament", "error", err, "tournament_id", cmd.TournamentID)
	}

	slog.InfoContext(ctx, "player registered for tournament",
		"tournament_id", cmd.TournamentID,
		"player_id", cmd.PlayerID,
		"display_name", cmd.DisplayName,
		"entry_fee", tournament.EntryFee.ToFloat64(),
	)

	return nil
}
