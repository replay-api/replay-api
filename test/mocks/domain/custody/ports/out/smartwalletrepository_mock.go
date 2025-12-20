package custody_out

import (
	"context"

	"github.com/google/uuid"
	custody_entities "github.com/replay-api/replay-api/pkg/domain/custody/entities"
	custody_out "github.com/replay-api/replay-api/pkg/domain/custody/ports/out"
	custody_vo "github.com/replay-api/replay-api/pkg/domain/custody/value-objects"
	"github.com/stretchr/testify/mock"
)

// MockSmartWalletRepository is a mock implementation of SmartWalletRepository
type MockSmartWalletRepository struct {
	mock.Mock
}

// Create provides a mock function
func (_m *MockSmartWalletRepository) Create(ctx context.Context, wallet *custody_entities.SmartWallet) error {
	ret := _m.Called(ctx, wallet)

	return ret.Error(0)
}

// GetByID provides a mock function
func (_m *MockSmartWalletRepository) GetByID(ctx context.Context, id uuid.UUID) (*custody_entities.SmartWallet, error) {
	ret := _m.Called(ctx, id)

	var r0 *custody_entities.SmartWallet
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*custody_entities.SmartWallet, error)); ok {
		return rf(ctx, id)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *custody_entities.SmartWallet); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_entities.SmartWallet)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetByUserID provides a mock function
func (_m *MockSmartWalletRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*custody_entities.SmartWallet, error) {
	ret := _m.Called(ctx, userID)

	var r0 []*custody_entities.SmartWallet
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) ([]*custody_entities.SmartWallet, error)); ok {
		return rf(ctx, userID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) []*custody_entities.SmartWallet); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*custody_entities.SmartWallet)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetByAddress provides a mock function
func (_m *MockSmartWalletRepository) GetByAddress(ctx context.Context, chainID custody_vo.ChainID, address string) (*custody_entities.SmartWallet, error) {
	ret := _m.Called(ctx, chainID, address)

	var r0 *custody_entities.SmartWallet
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, custody_vo.ChainID, string) (*custody_entities.SmartWallet, error)); ok {
		return rf(ctx, chainID, address)
	}

	if rf, ok := ret.Get(0).(func(context.Context, custody_vo.ChainID, string) *custody_entities.SmartWallet); ok {
		r0 = rf(ctx, chainID, address)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_entities.SmartWallet)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Update provides a mock function
func (_m *MockSmartWalletRepository) Update(ctx context.Context, wallet *custody_entities.SmartWallet) error {
	ret := _m.Called(ctx, wallet)

	return ret.Error(0)
}

// Delete provides a mock function
func (_m *MockSmartWalletRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ret := _m.Called(ctx, id)

	return ret.Error(0)
}

// List provides a mock function
func (_m *MockSmartWalletRepository) List(ctx context.Context, filter *custody_out.WalletFilter) (*custody_out.WalletListResult, error) {
	ret := _m.Called(ctx, filter)

	var r0 *custody_out.WalletListResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.WalletFilter) (*custody_out.WalletListResult, error)); ok {
		return rf(ctx, filter)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.WalletFilter) *custody_out.WalletListResult); ok {
		r0 = rf(ctx, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.WalletListResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetByMPCKeyID provides a mock function
func (_m *MockSmartWalletRepository) GetByMPCKeyID(ctx context.Context, keyID string) (*custody_entities.SmartWallet, error) {
	ret := _m.Called(ctx, keyID)

	var r0 *custody_entities.SmartWallet
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*custody_entities.SmartWallet, error)); ok {
		return rf(ctx, keyID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) *custody_entities.SmartWallet); ok {
		r0 = rf(ctx, keyID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_entities.SmartWallet)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetPendingRecoveries provides a mock function
func (_m *MockSmartWalletRepository) GetPendingRecoveries(ctx context.Context) ([]*custody_entities.SmartWallet, error) {
	ret := _m.Called(ctx)

	var r0 []*custody_entities.SmartWallet
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context) ([]*custody_entities.SmartWallet, error)); ok {
		return rf(ctx)
	}

	if rf, ok := ret.Get(0).(func(context.Context) []*custody_entities.SmartWallet); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*custody_entities.SmartWallet)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetFrozenWallets provides a mock function
func (_m *MockSmartWalletRepository) GetFrozenWallets(ctx context.Context) ([]*custody_entities.SmartWallet, error) {
	ret := _m.Called(ctx)

	var r0 []*custody_entities.SmartWallet
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context) ([]*custody_entities.SmartWallet, error)); ok {
		return rf(ctx)
	}

	if rf, ok := ret.Get(0).(func(context.Context) []*custody_entities.SmartWallet); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*custody_entities.SmartWallet)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// AddGuardian provides a mock function
func (_m *MockSmartWalletRepository) AddGuardian(ctx context.Context, walletID uuid.UUID, guardian *custody_entities.Guardian) error {
	ret := _m.Called(ctx, walletID, guardian)

	return ret.Error(0)
}

// RemoveGuardian provides a mock function
func (_m *MockSmartWalletRepository) RemoveGuardian(ctx context.Context, walletID uuid.UUID, guardianID uuid.UUID) error {
	ret := _m.Called(ctx, walletID, guardianID)

	return ret.Error(0)
}

// GetGuardians provides a mock function
func (_m *MockSmartWalletRepository) GetGuardians(ctx context.Context, walletID uuid.UUID) ([]*custody_entities.Guardian, error) {
	ret := _m.Called(ctx, walletID)

	var r0 []*custody_entities.Guardian
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) ([]*custody_entities.Guardian, error)); ok {
		return rf(ctx, walletID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) []*custody_entities.Guardian); ok {
		r0 = rf(ctx, walletID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*custody_entities.Guardian)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetGuardianByAddress provides a mock function
func (_m *MockSmartWalletRepository) GetGuardianByAddress(ctx context.Context, walletID uuid.UUID, address string) (*custody_entities.Guardian, error) {
	ret := _m.Called(ctx, walletID, address)

	var r0 *custody_entities.Guardian
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string) (*custody_entities.Guardian, error)); ok {
		return rf(ctx, walletID, address)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string) *custody_entities.Guardian); ok {
		r0 = rf(ctx, walletID, address)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_entities.Guardian)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// AddSessionKey provides a mock function
func (_m *MockSmartWalletRepository) AddSessionKey(ctx context.Context, walletID uuid.UUID, sessionKey *custody_entities.SessionKey) error {
	ret := _m.Called(ctx, walletID, sessionKey)

	return ret.Error(0)
}

// RevokeSessionKey provides a mock function
func (_m *MockSmartWalletRepository) RevokeSessionKey(ctx context.Context, walletID uuid.UUID, keyAddress string) error {
	ret := _m.Called(ctx, walletID, keyAddress)

	return ret.Error(0)
}

// GetActiveSessionKeys provides a mock function
func (_m *MockSmartWalletRepository) GetActiveSessionKeys(ctx context.Context, walletID uuid.UUID) ([]*custody_entities.SessionKey, error) {
	ret := _m.Called(ctx, walletID)

	var r0 []*custody_entities.SessionKey
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) ([]*custody_entities.SessionKey, error)); ok {
		return rf(ctx, walletID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) []*custody_entities.SessionKey); ok {
		r0 = rf(ctx, walletID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*custody_entities.SessionKey)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// SetPendingRecovery provides a mock function
func (_m *MockSmartWalletRepository) SetPendingRecovery(ctx context.Context, walletID uuid.UUID, recovery *custody_entities.PendingRecovery) error {
	ret := _m.Called(ctx, walletID, recovery)

	return ret.Error(0)
}

// ClearPendingRecovery provides a mock function
func (_m *MockSmartWalletRepository) ClearPendingRecovery(ctx context.Context, walletID uuid.UUID) error {
	ret := _m.Called(ctx, walletID)

	return ret.Error(0)
}

// AddRecoveryApproval provides a mock function
func (_m *MockSmartWalletRepository) AddRecoveryApproval(ctx context.Context, walletID uuid.UUID, guardianID uuid.UUID) error {
	ret := _m.Called(ctx, walletID, guardianID)

	return ret.Error(0)
}

// NewMockSmartWalletRepository creates a new instance of MockSmartWalletRepository
func NewMockSmartWalletRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockSmartWalletRepository {
	mock := &MockSmartWalletRepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
