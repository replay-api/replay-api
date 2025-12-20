package custody_in

import (
	"context"

	"github.com/google/uuid"
	custody_entities "github.com/replay-api/replay-api/pkg/domain/custody/entities"
	custody_in "github.com/replay-api/replay-api/pkg/domain/custody/ports/in"
	custody_vo "github.com/replay-api/replay-api/pkg/domain/custody/value-objects"
	"github.com/stretchr/testify/mock"
)

// MockWalletService is a mock implementation of WalletService
type MockWalletService struct {
	mock.Mock
}

// CreateWallet provides a mock function
func (_m *MockWalletService) CreateWallet(ctx context.Context, req *custody_in.CreateWalletRequest) (*custody_in.CreateWalletResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_in.CreateWalletResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.CreateWalletRequest) (*custody_in.CreateWalletResult, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.CreateWalletRequest) *custody_in.CreateWalletResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.CreateWalletResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetWallet provides a mock function
func (_m *MockWalletService) GetWallet(ctx context.Context, walletID uuid.UUID) (*custody_entities.SmartWallet, error) {
	ret := _m.Called(ctx, walletID)

	var r0 *custody_entities.SmartWallet
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*custody_entities.SmartWallet, error)); ok {
		return rf(ctx, walletID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *custody_entities.SmartWallet); ok {
		r0 = rf(ctx, walletID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_entities.SmartWallet)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetWalletByAddress provides a mock function
func (_m *MockWalletService) GetWalletByAddress(ctx context.Context, chainID custody_vo.ChainID, address string) (*custody_entities.SmartWallet, error) {
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

// GetUserWallets provides a mock function
func (_m *MockWalletService) GetUserWallets(ctx context.Context, userID uuid.UUID) ([]*custody_entities.SmartWallet, error) {
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

// DeployWallet provides a mock function
func (_m *MockWalletService) DeployWallet(ctx context.Context, walletID uuid.UUID, chainID custody_vo.ChainID) (*custody_in.DeployWalletResult, error) {
	ret := _m.Called(ctx, walletID, chainID)

	var r0 *custody_in.DeployWalletResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, custody_vo.ChainID) (*custody_in.DeployWalletResult, error)); ok {
		return rf(ctx, walletID, chainID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, custody_vo.ChainID) *custody_in.DeployWalletResult); ok {
		r0 = rf(ctx, walletID, chainID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.DeployWalletResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetBalance provides a mock function
func (_m *MockWalletService) GetBalance(ctx context.Context, walletID uuid.UUID, chainID custody_vo.ChainID) (*custody_in.WalletBalance, error) {
	ret := _m.Called(ctx, walletID, chainID)

	var r0 *custody_in.WalletBalance
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, custody_vo.ChainID) (*custody_in.WalletBalance, error)); ok {
		return rf(ctx, walletID, chainID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, custody_vo.ChainID) *custody_in.WalletBalance); ok {
		r0 = rf(ctx, walletID, chainID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.WalletBalance)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetAllBalances provides a mock function
func (_m *MockWalletService) GetAllBalances(ctx context.Context, walletID uuid.UUID) ([]*custody_in.WalletBalance, error) {
	ret := _m.Called(ctx, walletID)

	var r0 []*custody_in.WalletBalance
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) ([]*custody_in.WalletBalance, error)); ok {
		return rf(ctx, walletID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) []*custody_in.WalletBalance); ok {
		r0 = rf(ctx, walletID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*custody_in.WalletBalance)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetTokenBalance provides a mock function
func (_m *MockWalletService) GetTokenBalance(ctx context.Context, walletID uuid.UUID, chainID custody_vo.ChainID, tokenAddress string) (*custody_in.TokenBalance, error) {
	ret := _m.Called(ctx, walletID, chainID, tokenAddress)

	var r0 *custody_in.TokenBalance
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, custody_vo.ChainID, string) (*custody_in.TokenBalance, error)); ok {
		return rf(ctx, walletID, chainID, tokenAddress)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, custody_vo.ChainID, string) *custody_in.TokenBalance); ok {
		r0 = rf(ctx, walletID, chainID, tokenAddress)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.TokenBalance)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Transfer provides a mock function
func (_m *MockWalletService) Transfer(ctx context.Context, req *custody_in.TransferRequest) (*custody_in.TransferResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_in.TransferResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.TransferRequest) (*custody_in.TransferResult, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.TransferRequest) *custody_in.TransferResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.TransferResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// TransferToken provides a mock function
func (_m *MockWalletService) TransferToken(ctx context.Context, req *custody_in.TokenTransferRequest) (*custody_in.TransferResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_in.TransferResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.TokenTransferRequest) (*custody_in.TransferResult, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.TokenTransferRequest) *custody_in.TransferResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.TransferResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// ExecuteTransaction provides a mock function
func (_m *MockWalletService) ExecuteTransaction(ctx context.Context, req *custody_in.ExecuteTxRequest) (*custody_in.ExecuteTxResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_in.ExecuteTxResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.ExecuteTxRequest) (*custody_in.ExecuteTxResult, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.ExecuteTxRequest) *custody_in.ExecuteTxResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.ExecuteTxResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// BatchExecute provides a mock function
func (_m *MockWalletService) BatchExecute(ctx context.Context, req *custody_in.BatchExecuteRequest) (*custody_in.BatchExecuteResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_in.BatchExecuteResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.BatchExecuteRequest) (*custody_in.BatchExecuteResult, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.BatchExecuteRequest) *custody_in.BatchExecuteResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.BatchExecuteResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetTransactionStatus provides a mock function
func (_m *MockWalletService) GetTransactionStatus(ctx context.Context, txID uuid.UUID) (*custody_in.TxStatusResult, error) {
	ret := _m.Called(ctx, txID)

	var r0 *custody_in.TxStatusResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*custody_in.TxStatusResult, error)); ok {
		return rf(ctx, txID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *custody_in.TxStatusResult); ok {
		r0 = rf(ctx, txID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.TxStatusResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetPendingTransactions provides a mock function
func (_m *MockWalletService) GetPendingTransactions(ctx context.Context, walletID uuid.UUID) ([]*custody_in.PendingTx, error) {
	ret := _m.Called(ctx, walletID)

	var r0 []*custody_in.PendingTx
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) ([]*custody_in.PendingTx, error)); ok {
		return rf(ctx, walletID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) []*custody_in.PendingTx); ok {
		r0 = rf(ctx, walletID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*custody_in.PendingTx)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetSpendingStatus provides a mock function
func (_m *MockWalletService) GetSpendingStatus(ctx context.Context, walletID uuid.UUID) (*custody_in.SpendingStatus, error) {
	ret := _m.Called(ctx, walletID)

	var r0 *custody_in.SpendingStatus
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*custody_in.SpendingStatus, error)); ok {
		return rf(ctx, walletID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *custody_in.SpendingStatus); ok {
		r0 = rf(ctx, walletID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.SpendingStatus)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// UpdateSpendingLimits provides a mock function
func (_m *MockWalletService) UpdateSpendingLimits(ctx context.Context, walletID uuid.UUID, limits *custody_entities.TransactionLimits) error {
	ret := _m.Called(ctx, walletID, limits)

	return ret.Error(0)
}

// AddSessionKey provides a mock function
func (_m *MockWalletService) AddSessionKey(ctx context.Context, walletID uuid.UUID, req *custody_in.AddSessionKeyRequest) (*custody_in.SessionKeyResult, error) {
	ret := _m.Called(ctx, walletID, req)

	var r0 *custody_in.SessionKeyResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, *custody_in.AddSessionKeyRequest) (*custody_in.SessionKeyResult, error)); ok {
		return rf(ctx, walletID, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, *custody_in.AddSessionKeyRequest) *custody_in.SessionKeyResult); ok {
		r0 = rf(ctx, walletID, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.SessionKeyResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// RevokeSessionKey provides a mock function
func (_m *MockWalletService) RevokeSessionKey(ctx context.Context, walletID uuid.UUID, keyAddress string) error {
	ret := _m.Called(ctx, walletID, keyAddress)

	return ret.Error(0)
}

// GetSessionKeys provides a mock function
func (_m *MockWalletService) GetSessionKeys(ctx context.Context, walletID uuid.UUID) ([]*custody_entities.SessionKey, error) {
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

// FreezeWallet provides a mock function
func (_m *MockWalletService) FreezeWallet(ctx context.Context, walletID uuid.UUID, reason string) error {
	ret := _m.Called(ctx, walletID, reason)

	return ret.Error(0)
}

// UnfreezeWallet provides a mock function
func (_m *MockWalletService) UnfreezeWallet(ctx context.Context, walletID uuid.UUID) error {
	ret := _m.Called(ctx, walletID)

	return ret.Error(0)
}

// UpdateKYCStatus provides a mock function
func (_m *MockWalletService) UpdateKYCStatus(ctx context.Context, walletID uuid.UUID, status custody_entities.KYCStatus) error {
	ret := _m.Called(ctx, walletID, status)

	return ret.Error(0)
}

// NewMockWalletService creates a new instance of MockWalletService
func NewMockWalletService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockWalletService {
	mock := &MockWalletService{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
