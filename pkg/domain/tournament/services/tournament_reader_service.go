package tournament_services

import (
	"context"

	"github.com/google/uuid"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
	tournament_in "github.com/replay-api/replay-api/pkg/domain/tournament/ports/in"
)

// TournamentReaderService implements tournament query operations
type TournamentReaderService struct {
	tournamentQueryService *TournamentQueryService
}

// NewTournamentReaderService creates a new tournament reader service
func NewTournamentReaderService(
	tournamentQueryService *TournamentQueryService,
) tournament_in.TournamentReader {
	return &TournamentReaderService{
		tournamentQueryService: tournamentQueryService,
	}
}

// GetTournament retrieves a tournament by ID
func (s *TournamentReaderService) GetTournament(ctx context.Context, id uuid.UUID) (*tournament_entities.Tournament, error) {
	return s.tournamentQueryService.GetByID(ctx, id)
}

// ListTournaments retrieves tournaments with filters
func (s *TournamentReaderService) ListTournaments(ctx context.Context, gameID, region string, status []tournament_entities.TournamentStatus, limit int) ([]*tournament_entities.Tournament, error) {
	return s.tournamentQueryService.FindByGameAndRegion(ctx, gameID, region, status, limit)
}

// GetUpcomingTournaments retrieves tournaments accepting registrations or starting soon
func (s *TournamentReaderService) GetUpcomingTournaments(ctx context.Context, gameID string, limit int) ([]*tournament_entities.Tournament, error) {
	return s.tournamentQueryService.FindUpcoming(ctx, gameID, limit)
}

// GetPlayerTournaments retrieves tournaments a player is registered in
func (s *TournamentReaderService) GetPlayerTournaments(ctx context.Context, playerID uuid.UUID) ([]*tournament_entities.Tournament, error) {
	// This would need a more complex query involving tournament registrations
	// For now, return empty slice as this requires additional domain modeling
	return []*tournament_entities.Tournament{}, nil
}

// GetOrganizerTournaments retrieves tournaments created by a specific organizer
func (s *TournamentReaderService) GetOrganizerTournaments(ctx context.Context, organizerID uuid.UUID) ([]*tournament_entities.Tournament, error) {
	return s.tournamentQueryService.FindByOrganizer(ctx, organizerID)
}

// Ensure TournamentReaderService implements TournamentReader interface
var _ tournament_in.TournamentReader = (*TournamentReaderService)(nil)
