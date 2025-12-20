package matchmaking_in

import (
	"context"

	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockLobbyQuery is a mock implementation of LobbyQuery
type MockLobbyQuery struct {
	mock.Mock
}

// GetLobby provides a mock function
func (_m *MockLobbyQuery) GetLobby(ctx context.Context, query matchmaking_in.GetLobbyQuery) (*matchmaking_in.LobbyDTO, error) {
	ret := _m.Called(ctx, query)

	var r0 *matchmaking_in.LobbyDTO
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, matchmaking_in.GetLobbyQuery) (*matchmaking_in.LobbyDTO, error)); ok {
		return rf(ctx, query)
	}

	if rf, ok := ret.Get(0).(func(context.Context, matchmaking_in.GetLobbyQuery) *matchmaking_in.LobbyDTO); ok {
		r0 = rf(ctx, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*matchmaking_in.LobbyDTO)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetUserLobbies provides a mock function
func (_m *MockLobbyQuery) GetUserLobbies(ctx context.Context, query matchmaking_in.GetUserLobbiesQuery) (*matchmaking_in.LobbiesResult, error) {
	ret := _m.Called(ctx, query)

	var r0 *matchmaking_in.LobbiesResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, matchmaking_in.GetUserLobbiesQuery) (*matchmaking_in.LobbiesResult, error)); ok {
		return rf(ctx, query)
	}

	if rf, ok := ret.Get(0).(func(context.Context, matchmaking_in.GetUserLobbiesQuery) *matchmaking_in.LobbiesResult); ok {
		r0 = rf(ctx, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*matchmaking_in.LobbiesResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// SearchLobbies provides a mock function
func (_m *MockLobbyQuery) SearchLobbies(ctx context.Context, query matchmaking_in.SearchLobbiesQuery) (*matchmaking_in.LobbiesResult, error) {
	ret := _m.Called(ctx, query)

	var r0 *matchmaking_in.LobbiesResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, matchmaking_in.SearchLobbiesQuery) (*matchmaking_in.LobbiesResult, error)); ok {
		return rf(ctx, query)
	}

	if rf, ok := ret.Get(0).(func(context.Context, matchmaking_in.SearchLobbiesQuery) *matchmaking_in.LobbiesResult); ok {
		r0 = rf(ctx, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*matchmaking_in.LobbiesResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockLobbyQuery creates a new instance of MockLobbyQuery
func NewMockLobbyQuery(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockLobbyQuery {
	mock := &MockLobbyQuery{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
