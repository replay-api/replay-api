package matchmaking_services

import (
	"context"

	"github.com/google/uuid"
	matchmaking_services "github.com/replay-api/replay-api/pkg/domain/matchmaking/services"
	"github.com/stretchr/testify/mock"
)

// MockMatchStatsRepository is a mock implementation of MatchStatsRepository
type MockMatchStatsRepository struct {
	mock.Mock
}

// GetPlayerMatchStats provides a mock function
func (_m *MockMatchStatsRepository) GetPlayerMatchStats(ctx context.Context, playerID uuid.UUID, limit int) ([]matchmaking_services.MatchStatsSummary, error) {
	ret := _m.Called(ctx, playerID, limit)

	var r0 []matchmaking_services.MatchStatsSummary
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, int) ([]matchmaking_services.MatchStatsSummary, error)); ok {
		return rf(ctx, playerID, limit)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, int) []matchmaking_services.MatchStatsSummary); ok {
		r0 = rf(ctx, playerID, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]matchmaking_services.MatchStatsSummary)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetPlayerMatchCount provides a mock function
func (_m *MockMatchStatsRepository) GetPlayerMatchCount(ctx context.Context, playerID uuid.UUID) (int, error) {
	ret := _m.Called(ctx, playerID)

	var r0 int
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (int, error)); ok {
		return rf(ctx, playerID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) int); ok {
		r0 = rf(ctx, playerID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(int)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockMatchStatsRepository creates a new instance of MockMatchStatsRepository
func NewMockMatchStatsRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMatchStatsRepository {
	mock := &MockMatchStatsRepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
