package blockchain_ports

import (
	"context"

	"github.com/google/uuid"
	blockchain_entities "github.com/replay-api/replay-api/pkg/domain/blockchain/entities"
	"github.com/stretchr/testify/mock"
)

// MockPrizePoolRepository is a mock implementation of PrizePoolRepository
type MockPrizePoolRepository struct {
	mock.Mock
}

// Save provides a mock function
func (_m *MockPrizePoolRepository) Save(ctx context.Context, pool *blockchain_entities.OnChainPrizePool) error {
	ret := _m.Called(ctx, pool)

	return ret.Error(0)
}

// FindByID provides a mock function
func (_m *MockPrizePoolRepository) FindByID(ctx context.Context, id uuid.UUID) (*blockchain_entities.OnChainPrizePool, error) {
	ret := _m.Called(ctx, id)

	var r0 *blockchain_entities.OnChainPrizePool
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*blockchain_entities.OnChainPrizePool, error)); ok {
		return rf(ctx, id)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *blockchain_entities.OnChainPrizePool); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*blockchain_entities.OnChainPrizePool)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindByMatchID provides a mock function
func (_m *MockPrizePoolRepository) FindByMatchID(ctx context.Context, matchID uuid.UUID) (*blockchain_entities.OnChainPrizePool, error) {
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

// FindByStatus provides a mock function
func (_m *MockPrizePoolRepository) FindByStatus(ctx context.Context, status blockchain_entities.OnChainPrizePoolStatus) ([]*blockchain_entities.OnChainPrizePool, error) {
	ret := _m.Called(ctx, status)

	var r0 []*blockchain_entities.OnChainPrizePool
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, blockchain_entities.OnChainPrizePoolStatus) ([]*blockchain_entities.OnChainPrizePool, error)); ok {
		return rf(ctx, status)
	}

	if rf, ok := ret.Get(0).(func(context.Context, blockchain_entities.OnChainPrizePoolStatus) []*blockchain_entities.OnChainPrizePool); ok {
		r0 = rf(ctx, status)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*blockchain_entities.OnChainPrizePool)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindPendingDistribution provides a mock function
func (_m *MockPrizePoolRepository) FindPendingDistribution(ctx context.Context) ([]*blockchain_entities.OnChainPrizePool, error) {
	ret := _m.Called(ctx)

	var r0 []*blockchain_entities.OnChainPrizePool
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context) ([]*blockchain_entities.OnChainPrizePool, error)); ok {
		return rf(ctx)
	}

	if rf, ok := ret.Get(0).(func(context.Context) []*blockchain_entities.OnChainPrizePool); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*blockchain_entities.OnChainPrizePool)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// UpdateSyncState provides a mock function
func (_m *MockPrizePoolRepository) UpdateSyncState(ctx context.Context, id uuid.UUID, blockNumber uint64, synced bool) error {
	ret := _m.Called(ctx, id, blockNumber, synced)

	return ret.Error(0)
}

// NewMockPrizePoolRepository creates a new instance of MockPrizePoolRepository
func NewMockPrizePoolRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPrizePoolRepository {
	mock := &MockPrizePoolRepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
