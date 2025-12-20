package wallet_services

import (
	"context"

	big "math/big"
	time "time"

	"github.com/google/uuid"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	"github.com/stretchr/testify/mock"
)

// MockLedgerRepository is a mock implementation of LedgerRepository
type MockLedgerRepository struct {
	mock.Mock
}

// CreateAccount provides a mock function
func (_m *MockLedgerRepository) CreateAccount(ctx context.Context, account *wallet_entities.LedgerAccount) error {
	ret := _m.Called(ctx, account)

	return ret.Error(0)
}

// GetAccountByID provides a mock function
func (_m *MockLedgerRepository) GetAccountByID(ctx context.Context, id uuid.UUID) (*wallet_entities.LedgerAccount, error) {
	ret := _m.Called(ctx, id)

	var r0 *wallet_entities.LedgerAccount
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*wallet_entities.LedgerAccount, error)); ok {
		return rf(ctx, id)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *wallet_entities.LedgerAccount); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*wallet_entities.LedgerAccount)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetAccountByCode provides a mock function
func (_m *MockLedgerRepository) GetAccountByCode(ctx context.Context, code string) (*wallet_entities.LedgerAccount, error) {
	ret := _m.Called(ctx, code)

	var r0 *wallet_entities.LedgerAccount
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*wallet_entities.LedgerAccount, error)); ok {
		return rf(ctx, code)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) *wallet_entities.LedgerAccount); ok {
		r0 = rf(ctx, code)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*wallet_entities.LedgerAccount)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetAccountByUserID provides a mock function
func (_m *MockLedgerRepository) GetAccountByUserID(ctx context.Context, userID uuid.UUID, currency string) (*wallet_entities.LedgerAccount, error) {
	ret := _m.Called(ctx, userID, currency)

	var r0 *wallet_entities.LedgerAccount
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string) (*wallet_entities.LedgerAccount, error)); ok {
		return rf(ctx, userID, currency)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string) *wallet_entities.LedgerAccount); ok {
		r0 = rf(ctx, userID, currency)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*wallet_entities.LedgerAccount)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// UpdateAccountBalance provides a mock function
func (_m *MockLedgerRepository) UpdateAccountBalance(ctx context.Context, accountID uuid.UUID, balance *big.Float, available *big.Float, held *big.Float, version int) error {
	ret := _m.Called(ctx, accountID, balance, available, held, version)

	return ret.Error(0)
}

// CreateJournal provides a mock function
func (_m *MockLedgerRepository) CreateJournal(ctx context.Context, journal *wallet_entities.JournalEntry) error {
	ret := _m.Called(ctx, journal)

	return ret.Error(0)
}

// GetJournalByID provides a mock function
func (_m *MockLedgerRepository) GetJournalByID(ctx context.Context, id uuid.UUID) (*wallet_entities.JournalEntry, error) {
	ret := _m.Called(ctx, id)

	var r0 *wallet_entities.JournalEntry
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*wallet_entities.JournalEntry, error)); ok {
		return rf(ctx, id)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *wallet_entities.JournalEntry); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*wallet_entities.JournalEntry)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetLastJournalHash provides a mock function
func (_m *MockLedgerRepository) GetLastJournalHash(ctx context.Context) (string, error) {
	ret := _m.Called(ctx)

	var r0 string
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context) (string, error)); ok {
		return rf(ctx)
	}

	if rf, ok := ret.Get(0).(func(context.Context) string); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(string)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// UpdateJournalStatus provides a mock function
func (_m *MockLedgerRepository) UpdateJournalStatus(ctx context.Context, id uuid.UUID, status wallet_entities.JournalStatus) error {
	ret := _m.Called(ctx, id, status)

	return ret.Error(0)
}

// CreateWallet provides a mock function
func (_m *MockLedgerRepository) CreateWallet(ctx context.Context, wallet *wallet_entities.LedgerWallet) error {
	ret := _m.Called(ctx, wallet)

	return ret.Error(0)
}

// GetWalletByUserID provides a mock function
func (_m *MockLedgerRepository) GetWalletByUserID(ctx context.Context, userID uuid.UUID, currency string) (*wallet_entities.LedgerWallet, error) {
	ret := _m.Called(ctx, userID, currency)

	var r0 *wallet_entities.LedgerWallet
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string) (*wallet_entities.LedgerWallet, error)); ok {
		return rf(ctx, userID, currency)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string) *wallet_entities.LedgerWallet); ok {
		r0 = rf(ctx, userID, currency)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*wallet_entities.LedgerWallet)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// UpdateWallet provides a mock function
func (_m *MockLedgerRepository) UpdateWallet(ctx context.Context, wallet *wallet_entities.LedgerWallet) error {
	ret := _m.Called(ctx, wallet)

	return ret.Error(0)
}

// GetJournalsByDateRange provides a mock function
func (_m *MockLedgerRepository) GetJournalsByDateRange(ctx context.Context, from time.Time, to time.Time) ([]wallet_entities.JournalEntry, error) {
	ret := _m.Called(ctx, from, to)

	var r0 []wallet_entities.JournalEntry
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, time.Time, time.Time) ([]wallet_entities.JournalEntry, error)); ok {
		return rf(ctx, from, to)
	}

	if rf, ok := ret.Get(0).(func(context.Context, time.Time, time.Time) []wallet_entities.JournalEntry); ok {
		r0 = rf(ctx, from, to)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]wallet_entities.JournalEntry)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetAccountBalances provides a mock function
func (_m *MockLedgerRepository) GetAccountBalances(ctx context.Context) ([]wallet_entities.LedgerAccount, error) {
	ret := _m.Called(ctx)

	var r0 []wallet_entities.LedgerAccount
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context) ([]wallet_entities.LedgerAccount, error)); ok {
		return rf(ctx)
	}

	if rf, ok := ret.Get(0).(func(context.Context) []wallet_entities.LedgerAccount); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]wallet_entities.LedgerAccount)
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
