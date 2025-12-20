package replay_out

import (
	"context"

	io "io"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockReplayFileContentWriter is a mock implementation of ReplayFileContentWriter
type MockReplayFileContentWriter struct {
	mock.Mock
}

// Put provides a mock function
func (_m *MockReplayFileContentWriter) Put(createCtx context.Context, replayFileID uuid.UUID, reader io.ReadSeeker) (string, error) {
	ret := _m.Called(createCtx, replayFileID, reader)

	var r0 string
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, io.ReadSeeker) (string, error)); ok {
		return rf(createCtx, replayFileID, reader)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, io.ReadSeeker) string); ok {
		r0 = rf(createCtx, replayFileID, reader)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(string)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockReplayFileContentWriter creates a new instance of MockReplayFileContentWriter
func NewMockReplayFileContentWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockReplayFileContentWriter {
	mock := &MockReplayFileContentWriter{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
