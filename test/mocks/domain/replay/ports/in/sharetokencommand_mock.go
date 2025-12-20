package replay_in

import (
	"context"

	"github.com/google/uuid"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	"github.com/stretchr/testify/mock"
)

// MockShareTokenCommand is a mock implementation of ShareTokenCommand
type MockShareTokenCommand struct {
	mock.Mock
}

// Create provides a mock function
func (_m *MockShareTokenCommand) Create(ctx context.Context, token *replay_entity.ShareToken) error {
	ret := _m.Called(ctx, token)

	return ret.Error(0)
}

// Revoke provides a mock function
func (_m *MockShareTokenCommand) Revoke(ctx context.Context, tokenID uuid.UUID) error {
	ret := _m.Called(ctx, tokenID)

	return ret.Error(0)
}

// Update provides a mock function
func (_m *MockShareTokenCommand) Update(ctx context.Context, token *replay_entity.ShareToken) error {
	ret := _m.Called(ctx, token)

	return ret.Error(0)
}

// NewMockShareTokenCommand creates a new instance of MockShareTokenCommand
func NewMockShareTokenCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockShareTokenCommand {
	mock := &MockShareTokenCommand{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
