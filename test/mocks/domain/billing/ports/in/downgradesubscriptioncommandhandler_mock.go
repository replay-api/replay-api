package billing_in

import (
	"context"

	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockDowngradeSubscriptionCommandHandler is a mock implementation of DowngradeSubscriptionCommandHandler
type MockDowngradeSubscriptionCommandHandler struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockDowngradeSubscriptionCommandHandler) Exec(ctx context.Context, command billing_in.DowngradeSubscriptionCommand) error {
	ret := _m.Called(ctx, command)

	return ret.Error(0)
}

// NewMockDowngradeSubscriptionCommandHandler creates a new instance of MockDowngradeSubscriptionCommandHandler
func NewMockDowngradeSubscriptionCommandHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockDowngradeSubscriptionCommandHandler {
	mock := &MockDowngradeSubscriptionCommandHandler{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
