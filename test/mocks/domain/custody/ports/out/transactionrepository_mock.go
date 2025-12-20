package custody_out

import (
	"context"

	"time"

	"github.com/google/uuid"
	custody_out "github.com/replay-api/replay-api/pkg/domain/custody/ports/out"
	custody_vo "github.com/replay-api/replay-api/pkg/domain/custody/value-objects"
	"github.com/stretchr/testify/mock"
)

// MockTransactionRepository is a mock implementation of TransactionRepository
type MockTransactionRepository struct {
	mock.Mock	
}

// Create provides a mock function
func (_m *MockTransactionRepository) Create(ctx context.Context, tx *custody_out.CustodyTransaction) error {
	ret := _m.Called(ctx, tx)

	return ret.Error(0)
}

// GetByID provides a mock function
func (_m *MockTransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*custody_out.CustodyTransaction, error) {
	ret := _m.Called(ctx, id)

	var r0 *custody_out.CustodyTransaction
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*custody_out.CustodyTransaction, error)); ok {
		return rf(ctx, id)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *custody_out.CustodyTransaction); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.CustodyTransaction)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetByHash provides a mock function
func (_m *MockTransactionRepository) GetByHash(ctx context.Context, chainID custody_vo.ChainID, hash string) (*custody_out.CustodyTransaction, error) {
	ret := _m.Called(ctx, chainID, hash)

	var r0 *custody_out.CustodyTransaction
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, custody_vo.ChainID, string) (*custody_out.CustodyTransaction, error)); ok {
		return rf(ctx, chainID, hash)
	}

	if rf, ok := ret.Get(0).(func(context.Context, custody_vo.ChainID, string) *custody_out.CustodyTransaction); ok {
		r0 = rf(ctx, chainID, hash)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.CustodyTransaction)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Update provides a mock function
func (_m *MockTransactionRepository) Update(ctx context.Context, tx *custody_out.CustodyTransaction) error {
	ret := _m.Called(ctx, tx)

	return ret.Error(0)
}

// ListByWallet provides a mock function
func (_m *MockTransactionRepository) ListByWallet(ctx context.Context, walletID uuid.UUID, filter *custody_out.TxFilter) (*custody_out.TxListResult, error) {
	ret := _m.Called(ctx, walletID, filter)

	var r0 *custody_out.TxListResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, *custody_out.TxFilter) (*custody_out.TxListResult, error)); ok {
		return rf(ctx, walletID, filter)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, *custody_out.TxFilter) *custody_out.TxListResult); ok {
		r0 = rf(ctx, walletID, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.TxListResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetPendingTransactions provides a mock function
func (_m *MockTransactionRepository) GetPendingTransactions(ctx context.Context) ([]*custody_out.CustodyTransaction, error) {
	ret := _m.Called(ctx)

	var r0 []*custody_out.CustodyTransaction
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context) ([]*custody_out.CustodyTransaction, error)); ok {
		return rf(ctx)
	}

	if rf, ok := ret.Get(0).(func(context.Context) []*custody_out.CustodyTransaction); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*custody_out.CustodyTransaction)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetFailedTransactions provides a mock function
func (_m *MockTransactionRepository) GetFailedTransactions(ctx context.Context, since time.Time) ([]*custody_out.CustodyTransaction, error) {
	ret := _m.Called(ctx, since)

	var r0 []*custody_out.CustodyTransaction
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, time.Time) ([]*custody_out.CustodyTransaction, error)); ok {
		return rf(ctx, since)
	}

	if rf, ok := ret.Get(0).(func(context.Context, time.Time) []*custody_out.CustodyTransaction); ok {
		r0 = rf(ctx, since)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*custody_out.CustodyTransaction)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetDailySpending provides a mock function
func (_m *MockTransactionRepository) GetDailySpending(ctx context.Context, walletID uuid.UUID, date time.Time) (*custody_out.SpendingAggregate, error) {
	ret := _m.Called(ctx, walletID, date)

	var r0 *custody_out.SpendingAggregate
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, time.Time) (*custody_out.SpendingAggregate, error)); ok {
		return rf(ctx, walletID, date)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, time.Time) *custody_out.SpendingAggregate); ok {
		r0 = rf(ctx, walletID, date)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.SpendingAggregate)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetWeeklySpending provides a mock function
func (_m *MockTransactionRepository) GetWeeklySpending(ctx context.Context, walletID uuid.UUID, weekStart time.Time) (*custody_out.SpendingAggregate, error) {
	ret := _m.Called(ctx, walletID, weekStart)

	var r0 *custody_out.SpendingAggregate
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, time.Time) (*custody_out.SpendingAggregate, error)); ok {
		return rf(ctx, walletID, weekStart)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, time.Time) *custody_out.SpendingAggregate); ok {
		r0 = rf(ctx, walletID, weekStart)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.SpendingAggregate)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetMonthlySpending provides a mock function
func (_m *MockTransactionRepository) GetMonthlySpending(ctx context.Context, walletID uuid.UUID, month time.Time) (*custody_out.SpendingAggregate, error) {
	ret := _m.Called(ctx, walletID, month)

	var r0 *custody_out.SpendingAggregate
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, time.Time) (*custody_out.SpendingAggregate, error)); ok {
		return rf(ctx, walletID, month)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, time.Time) *custody_out.SpendingAggregate); ok {
		r0 = rf(ctx, walletID, month)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.SpendingAggregate)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockTransactionRepository creates a new instance of MockTransactionRepository
func NewMockTransactionRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockTransactionRepository {
	mock := &MockTransactionRepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
