package tournament_services

import (
	"context"

	"github.com/google/uuid"
	tournament_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/tournament/entities"
	tournament_in "github.com/psavelis/team-pro/replay-api/pkg/domain/tournament/ports/in"
	tournament_out "github.com/psavelis/team-pro/replay-api/pkg/domain/tournament/ports/out"
)

// TournamentReaderService implements tournament query operations
type TournamentReaderService struct {
	tournamentRepo tournament_out.TournamentRepository
}

// NewTournamentReaderService creates a new tournament reader service
func NewTournamentReaderService(
	tournamentRepo tournament_out.TournamentRepository,
) tournament_in.TournamentReader {
	return &TournamentReaderService{
		tournamentRepo: tournamentRepo,
	}
}

// GetTournament retrieves a tournament by ID
func (s *TournamentReaderService) GetTournament(ctx context.Context, id uuid.UUID) (*tournament_entities.Tournament, error) {
	return s.tournamentRepo.FindByID(ctx, id)
}

// ListTournaments retrieves tournaments with filters
func (s *TournamentReaderService) ListTournaments(ctx context.Context, gameID, region string, status []tournament_entities.TournamentStatus, limit int) ([]*tournament_entities.Tournament, error) {
	return s.tournamentRepo.FindByGameAndRegion(ctx, gameID, region, status, limit)
}

// GetUpcomingTournaments retrieves tournaments accepting registrations or starting soon
func (s *TournamentReaderService) GetUpcomingTournaments(ctx context.Context, gameID string, limit int) ([]*tournament_entities.Tournament, error) {
	return s.tournamentRepo.FindUpcoming(ctx, gameID, limit)
}

// GetPlayerTournaments retrieves tournaments a player is registered in
func (s *TournamentReaderService) GetPlayerTournaments(ctx context.Context, playerID uuid.UUID) ([]*tournament_entities.Tournament, error) {
	return s.tournamentRepo.FindPlayerTournaments(ctx, playerID, nil)
}

// GetOrganizerTournaments retrieves tournaments created by a specific organizer
func (s *TournamentReaderService) GetOrganizerTournaments(ctx context.Context, organizerID uuid.UUID) ([]*tournament_entities.Tournament, error) {
	return s.tournamentRepo.FindByOrganizer(ctx, organizerID)
}

// Ensure TournamentReaderService implements TournamentReader interface
var _ tournament_in.TournamentReader = (*TournamentReaderService)(nil)
