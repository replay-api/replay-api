package custody_out

import (
	"context"

	custody_out "github.com/replay-api/replay-api/pkg/domain/custody/ports/out"
	"github.com/stretchr/testify/mock"
)

// MockSigningSessionRepository is a mock implementation of SigningSessionRepository
type MockSigningSessionRepository struct {
	mock.Mock
}

// Create provides a mock function
func (_m *MockSigningSessionRepository) Create(ctx context.Context, session *custody_out.SigningSessionRecord) error {
	ret := _m.Called(ctx, session)

	return ret.Error(0)
}

// GetByID provides a mock function
func (_m *MockSigningSessionRepository) GetByID(ctx context.Context, sessionID string) (*custody_out.SigningSessionRecord, error) {
	ret := _m.Called(ctx, sessionID)

	var r0 *custody_out.SigningSessionRecord
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*custody_out.SigningSessionRecord, error)); ok {
		return rf(ctx, sessionID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) *custody_out.SigningSessionRecord); ok {
		r0 = rf(ctx, sessionID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.SigningSessionRecord)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Update provides a mock function
func (_m *MockSigningSessionRepository) Update(ctx context.Context, session *custody_out.SigningSessionRecord) error {
	ret := _m.Called(ctx, session)

	return ret.Error(0)
}

// ListActive provides a mock function
func (_m *MockSigningSessionRepository) ListActive(ctx context.Context) ([]*custody_out.SigningSessionRecord, error) {
	ret := _m.Called(ctx)

	var r0 []*custody_out.SigningSessionRecord
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context) ([]*custody_out.SigningSessionRecord, error)); ok {
		return rf(ctx)
	}

	if rf, ok := ret.Get(0).(func(context.Context) []*custody_out.SigningSessionRecord); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*custody_out.SigningSessionRecord)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// ListByKey provides a mock function
func (_m *MockSigningSessionRepository) ListByKey(ctx context.Context, keyID string) ([]*custody_out.SigningSessionRecord, error) {
	ret := _m.Called(ctx, keyID)

	var r0 []*custody_out.SigningSessionRecord
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) ([]*custody_out.SigningSessionRecord, error)); ok {
		return rf(ctx, keyID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) []*custody_out.SigningSessionRecord); ok {
		r0 = rf(ctx, keyID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*custody_out.SigningSessionRecord)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockSigningSessionRepository creates a new instance of MockSigningSessionRepository
func NewMockSigningSessionRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockSigningSessionRepository {
	mock := &MockSigningSessionRepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
