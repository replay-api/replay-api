package replay_in

import (
	"context"

	io "io"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockReplayContentReader is a mock implementation of ReplayContentReader
type MockReplayContentReader struct {
	mock.Mock
}

// GetByID provides a mock function
func (_m *MockReplayContentReader) GetByID(ctx context.Context, replayFileID uuid.UUID) (io.ReadSeekCloser, error) {
	ret := _m.Called(ctx, replayFileID)

	var r0 io.ReadSeekCloser
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (io.ReadSeekCloser, error)); ok {
		return rf(ctx, replayFileID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) io.ReadSeekCloser); ok {
		r0 = rf(ctx, replayFileID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.ReadSeekCloser)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockReplayContentReader creates a new instance of MockReplayContentReader
func NewMockReplayContentReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockReplayContentReader {
	mock := &MockReplayContentReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
