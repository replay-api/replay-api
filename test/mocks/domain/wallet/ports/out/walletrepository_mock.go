package wallet_out

import (
	"context"

	"github.com/google/uuid"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	"github.com/stretchr/testify/mock"
)

// MockWalletRepository is a mock implementation of WalletRepository
type MockWalletRepository struct {
	mock.Mock
}

// Save provides a mock function
func (_m *MockWalletRepository) Save(ctx context.Context, wallet *wallet_entities.UserWallet) error {
	ret := _m.Called(ctx, wallet)

	return ret.Error(0)
}

// FindByID provides a mock function
func (_m *MockWalletRepository) FindByID(ctx context.Context, id uuid.UUID) (*wallet_entities.UserWallet, error) {
	ret := _m.Called(ctx, id)

	var r0 *wallet_entities.UserWallet
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*wallet_entities.UserWallet, error)); ok {
		return rf(ctx, id)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *wallet_entities.UserWallet); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*wallet_entities.UserWallet)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindByUserID provides a mock function
func (_m *MockWalletRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*wallet_entities.UserWallet, error) {
	ret := _m.Called(ctx, userID)

	var r0 *wallet_entities.UserWallet
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*wallet_entities.UserWallet, error)); ok {
		return rf(ctx, userID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *wallet_entities.UserWallet); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*wallet_entities.UserWallet)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindByEVMAddress provides a mock function
func (_m *MockWalletRepository) FindByEVMAddress(ctx context.Context, address string) (*wallet_entities.UserWallet, error) {
	ret := _m.Called(ctx, address)

	var r0 *wallet_entities.UserWallet
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*wallet_entities.UserWallet, error)); ok {
		return rf(ctx, address)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) *wallet_entities.UserWallet); ok {
		r0 = rf(ctx, address)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*wallet_entities.UserWallet)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Update provides a mock function
func (_m *MockWalletRepository) Update(ctx context.Context, wallet *wallet_entities.UserWallet) error {
	ret := _m.Called(ctx, wallet)

	return ret.Error(0)
}

// Delete provides a mock function
func (_m *MockWalletRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ret := _m.Called(ctx, id)

	return ret.Error(0)
}

// NewMockWalletRepository creates a new instance of MockWalletRepository
func NewMockWalletRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockWalletRepository {
	mock := &MockWalletRepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
