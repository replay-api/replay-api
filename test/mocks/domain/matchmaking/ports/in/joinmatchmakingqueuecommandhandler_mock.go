package matchmaking_in

import (
	"context"

	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockJoinMatchmakingQueueCommandHandler is a mock implementation of JoinMatchmakingQueueCommandHandler
type MockJoinMatchmakingQueueCommandHandler struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockJoinMatchmakingQueueCommandHandler) Exec(ctx context.Context, cmd matchmaking_in.JoinMatchmakingQueueCommand) (*matchmaking_entities.MatchmakingSession, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *matchmaking_entities.MatchmakingSession
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, matchmaking_in.JoinMatchmakingQueueCommand) (*matchmaking_entities.MatchmakingSession, error)); ok {
		return rf(ctx, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, matchmaking_in.JoinMatchmakingQueueCommand) *matchmaking_entities.MatchmakingSession); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*matchmaking_entities.MatchmakingSession)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockJoinMatchmakingQueueCommandHandler creates a new instance of MockJoinMatchmakingQueueCommandHandler
func NewMockJoinMatchmakingQueueCommandHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockJoinMatchmakingQueueCommandHandler {
	mock := &MockJoinMatchmakingQueueCommandHandler{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
