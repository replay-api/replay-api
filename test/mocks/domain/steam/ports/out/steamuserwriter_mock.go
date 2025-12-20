package steam_out

import (
	"context"

	steam_entity "github.com/replay-api/replay-api/pkg/domain/steam/entities"
	"github.com/stretchr/testify/mock"
)

// MockSteamUserWriter is a mock implementation of SteamUserWriter
type MockSteamUserWriter struct {
	mock.Mock
}

// Create provides a mock function
func (_m *MockSteamUserWriter) Create(ctx context.Context, user *steam_entity.SteamUser) (*steam_entity.SteamUser, error) {
	ret := _m.Called(ctx, user)

	var r0 *steam_entity.SteamUser
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *steam_entity.SteamUser) (*steam_entity.SteamUser, error)); ok {
		return rf(ctx, user)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *steam_entity.SteamUser) *steam_entity.SteamUser); ok {
		r0 = rf(ctx, user)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*steam_entity.SteamUser)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockSteamUserWriter creates a new instance of MockSteamUserWriter
func NewMockSteamUserWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockSteamUserWriter {
	mock := &MockSteamUserWriter{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
