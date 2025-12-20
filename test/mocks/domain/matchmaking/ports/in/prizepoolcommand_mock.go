package matchmaking_in

import (
	"context"

	"github.com/google/uuid"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockPrizePoolCommand is a mock implementation of PrizePoolCommand
type MockPrizePoolCommand struct {
	mock.Mock
}

// CreatePrizePool provides a mock function
func (_m *MockPrizePoolCommand) CreatePrizePool(ctx context.Context, cmd matchmaking_in.CreatePrizePoolCommand) (*matchmaking_entities.PrizePool, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *matchmaking_entities.PrizePool
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, matchmaking_in.CreatePrizePoolCommand) (*matchmaking_entities.PrizePool, error)); ok {
		return rf(ctx, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, matchmaking_in.CreatePrizePoolCommand) *matchmaking_entities.PrizePool); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*matchmaking_entities.PrizePool)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// AddPlayerContribution provides a mock function
func (_m *MockPrizePoolCommand) AddPlayerContribution(ctx context.Context, cmd matchmaking_in.AddContributionCommand) error {
	ret := _m.Called(ctx, cmd)

	return ret.Error(0)
}

// LockPool provides a mock function
func (_m *MockPrizePoolCommand) LockPool(ctx context.Context, prizePoolID uuid.UUID) error {
	ret := _m.Called(ctx, prizePoolID)

	return ret.Error(0)
}

// DistributePrizes provides a mock function
func (_m *MockPrizePoolCommand) DistributePrizes(ctx context.Context, cmd matchmaking_in.DistributePrizesCommand) error {
	ret := _m.Called(ctx, cmd)

	return ret.Error(0)
}

// CancelPool provides a mock function
func (_m *MockPrizePoolCommand) CancelPool(ctx context.Context, prizePoolID uuid.UUID, reason string) error {
	ret := _m.Called(ctx, prizePoolID, reason)

	return ret.Error(0)
}

// NewMockPrizePoolCommand creates a new instance of MockPrizePoolCommand
func NewMockPrizePoolCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPrizePoolCommand {
	mock := &MockPrizePoolCommand{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
