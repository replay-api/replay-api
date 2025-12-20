package auth_in

import (
	"context"

	"github.com/google/uuid"
	auth_entities "github.com/replay-api/replay-api/pkg/domain/auth/entities"
	"github.com/stretchr/testify/mock"
)

// MockMFACommand is a mock implementation of MFACommand
type MockMFACommand struct {
	mock.Mock
}

// SetupTOTP provides a mock function
func (_m *MockMFACommand) SetupTOTP(ctx context.Context, userID uuid.UUID, email string) (*auth_entities.MFASetupResponse, error) {
	ret := _m.Called(ctx, userID, email)

	var r0 *auth_entities.MFASetupResponse
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string) (*auth_entities.MFASetupResponse, error)); ok {
		return rf(ctx, userID, email)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string) *auth_entities.MFASetupResponse); ok {
		r0 = rf(ctx, userID, email)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*auth_entities.MFASetupResponse)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// VerifyAndActivate provides a mock function
func (_m *MockMFACommand) VerifyAndActivate(ctx context.Context, userID uuid.UUID, code string) error {
	ret := _m.Called(ctx, userID, code)

	return ret.Error(0)
}

// Verify provides a mock function
func (_m *MockMFACommand) Verify(ctx context.Context, userID uuid.UUID, code string) error {
	ret := _m.Called(ctx, userID, code)

	return ret.Error(0)
}

// VerifyBackupCode provides a mock function
func (_m *MockMFACommand) VerifyBackupCode(ctx context.Context, userID uuid.UUID, code string) error {
	ret := _m.Called(ctx, userID, code)

	return ret.Error(0)
}

// Disable provides a mock function
func (_m *MockMFACommand) Disable(ctx context.Context, userID uuid.UUID) error {
	ret := _m.Called(ctx, userID)

	return ret.Error(0)
}

// GetStatus provides a mock function
func (_m *MockMFACommand) GetStatus(ctx context.Context, userID uuid.UUID) (*auth_entities.UserMFA, error) {
	ret := _m.Called(ctx, userID)

	var r0 *auth_entities.UserMFA
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*auth_entities.UserMFA, error)); ok {
		return rf(ctx, userID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *auth_entities.UserMFA); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*auth_entities.UserMFA)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// RegenerateBackupCodes provides a mock function
func (_m *MockMFACommand) RegenerateBackupCodes(ctx context.Context, userID uuid.UUID, code string) ([]string, error) {
	ret := _m.Called(ctx, userID, code)

	var r0 []string
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string) ([]string, error)); ok {
		return rf(ctx, userID, code)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string) []string); ok {
		r0 = rf(ctx, userID, code)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockMFACommand creates a new instance of MockMFACommand
func NewMockMFACommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMFACommand {
	mock := &MockMFACommand{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
