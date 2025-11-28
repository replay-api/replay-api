// Package tournament_out defines outbound repository interfaces
package tournament_out

import (
	"context"

	"github.com/google/uuid"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
)

// TournamentRepository defines persistence operations for tournaments
type TournamentRepository interface {
	// Save creates a new tournament
	Save(ctx context.Context, tournament *tournament_entities.Tournament) error

	// FindByID retrieves a tournament by ID
	FindByID(ctx context.Context, id uuid.UUID) (*tournament_entities.Tournament, error)

	// FindByOrganizer retrieves tournaments organized by a specific user
	FindByOrganizer(ctx context.Context, organizerID uuid.UUID) ([]*tournament_entities.Tournament, error)

	// FindByGameAndRegion retrieves tournaments for a game/region
	FindByGameAndRegion(ctx context.Context, gameID, region string, status []tournament_entities.TournamentStatus, limit int) ([]*tournament_entities.Tournament, error)

	// FindUpcoming retrieves upcoming tournaments (registration or ready status)
	FindUpcoming(ctx context.Context, gameID string, limit int) ([]*tournament_entities.Tournament, error)

	// FindInProgress retrieves currently running tournaments
	FindInProgress(ctx context.Context, limit int) ([]*tournament_entities.Tournament, error)

	// Update updates an existing tournament
	Update(ctx context.Context, tournament *tournament_entities.Tournament) error

	// Delete removes a tournament (soft delete recommended)
	Delete(ctx context.Context, id uuid.UUID) error

	// FindPlayerTournaments retrieves tournaments a player is registered in
	FindPlayerTournaments(ctx context.Context, playerID uuid.UUID, statusFilter []tournament_entities.TournamentStatus) ([]*tournament_entities.Tournament, error)
}
