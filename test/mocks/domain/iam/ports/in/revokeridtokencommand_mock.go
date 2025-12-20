package iam_in

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockRevokeRIDTokenCommand is a mock implementation of RevokeRIDTokenCommand
type MockRevokeRIDTokenCommand struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockRevokeRIDTokenCommand) Exec(ctx context.Context, tokenID uuid.UUID) error {
	ret := _m.Called(ctx, tokenID)

	return ret.Error(0)
}

// NewMockRevokeRIDTokenCommand creates a new instance of MockRevokeRIDTokenCommand
func NewMockRevokeRIDTokenCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockRevokeRIDTokenCommand {
	mock := &MockRevokeRIDTokenCommand{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
