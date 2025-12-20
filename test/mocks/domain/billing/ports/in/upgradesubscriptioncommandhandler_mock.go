package billing_in

import (
	"context"

	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockUpgradeSubscriptionCommandHandler is a mock implementation of UpgradeSubscriptionCommandHandler
type MockUpgradeSubscriptionCommandHandler struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockUpgradeSubscriptionCommandHandler) Exec(ctx context.Context, command billing_in.UpgradeSubscriptionCommand) error {
	ret := _m.Called(ctx, command)

	return ret.Error(0)
}

// NewMockUpgradeSubscriptionCommandHandler creates a new instance of MockUpgradeSubscriptionCommandHandler
func NewMockUpgradeSubscriptionCommandHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockUpgradeSubscriptionCommandHandler {
	mock := &MockUpgradeSubscriptionCommandHandler{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
