package matchmaking_out

import (
	"context"

	"github.com/google/uuid"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	"github.com/stretchr/testify/mock"
)

// MockLobbyRepository is a mock implementation of LobbyRepository
type MockLobbyRepository struct {
	mock.Mock
}

// Save provides a mock function
func (_m *MockLobbyRepository) Save(ctx context.Context, lobby *matchmaking_entities.MatchmakingLobby) error {
	ret := _m.Called(ctx, lobby)

	return ret.Error(0)
}

// FindByID provides a mock function
func (_m *MockLobbyRepository) FindByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.MatchmakingLobby, error) {
	ret := _m.Called(ctx, id)

	var r0 *matchmaking_entities.MatchmakingLobby
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*matchmaking_entities.MatchmakingLobby, error)); ok {
		return rf(ctx, id)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *matchmaking_entities.MatchmakingLobby); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*matchmaking_entities.MatchmakingLobby)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindByCreatorID provides a mock function
func (_m *MockLobbyRepository) FindByCreatorID(ctx context.Context, creatorID uuid.UUID) ([]*matchmaking_entities.MatchmakingLobby, error) {
	ret := _m.Called(ctx, creatorID)

	var r0 []*matchmaking_entities.MatchmakingLobby
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) ([]*matchmaking_entities.MatchmakingLobby, error)); ok {
		return rf(ctx, creatorID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) []*matchmaking_entities.MatchmakingLobby); ok {
		r0 = rf(ctx, creatorID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*matchmaking_entities.MatchmakingLobby)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindOpenLobbies provides a mock function
func (_m *MockLobbyRepository) FindOpenLobbies(ctx context.Context, gameID string, region string, tier string, limit int) ([]*matchmaking_entities.MatchmakingLobby, error) {
	ret := _m.Called(ctx, gameID, region, tier, limit)

	var r0 []*matchmaking_entities.MatchmakingLobby
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, int) ([]*matchmaking_entities.MatchmakingLobby, error)); ok {
		return rf(ctx, gameID, region, tier, limit)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, int) []*matchmaking_entities.MatchmakingLobby); ok {
		r0 = rf(ctx, gameID, region, tier, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*matchmaking_entities.MatchmakingLobby)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Update provides a mock function
func (_m *MockLobbyRepository) Update(ctx context.Context, lobby *matchmaking_entities.MatchmakingLobby) error {
	ret := _m.Called(ctx, lobby)

	return ret.Error(0)
}

// Delete provides a mock function
func (_m *MockLobbyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ret := _m.Called(ctx, id)

	return ret.Error(0)
}

// NewMockLobbyRepository creates a new instance of MockLobbyRepository
func NewMockLobbyRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockLobbyRepository {
	mock := &MockLobbyRepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
