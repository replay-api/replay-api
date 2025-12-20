package matchmaking_in

import (
	"context"

	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockLeaveMatchmakingQueueCommandHandler is a mock implementation of LeaveMatchmakingQueueCommandHandler
type MockLeaveMatchmakingQueueCommandHandler struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockLeaveMatchmakingQueueCommandHandler) Exec(ctx context.Context, cmd matchmaking_in.LeaveMatchmakingQueueCommand) error {
	ret := _m.Called(ctx, cmd)

	return ret.Error(0)
}

// NewMockLeaveMatchmakingQueueCommandHandler creates a new instance of MockLeaveMatchmakingQueueCommandHandler
func NewMockLeaveMatchmakingQueueCommandHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockLeaveMatchmakingQueueCommandHandler {
	mock := &MockLeaveMatchmakingQueueCommandHandler{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
