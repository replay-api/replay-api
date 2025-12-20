package matchmaking_services

import (
	"context"

	"github.com/google/uuid"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_services "github.com/replay-api/replay-api/pkg/domain/matchmaking/services"
	"github.com/stretchr/testify/mock"
)

// MockPlayerRatingRepository is a mock implementation of PlayerRatingRepository
type MockPlayerRatingRepository struct {
	mock.Mock
}

// GetPlayerRating provides a mock function
func (_m *MockPlayerRatingRepository) GetPlayerRating(ctx context.Context, playerID uuid.UUID) (*matchmaking_entities.PlayerRating, error) {
	ret := _m.Called(ctx, playerID)

	var r0 *matchmaking_entities.PlayerRating
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*matchmaking_entities.PlayerRating, error)); ok {
		return rf(ctx, playerID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *matchmaking_entities.PlayerRating); ok {
		r0 = rf(ctx, playerID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*matchmaking_entities.PlayerRating)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetRatingHistory provides a mock function
func (_m *MockPlayerRatingRepository) GetRatingHistory(ctx context.Context, playerID uuid.UUID, limit int) ([]matchmaking_services.RatingSnapshot, error) {
	ret := _m.Called(ctx, playerID, limit)

	var r0 []matchmaking_services.RatingSnapshot
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, int) ([]matchmaking_services.RatingSnapshot, error)); ok {
		return rf(ctx, playerID, limit)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, int) []matchmaking_services.RatingSnapshot); ok {
		r0 = rf(ctx, playerID, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]matchmaking_services.RatingSnapshot)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
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
