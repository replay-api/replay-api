package steam_in

import (
	"context"

	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	steam_entities "github.com/replay-api/replay-api/pkg/domain/steam/entities"
	"github.com/stretchr/testify/mock"
)

// MockOnboardSteamUserCommand is a mock implementation of OnboardSteamUserCommand
type MockOnboardSteamUserCommand struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockOnboardSteamUserCommand) Exec(ctx context.Context, steamUser *steam_entities.SteamUser) (*steam_entities.SteamUser, *iam_entities.RIDToken, error) {
	ret := _m.Called(ctx, steamUser)

	var r0 *steam_entities.SteamUser
	var r1 *iam_entities.RIDToken
	var r2 error

	if rf, ok := ret.Get(0).(func(context.Context, *steam_entities.SteamUser) (*steam_entities.SteamUser, *iam_entities.RIDToken, error)); ok {
		return rf(ctx, steamUser)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *steam_entities.SteamUser) *steam_entities.SteamUser); ok {
		r0 = rf(ctx, steamUser)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*steam_entities.SteamUser)
		}
	}
	if rf, ok := ret.Get(1).(func(context.Context, *steam_entities.SteamUser) *iam_entities.RIDToken); ok {
		r1 = rf(ctx, steamUser)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*iam_entities.RIDToken)
		}
	}
	r2 = ret.Error(2)

	return r0, r1, r2
}

// Validate provides a mock function
func (_m *MockOnboardSteamUserCommand) Validate(ctx context.Context, steamUser *steam_entities.SteamUser) error {
	ret := _m.Called(ctx, steamUser)

	return ret.Error(0)
}

// NewMockOnboardSteamUserCommand creates a new instance of MockOnboardSteamUserCommand
func NewMockOnboardSteamUserCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockOnboardSteamUserCommand {
	mock := &MockOnboardSteamUserCommand{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
