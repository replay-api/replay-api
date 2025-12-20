package billing_in

import (
	"context"

	"github.com/google/uuid"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockWithdrawalCommand is a mock implementation of WithdrawalCommand
type MockWithdrawalCommand struct {
	mock.Mock
}

// Create provides a mock function
func (_m *MockWithdrawalCommand) Create(ctx context.Context, cmd billing_in.CreateWithdrawalCommand) (*billing_entities.Withdrawal, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *billing_entities.Withdrawal
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, billing_in.CreateWithdrawalCommand) (*billing_entities.Withdrawal, error)); ok {
		return rf(ctx, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, billing_in.CreateWithdrawalCommand) *billing_entities.Withdrawal); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing_entities.Withdrawal)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Cancel provides a mock function
func (_m *MockWithdrawalCommand) Cancel(ctx context.Context, withdrawalID uuid.UUID) (*billing_entities.Withdrawal, error) {
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

// GetByID provides a mock function
func (_m *MockWithdrawalCommand) GetByID(ctx context.Context, withdrawalID uuid.UUID) (*billing_entities.Withdrawal, error) {
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

// GetByUserID provides a mock function
func (_m *MockWithdrawalCommand) GetByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]billing_entities.Withdrawal, error) {
	ret := _m.Called(ctx, userID, limit, offset)

	var r0 []billing_entities.Withdrawal
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, int, int) ([]billing_entities.Withdrawal, error)); ok {
		return rf(ctx, userID, limit, offset)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, int, int) []billing_entities.Withdrawal); ok {
		r0 = rf(ctx, userID, limit, offset)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]billing_entities.Withdrawal)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockWithdrawalCommand creates a new instance of MockWithdrawalCommand
func NewMockWithdrawalCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockWithdrawalCommand {
	mock := &MockWithdrawalCommand{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
