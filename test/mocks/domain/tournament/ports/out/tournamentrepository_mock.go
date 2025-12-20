package tournament_out

import (
	"context"

	"github.com/google/uuid"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
	"github.com/stretchr/testify/mock"
)

// MockTournamentRepository is a mock implementation of TournamentRepository
type MockTournamentRepository struct {
	mock.Mock
}

// Save provides a mock function
func (_m *MockTournamentRepository) Save(ctx context.Context, tournament *tournament_entities.Tournament) error {
	ret := _m.Called(ctx, tournament)

	return ret.Error(0)
}

// FindByID provides a mock function
func (_m *MockTournamentRepository) FindByID(ctx context.Context, id uuid.UUID) (*tournament_entities.Tournament, error) {
	ret := _m.Called(ctx, id)

	var r0 *tournament_entities.Tournament
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*tournament_entities.Tournament, error)); ok {
		return rf(ctx, id)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *tournament_entities.Tournament); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tournament_entities.Tournament)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindByOrganizer provides a mock function
func (_m *MockTournamentRepository) FindByOrganizer(ctx context.Context, organizerID uuid.UUID) ([]*tournament_entities.Tournament, error) {
	ret := _m.Called(ctx, organizerID)

	var r0 []*tournament_entities.Tournament
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) ([]*tournament_entities.Tournament, error)); ok {
		return rf(ctx, organizerID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) []*tournament_entities.Tournament); ok {
		r0 = rf(ctx, organizerID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*tournament_entities.Tournament)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindByGameAndRegion provides a mock function
func (_m *MockTournamentRepository) FindByGameAndRegion(ctx context.Context, gameID string, region string, status []tournament_entities.TournamentStatus, limit int) ([]*tournament_entities.Tournament, error) {
	ret := _m.Called(ctx, gameID, region, status, limit)

	var r0 []*tournament_entities.Tournament
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string, string, []tournament_entities.TournamentStatus, int) ([]*tournament_entities.Tournament, error)); ok {
		return rf(ctx, gameID, region, status, limit)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string, string, []tournament_entities.TournamentStatus, int) []*tournament_entities.Tournament); ok {
		r0 = rf(ctx, gameID, region, status, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*tournament_entities.Tournament)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindUpcoming provides a mock function
func (_m *MockTournamentRepository) FindUpcoming(ctx context.Context, gameID string, limit int) ([]*tournament_entities.Tournament, error) {
	ret := _m.Called(ctx, gameID, limit)

	var r0 []*tournament_entities.Tournament
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string, int) ([]*tournament_entities.Tournament, error)); ok {
		return rf(ctx, gameID, limit)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string, int) []*tournament_entities.Tournament); ok {
		r0 = rf(ctx, gameID, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*tournament_entities.Tournament)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindInProgress provides a mock function
func (_m *MockTournamentRepository) FindInProgress(ctx context.Context, limit int) ([]*tournament_entities.Tournament, error) {
	ret := _m.Called(ctx, limit)

	var r0 []*tournament_entities.Tournament
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, int) ([]*tournament_entities.Tournament, error)); ok {
		return rf(ctx, limit)
	}

	if rf, ok := ret.Get(0).(func(context.Context, int) []*tournament_entities.Tournament); ok {
		r0 = rf(ctx, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*tournament_entities.Tournament)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Update provides a mock function
func (_m *MockTournamentRepository) Update(ctx context.Context, tournament *tournament_entities.Tournament) error {
	ret := _m.Called(ctx, tournament)

	return ret.Error(0)
}

// Delete provides a mock function
func (_m *MockTournamentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ret := _m.Called(ctx, id)

	return ret.Error(0)
}

// FindPlayerTournaments provides a mock function
func (_m *MockTournamentRepository) FindPlayerTournaments(ctx context.Context, playerID uuid.UUID, statusFilter []tournament_entities.TournamentStatus) ([]*tournament_entities.Tournament, error) {
	ret := _m.Called(ctx, playerID, statusFilter)

	var r0 []*tournament_entities.Tournament
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, []tournament_entities.TournamentStatus) ([]*tournament_entities.Tournament, error)); ok {
		return rf(ctx, playerID, statusFilter)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, []tournament_entities.TournamentStatus) []*tournament_entities.Tournament); ok {
		r0 = rf(ctx, playerID, statusFilter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*tournament_entities.Tournament)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockTournamentRepository creates a new instance of MockTournamentRepository
func NewMockTournamentRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockTournamentRepository {
	mock := &MockTournamentRepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
