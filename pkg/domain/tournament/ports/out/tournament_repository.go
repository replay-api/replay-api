// Package tournament_out defines outbound repository interfaces
package tournament_out

import (
	"context"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
)

// TournamentRepository defines persistence operations for tournaments
type TournamentRepository interface {
	shared.Searchable[tournament_entities.Tournament]

	// Save creates a new tournament
	Save(ctx context.Context, tournament *tournament_entities.Tournament) error

	// FindByID retrieves a tournament by ID
	FindByID(ctx context.Context, id uuid.UUID) (*tournament_entities.Tournament, error)

	// Update updates an existing tournament
	Update(ctx context.Context, tournament *tournament_entities.Tournament) error

	// Delete removes a tournament (soft delete recommended)
	Delete(ctx context.Context, id uuid.UUID) error

	// FindPlayerTournaments retrieves tournaments where the player is a participant
	FindPlayerTournaments(ctx context.Context, playerID uuid.UUID, statusFilter []tournament_entities.TournamentStatus) ([]*tournament_entities.Tournament, error)

	// FindByMatchID finds the tournament containing a specific match
	FindByMatchID(ctx context.Context, matchID uuid.UUID) (*tournament_entities.Tournament, error)
}
