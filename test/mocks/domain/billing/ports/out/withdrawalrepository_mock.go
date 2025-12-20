package billing_out

import (
	"context"

	"github.com/google/uuid"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	"github.com/stretchr/testify/mock"
)

// MockWithdrawalRepository is a mock implementation of WithdrawalRepository
type MockWithdrawalRepository struct {
	mock.Mock
}

// Create provides a mock function
func (_m *MockWithdrawalRepository) Create(ctx context.Context, withdrawal *billing_entities.Withdrawal) (*billing_entities.Withdrawal, error) {
	ret := _m.Called(ctx, withdrawal)

	var r0 *billing_entities.Withdrawal
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *billing_entities.Withdrawal) (*billing_entities.Withdrawal, error)); ok {
		return rf(ctx, withdrawal)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *billing_entities.Withdrawal) *billing_entities.Withdrawal); ok {
		r0 = rf(ctx, withdrawal)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing_entities.Withdrawal)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Update provides a mock function
func (_m *MockWithdrawalRepository) Update(ctx context.Context, withdrawal *billing_entities.Withdrawal) (*billing_entities.Withdrawal, error) {
	ret := _m.Called(ctx, withdrawal)

	var r0 *billing_entities.Withdrawal
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *billing_entities.Withdrawal) (*billing_entities.Withdrawal, error)); ok {
		return rf(ctx, withdrawal)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *billing_entities.Withdrawal) *billing_entities.Withdrawal); ok {
		r0 = rf(ctx, withdrawal)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing_entities.Withdrawal)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetByID provides a mock function
func (_m *MockWithdrawalRepository) GetByID(ctx context.Context, id uuid.UUID) (*billing_entities.Withdrawal, error) {
	ret := _m.Called(ctx, id)

	var r0 *billing_entities.Withdrawal
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*billing_entities.Withdrawal, error)); ok {
		return rf(ctx, id)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *billing_entities.Withdrawal); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing_entities.Withdrawal)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetByUserID provides a mock function
func (_m *MockWithdrawalRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]billing_entities.Withdrawal, error) {
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

// GetByStatus provides a mock function
func (_m *MockWithdrawalRepository) GetByStatus(ctx context.Context, status billing_entities.WithdrawalStatus, limit int, offset int) ([]billing_entities.Withdrawal, error) {
	ret := _m.Called(ctx, status, limit, offset)

	var r0 []billing_entities.Withdrawal
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, billing_entities.WithdrawalStatus, int, int) ([]billing_entities.Withdrawal, error)); ok {
		return rf(ctx, status, limit, offset)
	}

	if rf, ok := ret.Get(0).(func(context.Context, billing_entities.WithdrawalStatus, int, int) []billing_entities.Withdrawal); ok {
		r0 = rf(ctx, status, limit, offset)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]billing_entities.Withdrawal)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetPending provides a mock function
func (_m *MockWithdrawalRepository) GetPending(ctx context.Context, limit int, offset int) ([]billing_entities.Withdrawal, error) {
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

// NewMockWithdrawalRepository creates a new instance of MockWithdrawalRepository
func NewMockWithdrawalRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockWithdrawalRepository {
	mock := &MockWithdrawalRepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
