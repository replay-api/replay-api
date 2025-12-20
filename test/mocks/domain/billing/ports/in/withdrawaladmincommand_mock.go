package billing_in

import (
	"context"

	"github.com/google/uuid"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	"github.com/stretchr/testify/mock"
)

// MockWithdrawalAdminCommand is a mock implementation of WithdrawalAdminCommand
type MockWithdrawalAdminCommand struct {
	mock.Mock
}

// Approve provides a mock function
func (_m *MockWithdrawalAdminCommand) Approve(ctx context.Context, withdrawalID uuid.UUID, reviewerID uuid.UUID) (*billing_entities.Withdrawal, error) {
	ret := _m.Called(ctx, withdrawalID, reviewerID)

	var r0 *billing_entities.Withdrawal
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, uuid.UUID) (*billing_entities.Withdrawal, error)); ok {
		return rf(ctx, withdrawalID, reviewerID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, uuid.UUID) *billing_entities.Withdrawal); ok {
		r0 = rf(ctx, withdrawalID, reviewerID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing_entities.Withdrawal)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Reject provides a mock function
func (_m *MockWithdrawalAdminCommand) Reject(ctx context.Context, withdrawalID uuid.UUID, reviewerID uuid.UUID, reason string) (*billing_entities.Withdrawal, error) {
	ret := _m.Called(ctx, withdrawalID, reviewerID, reason)

	var r0 *billing_entities.Withdrawal
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, uuid.UUID, string) (*billing_entities.Withdrawal, error)); ok {
		return rf(ctx, withdrawalID, reviewerID, reason)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, uuid.UUID, string) *billing_entities.Withdrawal); ok {
		r0 = rf(ctx, withdrawalID, reviewerID, reason)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing_entities.Withdrawal)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Process provides a mock function
func (_m *MockWithdrawalAdminCommand) Process(ctx context.Context, withdrawalID uuid.UUID) (*billing_entities.Withdrawal, error) {
	ret := _m.Called(ctx, withdrawalID)

	var r0 *billing_entities.Withdrawal
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*billing_entities.Withdrawal, error)); ok {
		return rf(ctx, withdrawalID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *billing_entities.Withdrawal); ok {
		r0 = rf(ctx, withdrawalID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing_entities.Withdrawal)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Complete provides a mock function
func (_m *MockWithdrawalAdminCommand) Complete(ctx context.Context, withdrawalID uuid.UUID, providerRef string) (*billing_entities.Withdrawal, error) {
	ret := _m.Called(ctx, withdrawalID, providerRef)

	var r0 *billing_entities.Withdrawal
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string) (*billing_entities.Withdrawal, error)); ok {
		return rf(ctx, withdrawalID, providerRef)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string) *billing_entities.Withdrawal); ok {
		r0 = rf(ctx, withdrawalID, providerRef)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing_entities.Withdrawal)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Fail provides a mock function
func (_m *MockWithdrawalAdminCommand) Fail(ctx context.Context, withdrawalID uuid.UUID, reason string) (*billing_entities.Withdrawal, error) {
	ret := _m.Called(ctx, withdrawalID, reason)

	var r0 *billing_entities.Withdrawal
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string) (*billing_entities.Withdrawal, error)); ok {
		return rf(ctx, withdrawalID, reason)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string) *billing_entities.Withdrawal); ok {
		r0 = rf(ctx, withdrawalID, reason)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing_entities.Withdrawal)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetPending provides a mock function
func (_m *MockWithdrawalAdminCommand) GetPending(ctx context.Context, limit int, offset int) ([]billing_entities.Withdrawal, error) {
	ret := _m.Called(ctx, limit, offset)

	var r0 []billing_entities.Withdrawal
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, int, int) ([]billing_entities.Withdrawal, error)); ok {
		return rf(ctx, limit, offset)
	}

	if rf, ok := ret.Get(0).(func(context.Context, int, int) []billing_entities.Withdrawal); ok {
		r0 = rf(ctx, limit, offset)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]billing_entities.Withdrawal)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockWithdrawalAdminCommand creates a new instance of MockWithdrawalAdminCommand
func NewMockWithdrawalAdminCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockWithdrawalAdminCommand {
	mock := &MockWithdrawalAdminCommand{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
