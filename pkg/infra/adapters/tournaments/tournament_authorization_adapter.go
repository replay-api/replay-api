package tournaments

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	tournament_out "github.com/replay-api/replay-api/pkg/domain/tournament/ports/out"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// tournamentAuthorizationAdapter implements TournamentAuthorization using the tournament repository
// and resource ownership context to enforce financial-grade RBAC on tournament operations.
type tournamentAuthorizationAdapter struct {
	tournamentRepo tournament_out.TournamentRepository
}

// NewTournamentAuthorizationAdapter creates a new authorization adapter
func NewTournamentAuthorizationAdapter(
	tournamentRepo tournament_out.TournamentRepository,
) tournament_out.TournamentAuthorization {
	return &tournamentAuthorizationAdapter{
		tournamentRepo: tournamentRepo,
	}
}

// IsOrganizer checks if the user is the organizer of the specified tournament
func (a *tournamentAuthorizationAdapter) IsOrganizer(ctx context.Context, userID uuid.UUID, tournamentID uuid.UUID) (bool, error) {
	if userID == uuid.Nil || tournamentID == uuid.Nil {
		return false, nil
	}

	tournament, err := a.tournamentRepo.FindByID(ctx, tournamentID)
	if err != nil {
		return false, fmt.Errorf("failed to find tournament %s: %w", tournamentID, err)
	}

	if tournament == nil {
		return false, nil
	}

	return tournament.OrganizerID == userID, nil
}

// IsPlatformAdmin checks if the current context has platform admin privileges
func (a *tournamentAuthorizationAdapter) IsPlatformAdmin(ctx context.Context) bool {
	return shared.IsAdmin(ctx)
}

// IsParticipant checks if the user is a registered participant in the tournament
func (a *tournamentAuthorizationAdapter) IsParticipant(ctx context.Context, userID uuid.UUID, tournamentID uuid.UUID) (bool, error) {
	if userID == uuid.Nil || tournamentID == uuid.Nil {
		return false, nil
	}

	tournament, err := a.tournamentRepo.FindByID(ctx, tournamentID)
	if err != nil {
		return false, fmt.Errorf("failed to find tournament %s: %w", tournamentID, err)
	}

	if tournament == nil {
		return false, nil
	}

	return tournament.IsPlayerRegistered(userID), nil
}
