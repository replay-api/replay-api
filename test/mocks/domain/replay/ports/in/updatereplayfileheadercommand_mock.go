package replay_in

import (
	"context"

	"github.com/google/uuid"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	"github.com/stretchr/testify/mock"
)

// MockUpdateReplayFileHeaderCommand is a mock implementation of UpdateReplayFileHeaderCommand
type MockUpdateReplayFileHeaderCommand struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockUpdateReplayFileHeaderCommand) Exec(ctx context.Context, replayFileID uuid.UUID) (*replay_entity.ReplayFile, error) {
	ret := _m.Called(ctx, replayFileID)

	var r0 *replay_entity.ReplayFile
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*replay_entity.ReplayFile, error)); ok {
		return rf(ctx, replayFileID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *replay_entity.ReplayFile); ok {
		r0 = rf(ctx, replayFileID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*replay_entity.ReplayFile)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockUpdateReplayFileHeaderCommand creates a new instance of MockUpdateReplayFileHeaderCommand
func NewMockUpdateReplayFileHeaderCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockUpdateReplayFileHeaderCommand {
	mock := &MockUpdateReplayFileHeaderCommand{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
