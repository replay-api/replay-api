package wallet_in

import (
	"context"

	wallet_in "github.com/replay-api/replay-api/pkg/domain/wallet/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockWalletQuery is a mock implementation of WalletQuery
type MockWalletQuery struct {
	mock.Mock
}

// GetBalance provides a mock function
func (_m *MockWalletQuery) GetBalance(ctx context.Context, query wallet_in.GetWalletBalanceQuery) (*wallet_in.WalletBalanceResult, error) {
	ret := _m.Called(ctx, query)

	var r0 *wallet_in.WalletBalanceResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, wallet_in.GetWalletBalanceQuery) (*wallet_in.WalletBalanceResult, error)); ok {
		return rf(ctx, query)
	}

	if rf, ok := ret.Get(0).(func(context.Context, wallet_in.GetWalletBalanceQuery) *wallet_in.WalletBalanceResult); ok {
		r0 = rf(ctx, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*wallet_in.WalletBalanceResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetTransactions provides a mock function
func (_m *MockWalletQuery) GetTransactions(ctx context.Context, query wallet_in.GetTransactionsQuery) (*wallet_in.TransactionsResult, error) {
	ret := _m.Called(ctx, query)

	var r0 *wallet_in.TransactionsResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, wallet_in.GetTransactionsQuery) (*wallet_in.TransactionsResult, error)); ok {
		return rf(ctx, query)
	}

	if rf, ok := ret.Get(0).(func(context.Context, wallet_in.GetTransactionsQuery) *wallet_in.TransactionsResult); ok {
		r0 = rf(ctx, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*wallet_in.TransactionsResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockWalletQuery creates a new instance of MockWalletQuery
func NewMockWalletQuery(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockWalletQuery {
	mock := &MockWalletQuery{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
