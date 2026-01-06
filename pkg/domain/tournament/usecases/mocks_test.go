package tournament_usecases_test

import (
	"context"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
	"github.com/stretchr/testify/mock"
)

// MockTournamentRepository implements tournament_out.TournamentRepository
type MockTournamentRepository struct {
	mock.Mock
}

func (m *MockTournamentRepository) Save(ctx context.Context, tournament *tournament_entities.Tournament) error {
	args := m.Called(ctx, tournament)
	return args.Error(0)
}

func (m *MockTournamentRepository) Update(ctx context.Context, tournament *tournament_entities.Tournament) error {
	args := m.Called(ctx, tournament)
	return args.Error(0)
}

func (m *MockTournamentRepository) FindByID(ctx context.Context, id uuid.UUID) (*tournament_entities.Tournament, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tournament_entities.Tournament), args.Error(1)
}

func (m *MockTournamentRepository) FindByOrganizer(ctx context.Context, organizerID uuid.UUID) ([]*tournament_entities.Tournament, error) {
	args := m.Called(ctx, organizerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*tournament_entities.Tournament), args.Error(1)
}

func (m *MockTournamentRepository) FindByGameAndRegion(ctx context.Context, gameID, region string, status []tournament_entities.TournamentStatus, limit int) ([]*tournament_entities.Tournament, error) {
	args := m.Called(ctx, gameID, region, status, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*tournament_entities.Tournament), args.Error(1)
}

func (m *MockTournamentRepository) FindUpcoming(ctx context.Context, gameID string, limit int) ([]*tournament_entities.Tournament, error) {
	args := m.Called(ctx, gameID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*tournament_entities.Tournament), args.Error(1)
}

func (m *MockTournamentRepository) FindInProgress(ctx context.Context, limit int) ([]*tournament_entities.Tournament, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*tournament_entities.Tournament), args.Error(1)
}

func (m *MockTournamentRepository) FindPlayerTournaments(ctx context.Context, playerID uuid.UUID, statusFilter []tournament_entities.TournamentStatus) ([]*tournament_entities.Tournament, error) {
	args := m.Called(ctx, playerID, statusFilter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*tournament_entities.Tournament), args.Error(1)
}

func (m *MockTournamentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTournamentRepository) GetByID(ctx context.Context, id uuid.UUID) (*tournament_entities.Tournament, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tournament_entities.Tournament), args.Error(1)
}

func (m *MockTournamentRepository) Search(ctx context.Context, s shared.Search) ([]tournament_entities.Tournament, error) {
	args := m.Called(ctx, s)
	if args.Get(0) == nil {
		return []tournament_entities.Tournament{}, args.Error(1)
	}
	return args.Get(0).([]tournament_entities.Tournament), args.Error(1)
}

func (m *MockTournamentRepository) Compile(ctx context.Context, searchParams []shared.SearchAggregation, resultOptions shared.SearchResultOptions) (*shared.Search, error) {
	args := m.Called(ctx, searchParams, resultOptions)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shared.Search), args.Error(1)
}