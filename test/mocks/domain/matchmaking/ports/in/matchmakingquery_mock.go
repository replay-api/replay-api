package matchmaking_in

import (
	"context"

	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockMatchmakingQuery is a mock implementation of MatchmakingQuery
type MockMatchmakingQuery struct {
	mock.Mock
}

// GetSession provides a mock function
func (_m *MockMatchmakingQuery) GetSession(ctx context.Context, query matchmaking_in.GetSessionQuery) (*matchmaking_in.SessionDTO, error) {
	ret := _m.Called(ctx, query)

	var r0 *matchmaking_in.SessionDTO
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, matchmaking_in.GetSessionQuery) (*matchmaking_in.SessionDTO, error)); ok {
		return rf(ctx, query)
	}

	if rf, ok := ret.Get(0).(func(context.Context, matchmaking_in.GetSessionQuery) *matchmaking_in.SessionDTO); ok {
		r0 = rf(ctx, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*matchmaking_in.SessionDTO)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetPoolStats provides a mock function
func (_m *MockMatchmakingQuery) GetPoolStats(ctx context.Context, query matchmaking_in.GetPoolStatsQuery) (*matchmaking_in.PoolStatsDTO, error) {
	ret := _m.Called(ctx, query)

	var r0 *matchmaking_in.PoolStatsDTO
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, matchmaking_in.GetPoolStatsQuery) (*matchmaking_in.PoolStatsDTO, error)); ok {
		return rf(ctx, query)
	}

	if rf, ok := ret.Get(0).(func(context.Context, matchmaking_in.GetPoolStatsQuery) *matchmaking_in.PoolStatsDTO); ok {
		r0 = rf(ctx, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*matchmaking_in.PoolStatsDTO)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockMatchmakingQuery creates a new instance of MockMatchmakingQuery
func NewMockMatchmakingQuery(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMatchmakingQuery {
	mock := &MockMatchmakingQuery{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
