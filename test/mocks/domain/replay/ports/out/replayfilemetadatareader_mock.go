package replay_out

import (
	"context"

	"github.com/google/uuid"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	"github.com/stretchr/testify/mock"
)

// MockReplayFileMetadataReader is a mock implementation of ReplayFileMetadataReader
type MockReplayFileMetadataReader struct {
	mock.Mock
}

// GetByID provides a mock function
func (_m *MockReplayFileMetadataReader) GetByID(ctx context.Context, replayFileID uuid.UUID) (*replay_entity.ReplayFile, error) {
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

// NewMockReplayFileMetadataReader creates a new instance of MockReplayFileMetadataReader
func NewMockReplayFileMetadataReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockReplayFileMetadataReader {
	mock := &MockReplayFileMetadataReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
