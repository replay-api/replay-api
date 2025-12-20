package custody_out

import (
	"context"

	custody_out "github.com/replay-api/replay-api/pkg/domain/custody/ports/out"
	"github.com/stretchr/testify/mock"
)

// MockHSMProvider is a mock implementation of HSMProvider
type MockHSMProvider struct {
	mock.Mock
}

// StoreKeyShare provides a mock function
func (_m *MockHSMProvider) StoreKeyShare(ctx context.Context, req *custody_out.StoreKeyShareRequest) (*custody_out.StoredKeyShare, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_out.StoredKeyShare
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.StoreKeyShareRequest) (*custody_out.StoredKeyShare, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.StoreKeyShareRequest) *custody_out.StoredKeyShare); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.StoredKeyShare)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// RetrieveKeyShare provides a mock function
func (_m *MockHSMProvider) RetrieveKeyShare(ctx context.Context, keyID string, shareIndex uint8) (*custody_out.RetrievedKeyShare, error) {
	ret := _m.Called(ctx, keyID, shareIndex)

	var r0 *custody_out.RetrievedKeyShare
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string, uint8) (*custody_out.RetrievedKeyShare, error)); ok {
		return rf(ctx, keyID, shareIndex)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string, uint8) *custody_out.RetrievedKeyShare); ok {
		r0 = rf(ctx, keyID, shareIndex)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.RetrievedKeyShare)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// DeleteKeyShare provides a mock function
func (_m *MockHSMProvider) DeleteKeyShare(ctx context.Context, keyID string, shareIndex uint8) error {
	ret := _m.Called(ctx, keyID, shareIndex)

	return ret.Error(0)
}

// SignInsideHSM provides a mock function
func (_m *MockHSMProvider) SignInsideHSM(ctx context.Context, req *custody_out.HSMSignRequest) (*custody_out.HSMSignResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_out.HSMSignResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.HSMSignRequest) (*custody_out.HSMSignResult, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.HSMSignRequest) *custody_out.HSMSignResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.HSMSignResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// DeriveChildKey provides a mock function
func (_m *MockHSMProvider) DeriveChildKey(ctx context.Context, req *custody_out.DeriveKeyRequest) (*custody_out.DerivedKey, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_out.DerivedKey
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.DeriveKeyRequest) (*custody_out.DerivedKey, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.DeriveKeyRequest) *custody_out.DerivedKey); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.DerivedKey)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// ListKeys provides a mock function
func (_m *MockHSMProvider) ListKeys(ctx context.Context, filter *custody_out.KeyFilter) ([]*custody_out.HSMKeyInfo, error) {
	ret := _m.Called(ctx, filter)

	var r0 []*custody_out.HSMKeyInfo
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.KeyFilter) ([]*custody_out.HSMKeyInfo, error)); ok {
		return rf(ctx, filter)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.KeyFilter) []*custody_out.HSMKeyInfo); ok {
		r0 = rf(ctx, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*custody_out.HSMKeyInfo)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetKeyInfo provides a mock function
func (_m *MockHSMProvider) GetKeyInfo(ctx context.Context, keyID string) (*custody_out.HSMKeyInfo, error) {
	ret := _m.Called(ctx, keyID)

	var r0 *custody_out.HSMKeyInfo
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*custody_out.HSMKeyInfo, error)); ok {
		return rf(ctx, keyID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) *custody_out.HSMKeyInfo); ok {
		r0 = rf(ctx, keyID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.HSMKeyInfo)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// RotateWrappingKey provides a mock function
func (_m *MockHSMProvider) RotateWrappingKey(ctx context.Context) error {
	ret := _m.Called(ctx)

	return ret.Error(0)
}

// GetAuditLogs provides a mock function
func (_m *MockHSMProvider) GetAuditLogs(ctx context.Context, filter *custody_out.AuditFilter) ([]*custody_out.HSMAuditLog, error) {
	ret := _m.Called(ctx, filter)

	var r0 []*custody_out.HSMAuditLog
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.AuditFilter) ([]*custody_out.HSMAuditLog, error)); ok {
		return rf(ctx, filter)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.AuditFilter) []*custody_out.HSMAuditLog); ok {
		r0 = rf(ctx, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*custody_out.HSMAuditLog)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// HealthCheck provides a mock function
func (_m *MockHSMProvider) HealthCheck(ctx context.Context) (*custody_out.HSMHealthStatus, error) {
	ret := _m.Called(ctx)

	var r0 *custody_out.HSMHealthStatus
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context) (*custody_out.HSMHealthStatus, error)); ok {
		return rf(ctx)
	}

	if rf, ok := ret.Get(0).(func(context.Context) *custody_out.HSMHealthStatus); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.HSMHealthStatus)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetProviderInfo provides a mock function
func (_m *MockHSMProvider) GetProviderInfo() custody_out.HSMProviderInfo {
	ret := _m.Called()

	var r0 custody_out.HSMProviderInfo
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(custody_out.HSMProviderInfo)
	}
	return r0
}

// NewMockHSMProvider creates a new instance of MockHSMProvider
func NewMockHSMProvider(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockHSMProvider {
	mock := &MockHSMProvider{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
