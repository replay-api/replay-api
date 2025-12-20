package billing_out

import (
	"context"

	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	"github.com/stretchr/testify/mock"
)

// MockSubscriptionWriter is a mock implementation of SubscriptionWriter
type MockSubscriptionWriter struct {
	mock.Mock
}

// Create provides a mock function
func (_m *MockSubscriptionWriter) Create(ctx context.Context, subscription *billing_entities.Subscription) (*billing_entities.Subscription, error) {
	ret := _m.Called(ctx, subscription)

	var r0 *billing_entities.Subscription
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *billing_entities.Subscription) (*billing_entities.Subscription, error)); ok {
		return rf(ctx, subscription)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *billing_entities.Subscription) *billing_entities.Subscription); ok {
		r0 = rf(ctx, subscription)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing_entities.Subscription)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Update provides a mock function
func (_m *MockSubscriptionWriter) Update(ctx context.Context, subscription *billing_entities.Subscription) (*billing_entities.Subscription, error) {
	ret := _m.Called(ctx, subscription)

	var r0 *billing_entities.Subscription
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *billing_entities.Subscription) (*billing_entities.Subscription, error)); ok {
		return rf(ctx, subscription)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *billing_entities.Subscription) *billing_entities.Subscription); ok {
		r0 = rf(ctx, subscription)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing_entities.Subscription)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Cancel provides a mock function
func (_m *MockSubscriptionWriter) Cancel(ctx context.Context, subscription *billing_entities.Subscription) (*billing_entities.Subscription, error) {
	ret := _m.Called(ctx, subscription)

	var r0 *billing_entities.Subscription
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *billing_entities.Subscription) (*billing_entities.Subscription, error)); ok {
		return rf(ctx, subscription)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *billing_entities.Subscription) *billing_entities.Subscription); ok {
		r0 = rf(ctx, subscription)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing_entities.Subscription)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockSubscriptionWriter creates a new instance of MockSubscriptionWriter
func NewMockSubscriptionWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockSubscriptionWriter {
	mock := &MockSubscriptionWriter{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
