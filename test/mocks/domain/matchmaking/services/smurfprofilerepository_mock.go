package matchmaking_services

import (
	"context"

	"github.com/google/uuid"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	"github.com/stretchr/testify/mock"
)

// MockSmurfProfileRepository is a mock implementation of SmurfProfileRepository
type MockSmurfProfileRepository struct {
	mock.Mock
}

// Create provides a mock function
func (_m *MockSmurfProfileRepository) Create(ctx context.Context, profile *matchmaking_entities.SmurfProfile) error {
	ret := _m.Called(ctx, profile)

	return ret.Error(0)
}

// Update provides a mock function
func (_m *MockSmurfProfileRepository) Update(ctx context.Context, profile *matchmaking_entities.SmurfProfile) error {
	ret := _m.Called(ctx, profile)

	return ret.Error(0)
}

// GetByPlayerID provides a mock function
func (_m *MockSmurfProfileRepository) GetByPlayerID(ctx context.Context, playerID uuid.UUID) (*matchmaking_entities.SmurfProfile, error) {
	ret := _m.Called(ctx, playerID)

	var r0 *matchmaking_entities.SmurfProfile
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*matchmaking_entities.SmurfProfile, error)); ok {
		return rf(ctx, playerID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *matchmaking_entities.SmurfProfile); ok {
		r0 = rf(ctx, playerID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*matchmaking_entities.SmurfProfile)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetFlaggedProfiles provides a mock function
func (_m *MockSmurfProfileRepository) GetFlaggedProfiles(ctx context.Context, limit int, offset int) ([]matchmaking_entities.SmurfProfile, error) {
	ret := _m.Called(ctx, limit, offset)

	var r0 []matchmaking_entities.SmurfProfile
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, int, int) ([]matchmaking_entities.SmurfProfile, error)); ok {
		return rf(ctx, limit, offset)
	}

	if rf, ok := ret.Get(0).(func(context.Context, int, int) []matchmaking_entities.SmurfProfile); ok {
		r0 = rf(ctx, limit, offset)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]matchmaking_entities.SmurfProfile)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockSmurfProfileRepository creates a new instance of MockSmurfProfileRepository
func NewMockSmurfProfileRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockSmurfProfileRepository {
	mock := &MockSmurfProfileRepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
