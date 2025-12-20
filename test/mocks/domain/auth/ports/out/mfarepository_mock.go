package auth_out

import (
	"context"

	"github.com/google/uuid"
	auth_entities "github.com/replay-api/replay-api/pkg/domain/auth/entities"
	"github.com/stretchr/testify/mock"
)

// MockMFARepository is a mock implementation of MFARepository
type MockMFARepository struct {
	mock.Mock
}

// Create provides a mock function
func (_m *MockMFARepository) Create(ctx context.Context, mfa *auth_entities.UserMFA) (*auth_entities.UserMFA, error) {
	ret := _m.Called(ctx, mfa)

	var r0 *auth_entities.UserMFA
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *auth_entities.UserMFA) (*auth_entities.UserMFA, error)); ok {
		return rf(ctx, mfa)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *auth_entities.UserMFA) *auth_entities.UserMFA); ok {
		r0 = rf(ctx, mfa)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*auth_entities.UserMFA)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetByUserID provides a mock function
func (_m *MockMFARepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*auth_entities.UserMFA, error) {
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

// Update provides a mock function
func (_m *MockMFARepository) Update(ctx context.Context, mfa *auth_entities.UserMFA) (*auth_entities.UserMFA, error) {
	ret := _m.Called(ctx, mfa)

	var r0 *auth_entities.UserMFA
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *auth_entities.UserMFA) (*auth_entities.UserMFA, error)); ok {
		return rf(ctx, mfa)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *auth_entities.UserMFA) *auth_entities.UserMFA); ok {
		r0 = rf(ctx, mfa)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*auth_entities.UserMFA)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Delete provides a mock function
func (_m *MockMFARepository) Delete(ctx context.Context, id uuid.UUID) error {
	ret := _m.Called(ctx, id)

	return ret.Error(0)
}

// NewMockMFARepository creates a new instance of MockMFARepository
func NewMockMFARepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMFARepository {
	mock := &MockMFARepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
