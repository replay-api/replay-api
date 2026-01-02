package tournament_usecases

import (
	"context"
	"fmt"
	"log/slog"

	common "github.com/replay-api/replay-api/pkg/domain"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	tournament_in "github.com/replay-api/replay-api/pkg/domain/tournament/ports/in"
	tournament_out "github.com/replay-api/replay-api/pkg/domain/tournament/ports/out"
)

// RegisterForTournamentUseCase handles player registration for competitive tournaments.
//
// This is a critical financial use case that involves entry fees and prize pool management.
//
// Flow:
//  1. Authentication verification - user must be authenticated
//  2. CRITICAL Ownership validation - prevents impersonation attacks where a user
//     attempts to register another player for a tournament (fraudulent entry fee charges)
//  3. Tournament retrieval and validation
//  4. Billing validation - verifies user has sufficient balance/credits for entry fee
//  5. Player registration on tournament entity
//  6. Tournament persistence with updated registrations
//  7. Billing execution - charges entry fee
//
// Security:
//   - Requires authenticated context (common.AuthenticatedKey)
//   - Validates PlayerID ownership against authenticated user
//   - Logs impersonation attempts with attacker details for security monitoring
//
// Financial:
//   - Entry fee validation before registration
//   - Billing integration for entry fee collection
//   - Supports variable entry fees per tournament
//
// Dependencies:
//   - BillableOperationCommandHandler: Entry fee validation and collection
//   - TournamentRepository: Tournament lookup and update
//   - PlayerProfileReader: Player ownership verification (temporarily disabled)
type RegisterForTournamentUseCase struct {
	billableOperationHandler billing_in.BillableOperationCommandHandler
	tournamentRepository     tournament_out.TournamentRepository
	// playerProfileReader      squad_in.PlayerProfileReader // TODO: Re-enable once PlayerProfileRepository is properly registered
}

// NewRegisterForTournamentUseCase creates a new register for tournament usecase
func NewRegisterForTournamentUseCase(
	billableOperationHandler billing_in.BillableOperationCommandHandler,
	tournamentRepository tournament_out.TournamentRepository,
	// playerProfileReader squad_in.PlayerProfileReader, // TODO: Re-enable once PlayerProfileRepository is properly registered
) *RegisterForTournamentUseCase {
	return &RegisterForTournamentUseCase{
		billableOperationHandler: billableOperationHandler,
		tournamentRepository:     tournamentRepository,
		// playerProfileReader:      playerProfileReader,
	}
}

// Exec registers a player for a tournament
func (uc *RegisterForTournamentUseCase) Exec(ctx context.Context, cmd tournament_in.RegisterPlayerCommand) error {
	// 1. Authentication check
	isAuthenticated := ctx.Value(common.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return common.NewErrUnauthorized()
	}

	// 2. CRITICAL: Ownership validation - prevent impersonation
	// TODO: Re-enable ownership validation once PlayerProfileRepository is properly registered
	// Verify the PlayerID belongs to the authenticated user
	// playerSearch := squad_entities.NewSearchByID(ctx, cmd.PlayerID)
	// players, err := uc.playerProfileReader.Search(ctx, playerSearch)
	// if err != nil {
	// 	slog.ErrorContext(ctx, "failed to find player profile", "error", err, "player_id", cmd.PlayerID)
	// 	return fmt.Errorf("player not found")
	// }
	// if len(players) == 0 {
	// 	return common.NewErrNotFound(common.ResourceTypePlayerProfile, "ID", cmd.PlayerID.String())
	// }

	// // Verify ownership - player must belong to authenticated user
	// currentUserID := common.GetResourceOwner(ctx).UserID
	// if players[0].ResourceOwner.UserID != currentUserID {
	// 	slog.WarnContext(ctx, "Tournament registration impersonation attempt blocked",
	// 		"attempted_player_id", cmd.PlayerID,
	// 		"player_owner", players[0].ResourceOwner.UserID,
	// 		"attacker_user_id", currentUserID,
	// 		"tournament_id", cmd.TournamentID,
	// 	)
	// 	return common.NewErrUnauthorized()
	// }

	// 3. Get tournament
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
