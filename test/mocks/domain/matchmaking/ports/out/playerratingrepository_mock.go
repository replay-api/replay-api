package matchmaking_out

import (
	"context"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	"github.com/stretchr/testify/mock"
)

// MockPlayerRatingRepository is a mock implementation of PlayerRatingRepository
type MockPlayerRatingRepository struct {
	mock.Mock
}

// Save provides a mock function
func (_m *MockPlayerRatingRepository) Save(ctx context.Context, rating *matchmaking_entities.PlayerRating) error {
	ret := _m.Called(ctx, rating)

	return ret.Error(0)
}

// Update provides a mock function
func (_m *MockPlayerRatingRepository) Update(ctx context.Context, rating *matchmaking_entities.PlayerRating) error {
	ret := _m.Called(ctx, rating)

	return ret.Error(0)
}

// FindByPlayerAndGame provides a mock function
func (_m *MockPlayerRatingRepository) FindByPlayerAndGame(ctx context.Context, playerID uuid.UUID, gameID common.GameIDKey) (*matchmaking_entities.PlayerRating, error) {
	ret := _m.Called(ctx, playerID, gameID)

	var r0 *matchmaking_entities.PlayerRating
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, common.GameIDKey) (*matchmaking_entities.PlayerRating, error)); ok {
		return rf(ctx, playerID, gameID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, common.GameIDKey) *matchmaking_entities.PlayerRating); ok {
		r0 = rf(ctx, playerID, gameID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*matchmaking_entities.PlayerRating)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetByID provides a mock function
func (_m *MockPlayerRatingRepository) GetByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.PlayerRating, error) {
	ret := _m.Called(ctx, id)

	var r0 *matchmaking_entities.PlayerRating
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*matchmaking_entities.PlayerRating, error)); ok {
		return rf(ctx, id)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *matchmaking_entities.PlayerRating); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*matchmaking_entities.PlayerRating)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetTopPlayers provides a mock function
func (_m *MockPlayerRatingRepository) GetTopPlayers(ctx context.Context, gameID common.GameIDKey, limit int) ([]*matchmaking_entities.PlayerRating, error) {
	ret := _m.Called(ctx, gameID, limit)

	var r0 []*matchmaking_entities.PlayerRating
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, common.GameIDKey, int) ([]*matchmaking_entities.PlayerRating, error)); ok {
		return rf(ctx, gameID, limit)
	}

	if rf, ok := ret.Get(0).(func(context.Context, common.GameIDKey, int) []*matchmaking_entities.PlayerRating); ok {
		r0 = rf(ctx, gameID, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*matchmaking_entities.PlayerRating)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetRankDistribution provides a mock function
func (_m *MockPlayerRatingRepository) GetRankDistribution(ctx context.Context, gameID common.GameIDKey) (map[matchmaking_entities.Rank]int, error) {
	ret := _m.Called(ctx, gameID)

	var r0 map[matchmaking_entities.Rank]int
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, common.GameIDKey) (map[matchmaking_entities.Rank]int, error)); ok {
		return rf(ctx, gameID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, common.GameIDKey) map[matchmaking_entities.Rank]int); ok {
		r0 = rf(ctx, gameID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[matchmaking_entities.Rank]int)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Delete provides a mock function
func (_m *MockPlayerRatingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ret := _m.Called(ctx, id)

	return ret.Error(0)
}

// NewMockPlayerRatingRepository creates a new instance of MockPlayerRatingRepository
func NewMockPlayerRatingRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPlayerRatingRepository {
	mock := &MockPlayerRatingRepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
