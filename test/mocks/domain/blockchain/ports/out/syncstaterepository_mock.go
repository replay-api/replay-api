package blockchain_ports

import (
	"context"

	blockchain_vo "github.com/replay-api/replay-api/pkg/domain/blockchain/value-objects"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	"github.com/stretchr/testify/mock"
)

// MockSyncStateRepository is a mock implementation of SyncStateRepository
type MockSyncStateRepository struct {
	mock.Mock
}

// GetLastSyncedBlock provides a mock function
func (_m *MockSyncStateRepository) GetLastSyncedBlock(ctx context.Context, chainID blockchain_vo.ChainID, contractAddr wallet_vo.EVMAddress) (uint64, error) {
	ret := _m.Called(ctx, chainID, contractAddr)

	var r0 uint64
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, blockchain_vo.ChainID, wallet_vo.EVMAddress) (uint64, error)); ok {
		return rf(ctx, chainID, contractAddr)
	}

	if rf, ok := ret.Get(0).(func(context.Context, blockchain_vo.ChainID, wallet_vo.EVMAddress) uint64); ok {
		r0 = rf(ctx, chainID, contractAddr)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(uint64)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// SetLastSyncedBlock provides a mock function
func (_m *MockSyncStateRepository) SetLastSyncedBlock(ctx context.Context, chainID blockchain_vo.ChainID, contractAddr wallet_vo.EVMAddress, blockNumber uint64) error {
	ret := _m.Called(ctx, chainID, contractAddr, blockNumber)

	return ret.Error(0)
}

// NewMockSyncStateRepository creates a new instance of MockSyncStateRepository
func NewMockSyncStateRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockSyncStateRepository {
	mock := &MockSyncStateRepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
