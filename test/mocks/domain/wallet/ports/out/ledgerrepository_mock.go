package wallet_out

import (
	"context"

	"time"

	"github.com/google/uuid"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_out "github.com/replay-api/replay-api/pkg/domain/wallet/ports/out"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	"github.com/stretchr/testify/mock"
)

// MockLedgerRepository is a mock implementation of LedgerRepository
type MockLedgerRepository struct {
	mock.Mock
}

// CreateTransaction provides a mock function
func (_m *MockLedgerRepository) CreateTransaction(ctx context.Context, entries []*wallet_entities.LedgerEntry) error {
	ret := _m.Called(ctx, entries)

	return ret.Error(0)
}

// CreateEntry provides a mock function
func (_m *MockLedgerRepository) CreateEntry(ctx context.Context, entry *wallet_entities.LedgerEntry) error {
	ret := _m.Called(ctx, entry)

	return ret.Error(0)
}

// FindByID provides a mock function
func (_m *MockLedgerRepository) FindByID(ctx context.Context, id uuid.UUID) (*wallet_entities.LedgerEntry, error) {
	ret := _m.Called(ctx, id)

	var r0 *wallet_entities.LedgerEntry
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*wallet_entities.LedgerEntry, error)); ok {
		return rf(ctx, id)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *wallet_entities.LedgerEntry); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*wallet_entities.LedgerEntry)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindByTransactionID provides a mock function
func (_m *MockLedgerRepository) FindByTransactionID(ctx context.Context, txID uuid.UUID) ([]*wallet_entities.LedgerEntry, error) {
	ret := _m.Called(ctx, txID)

	var r0 []*wallet_entities.LedgerEntry
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) ([]*wallet_entities.LedgerEntry, error)); ok {
		return rf(ctx, txID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) []*wallet_entities.LedgerEntry); ok {
		r0 = rf(ctx, txID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*wallet_entities.LedgerEntry)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindByAccountID provides a mock function
func (_m *MockLedgerRepository) FindByAccountID(ctx context.Context, accountID uuid.UUID, limit int, offset int) ([]*wallet_entities.LedgerEntry, error) {
	ret := _m.Called(ctx, accountID, limit, offset)

	var r0 []*wallet_entities.LedgerEntry
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, int, int) ([]*wallet_entities.LedgerEntry, error)); ok {
		return rf(ctx, accountID, limit, offset)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, int, int) []*wallet_entities.LedgerEntry); ok {
		r0 = rf(ctx, accountID, limit, offset)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*wallet_entities.LedgerEntry)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindByAccountAndCurrency provides a mock function
func (_m *MockLedgerRepository) FindByAccountAndCurrency(ctx context.Context, accountID uuid.UUID, currency wallet_vo.Currency) ([]*wallet_entities.LedgerEntry, error) {
	ret := _m.Called(ctx, accountID, currency)

	var r0 []*wallet_entities.LedgerEntry
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, wallet_vo.Currency) ([]*wallet_entities.LedgerEntry, error)); ok {
		return rf(ctx, accountID, currency)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, wallet_vo.Currency) []*wallet_entities.LedgerEntry); ok {
		r0 = rf(ctx, accountID, currency)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*wallet_entities.LedgerEntry)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindByIdempotencyKey provides a mock function
func (_m *MockLedgerRepository) FindByIdempotencyKey(ctx context.Context, key string) (*wallet_entities.LedgerEntry, error) {
	ret := _m.Called(ctx, key)

	var r0 *wallet_entities.LedgerEntry
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*wallet_entities.LedgerEntry, error)); ok {
		return rf(ctx, key)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) *wallet_entities.LedgerEntry); ok {
		r0 = rf(ctx, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*wallet_entities.LedgerEntry)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// ExistsByIdempotencyKey provides a mock function
func (_m *MockLedgerRepository) ExistsByIdempotencyKey(ctx context.Context, key string) bool {
	ret := _m.Called(ctx, key)

	var r0 bool
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(bool)
	}
	return r0
}

// FindByDateRange provides a mock function
func (_m *MockLedgerRepository) FindByDateRange(ctx context.Context, accountID uuid.UUID, from time.Time, to time.Time) ([]*wallet_entities.LedgerEntry, error) {
	ret := _m.Called(ctx, accountID, from, to)

	var r0 []*wallet_entities.LedgerEntry
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, time.Time, time.Time) ([]*wallet_entities.LedgerEntry, error)); ok {
		return rf(ctx, accountID, from, to)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, time.Time, time.Time) []*wallet_entities.LedgerEntry); ok {
		r0 = rf(ctx, accountID, from, to)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*wallet_entities.LedgerEntry)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// CalculateBalance provides a mock function
func (_m *MockLedgerRepository) CalculateBalance(ctx context.Context, accountID uuid.UUID, currency wallet_vo.Currency) (wallet_vo.Amount, error) {
	ret := _m.Called(ctx, accountID, currency)

	var r0 wallet_vo.Amount
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, wallet_vo.Currency) (wallet_vo.Amount, error)); ok {
		return rf(ctx, accountID, currency)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, wallet_vo.Currency) wallet_vo.Amount); ok {
		r0 = rf(ctx, accountID, currency)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(wallet_vo.Amount)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetAccountHistory provides a mock function
func (_m *MockLedgerRepository) GetAccountHistory(ctx context.Context, accountID uuid.UUID, filters wallet_out.HistoryFilters) ([]*wallet_entities.LedgerEntry, int64, error) {
	ret := _m.Called(ctx, accountID, filters)

	var r0 []*wallet_entities.LedgerEntry
	var r1 int64
	var r2 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, wallet_out.HistoryFilters) ([]*wallet_entities.LedgerEntry, int64, error)); ok {
		return rf(ctx, accountID, filters)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, wallet_out.HistoryFilters) []*wallet_entities.LedgerEntry); ok {
		r0 = rf(ctx, accountID, filters)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*wallet_entities.LedgerEntry)
		}
	}
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, wallet_out.HistoryFilters) int64); ok {
		r1 = rf(ctx, accountID, filters)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(int64)
		}
	}
	r2 = ret.Error(2)

	return r0, r1, r2
}

// FindPendingApprovals provides a mock function
func (_m *MockLedgerRepository) FindPendingApprovals(ctx context.Context, limit int) ([]*wallet_entities.LedgerEntry, error) {
	ret := _m.Called(ctx, limit)

	var r0 []*wallet_entities.LedgerEntry
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, int) ([]*wallet_entities.LedgerEntry, error)); ok {
		return rf(ctx, limit)
	}

	if rf, ok := ret.Get(0).(func(context.Context, int) []*wallet_entities.LedgerEntry); ok {
		r0 = rf(ctx, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*wallet_entities.LedgerEntry)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// UpdateApprovalStatus provides a mock function
func (_m *MockLedgerRepository) UpdateApprovalStatus(ctx context.Context, entryID uuid.UUID, status wallet_entities.ApprovalStatus, approverID uuid.UUID) error {
	ret := _m.Called(ctx, entryID, status, approverID)

	return ret.Error(0)
}

// MarkAsReversed provides a mock function
func (_m *MockLedgerRepository) MarkAsReversed(ctx context.Context, entryID uuid.UUID, reversalEntryID uuid.UUID) error {
	ret := _m.Called(ctx, entryID, reversalEntryID)

	return ret.Error(0)
}

// GetDailyTransactionCount provides a mock function
func (_m *MockLedgerRepository) GetDailyTransactionCount(ctx context.Context, accountID uuid.UUID) (int64, error) {
	ret := _m.Called(ctx, accountID)

	var r0 int64
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (int64, error)); ok {
		return rf(ctx, accountID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) int64); ok {
		r0 = rf(ctx, accountID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(int64)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetDailyTransactionVolume provides a mock function
func (_m *MockLedgerRepository) GetDailyTransactionVolume(ctx context.Context, accountID uuid.UUID, currency wallet_vo.Currency) (wallet_vo.Amount, error) {
	ret := _m.Called(ctx, accountID, currency)

	var r0 wallet_vo.Amount
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, wallet_vo.Currency) (wallet_vo.Amount, error)); ok {
		return rf(ctx, accountID, currency)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, wallet_vo.Currency) wallet_vo.Amount); ok {
		r0 = rf(ctx, accountID, currency)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(wallet_vo.Amount)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindByUserAndDateRange provides a mock function
func (_m *MockLedgerRepository) FindByUserAndDateRange(ctx context.Context, userID uuid.UUID, from time.Time, to time.Time) ([]*wallet_entities.LedgerEntry, error) {
	ret := _m.Called(ctx, userID, from, to)

	var r0 []*wallet_entities.LedgerEntry
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, time.Time, time.Time) ([]*wallet_entities.LedgerEntry, error)); ok {
		return rf(ctx, userID, from, to)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, time.Time, time.Time) []*wallet_entities.LedgerEntry); ok {
		r0 = rf(ctx, userID, from, to)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*wallet_entities.LedgerEntry)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockLedgerRepository creates a new instance of MockLedgerRepository
func NewMockLedgerRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockLedgerRepository {
	mock := &MockLedgerRepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
