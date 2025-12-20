package billing_in

import (
	"context"

	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockCreateSubscriptionCommandHandler is a mock implementation of CreateSubscriptionCommandHandler
type MockCreateSubscriptionCommandHandler struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockCreateSubscriptionCommandHandler) Exec(ctx context.Context, command billing_in.CreateSubscriptionCommand) error {
	ret := _m.Called(ctx, command)

	return ret.Error(0)
}

// NewMockCreateSubscriptionCommandHandler creates a new instance of MockCreateSubscriptionCommandHandler
func NewMockCreateSubscriptionCommandHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockCreateSubscriptionCommandHandler {
	mock := &MockCreateSubscriptionCommandHandler{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
