package blockchain_ports

import (
	"context"

	"github.com/google/uuid"
	blockchain_entities "github.com/replay-api/replay-api/pkg/domain/blockchain/entities"
	blockchain_vo "github.com/replay-api/replay-api/pkg/domain/blockchain/value-objects"
	"github.com/stretchr/testify/mock"
)

// MockTransactionRepository is a mock implementation of TransactionRepository
type MockTransactionRepository struct {
	mock.Mock
}

// Save provides a mock function
func (_m *MockTransactionRepository) Save(ctx context.Context, tx *blockchain_entities.BlockchainTransaction) error {
	ret := _m.Called(ctx, tx)

	return ret.Error(0)
}

// FindByID provides a mock function
func (_m *MockTransactionRepository) FindByID(ctx context.Context, id uuid.UUID) (*blockchain_entities.BlockchainTransaction, error) {
	ret := _m.Called(ctx, id)

	var r0 *blockchain_entities.BlockchainTransaction
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*blockchain_entities.BlockchainTransaction, error)); ok {
		return rf(ctx, id)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *blockchain_entities.BlockchainTransaction); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*blockchain_entities.BlockchainTransaction)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindByTxHash provides a mock function
func (_m *MockTransactionRepository) FindByTxHash(ctx context.Context, chainID blockchain_vo.ChainID, txHash blockchain_vo.TxHash) (*blockchain_entities.BlockchainTransaction, error) {
	ret := _m.Called(ctx, chainID, txHash)

	var r0 *blockchain_entities.BlockchainTransaction
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, blockchain_vo.ChainID, blockchain_vo.TxHash) (*blockchain_entities.BlockchainTransaction, error)); ok {
		return rf(ctx, chainID, txHash)
	}

	if rf, ok := ret.Get(0).(func(context.Context, blockchain_vo.ChainID, blockchain_vo.TxHash) *blockchain_entities.BlockchainTransaction); ok {
		r0 = rf(ctx, chainID, txHash)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*blockchain_entities.BlockchainTransaction)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindPending provides a mock function
func (_m *MockTransactionRepository) FindPending(ctx context.Context, chainID blockchain_vo.ChainID) ([]*blockchain_entities.BlockchainTransaction, error) {
	ret := _m.Called(ctx, chainID)

	var r0 []*blockchain_entities.BlockchainTransaction
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, blockchain_vo.ChainID) ([]*blockchain_entities.BlockchainTransaction, error)); ok {
		return rf(ctx, chainID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, blockchain_vo.ChainID) []*blockchain_entities.BlockchainTransaction); ok {
		r0 = rf(ctx, chainID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*blockchain_entities.BlockchainTransaction)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindByWallet provides a mock function
func (_m *MockTransactionRepository) FindByWallet(ctx context.Context, walletID uuid.UUID, limit int, offset int) ([]*blockchain_entities.BlockchainTransaction, int64, error) {
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

// FindByMatch provides a mock function
func (_m *MockTransactionRepository) FindByMatch(ctx context.Context, matchID uuid.UUID) ([]*blockchain_entities.BlockchainTransaction, error) {
	ret := _m.Called(ctx, matchID)

	var r0 []*blockchain_entities.BlockchainTransaction
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) ([]*blockchain_entities.BlockchainTransaction, error)); ok {
		return rf(ctx, matchID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) []*blockchain_entities.BlockchainTransaction); ok {
		r0 = rf(ctx, matchID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*blockchain_entities.BlockchainTransaction)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// UpdateStatus provides a mock function
func (_m *MockTransactionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status blockchain_entities.TransactionStatus, confirmations uint64) error {
	ret := _m.Called(ctx, id, status, confirmations)

	return ret.Error(0)
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
