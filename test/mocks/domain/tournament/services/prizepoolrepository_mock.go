package tournament_services

import (
	"context"

	"github.com/google/uuid"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
	"github.com/stretchr/testify/mock"
)

// MockPrizePoolRepository is a mock implementation of PrizePoolRepository
type MockPrizePoolRepository struct {
	mock.Mock
}

// Create provides a mock function
func (_m *MockPrizePoolRepository) Create(ctx context.Context, pool *tournament_entities.PrizePool) error {
	ret := _m.Called(ctx, pool)

	return ret.Error(0)
}

// Update provides a mock function
func (_m *MockPrizePoolRepository) Update(ctx context.Context, pool *tournament_entities.PrizePool) error {
	ret := _m.Called(ctx, pool)

	return ret.Error(0)
}

// GetByID provides a mock function
func (_m *MockPrizePoolRepository) GetByID(ctx context.Context, id uuid.UUID) (*tournament_entities.PrizePool, error) {
	ret := _m.Called(ctx, id)

	var r0 *tournament_entities.PrizePool
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*tournament_entities.PrizePool, error)); ok {
		return rf(ctx, id)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *tournament_entities.PrizePool); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tournament_entities.PrizePool)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetByTournamentID provides a mock function
func (_m *MockPrizePoolRepository) GetByTournamentID(ctx context.Context, tournamentID uuid.UUID) (*tournament_entities.PrizePool, error) {
	ret := _m.Called(ctx, tournamentID)

	var r0 *tournament_entities.PrizePool
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*tournament_entities.PrizePool, error)); ok {
		return rf(ctx, tournamentID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *tournament_entities.PrizePool); ok {
		r0 = rf(ctx, tournamentID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tournament_entities.PrizePool)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
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
