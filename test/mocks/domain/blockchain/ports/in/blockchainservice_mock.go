package blockchain_ports

import (
	"context"

	big "math/big"

	"github.com/google/uuid"
	blockchain_entities "github.com/replay-api/replay-api/pkg/domain/blockchain/entities"
	blockchain_in "github.com/replay-api/replay-api/pkg/domain/blockchain/ports/in"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	"github.com/stretchr/testify/mock"
)

// MockBlockchainService is a mock implementation of BlockchainService
type MockBlockchainService struct {
	mock.Mock
}

// CreatePrizePool provides a mock function
func (_m *MockBlockchainService) CreatePrizePool(ctx context.Context, cmd blockchain_in.CreatePrizePoolCommand) (*blockchain_entities.OnChainPrizePool, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *blockchain_entities.OnChainPrizePool
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, blockchain_in.CreatePrizePoolCommand) (*blockchain_entities.OnChainPrizePool, error)); ok {
		return rf(ctx, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, blockchain_in.CreatePrizePoolCommand) *blockchain_entities.OnChainPrizePool); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*blockchain_entities.OnChainPrizePool)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// JoinPrizePool provides a mock function
func (_m *MockBlockchainService) JoinPrizePool(ctx context.Context, cmd blockchain_in.JoinPrizePoolCommand) error {
	ret := _m.Called(ctx, cmd)

	return ret.Error(0)
}

// LockPrizePool provides a mock function
func (_m *MockBlockchainService) LockPrizePool(ctx context.Context, matchID uuid.UUID) error {
	ret := _m.Called(ctx, matchID)

	return ret.Error(0)
}

// DistributePrizes provides a mock function
func (_m *MockBlockchainService) DistributePrizes(ctx context.Context, cmd blockchain_in.DistributePrizesCommand) error {
	ret := _m.Called(ctx, cmd)

	return ret.Error(0)
}

// CancelPrizePool provides a mock function
func (_m *MockBlockchainService) CancelPrizePool(ctx context.Context, matchID uuid.UUID) error {
	ret := _m.Called(ctx, matchID)

	return ret.Error(0)
}

// DepositToVault provides a mock function
func (_m *MockBlockchainService) DepositToVault(ctx context.Context, cmd blockchain_in.DepositCommand) (*blockchain_entities.BlockchainTransaction, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *blockchain_entities.BlockchainTransaction
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, blockchain_in.DepositCommand) (*blockchain_entities.BlockchainTransaction, error)); ok {
		return rf(ctx, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, blockchain_in.DepositCommand) *blockchain_entities.BlockchainTransaction); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*blockchain_entities.BlockchainTransaction)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// WithdrawFromVault provides a mock function
func (_m *MockBlockchainService) WithdrawFromVault(ctx context.Context, cmd blockchain_in.WithdrawCommand) (*blockchain_entities.BlockchainTransaction, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *blockchain_entities.BlockchainTransaction
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, blockchain_in.WithdrawCommand) (*blockchain_entities.BlockchainTransaction, error)); ok {
		return rf(ctx, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, blockchain_in.WithdrawCommand) *blockchain_entities.BlockchainTransaction); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*blockchain_entities.BlockchainTransaction)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// RecordTransaction provides a mock function
func (_m *MockBlockchainService) RecordTransaction(ctx context.Context, cmd blockchain_in.RecordLedgerEntryCommand) error {
	ret := _m.Called(ctx, cmd)

	return ret.Error(0)
}

// GetLedgerBalance provides a mock function
func (_m *MockBlockchainService) GetLedgerBalance(ctx context.Context, account wallet_vo.EVMAddress, token wallet_vo.EVMAddress) (*big.Int, error) {
	ret := _m.Called(ctx, account, token)

	var r0 *big.Int
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, wallet_vo.EVMAddress, wallet_vo.EVMAddress) (*big.Int, error)); ok {
		return rf(ctx, account, token)
	}

	if rf, ok := ret.Get(0).(func(context.Context, wallet_vo.EVMAddress, wallet_vo.EVMAddress) *big.Int); ok {
		r0 = rf(ctx, account, token)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*big.Int)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// SyncPrizePool provides a mock function
func (_m *MockBlockchainService) SyncPrizePool(ctx context.Context, matchID uuid.UUID) error {
	ret := _m.Called(ctx, matchID)

	return ret.Error(0)
}

// SyncAllPendingPools provides a mock function
func (_m *MockBlockchainService) SyncAllPendingPools(ctx context.Context) error {
	ret := _m.Called(ctx)

	return ret.Error(0)
}

// GetPrizePool provides a mock function
func (_m *MockBlockchainService) GetPrizePool(ctx context.Context, matchID uuid.UUID) (*blockchain_entities.OnChainPrizePool, error) {
	ret := _m.Called(ctx, matchID)

	var r0 *blockchain_entities.OnChainPrizePool
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*blockchain_entities.OnChainPrizePool, error)); ok {
		return rf(ctx, matchID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *blockchain_entities.OnChainPrizePool); ok {
		r0 = rf(ctx, matchID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*blockchain_entities.OnChainPrizePool)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetTransaction provides a mock function
func (_m *MockBlockchainService) GetTransaction(ctx context.Context, txID uuid.UUID) (*blockchain_entities.BlockchainTransaction, error) {
	ret := _m.Called(ctx, txID)

	var r0 *blockchain_entities.BlockchainTransaction
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*blockchain_entities.BlockchainTransaction, error)); ok {
		return rf(ctx, txID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *blockchain_entities.BlockchainTransaction); ok {
		r0 = rf(ctx, txID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*blockchain_entities.BlockchainTransaction)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetTransactionsByWallet provides a mock function
func (_m *MockBlockchainService) GetTransactionsByWallet(ctx context.Context, walletID uuid.UUID, limit int, offset int) ([]*blockchain_entities.BlockchainTransaction, int64, error) {
	ret := _m.Called(ctx, walletID, limit, offset)

	var r0 []*blockchain_entities.BlockchainTransaction
	var r1 int64
	var r2 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, int, int) ([]*blockchain_entities.BlockchainTransaction, int64, error)); ok {
		return rf(ctx, walletID, limit, offset)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, int, int) []*blockchain_entities.BlockchainTransaction); ok {
		r0 = rf(ctx, walletID, limit, offset)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*blockchain_entities.BlockchainTransaction)
		}
	}
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, int, int) int64); ok {
		r1 = rf(ctx, walletID, limit, offset)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(int64)
		}
	}
	r2 = ret.Error(2)

	return r0, r1, r2
}

// NewMockBlockchainService creates a new instance of MockBlockchainService
func NewMockBlockchainService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockBlockchainService {
	mock := &MockBlockchainService{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
