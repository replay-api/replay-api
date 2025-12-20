package matchmaking_in

import (
	"context"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockRatingService is a mock implementation of RatingService
type MockRatingService struct {
	mock.Mock
}

// GetPlayerRating provides a mock function
func (_m *MockRatingService) GetPlayerRating(ctx context.Context, playerID uuid.UUID, gameID common.GameIDKey) (*matchmaking_entities.PlayerRating, error) {
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

// UpdateRatingsAfterMatch provides a mock function
func (_m *MockRatingService) UpdateRatingsAfterMatch(ctx context.Context, cmd matchmaking_in.UpdateRatingsCommand) error {
	ret := _m.Called(ctx, cmd)

	return ret.Error(0)
}

// GetLeaderboard provides a mock function
func (_m *MockRatingService) GetLeaderboard(ctx context.Context, gameID common.GameIDKey, limit int) ([]*matchmaking_entities.PlayerRating, error) {
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
func (_m *MockRatingService) GetRankDistribution(ctx context.Context, gameID common.GameIDKey) (map[matchmaking_entities.Rank]int, error) {
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

// NewMockRatingService creates a new instance of MockRatingService
func NewMockRatingService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockRatingService {
	mock := &MockRatingService{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
