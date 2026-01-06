package tournament_services

import (
	"context"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
)

// TournamentQueryService provides domain query operations for tournaments using technology-agnostic search patterns
type TournamentQueryService struct {
	reader          shared.Searchable[tournament_entities.Tournament]
	queryableFields map[string]bool
}

// NewTournamentQueryService creates a new tournament query service
func NewTournamentQueryService(tournamentReader shared.Searchable[tournament_entities.Tournament]) *TournamentQueryService {
	queryableFields := map[string]bool{
		"ID":                true,
		"Name":              true,
		"Description":       true,
		"GameID":            true,
		"GameMode":          true,
		"Region":            true,
		"Format":            true,
		"MaxParticipants":   true,
		"MinParticipants":   true,
		"Status":            true,
		"StartTime":         true,
		"EndTime":           true,
		"RegistrationOpen":  true,
		"RegistrationClose": true,
		"OrganizerID":       true,
		"CreatedAt":         true,
		"UpdatedAt":         true,
	}

	return &TournamentQueryService{
		reader:          tournamentReader,
		queryableFields: queryableFields,
	}
}

// GetByID retrieves a single tournament by its ID
func (s *TournamentQueryService) GetByID(ctx context.Context, id uuid.UUID) (*tournament_entities.Tournament, error) {
	return s.reader.GetByID(ctx, id)
}

// FindByOrganizer finds tournaments created by a specific organizer
// Business rule: Tournaments sorted by most recent start time first
func (s *TournamentQueryService) FindByOrganizer(ctx context.Context, organizerID uuid.UUID) ([]*tournament_entities.Tournament, error) {
	search := shared.NewSearchBuilder().
		WithAggregation(shared.NewSearchAggregation().
			WithValueParam("OrganizerID", organizerID).
			Build()).
		WithSort("StartTime", shared.DescendingIDKey).
		Build()

	entities, err := s.reader.Search(ctx, search)
	if err != nil {
		return nil, err
	}

	// Convert to pointer slice
	tournaments := make([]*tournament_entities.Tournament, len(entities))
	for i := range entities {
		tournaments[i] = &entities[i]
	}

	return tournaments, nil
}

// FindByGameAndRegion finds tournaments by game and region with optional status filtering
// Business rule: Tournaments sorted by start time (upcoming first)
func (s *TournamentQueryService) FindByGameAndRegion(ctx context.Context, gameID, region string, statusFilter []tournament_entities.TournamentStatus, limit int) ([]*tournament_entities.Tournament, error) {
	// Build aggregation with filters
	aggBuilder := shared.NewSearchAggregation().NewParam().
		WithValueParam("GameID", gameID)

	if region != "" {
		aggBuilder.WithValueParam("Region", region)
	}

	if len(statusFilter) > 0 {
		statusValues := make([]interface{}, len(statusFilter))
		for i, status := range statusFilter {
			statusValues[i] = status
		}
		aggBuilder.WithValueParam("Status", statusValues...)
	}

	builder := shared.NewSearchBuilder().
		WithAggregation(aggBuilder.Build()).
		WithSort("StartTime", shared.AscendingIDKey)

	if limit > 0 {
		builder.WithLimit(uint(limit))
	}

	search := builder.Build()

	entities, err := s.reader.Search(ctx, search)
	if err != nil {
		return nil, err
	}

	// Convert to pointer slice
	tournaments := make([]*tournament_entities.Tournament, len(entities))
	for i := range entities {
		tournaments[i] = &entities[i]
	}

	return tournaments, nil
}

// FindUpcoming finds tournaments that are accepting registrations or starting soon
// Business rule: Tournaments in registration or ready status with start time in the future
func (s *TournamentQueryService) FindUpcoming(ctx context.Context, gameID string, limit int) ([]*tournament_entities.Tournament, error) {
	now := time.Now().UTC()
	search := shared.NewSearchBuilder().
		WithAggregation(shared.NewSearchAggregation().
			NewParam().
			WithValueParam("GameID", gameID).
			WithValueParam("Status", tournament_entities.TournamentStatusRegistration, tournament_entities.TournamentStatusReady).
			WithDateParam("StartTime", &now, nil). // StartTime > now
			Build()).
		WithSort("StartTime", shared.AscendingIDKey).
		WithLimit(uint(limit)).
		Build()

	entities, err := s.reader.Search(ctx, search)
	if err != nil {
		return nil, err
	}

	// Convert to pointer slice
	tournaments := make([]*tournament_entities.Tournament, len(entities))
	for i := range entities {
		tournaments[i] = &entities[i]
	}

	return tournaments, nil
}

// FindByStatus finds tournaments by status
// Business rule: General status-based filtering
func (s *TournamentQueryService) FindByStatus(ctx context.Context, status tournament_entities.TournamentStatus, limit int) ([]*tournament_entities.Tournament, error) {
	search := shared.NewSearchBuilder().
		WithAggregation(shared.NewSearchAggregation().
			WithValueParam("Status", status).
			Build()).
		WithSort("StartTime", shared.AscendingIDKey).
		WithLimit(uint(limit)).
		Build()

	entities, err := s.reader.Search(ctx, search)
	if err != nil {
		return nil, err
	}

	// Convert to pointer slice
	tournaments := make([]*tournament_entities.Tournament, len(entities))
	for i := range entities {
		tournaments[i] = &entities[i]
	}

	return tournaments, nil
}
