package replay_in

import (
	"context"

	io "io"

	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	"github.com/stretchr/testify/mock"
)

// MockUploadReplayFileCommand is a mock implementation of UploadReplayFileCommand
type MockUploadReplayFileCommand struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockUploadReplayFileCommand) Exec(c context.Context, file io.Reader) (*replay_entity.ReplayFile, error) {
	ret := _m.Called(c, file)

	var r0 *replay_entity.ReplayFile
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, io.Reader) (*replay_entity.ReplayFile, error)); ok {
		return rf(c, file)
	}

	if rf, ok := ret.Get(0).(func(context.Context, io.Reader) *replay_entity.ReplayFile); ok {
		r0 = rf(c, file)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*replay_entity.ReplayFile)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockUploadReplayFileCommand creates a new instance of MockUploadReplayFileCommand
func NewMockUploadReplayFileCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockUploadReplayFileCommand {
	mock := &MockUploadReplayFileCommand{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
