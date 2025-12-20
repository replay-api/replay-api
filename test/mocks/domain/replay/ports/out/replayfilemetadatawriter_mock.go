package replay_out

import (
	"context"

	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	"github.com/stretchr/testify/mock"
)

// MockReplayFileMetadataWriter is a mock implementation of ReplayFileMetadataWriter
type MockReplayFileMetadataWriter struct {
	mock.Mock
}

// Create provides a mock function
func (_m *MockReplayFileMetadataWriter) Create(createCtx context.Context, replayFile *replay_entity.ReplayFile) (*replay_entity.ReplayFile, error) {
	ret := _m.Called(createCtx, replayFile)

	var r0 *replay_entity.ReplayFile
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *replay_entity.ReplayFile) (*replay_entity.ReplayFile, error)); ok {
		return rf(createCtx, replayFile)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *replay_entity.ReplayFile) *replay_entity.ReplayFile); ok {
		r0 = rf(createCtx, replayFile)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*replay_entity.ReplayFile)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Update provides a mock function
func (_m *MockReplayFileMetadataWriter) Update(createCtx context.Context, replayFile *replay_entity.ReplayFile) (*replay_entity.ReplayFile, error) {
	ret := _m.Called(createCtx, replayFile)

	var r0 *replay_entity.ReplayFile
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *replay_entity.ReplayFile) (*replay_entity.ReplayFile, error)); ok {
		return rf(createCtx, replayFile)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *replay_entity.ReplayFile) *replay_entity.ReplayFile); ok {
		r0 = rf(createCtx, replayFile)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*replay_entity.ReplayFile)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockReplayFileMetadataWriter creates a new instance of MockReplayFileMetadataWriter
func NewMockReplayFileMetadataWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockReplayFileMetadataWriter {
	mock := &MockReplayFileMetadataWriter{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
