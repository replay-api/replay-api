package wallet_out

import (
	"context"

	time "time"

	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	"github.com/stretchr/testify/mock"
)

// MockIdempotencyRepository is a mock implementation of IdempotencyRepository
type MockIdempotencyRepository struct {
	mock.Mock
}

// Create provides a mock function
func (_m *MockIdempotencyRepository) Create(ctx context.Context, op *wallet_entities.IdempotentOperation) error {
	ret := _m.Called(ctx, op)

	return ret.Error(0)
}

// FindByKey provides a mock function
func (_m *MockIdempotencyRepository) FindByKey(ctx context.Context, key string) (*wallet_entities.IdempotentOperation, error) {
	ret := _m.Called(ctx, key)

	var r0 *wallet_entities.IdempotentOperation
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*wallet_entities.IdempotentOperation, error)); ok {
		return rf(ctx, key)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) *wallet_entities.IdempotentOperation); ok {
		r0 = rf(ctx, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*wallet_entities.IdempotentOperation)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Update provides a mock function
func (_m *MockIdempotencyRepository) Update(ctx context.Context, op *wallet_entities.IdempotentOperation) error {
	ret := _m.Called(ctx, op)

	return ret.Error(0)
}

// Delete provides a mock function
func (_m *MockIdempotencyRepository) Delete(ctx context.Context, key string) error {
	ret := _m.Called(ctx, key)

	return ret.Error(0)
}

// FindStaleOperations provides a mock function
func (_m *MockIdempotencyRepository) FindStaleOperations(ctx context.Context, threshold time.Duration) ([]*wallet_entities.IdempotentOperation, error) {
	ret := _m.Called(ctx, threshold)

	var r0 []*wallet_entities.IdempotentOperation
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, time.Duration) ([]*wallet_entities.IdempotentOperation, error)); ok {
		return rf(ctx, threshold)
	}

	if rf, ok := ret.Get(0).(func(context.Context, time.Duration) []*wallet_entities.IdempotentOperation); ok {
		r0 = rf(ctx, threshold)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*wallet_entities.IdempotentOperation)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// CleanupExpired provides a mock function
func (_m *MockIdempotencyRepository) CleanupExpired(ctx context.Context) (int64, error) {
	ret := _m.Called(ctx)

	var r0 int64
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context) (int64, error)); ok {
		return rf(ctx)
	}

	if rf, ok := ret.Get(0).(func(context.Context) int64); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(int64)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockIdempotencyRepository creates a new instance of MockIdempotencyRepository
func NewMockIdempotencyRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockIdempotencyRepository {
	mock := &MockIdempotencyRepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
