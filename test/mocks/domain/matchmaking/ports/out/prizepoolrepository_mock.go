package matchmaking_out

import (
	"context"

	"github.com/google/uuid"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	"github.com/stretchr/testify/mock"
)

// MockPrizePoolRepository is a mock implementation of PrizePoolRepository
type MockPrizePoolRepository struct {
	mock.Mock
}

// Save provides a mock function
func (_m *MockPrizePoolRepository) Save(ctx context.Context, pool *matchmaking_entities.PrizePool) error {
	ret := _m.Called(ctx, pool)

	return ret.Error(0)
}

// FindByID provides a mock function
func (_m *MockPrizePoolRepository) FindByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.PrizePool, error) {
	ret := _m.Called(ctx, id)

	var r0 *matchmaking_entities.PrizePool
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*matchmaking_entities.PrizePool, error)); ok {
		return rf(ctx, id)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *matchmaking_entities.PrizePool); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*matchmaking_entities.PrizePool)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindByMatchID provides a mock function
func (_m *MockPrizePoolRepository) FindByMatchID(ctx context.Context, matchID uuid.UUID) (*matchmaking_entities.PrizePool, error) {
	ret := _m.Called(ctx, matchID)

	var r0 *matchmaking_entities.PrizePool
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*matchmaking_entities.PrizePool, error)); ok {
		return rf(ctx, matchID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *matchmaking_entities.PrizePool); ok {
		r0 = rf(ctx, matchID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*matchmaking_entities.PrizePool)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindPendingDistributions provides a mock function
func (_m *MockPrizePoolRepository) FindPendingDistributions(ctx context.Context, limit int) ([]*matchmaking_entities.PrizePool, error) {
	ret := _m.Called(ctx, limit)

	var r0 []*matchmaking_entities.PrizePool
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, int) ([]*matchmaking_entities.PrizePool, error)); ok {
		return rf(ctx, limit)
	}

	if rf, ok := ret.Get(0).(func(context.Context, int) []*matchmaking_entities.PrizePool); ok {
		r0 = rf(ctx, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*matchmaking_entities.PrizePool)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Update provides a mock function
func (_m *MockPrizePoolRepository) Update(ctx context.Context, pool *matchmaking_entities.PrizePool) error {
	ret := _m.Called(ctx, pool)

	return ret.Error(0)
}

// Delete provides a mock function
func (_m *MockPrizePoolRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ret := _m.Called(ctx, id)

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
