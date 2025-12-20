package auth_out

import (
	"context"

	"github.com/google/uuid"
	auth_entities "github.com/replay-api/replay-api/pkg/domain/auth/entities"
	"github.com/stretchr/testify/mock"
)

// MockPasswordResetRepository is a mock implementation of PasswordResetRepository
type MockPasswordResetRepository struct {
	mock.Mock
}

// Save provides a mock function
func (_m *MockPasswordResetRepository) Save(ctx context.Context, reset *auth_entities.PasswordReset) error {
	ret := _m.Called(ctx, reset)

	return ret.Error(0)
}

// FindByID provides a mock function
func (_m *MockPasswordResetRepository) FindByID(ctx context.Context, id uuid.UUID) (*auth_entities.PasswordReset, error) {
	ret := _m.Called(ctx, id)

	var r0 *auth_entities.PasswordReset
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*auth_entities.PasswordReset, error)); ok {
		return rf(ctx, id)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *auth_entities.PasswordReset); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*auth_entities.PasswordReset)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindByToken provides a mock function
func (_m *MockPasswordResetRepository) FindByToken(ctx context.Context, token string) (*auth_entities.PasswordReset, error) {
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

// FindByUserID provides a mock function
func (_m *MockPasswordResetRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*auth_entities.PasswordReset, error) {
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

// FindPendingByEmail provides a mock function
func (_m *MockPasswordResetRepository) FindPendingByEmail(ctx context.Context, email string) (*auth_entities.PasswordReset, error) {
	ret := _m.Called(ctx, email)

	var r0 *auth_entities.PasswordReset
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*auth_entities.PasswordReset, error)); ok {
		return rf(ctx, email)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) *auth_entities.PasswordReset); ok {
		r0 = rf(ctx, email)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*auth_entities.PasswordReset)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Update provides a mock function
func (_m *MockPasswordResetRepository) Update(ctx context.Context, reset *auth_entities.PasswordReset) error {
	ret := _m.Called(ctx, reset)

	return ret.Error(0)
}

// InvalidatePreviousResets provides a mock function
func (_m *MockPasswordResetRepository) InvalidatePreviousResets(ctx context.Context, userID uuid.UUID, email string) error {
	ret := _m.Called(ctx, userID, email)

	return ret.Error(0)
}

// CountRecentAttempts provides a mock function
func (_m *MockPasswordResetRepository) CountRecentAttempts(ctx context.Context, email string, minutes int) (int, error) {
	ret := _m.Called(ctx, email, minutes)

	var r0 int
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string, int) (int, error)); ok {
		return rf(ctx, email, minutes)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string, int) int); ok {
		r0 = rf(ctx, email, minutes)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(int)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockPasswordResetRepository creates a new instance of MockPasswordResetRepository
func NewMockPasswordResetRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPasswordResetRepository {
	mock := &MockPasswordResetRepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
