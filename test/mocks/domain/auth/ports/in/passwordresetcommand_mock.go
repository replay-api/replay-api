package auth_in

import (
	"context"

	"github.com/google/uuid"
	auth_entities "github.com/replay-api/replay-api/pkg/domain/auth/entities"
	auth_in "github.com/replay-api/replay-api/pkg/domain/auth/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockPasswordResetCommand is a mock implementation of PasswordResetCommand
type MockPasswordResetCommand struct {
	mock.Mock
}

// RequestPasswordReset provides a mock function
func (_m *MockPasswordResetCommand) RequestPasswordReset(ctx context.Context, cmd auth_in.RequestPasswordResetCommand) (*auth_entities.PasswordReset, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *auth_entities.PasswordReset
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, auth_in.RequestPasswordResetCommand) (*auth_entities.PasswordReset, error)); ok {
		return rf(ctx, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, auth_in.RequestPasswordResetCommand) *auth_entities.PasswordReset); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*auth_entities.PasswordReset)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// ConfirmPasswordReset provides a mock function
func (_m *MockPasswordResetCommand) ConfirmPasswordReset(ctx context.Context, cmd auth_in.ConfirmPasswordResetCommand) (*auth_in.PasswordResetResult, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *auth_in.PasswordResetResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, auth_in.ConfirmPasswordResetCommand) (*auth_in.PasswordResetResult, error)); ok {
		return rf(ctx, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, auth_in.ConfirmPasswordResetCommand) *auth_in.PasswordResetResult); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*auth_in.PasswordResetResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// ValidateResetToken provides a mock function
func (_m *MockPasswordResetCommand) ValidateResetToken(ctx context.Context, token string) (*auth_entities.PasswordReset, error) {
	ret := _m.Called(ctx, token)

	var r0 *auth_entities.PasswordReset
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*auth_entities.PasswordReset, error)); ok {
		return rf(ctx, token)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) *auth_entities.PasswordReset); ok {
		r0 = rf(ctx, token)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*auth_entities.PasswordReset)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetResetStatus provides a mock function
func (_m *MockPasswordResetCommand) GetResetStatus(ctx context.Context, userID uuid.UUID) (*auth_entities.PasswordReset, error) {
	ret := _m.Called(ctx, userID)

	var r0 *auth_entities.PasswordReset
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*auth_entities.PasswordReset, error)); ok {
		return rf(ctx, userID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *auth_entities.PasswordReset); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*auth_entities.PasswordReset)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// CancelPasswordReset provides a mock function
func (_m *MockPasswordResetCommand) CancelPasswordReset(ctx context.Context, resetID uuid.UUID) error {
	ret := _m.Called(ctx, resetID)

	return ret.Error(0)
}

// NewMockPasswordResetCommand creates a new instance of MockPasswordResetCommand
func NewMockPasswordResetCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPasswordResetCommand {
	mock := &MockPasswordResetCommand{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
