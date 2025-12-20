package custody_in

import (
	"context"

	in_custody "github.com/replay-api/replay-api/pkg/domain/custody/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockKeyGenerationService is a mock implementation of KeyGenerationService
type MockKeyGenerationService struct {
	mock.Mock
}

// GenerateKey provides a mock function
func (_m *MockKeyGenerationService) GenerateKey(ctx context.Context, req *in_custody.KeyGenRequest) (*in_custody.KeyGenResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *in_custody.KeyGenResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *in_custody.KeyGenRequest) (*in_custody.KeyGenResult, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *in_custody.KeyGenRequest) *in_custody.KeyGenResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*in_custody.KeyGenResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetKeyGenStatus provides a mock function
func (_m *MockKeyGenerationService) GetKeyGenStatus(ctx context.Context, sessionID string) (*in_custody.KeyGenStatus, error) {
	ret := _m.Called(ctx, sessionID)

	var r0 *in_custody.KeyGenStatus
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*in_custody.KeyGenStatus, error)); ok {
		return rf(ctx, sessionID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) *in_custody.KeyGenStatus); ok {
		r0 = rf(ctx, sessionID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*in_custody.KeyGenStatus)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetKeyInfo provides a mock function
func (_m *MockKeyGenerationService) GetKeyInfo(ctx context.Context, keyID string) (*in_custody.KeyInfo, error) {
	ret := _m.Called(ctx, keyID)

	var r0 *in_custody.KeyInfo
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*in_custody.KeyInfo, error)); ok {
		return rf(ctx, keyID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) *in_custody.KeyInfo); ok {
		r0 = rf(ctx, keyID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*in_custody.KeyInfo)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// RotateKey provides a mock function
func (_m *MockKeyGenerationService) RotateKey(ctx context.Context, keyID string) (*in_custody.KeyRotationResult, error) {
	ret := _m.Called(ctx, keyID)

	var r0 *in_custody.KeyRotationResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*in_custody.KeyRotationResult, error)); ok {
		return rf(ctx, keyID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) *in_custody.KeyRotationResult); ok {
		r0 = rf(ctx, keyID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*in_custody.KeyRotationResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// RevokeKey provides a mock function
func (_m *MockKeyGenerationService) RevokeKey(ctx context.Context, keyID string) error {
	ret := _m.Called(ctx, keyID)

	return ret.Error(0)
}

// DeriveChildKey provides a mock function
func (_m *MockKeyGenerationService) DeriveChildKey(ctx context.Context, req *in_custody.DeriveKeyRequest) (*in_custody.DerivedKeyResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *in_custody.DerivedKeyResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *in_custody.DeriveKeyRequest) (*in_custody.DerivedKeyResult, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *in_custody.DeriveKeyRequest) *in_custody.DerivedKeyResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*in_custody.DerivedKeyResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockKeyGenerationService creates a new instance of MockKeyGenerationService
func NewMockKeyGenerationService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockKeyGenerationService {
	mock := &MockKeyGenerationService{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
