// Package tournament_out defines outbound interfaces for tournament domain
package tournament_out

import (
	"context"

	"github.com/google/uuid"
)

// TournamentAuthorization defines the outbound port for tournament-related authorization checks.
// This port enables the tournament domain to verify permissions without directly depending
// on external auth systems, maintaining Clean Architecture boundaries.
//
// Permission Model (Financial-Grade):
//   - Create: Any authenticated user with appropriate subscription tier
//   - Update/Delete: Tournament organizer only, OR platform admin
//   - OpenRegistration/CloseRegistration: Tournament organizer only, OR platform admin
//   - Start/Complete/Cancel: Tournament organizer only, OR platform admin
//   - Register/Unregister: Any authenticated user (self-service)
//   - CheckIn: Registered participant only
//   - RecordMatchResult: Tournament organizer, OR platform admin
//   - AdvanceBracket: Tournament organizer, OR platform admin
type TournamentAuthorization interface {
	// IsOrganizer checks if the given user is the organizer of the specified tournament
	IsOrganizer(ctx context.Context, userID uuid.UUID, tournamentID uuid.UUID) (bool, error)

	// IsPlatformAdmin checks if the current context has platform admin privileges
	IsPlatformAdmin(ctx context.Context) bool

	// IsParticipant checks if the given user is a registered participant
	IsParticipant(ctx context.Context, userID uuid.UUID, tournamentID uuid.UUID) (bool, error)
}
