package custody_out

import (
	"context"

	"github.com/google/uuid"
	custody_out "github.com/replay-api/replay-api/pkg/domain/custody/ports/out"
	"github.com/stretchr/testify/mock"
)

// MockKeyRepository is a mock implementation of KeyRepository
type MockKeyRepository struct {
	mock.Mock
}

// Create provides a mock function
func (_m *MockKeyRepository) Create(ctx context.Context, key *custody_out.KeyRecord) error {
	ret := _m.Called(ctx, key)

	return ret.Error(0)
}

// GetByID provides a mock function
func (_m *MockKeyRepository) GetByID(ctx context.Context, keyID string) (*custody_out.KeyRecord, error) {
	ret := _m.Called(ctx, keyID)

	var r0 *custody_out.KeyRecord
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*custody_out.KeyRecord, error)); ok {
		return rf(ctx, keyID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) *custody_out.KeyRecord); ok {
		r0 = rf(ctx, keyID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.KeyRecord)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetByWallet provides a mock function
func (_m *MockKeyRepository) GetByWallet(ctx context.Context, walletID uuid.UUID) ([]*custody_out.KeyRecord, error) {
	ret := _m.Called(ctx, walletID)

	var r0 []*custody_out.KeyRecord
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) ([]*custody_out.KeyRecord, error)); ok {
		return rf(ctx, walletID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) []*custody_out.KeyRecord); ok {
		r0 = rf(ctx, walletID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*custody_out.KeyRecord)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Update provides a mock function
func (_m *MockKeyRepository) Update(ctx context.Context, key *custody_out.KeyRecord) error {
	ret := _m.Called(ctx, key)

	return ret.Error(0)
}

// Deactivate provides a mock function
func (_m *MockKeyRepository) Deactivate(ctx context.Context, keyID string) error {
	ret := _m.Called(ctx, keyID)

	return ret.Error(0)
}

// ListActive provides a mock function
func (_m *MockKeyRepository) ListActive(ctx context.Context) ([]*custody_out.KeyRecord, error) {
	ret := _m.Called(ctx)

	var r0 []*custody_out.KeyRecord
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context) ([]*custody_out.KeyRecord, error)); ok {
		return rf(ctx)
	}

	if rf, ok := ret.Get(0).(func(context.Context) []*custody_out.KeyRecord); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*custody_out.KeyRecord)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockKeyRepository creates a new instance of MockKeyRepository
func NewMockKeyRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockKeyRepository {
	mock := &MockKeyRepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
