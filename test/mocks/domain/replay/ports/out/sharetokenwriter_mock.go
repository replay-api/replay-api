package replay_out

import (
	"context"

	"github.com/google/uuid"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	"github.com/stretchr/testify/mock"
)

// MockShareTokenWriter is a mock implementation of ShareTokenWriter
type MockShareTokenWriter struct {
	mock.Mock
}

// Create provides a mock function
func (_m *MockShareTokenWriter) Create(ctx context.Context, token *replay_entity.ShareToken) error {
	ret := _m.Called(ctx, token)

	return ret.Error(0)
}

// Update provides a mock function
func (_m *MockShareTokenWriter) Update(ctx context.Context, token *replay_entity.ShareToken) error {
	ret := _m.Called(ctx, token)

	return ret.Error(0)
}

// Delete provides a mock function
func (_m *MockShareTokenWriter) Delete(ctx context.Context, tokenID uuid.UUID) error {
	ret := _m.Called(ctx, tokenID)

	return ret.Error(0)
}

// NewMockShareTokenWriter creates a new instance of MockShareTokenWriter
func NewMockShareTokenWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockShareTokenWriter {
	mock := &MockShareTokenWriter{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
