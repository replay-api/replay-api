package email_in

import (
	"context"

	email_entities "github.com/replay-api/replay-api/pkg/domain/email/entities"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	"github.com/stretchr/testify/mock"
)

// MockLoginEmailUserCommand is a mock implementation of LoginEmailUserCommand
type MockLoginEmailUserCommand struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockLoginEmailUserCommand) Exec(ctx context.Context, email string, password string, vHash string) (*email_entities.EmailUser, *iam_entities.RIDToken, error) {
	ret := _m.Called(ctx, email, password, vHash)

	var r0 *email_entities.EmailUser
	var r1 *iam_entities.RIDToken
	var r2 error

	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) (*email_entities.EmailUser, *iam_entities.RIDToken, error)); ok {
		return rf(ctx, email, password, vHash)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) *email_entities.EmailUser); ok {
		r0 = rf(ctx, email, password, vHash)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*email_entities.EmailUser)
		}
	}
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) *iam_entities.RIDToken); ok {
		r1 = rf(ctx, email, password, vHash)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*iam_entities.RIDToken)
		}
	}
	r2 = ret.Error(2)

	return r0, r1, r2
}

// Validate provides a mock function
func (_m *MockLoginEmailUserCommand) Validate(ctx context.Context, email string, password string, vHash string) error {
	ret := _m.Called(ctx, email, password, vHash)

	return ret.Error(0)
}

// NewMockLoginEmailUserCommand creates a new instance of MockLoginEmailUserCommand
func NewMockLoginEmailUserCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockLoginEmailUserCommand {
	mock := &MockLoginEmailUserCommand{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
