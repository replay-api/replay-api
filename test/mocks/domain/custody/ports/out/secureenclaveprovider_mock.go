package custody_out

import (
	"context"

	custody_out "github.com/replay-api/replay-api/pkg/domain/custody/ports/out"
	"github.com/stretchr/testify/mock"
)

// MockSecureEnclaveProvider is a mock implementation of SecureEnclaveProvider
type MockSecureEnclaveProvider struct {
	mock.Mock
}

// ExecuteInEnclave provides a mock function
func (_m *MockSecureEnclaveProvider) ExecuteInEnclave(ctx context.Context, req *custody_out.EnclaveRequest) (*custody_out.EnclaveResponse, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_out.EnclaveResponse
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.EnclaveRequest) (*custody_out.EnclaveResponse, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.EnclaveRequest) *custody_out.EnclaveResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.EnclaveResponse)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetAttestation provides a mock function
func (_m *MockSecureEnclaveProvider) GetAttestation(ctx context.Context) (*custody_out.AttestationDocument, error) {
	ret := _m.Called(ctx)

	var r0 *custody_out.AttestationDocument
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context) (*custody_out.AttestationDocument, error)); ok {
		return rf(ctx)
	}

	if rf, ok := ret.Get(0).(func(context.Context) *custody_out.AttestationDocument); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.AttestationDocument)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// VerifyAttestation provides a mock function
func (_m *MockSecureEnclaveProvider) VerifyAttestation(ctx context.Context, doc *custody_out.AttestationDocument) error {
	ret := _m.Called(ctx, doc)

	return ret.Error(0)
}

// GenerateKeyInEnclave provides a mock function
func (_m *MockSecureEnclaveProvider) GenerateKeyInEnclave(ctx context.Context, req *custody_out.EnclaveKeyGenRequest) (*custody_out.EnclaveKeyResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_out.EnclaveKeyResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.EnclaveKeyGenRequest) (*custody_out.EnclaveKeyResult, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.EnclaveKeyGenRequest) *custody_out.EnclaveKeyResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.EnclaveKeyResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// SignInEnclave provides a mock function
func (_m *MockSecureEnclaveProvider) SignInEnclave(ctx context.Context, req *custody_out.EnclaveSignRequest) (*custody_out.EnclaveSignResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_out.EnclaveSignResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.EnclaveSignRequest) (*custody_out.EnclaveSignResult, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.EnclaveSignRequest) *custody_out.EnclaveSignResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.EnclaveSignResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// HealthCheck provides a mock function
func (_m *MockSecureEnclaveProvider) HealthCheck(ctx context.Context) (*custody_out.EnclaveHealthStatus, error) {
	ret := _m.Called(ctx)

	var r0 *custody_out.EnclaveHealthStatus
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context) (*custody_out.EnclaveHealthStatus, error)); ok {
		return rf(ctx)
	}

	if rf, ok := ret.Get(0).(func(context.Context) *custody_out.EnclaveHealthStatus); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.EnclaveHealthStatus)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockSecureEnclaveProvider creates a new instance of MockSecureEnclaveProvider
func NewMockSecureEnclaveProvider(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockSecureEnclaveProvider {
	mock := &MockSecureEnclaveProvider{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
