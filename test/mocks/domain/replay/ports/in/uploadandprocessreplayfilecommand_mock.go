package replay_in

import (
	"context"

	io "io"

	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	"github.com/stretchr/testify/mock"
)

// MockUploadAndProcessReplayFileCommand is a mock implementation of UploadAndProcessReplayFileCommand
type MockUploadAndProcessReplayFileCommand struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockUploadAndProcessReplayFileCommand) Exec(c context.Context, file io.Reader) (*replay_entity.Match, error) {
	ret := _m.Called(c, file)

	var r0 *replay_entity.Match
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, io.Reader) (*replay_entity.Match, error)); ok {
		return rf(c, file)
	}

	if rf, ok := ret.Get(0).(func(context.Context, io.Reader) *replay_entity.Match); ok {
		r0 = rf(c, file)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*replay_entity.Match)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockUploadAndProcessReplayFileCommand creates a new instance of MockUploadAndProcessReplayFileCommand
func NewMockUploadAndProcessReplayFileCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockUploadAndProcessReplayFileCommand {
	mock := &MockUploadAndProcessReplayFileCommand{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
