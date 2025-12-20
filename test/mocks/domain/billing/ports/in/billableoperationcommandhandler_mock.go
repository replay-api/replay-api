package billing_in

import (
	"context"

	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockBillableOperationCommandHandler is a mock implementation of BillableOperationCommandHandler
type MockBillableOperationCommandHandler struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockBillableOperationCommandHandler) Exec(ctx context.Context, command billing_in.BillableOperationCommand) (*billing_entities.BillableEntry, *billing_entities.Subscription, error) {
	ret := _m.Called(ctx, command)

	var r0 *billing_entities.BillableEntry
	var r1 *billing_entities.Subscription
	var r2 error

	if rf, ok := ret.Get(0).(func(context.Context, billing_in.BillableOperationCommand) (*billing_entities.BillableEntry, *billing_entities.Subscription, error)); ok {
		return rf(ctx, command)
	}

	if rf, ok := ret.Get(0).(func(context.Context, billing_in.BillableOperationCommand) *billing_entities.BillableEntry); ok {
		r0 = rf(ctx, command)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing_entities.BillableEntry)
		}
	}
	if rf, ok := ret.Get(1).(func(context.Context, billing_in.BillableOperationCommand) *billing_entities.Subscription); ok {
		r1 = rf(ctx, command)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*billing_entities.Subscription)
		}
	}
	r2 = ret.Error(2)

	return r0, r1, r2
}

// Validate provides a mock function
func (_m *MockBillableOperationCommandHandler) Validate(ctx context.Context, command billing_in.BillableOperationCommand) error {
	ret := _m.Called(ctx, command)

	return ret.Error(0)
}

// NewMockBillableOperationCommandHandler creates a new instance of MockBillableOperationCommandHandler
func NewMockBillableOperationCommandHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockBillableOperationCommandHandler {
	mock := &MockBillableOperationCommandHandler{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
