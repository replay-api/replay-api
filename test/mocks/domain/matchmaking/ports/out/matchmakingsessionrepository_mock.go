package matchmaking_out

import (
	"context"

	"github.com/google/uuid"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_out "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/out"
	"github.com/stretchr/testify/mock"
)

// MockMatchmakingSessionRepository is a mock implementation of MatchmakingSessionRepository
type MockMatchmakingSessionRepository struct {
	mock.Mock
}

// Save provides a mock function
func (_m *MockMatchmakingSessionRepository) Save(ctx context.Context, session *matchmaking_entities.MatchmakingSession) error {
	ret := _m.Called(ctx, session)

	return ret.Error(0)
}

// GetByID provides a mock function
func (_m *MockMatchmakingSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.MatchmakingSession, error) {
	ret := _m.Called(ctx, id)

	var r0 *matchmaking_entities.MatchmakingSession
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*matchmaking_entities.MatchmakingSession, error)); ok {
		return rf(ctx, id)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *matchmaking_entities.MatchmakingSession); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*matchmaking_entities.MatchmakingSession)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetByPlayerID provides a mock function
func (_m *MockMatchmakingSessionRepository) GetByPlayerID(ctx context.Context, playerID uuid.UUID) ([]*matchmaking_entities.MatchmakingSession, error) {
	ret := _m.Called(ctx, playerID)

	var r0 []*matchmaking_entities.MatchmakingSession
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) ([]*matchmaking_entities.MatchmakingSession, error)); ok {
		return rf(ctx, playerID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) []*matchmaking_entities.MatchmakingSession); ok {
		r0 = rf(ctx, playerID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*matchmaking_entities.MatchmakingSession)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetActiveSessions provides a mock function
func (_m *MockMatchmakingSessionRepository) GetActiveSessions(ctx context.Context, filters matchmaking_out.SessionFilters) ([]*matchmaking_entities.MatchmakingSession, error) {
	ret := _m.Called(ctx, filters)

	var r0 []*matchmaking_entities.MatchmakingSession
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, matchmaking_out.SessionFilters) ([]*matchmaking_entities.MatchmakingSession, error)); ok {
		return rf(ctx, filters)
	}

	if rf, ok := ret.Get(0).(func(context.Context, matchmaking_out.SessionFilters) []*matchmaking_entities.MatchmakingSession); ok {
		r0 = rf(ctx, filters)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*matchmaking_entities.MatchmakingSession)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// UpdateStatus provides a mock function
func (_m *MockMatchmakingSessionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status matchmaking_entities.SessionStatus) error {
	ret := _m.Called(ctx, id, status)

	return ret.Error(0)
}

// Delete provides a mock function
func (_m *MockMatchmakingSessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ret := _m.Called(ctx, id)

	return ret.Error(0)
}

// DeleteExpired provides a mock function
func (_m *MockMatchmakingSessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	ret := _m.Called(ctx)

	var r0 int64
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context) (int64, error)); ok {
		return rf(ctx)
	}

	if rf, ok := ret.Get(0).(func(context.Context) int64); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(int64)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockMatchmakingSessionRepository creates a new instance of MockMatchmakingSessionRepository
func NewMockMatchmakingSessionRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMatchmakingSessionRepository {
	mock := &MockMatchmakingSessionRepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
