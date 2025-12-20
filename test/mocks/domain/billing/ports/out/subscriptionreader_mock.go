package billing_out

import (
	"context"

	common "github.com/replay-api/replay-api/pkg/domain"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	"github.com/stretchr/testify/mock"
)

// MockSubscriptionReader is a mock implementation of SubscriptionReader
type MockSubscriptionReader struct {
	mock.Mock
}

// GetCurrentSubscription provides a mock function
func (_m *MockSubscriptionReader) GetCurrentSubscription(ctx context.Context, rxn common.ResourceOwner) (*billing_entities.Subscription, error) {
	ret := _m.Called(ctx, rxn)

	var r0 *billing_entities.Subscription
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, common.ResourceOwner) (*billing_entities.Subscription, error)); ok {
		return rf(ctx, rxn)
	}

	if rf, ok := ret.Get(0).(func(context.Context, common.ResourceOwner) *billing_entities.Subscription); ok {
		r0 = rf(ctx, rxn)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing_entities.Subscription)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockSubscriptionReader creates a new instance of MockSubscriptionReader
func NewMockSubscriptionReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockSubscriptionReader {
	mock := &MockSubscriptionReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
