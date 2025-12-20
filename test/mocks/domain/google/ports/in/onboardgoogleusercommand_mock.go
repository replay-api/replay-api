package google_in

import (
	"context"

	google_entities "github.com/replay-api/replay-api/pkg/domain/google/entities"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	"github.com/stretchr/testify/mock"
)

// MockOnboardGoogleUserCommand is a mock implementation of OnboardGoogleUserCommand
type MockOnboardGoogleUserCommand struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockOnboardGoogleUserCommand) Exec(ctx context.Context, googleUser *google_entities.GoogleUser) (*google_entities.GoogleUser, *iam_entities.RIDToken, error) {
	ret := _m.Called(ctx, googleUser)

	var r0 *google_entities.GoogleUser
	var r1 *iam_entities.RIDToken
	var r2 error

	if rf, ok := ret.Get(0).(func(context.Context, *google_entities.GoogleUser) (*google_entities.GoogleUser, *iam_entities.RIDToken, error)); ok {
		return rf(ctx, googleUser)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *google_entities.GoogleUser) *google_entities.GoogleUser); ok {
		r0 = rf(ctx, googleUser)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*google_entities.GoogleUser)
		}
	}
	if rf, ok := ret.Get(1).(func(context.Context, *google_entities.GoogleUser) *iam_entities.RIDToken); ok {
		r1 = rf(ctx, googleUser)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*iam_entities.RIDToken)
		}
	}
	r2 = ret.Error(2)

	return r0, r1, r2
}

// Validate provides a mock function
func (_m *MockOnboardGoogleUserCommand) Validate(ctx context.Context, googleUser *google_entities.GoogleUser) error {
	ret := _m.Called(ctx, googleUser)

	return ret.Error(0)
}

// NewMockOnboardGoogleUserCommand creates a new instance of MockOnboardGoogleUserCommand
func NewMockOnboardGoogleUserCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockOnboardGoogleUserCommand {
	mock := &MockOnboardGoogleUserCommand{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
