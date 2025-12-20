package payment_out

import (
	"context"

	payment_entities "github.com/replay-api/replay-api/pkg/domain/payment/entities"
	payment_out "github.com/replay-api/replay-api/pkg/domain/payment/ports/out"
	"github.com/stretchr/testify/mock"
)

// MockPaymentProviderAdapter is a mock implementation of PaymentProviderAdapter
type MockPaymentProviderAdapter struct {
	mock.Mock
}

// GetProvider provides a mock function
func (_m *MockPaymentProviderAdapter) GetProvider() payment_entities.PaymentProvider {
	ret := _m.Called()

	var r0 payment_entities.PaymentProvider
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(payment_entities.PaymentProvider)
	}
	return r0
}

// CreatePaymentIntent provides a mock function
func (_m *MockPaymentProviderAdapter) CreatePaymentIntent(ctx context.Context, req payment_out.CreateIntentRequest) (*payment_out.CreateIntentResponse, error) {
	ret := _m.Called(ctx, req)

	var r0 *payment_out.CreateIntentResponse
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, payment_out.CreateIntentRequest) (*payment_out.CreateIntentResponse, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, payment_out.CreateIntentRequest) *payment_out.CreateIntentResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*payment_out.CreateIntentResponse)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// ConfirmPayment provides a mock function
func (_m *MockPaymentProviderAdapter) ConfirmPayment(ctx context.Context, req payment_out.ConfirmPaymentRequest) (*payment_out.ConfirmPaymentResponse, error) {
	ret := _m.Called(ctx, req)

	var r0 *payment_out.ConfirmPaymentResponse
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, payment_out.ConfirmPaymentRequest) (*payment_out.ConfirmPaymentResponse, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, payment_out.ConfirmPaymentRequest) *payment_out.ConfirmPaymentResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*payment_out.ConfirmPaymentResponse)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// RefundPayment provides a mock function
func (_m *MockPaymentProviderAdapter) RefundPayment(ctx context.Context, req payment_out.RefundRequest) (*payment_out.RefundResponse, error) {
	ret := _m.Called(ctx, req)

	var r0 *payment_out.RefundResponse
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, payment_out.RefundRequest) (*payment_out.RefundResponse, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, payment_out.RefundRequest) *payment_out.RefundResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*payment_out.RefundResponse)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// CancelPayment provides a mock function
func (_m *MockPaymentProviderAdapter) CancelPayment(ctx context.Context, req payment_out.CancelRequest) (*payment_out.CancelResponse, error) {
	ret := _m.Called(ctx, req)

	var r0 *payment_out.CancelResponse
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, payment_out.CancelRequest) (*payment_out.CancelResponse, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, payment_out.CancelRequest) *payment_out.CancelResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*payment_out.CancelResponse)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// ParseWebhook provides a mock function
func (_m *MockPaymentProviderAdapter) ParseWebhook(payload []byte, signature string) (*payment_out.WebhookEvent, error) {
	ret := _m.Called(payload, signature)

	var r0 *payment_out.WebhookEvent
	var r1 error

	if rf, ok := ret.Get(0).(func([]byte, string) (*payment_out.WebhookEvent, error)); ok {
		return rf(payload, signature)
	}

	if rf, ok := ret.Get(0).(func([]byte, string) *payment_out.WebhookEvent); ok {
		r0 = rf(payload, signature)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*payment_out.WebhookEvent)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// CreateOrGetCustomer provides a mock function
func (_m *MockPaymentProviderAdapter) CreateOrGetCustomer(ctx context.Context, req payment_out.CustomerRequest) (*payment_out.CustomerResponse, error) {
	ret := _m.Called(ctx, req)

	var r0 *payment_out.CustomerResponse
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, payment_out.CustomerRequest) (*payment_out.CustomerResponse, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, payment_out.CustomerRequest) *payment_out.CustomerResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*payment_out.CustomerResponse)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockPaymentProviderAdapter creates a new instance of MockPaymentProviderAdapter
func NewMockPaymentProviderAdapter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPaymentProviderAdapter {
	mock := &MockPaymentProviderAdapter{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
