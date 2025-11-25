package tournament_usecases

import (
	"context"
	"fmt"
	"log/slog"

	common "github.com/replay-api/replay-api/pkg/domain"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	tournament_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/tournament/entities"
	tournament_in "github.com/psavelis/team-pro/replay-api/pkg/domain/tournament/ports/in"
	tournament_out "github.com/psavelis/team-pro/replay-api/pkg/domain/tournament/ports/out"
)

// CreateTournamentUseCase handles tournament creation
type CreateTournamentUseCase struct {
	billableOperationHandler billing_in.BillableOperationCommandHandler
	tournamentRepository     tournament_out.TournamentRepository
}

// NewCreateTournamentUseCase creates a new create tournament usecase
func NewCreateTournamentUseCase(
	billableOperationHandler billing_in.BillableOperationCommandHandler,
	tournamentRepository tournament_out.TournamentRepository,
) *CreateTournamentUseCase {
	return &CreateTournamentUseCase{
		billableOperationHandler: billableOperationHandler,
		tournamentRepository:     tournamentRepository,
	}
}

// Exec creates a new tournament
func (uc *CreateTournamentUseCase) Exec(ctx context.Context, cmd tournament_in.CreateTournamentCommand) (*tournament_entities.Tournament, error) {
	// auth check
	isAuthenticated := ctx.Value(common.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, common.NewErrUnauthorized()
	}

	// validate format
	switch cmd.Format {
	case tournament_entities.TournamentFormatSingleElimination,
		tournament_entities.TournamentFormatDoubleElimination,
		tournament_entities.TournamentFormatRoundRobin,
		tournament_entities.TournamentFormatSwiss:
		// valid
	default:
		return nil, fmt.Errorf("invalid tournament format: %s", cmd.Format)
	}

	// billing validation BEFORE creating tournament
	billingCmd := billing_in.BillableOperationCommand{
		OperationID: billing_entities.OperationTypeCreateTournament,
		UserID:      common.GetResourceOwner(ctx).UserID,
		Amount:      1,
		Args: map[string]interface{}{
			"name":             cmd.Name,
			"format":           cmd.Format,
			"max_participants": cmd.MaxParticipants,
			"entry_fee":        cmd.EntryFee.ToFloat64(),
		},
	}

	err := uc.billableOperationHandler.Validate(ctx, billingCmd)
	if err != nil {
		slog.ErrorContext(ctx, "billing validation failed for create tournament", "error", err, "organizer_id", cmd.OrganizerID)
		return nil, err
	}

	// create tournament entity
	tournament, err := tournament_entities.NewTournament(
		cmd.ResourceOwner,
		cmd.Name,
		cmd.Description,
		cmd.GameID,
		cmd.GameMode,
		cmd.Region,
		cmd.Format,
		cmd.MaxParticipants,
		cmd.MinParticipants,
		cmd.EntryFee,
		cmd.Currency,
		cmd.StartTime,
		cmd.RegistrationOpen,
		cmd.RegistrationClose,
		cmd.Rules,
		cmd.OrganizerID,
	)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create tournament entity", "error", err)
		return nil, err
	}

	// validate tournament
	err = tournament.Validate()
	if err != nil {
		slog.ErrorContext(ctx, "tournament validation failed", "error", err)
		return nil, err
	}

	// save tournament
	err = uc.tournamentRepository.Save(ctx, tournament)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save tournament", "error", err)
		return nil, fmt.Errorf("failed to create tournament")
	}

	// billing execution AFTER successful operation
	_, _, err = uc.billableOperationHandler.Exec(ctx, billingCmd)
	if err != nil {
		slog.WarnContext(ctx, "failed to execute billing for create tournament", "error", err, "tournament_id", tournament.ID)
	}

	slog.InfoContext(ctx, "tournament created",
		"tournament_id", tournament.ID,
		"name", cmd.Name,
		"format", cmd.Format,
		"max_participants", cmd.MaxParticipants,
		"entry_fee", cmd.EntryFee.ToFloat64(),
	)

	return tournament, nil
}
