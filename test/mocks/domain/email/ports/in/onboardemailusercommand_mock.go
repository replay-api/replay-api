package email_in

import (
	"context"

	email_entities "github.com/replay-api/replay-api/pkg/domain/email/entities"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	"github.com/stretchr/testify/mock"
)

// MockOnboardEmailUserCommand is a mock implementation of OnboardEmailUserCommand
type MockOnboardEmailUserCommand struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockOnboardEmailUserCommand) Exec(ctx context.Context, emailUser *email_entities.EmailUser, password string) (*email_entities.EmailUser, *iam_entities.RIDToken, error) {
	ret := _m.Called(ctx, emailUser, password)

	var r0 *email_entities.EmailUser
	var r1 *iam_entities.RIDToken
	var r2 error

	if rf, ok := ret.Get(0).(func(context.Context, *email_entities.EmailUser, string) (*email_entities.EmailUser, *iam_entities.RIDToken, error)); ok {
		return rf(ctx, emailUser, password)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *email_entities.EmailUser, string) *email_entities.EmailUser); ok {
		r0 = rf(ctx, emailUser, password)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*email_entities.EmailUser)
		}
	}
	if rf, ok := ret.Get(1).(func(context.Context, *email_entities.EmailUser, string) *iam_entities.RIDToken); ok {
		r1 = rf(ctx, emailUser, password)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*iam_entities.RIDToken)
		}
	}
	r2 = ret.Error(2)

	return r0, r1, r2
}

// Validate provides a mock function
func (_m *MockOnboardEmailUserCommand) Validate(ctx context.Context, emailUser *email_entities.EmailUser, password string) error {
	ret := _m.Called(ctx, emailUser, password)

	return ret.Error(0)
}

// NewMockOnboardEmailUserCommand creates a new instance of MockOnboardEmailUserCommand
func NewMockOnboardEmailUserCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockOnboardEmailUserCommand {
	mock := &MockOnboardEmailUserCommand{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
