package auth_in

import (
	"context"

	"github.com/google/uuid"
	auth_entities "github.com/replay-api/replay-api/pkg/domain/auth/entities"
	auth_in "github.com/replay-api/replay-api/pkg/domain/auth/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockEmailVerificationCommand is a mock implementation of EmailVerificationCommand
type MockEmailVerificationCommand struct {
	mock.Mock
}

// SendVerificationEmail provides a mock function
func (_m *MockEmailVerificationCommand) SendVerificationEmail(ctx context.Context, cmd auth_in.SendVerificationEmailCommand) (*auth_entities.EmailVerification, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *auth_entities.EmailVerification
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, auth_in.SendVerificationEmailCommand) (*auth_entities.EmailVerification, error)); ok {
		return rf(ctx, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, auth_in.SendVerificationEmailCommand) *auth_entities.EmailVerification); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*auth_entities.EmailVerification)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// VerifyEmail provides a mock function
func (_m *MockEmailVerificationCommand) VerifyEmail(ctx context.Context, cmd auth_in.VerifyEmailCommand) (*auth_in.VerificationResult, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *auth_in.VerificationResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, auth_in.VerifyEmailCommand) (*auth_in.VerificationResult, error)); ok {
		return rf(ctx, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, auth_in.VerifyEmailCommand) *auth_in.VerificationResult); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*auth_in.VerificationResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// ResendVerification provides a mock function
func (_m *MockEmailVerificationCommand) ResendVerification(ctx context.Context, cmd auth_in.ResendVerificationCommand) (*auth_entities.EmailVerification, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *auth_entities.EmailVerification
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, auth_in.ResendVerificationCommand) (*auth_entities.EmailVerification, error)); ok {
		return rf(ctx, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, auth_in.ResendVerificationCommand) *auth_entities.EmailVerification); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*auth_entities.EmailVerification)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetVerificationStatus provides a mock function
func (_m *MockEmailVerificationCommand) GetVerificationStatus(ctx context.Context, userID uuid.UUID) (*auth_entities.EmailVerification, error) {
	ret := _m.Called(ctx, userID)

	var r0 *auth_entities.EmailVerification
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*auth_entities.EmailVerification, error)); ok {
		return rf(ctx, userID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *auth_entities.EmailVerification); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*auth_entities.EmailVerification)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// CancelVerification provides a mock function
func (_m *MockEmailVerificationCommand) CancelVerification(ctx context.Context, verificationID uuid.UUID) error {
	ret := _m.Called(ctx, verificationID)

	return ret.Error(0)
}

// NewMockEmailVerificationCommand creates a new instance of MockEmailVerificationCommand
func NewMockEmailVerificationCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockEmailVerificationCommand {
	mock := &MockEmailVerificationCommand{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
