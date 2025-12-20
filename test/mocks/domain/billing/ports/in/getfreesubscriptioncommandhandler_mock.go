package billing_in

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockGetFreeSubscriptionCommandHandler is a mock implementation of GetFreeSubscriptionCommandHandler
type MockGetFreeSubscriptionCommandHandler struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockGetFreeSubscriptionCommandHandler) Exec(ctx context.Context) (uuid.UUID, error) {
	ret := _m.Called(ctx)

	var r0 uuid.UUID
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context) (uuid.UUID, error)); ok {
		return rf(ctx)
	}

	if rf, ok := ret.Get(0).(func(context.Context) uuid.UUID); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(uuid.UUID)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockGetFreeSubscriptionCommandHandler creates a new instance of MockGetFreeSubscriptionCommandHandler
func NewMockGetFreeSubscriptionCommandHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockGetFreeSubscriptionCommandHandler {
	mock := &MockGetFreeSubscriptionCommandHandler{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
