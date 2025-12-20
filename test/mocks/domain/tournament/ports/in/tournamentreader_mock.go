package tournament_in

import (
	"context"

	"github.com/google/uuid"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
	"github.com/stretchr/testify/mock"
)

// MockTournamentReader is a mock implementation of TournamentReader
type MockTournamentReader struct {
	mock.Mock
}

// GetTournament provides a mock function
func (_m *MockTournamentReader) GetTournament(ctx context.Context, id uuid.UUID) (*tournament_entities.Tournament, error) {
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

// ListTournaments provides a mock function
func (_m *MockTournamentReader) ListTournaments(ctx context.Context, gameID string, region string, status []tournament_entities.TournamentStatus, limit int) ([]*tournament_entities.Tournament, error) {
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

// GetUpcomingTournaments provides a mock function
func (_m *MockTournamentReader) GetUpcomingTournaments(ctx context.Context, gameID string, limit int) ([]*tournament_entities.Tournament, error) {
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

// GetPlayerTournaments provides a mock function
func (_m *MockTournamentReader) GetPlayerTournaments(ctx context.Context, playerID uuid.UUID) ([]*tournament_entities.Tournament, error) {
	ret := _m.Called(ctx, playerID)

	var r0 []*tournament_entities.Tournament
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) ([]*tournament_entities.Tournament, error)); ok {
		return rf(ctx, playerID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) []*tournament_entities.Tournament); ok {
		r0 = rf(ctx, playerID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*tournament_entities.Tournament)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetOrganizerTournaments provides a mock function
func (_m *MockTournamentReader) GetOrganizerTournaments(ctx context.Context, organizerID uuid.UUID) ([]*tournament_entities.Tournament, error) {
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

// NewMockTournamentReader creates a new instance of MockTournamentReader
func NewMockTournamentReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockTournamentReader {
	mock := &MockTournamentReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
