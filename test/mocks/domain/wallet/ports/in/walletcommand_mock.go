package wallet_in

import (
	"context"

	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_in "github.com/replay-api/replay-api/pkg/domain/wallet/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockWalletCommand is a mock implementation of WalletCommand
type MockWalletCommand struct {
	mock.Mock
}

// CreateWallet provides a mock function
func (_m *MockWalletCommand) CreateWallet(ctx context.Context, cmd wallet_in.CreateWalletCommand) (*wallet_entities.UserWallet, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *wallet_entities.UserWallet
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, wallet_in.CreateWalletCommand) (*wallet_entities.UserWallet, error)); ok {
		return rf(ctx, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, wallet_in.CreateWalletCommand) *wallet_entities.UserWallet); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*wallet_entities.UserWallet)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Deposit provides a mock function
func (_m *MockWalletCommand) Deposit(ctx context.Context, cmd wallet_in.DepositCommand) error {
	ret := _m.Called(ctx, cmd)

	return ret.Error(0)
}

// Withdraw provides a mock function
func (_m *MockWalletCommand) Withdraw(ctx context.Context, cmd wallet_in.WithdrawCommand) error {
	ret := _m.Called(ctx, cmd)

	return ret.Error(0)
}

// DeductEntryFee provides a mock function
func (_m *MockWalletCommand) DeductEntryFee(ctx context.Context, cmd wallet_in.DeductEntryFeeCommand) error {
	ret := _m.Called(ctx, cmd)

	return ret.Error(0)
}

// AddPrize provides a mock function
func (_m *MockWalletCommand) AddPrize(ctx context.Context, cmd wallet_in.AddPrizeCommand) error {
	ret := _m.Called(ctx, cmd)

	return ret.Error(0)
}

// Refund provides a mock function
func (_m *MockWalletCommand) Refund(ctx context.Context, cmd wallet_in.RefundCommand) error {
	ret := _m.Called(ctx, cmd)

	return ret.Error(0)
}

// DebitWallet provides a mock function
func (_m *MockWalletCommand) DebitWallet(ctx context.Context, cmd wallet_in.DebitWalletCommand) (*wallet_entities.WalletTransaction, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *wallet_entities.WalletTransaction
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, wallet_in.DebitWalletCommand) (*wallet_entities.WalletTransaction, error)); ok {
		return rf(ctx, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, wallet_in.DebitWalletCommand) *wallet_entities.WalletTransaction); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*wallet_entities.WalletTransaction)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// CreditWallet provides a mock function
func (_m *MockWalletCommand) CreditWallet(ctx context.Context, cmd wallet_in.CreditWalletCommand) (*wallet_entities.WalletTransaction, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *wallet_entities.WalletTransaction
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, wallet_in.CreditWalletCommand) (*wallet_entities.WalletTransaction, error)); ok {
		return rf(ctx, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, wallet_in.CreditWalletCommand) *wallet_entities.WalletTransaction); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*wallet_entities.WalletTransaction)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockWalletCommand creates a new instance of MockWalletCommand
func NewMockWalletCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockWalletCommand {
	mock := &MockWalletCommand{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
