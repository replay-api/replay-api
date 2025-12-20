package replay_out

import (
	"context"

	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	"github.com/stretchr/testify/mock"
)

// MockMatchMetadataWriter is a mock implementation of MatchMetadataWriter
type MockMatchMetadataWriter struct {
	mock.Mock
}

// Create provides a mock function
func (_m *MockMatchMetadataWriter) Create(createCtx context.Context, match replay_entity.Match) error {
	ret := _m.Called(createCtx, match)

	return ret.Error(0)
}

// CreateMany provides a mock function
func (_m *MockMatchMetadataWriter) CreateMany(createCtx context.Context, matches []replay_entity.Match) error {
	ret := _m.Called(createCtx, matches)

	return ret.Error(0)
}

// NewMockMatchMetadataWriter creates a new instance of MockMatchMetadataWriter
func NewMockMatchMetadataWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMatchMetadataWriter {
	mock := &MockMatchMetadataWriter{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
