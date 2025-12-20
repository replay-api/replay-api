package replay_out

import (
	"context"

	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	"github.com/stretchr/testify/mock"
)

// MockPlayerMetadataWriter is a mock implementation of PlayerMetadataWriter
type MockPlayerMetadataWriter struct {
	mock.Mock
}

// Create provides a mock function
func (_m *MockPlayerMetadataWriter) Create(createCtx context.Context, player replay_entity.PlayerMetadata) error {
	ret := _m.Called(createCtx, player)

	return ret.Error(0)
}

// CreateMany provides a mock function
func (_m *MockPlayerMetadataWriter) CreateMany(createCtx context.Context, players []replay_entity.PlayerMetadata) error {
	ret := _m.Called(createCtx, players)

	return ret.Error(0)
}

// NewMockPlayerMetadataWriter creates a new instance of MockPlayerMetadataWriter
func NewMockPlayerMetadataWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPlayerMetadataWriter {
	mock := &MockPlayerMetadataWriter{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
