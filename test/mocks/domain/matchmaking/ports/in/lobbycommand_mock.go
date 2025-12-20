package matchmaking_in

import (
	"context"

	"github.com/google/uuid"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockLobbyCommand is a mock implementation of LobbyCommand
type MockLobbyCommand struct {
	mock.Mock
}

// CreateLobby provides a mock function
func (_m *MockLobbyCommand) CreateLobby(ctx context.Context, cmd matchmaking_in.CreateLobbyCommand) (*matchmaking_entities.MatchmakingLobby, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *matchmaking_entities.MatchmakingLobby
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, matchmaking_in.CreateLobbyCommand) (*matchmaking_entities.MatchmakingLobby, error)); ok {
		return rf(ctx, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, matchmaking_in.CreateLobbyCommand) *matchmaking_entities.MatchmakingLobby); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*matchmaking_entities.MatchmakingLobby)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// JoinLobby provides a mock function
func (_m *MockLobbyCommand) JoinLobby(ctx context.Context, cmd matchmaking_in.JoinLobbyCommand) error {
	ret := _m.Called(ctx, cmd)

	return ret.Error(0)
}

// LeaveLobby provides a mock function
func (_m *MockLobbyCommand) LeaveLobby(ctx context.Context, cmd matchmaking_in.LeaveLobbyCommand) error {
	ret := _m.Called(ctx, cmd)

	return ret.Error(0)
}

// SetPlayerReady provides a mock function
func (_m *MockLobbyCommand) SetPlayerReady(ctx context.Context, cmd matchmaking_in.SetPlayerReadyCommand) error {
	ret := _m.Called(ctx, cmd)

	return ret.Error(0)
}

// StartReadyCheck provides a mock function
func (_m *MockLobbyCommand) StartReadyCheck(ctx context.Context, cmd matchmaking_in.StartReadyCheckCommand) error {
	ret := _m.Called(ctx, cmd)

	return ret.Error(0)
}

// StartMatch provides a mock function
func (_m *MockLobbyCommand) StartMatch(ctx context.Context, lobbyID uuid.UUID) (uuid.UUID, error) {
	ret := _m.Called(ctx, lobbyID)

	var r0 uuid.UUID
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (uuid.UUID, error)); ok {
		return rf(ctx, lobbyID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) uuid.UUID); ok {
		r0 = rf(ctx, lobbyID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(uuid.UUID)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// CancelLobby provides a mock function
func (_m *MockLobbyCommand) CancelLobby(ctx context.Context, lobbyID uuid.UUID, reason string) error {
	ret := _m.Called(ctx, lobbyID, reason)

	return ret.Error(0)
}

// NewMockLobbyCommand creates a new instance of MockLobbyCommand
func NewMockLobbyCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockLobbyCommand {
	mock := &MockLobbyCommand{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
