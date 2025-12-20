package matchmaking_out

import (
	"context"

	"github.com/google/uuid"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	"github.com/stretchr/testify/mock"
)

// MockMatchmakingPoolRepository is a mock implementation of MatchmakingPoolRepository
type MockMatchmakingPoolRepository struct {
	mock.Mock
}

// Save provides a mock function
func (_m *MockMatchmakingPoolRepository) Save(ctx context.Context, pool *matchmaking_entities.MatchmakingPool) error {
	ret := _m.Called(ctx, pool)

	return ret.Error(0)
}

// GetByID provides a mock function
func (_m *MockMatchmakingPoolRepository) GetByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.MatchmakingPool, error) {
	ret := _m.Called(ctx, id)

	var r0 *matchmaking_entities.MatchmakingPool
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*matchmaking_entities.MatchmakingPool, error)); ok {
		return rf(ctx, id)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *matchmaking_entities.MatchmakingPool); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*matchmaking_entities.MatchmakingPool)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetByGameModeRegion provides a mock function
func (_m *MockMatchmakingPoolRepository) GetByGameModeRegion(ctx context.Context, gameID string, gameMode string, region string) (*matchmaking_entities.MatchmakingPool, error) {
	ret := _m.Called(ctx, gameID, gameMode, region)

	var r0 *matchmaking_entities.MatchmakingPool
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) (*matchmaking_entities.MatchmakingPool, error)); ok {
		return rf(ctx, gameID, gameMode, region)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) *matchmaking_entities.MatchmakingPool); ok {
		r0 = rf(ctx, gameID, gameMode, region)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*matchmaking_entities.MatchmakingPool)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// UpdateStats provides a mock function
func (_m *MockMatchmakingPoolRepository) UpdateStats(ctx context.Context, poolID uuid.UUID, stats matchmaking_entities.PoolStatistics) error {
	ret := _m.Called(ctx, poolID, stats)

	return ret.Error(0)
}

// GetAllActive provides a mock function
func (_m *MockMatchmakingPoolRepository) GetAllActive(ctx context.Context) ([]*matchmaking_entities.MatchmakingPool, error) {
	ret := _m.Called(ctx)

	var r0 []*matchmaking_entities.MatchmakingPool
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context) ([]*matchmaking_entities.MatchmakingPool, error)); ok {
		return rf(ctx)
	}

	if rf, ok := ret.Get(0).(func(context.Context) []*matchmaking_entities.MatchmakingPool); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*matchmaking_entities.MatchmakingPool)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockMatchmakingPoolRepository creates a new instance of MockMatchmakingPoolRepository
func NewMockMatchmakingPoolRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMatchmakingPoolRepository {
	mock := &MockMatchmakingPoolRepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
