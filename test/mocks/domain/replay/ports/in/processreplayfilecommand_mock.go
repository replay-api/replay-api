package replay_in

import (
	"context"

	"github.com/google/uuid"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	"github.com/stretchr/testify/mock"
)

// MockProcessReplayFileCommand is a mock implementation of ProcessReplayFileCommand
type MockProcessReplayFileCommand struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockProcessReplayFileCommand) Exec(c context.Context, replayFileID uuid.UUID) (*replay_entity.Match, error) {
	ret := _m.Called(c, replayFileID)

	var r0 *replay_entity.Match
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*replay_entity.Match, error)); ok {
		return rf(c, replayFileID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *replay_entity.Match); ok {
		r0 = rf(c, replayFileID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*replay_entity.Match)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockProcessReplayFileCommand creates a new instance of MockProcessReplayFileCommand
func NewMockProcessReplayFileCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockProcessReplayFileCommand {
	mock := &MockProcessReplayFileCommand{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
