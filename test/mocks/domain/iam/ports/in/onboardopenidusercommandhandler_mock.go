package iam_in

import (
	"context"

	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	iam_in "github.com/replay-api/replay-api/pkg/domain/iam/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockOnboardOpenIDUserCommandHandler is a mock implementation of OnboardOpenIDUserCommandHandler
type MockOnboardOpenIDUserCommandHandler struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockOnboardOpenIDUserCommandHandler) Exec(ctx context.Context, cmd iam_in.OnboardOpenIDUserCommand) (*iam_entities.Profile, *iam_entities.RIDToken, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *iam_entities.Profile
	var r1 *iam_entities.RIDToken
	var r2 error

	if rf, ok := ret.Get(0).(func(context.Context, iam_in.OnboardOpenIDUserCommand) (*iam_entities.Profile, *iam_entities.RIDToken, error)); ok {
		return rf(ctx, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, iam_in.OnboardOpenIDUserCommand) *iam_entities.Profile); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*iam_entities.Profile)
		}
	}
	if rf, ok := ret.Get(1).(func(context.Context, iam_in.OnboardOpenIDUserCommand) *iam_entities.RIDToken); ok {
		r1 = rf(ctx, cmd)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*iam_entities.RIDToken)
		}
	}
	r2 = ret.Error(2)

	return r0, r1, r2
}

// NewMockOnboardOpenIDUserCommandHandler creates a new instance of MockOnboardOpenIDUserCommandHandler
func NewMockOnboardOpenIDUserCommandHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockOnboardOpenIDUserCommandHandler {
	mock := &MockOnboardOpenIDUserCommandHandler{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
