package custody_out

import (
	"context"

	custody_out "github.com/replay-api/replay-api/pkg/domain/custody/ports/out"
	"github.com/stretchr/testify/mock"
)

// MockMPCProvider is a mock implementation of MPCProvider
type MockMPCProvider struct {
	mock.Mock
}

// InitiateKeyGeneration provides a mock function
func (_m *MockMPCProvider) InitiateKeyGeneration(ctx context.Context, req *custody_out.KeyGenRequest) (*custody_out.KeyGenSession, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_out.KeyGenSession
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.KeyGenRequest) (*custody_out.KeyGenSession, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.KeyGenRequest) *custody_out.KeyGenSession); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.KeyGenSession)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetKeyGenStatus provides a mock function
func (_m *MockMPCProvider) GetKeyGenStatus(ctx context.Context, sessionID string) (*custody_out.KeyGenSession, error) {
	ret := _m.Called(ctx, sessionID)

	var r0 *custody_out.KeyGenSession
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*custody_out.KeyGenSession, error)); ok {
		return rf(ctx, sessionID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) *custody_out.KeyGenSession); ok {
		r0 = rf(ctx, sessionID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.KeyGenSession)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FinalizeKeyGeneration provides a mock function
func (_m *MockMPCProvider) FinalizeKeyGeneration(ctx context.Context, sessionID string) (*custody_out.GeneratedKey, error) {
	ret := _m.Called(ctx, sessionID)

	var r0 *custody_out.GeneratedKey
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*custody_out.GeneratedKey, error)); ok {
		return rf(ctx, sessionID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) *custody_out.GeneratedKey); ok {
		r0 = rf(ctx, sessionID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.GeneratedKey)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// InitiateSigning provides a mock function
func (_m *MockMPCProvider) InitiateSigning(ctx context.Context, req *custody_out.SigningRequest) (*custody_out.SigningSession, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_out.SigningSession
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.SigningRequest) (*custody_out.SigningSession, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.SigningRequest) *custody_out.SigningSession); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.SigningSession)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetSigningStatus provides a mock function
func (_m *MockMPCProvider) GetSigningStatus(ctx context.Context, sessionID string) (*custody_out.SigningSession, error) {
	ret := _m.Called(ctx, sessionID)

	var r0 *custody_out.SigningSession
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*custody_out.SigningSession, error)); ok {
		return rf(ctx, sessionID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) *custody_out.SigningSession); ok {
		r0 = rf(ctx, sessionID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.SigningSession)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetSignature provides a mock function
func (_m *MockMPCProvider) GetSignature(ctx context.Context, sessionID string) (*custody_out.SignatureResult, error) {
	ret := _m.Called(ctx, sessionID)

	var r0 *custody_out.SignatureResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*custody_out.SignatureResult, error)); ok {
		return rf(ctx, sessionID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) *custody_out.SignatureResult); ok {
		r0 = rf(ctx, sessionID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.SignatureResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetPublicKey provides a mock function
func (_m *MockMPCProvider) GetPublicKey(ctx context.Context, keyID string) (*custody_out.PublicKeyInfo, error) {
	ret := _m.Called(ctx, keyID)

	var r0 *custody_out.PublicKeyInfo
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*custody_out.PublicKeyInfo, error)); ok {
		return rf(ctx, keyID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) *custody_out.PublicKeyInfo); ok {
		r0 = rf(ctx, keyID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.PublicKeyInfo)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// RefreshKeyShares provides a mock function
func (_m *MockMPCProvider) RefreshKeyShares(ctx context.Context, keyID string) error {
	ret := _m.Called(ctx, keyID)

	return ret.Error(0)
}

// RevokeKey provides a mock function
func (_m *MockMPCProvider) RevokeKey(ctx context.Context, keyID string) error {
	ret := _m.Called(ctx, keyID)

	return ret.Error(0)
}

// GetProviderInfo provides a mock function
func (_m *MockMPCProvider) GetProviderInfo() custody_out.ProviderInfo {
	ret := _m.Called()

	var r0 custody_out.ProviderInfo
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(custody_out.ProviderInfo)
	}
	return r0
}

// HealthCheck provides a mock function
func (_m *MockMPCProvider) HealthCheck(ctx context.Context) error {
	ret := _m.Called(ctx)

	return ret.Error(0)
}

// NewMockMPCProvider creates a new instance of MockMPCProvider
func NewMockMPCProvider(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMPCProvider {
	mock := &MockMPCProvider{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
