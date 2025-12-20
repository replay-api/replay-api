package custody_in

import (
	"context"

	"github.com/google/uuid"
	custody_in "github.com/replay-api/replay-api/pkg/domain/custody/ports/in"
	custody_vo "github.com/replay-api/replay-api/pkg/domain/custody/value-objects"
	"github.com/stretchr/testify/mock"
)

// MockSigningService is a mock implementation of SigningService
type MockSigningService struct {
	mock.Mock
}

// RequestSigning provides a mock function
func (_m *MockSigningService) RequestSigning(ctx context.Context, req *custody_in.SigningRequest) (*custody_in.SigningResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_in.SigningResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.SigningRequest) (*custody_in.SigningResult, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.SigningRequest) *custody_in.SigningResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.SigningResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetSigningStatus provides a mock function
func (_m *MockSigningService) GetSigningStatus(ctx context.Context, sessionID string) (*custody_in.SigningStatus, error) {
	ret := _m.Called(ctx, sessionID)

	var r0 *custody_in.SigningStatus
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*custody_in.SigningStatus, error)); ok {
		return rf(ctx, sessionID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) *custody_in.SigningStatus); ok {
		r0 = rf(ctx, sessionID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.SigningStatus)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// CancelSigning provides a mock function
func (_m *MockSigningService) CancelSigning(ctx context.Context, sessionID string) error {
	ret := _m.Called(ctx, sessionID)

	return ret.Error(0)
}

// SignTypedData provides a mock function
func (_m *MockSigningService) SignTypedData(ctx context.Context, req *custody_in.TypedDataSigningRequest) (*custody_in.SigningResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_in.SigningResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.TypedDataSigningRequest) (*custody_in.SigningResult, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.TypedDataSigningRequest) *custody_in.SigningResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.SigningResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// PersonalSign provides a mock function
func (_m *MockSigningService) PersonalSign(ctx context.Context, req *custody_in.PersonalSignRequest) (*custody_in.SigningResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_in.SigningResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.PersonalSignRequest) (*custody_in.SigningResult, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.PersonalSignRequest) *custody_in.SigningResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.SigningResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// SignSolanaTransaction provides a mock function
func (_m *MockSigningService) SignSolanaTransaction(ctx context.Context, req *custody_in.SolanaSigningRequest) (*custody_in.SigningResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_in.SigningResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.SolanaSigningRequest) (*custody_in.SigningResult, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.SolanaSigningRequest) *custody_in.SigningResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.SigningResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// SignSolanaMessage provides a mock function
func (_m *MockSigningService) SignSolanaMessage(ctx context.Context, req *custody_in.SolanaMessageRequest) (*custody_in.SigningResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_in.SigningResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.SolanaMessageRequest) (*custody_in.SigningResult, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.SolanaMessageRequest) *custody_in.SigningResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.SigningResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// SignTransaction provides a mock function
func (_m *MockSigningService) SignTransaction(ctx context.Context, req *custody_in.TxSigningRequest) (*custody_in.SignedTxResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_in.SignedTxResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.TxSigningRequest) (*custody_in.SignedTxResult, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.TxSigningRequest) *custody_in.SignedTxResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.SignedTxResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// SignUserOperation provides a mock function
func (_m *MockSigningService) SignUserOperation(ctx context.Context, req *custody_in.UserOpSigningRequest) (*custody_in.SignedUserOpResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_in.SignedUserOpResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.UserOpSigningRequest) (*custody_in.SignedUserOpResult, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.UserOpSigningRequest) *custody_in.SignedUserOpResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.SignedUserOpResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetSignerAddress provides a mock function
func (_m *MockSigningService) GetSignerAddress(ctx context.Context, walletID uuid.UUID, chainID custody_vo.ChainID) (string, error) {
	ret := _m.Called(ctx, walletID, chainID)

	var r0 string
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, custody_vo.ChainID) (string, error)); ok {
		return rf(ctx, walletID, chainID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, custody_vo.ChainID) string); ok {
		r0 = rf(ctx, walletID, chainID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(string)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetPublicKey provides a mock function
func (_m *MockSigningService) GetPublicKey(ctx context.Context, walletID uuid.UUID) (*custody_in.PublicKeyInfo, error) {
	ret := _m.Called(ctx, walletID)

	var r0 *custody_in.PublicKeyInfo
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*custody_in.PublicKeyInfo, error)); ok {
		return rf(ctx, walletID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *custody_in.PublicKeyInfo); ok {
		r0 = rf(ctx, walletID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.PublicKeyInfo)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockSigningService creates a new instance of MockSigningService
func NewMockSigningService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockSigningService {
	mock := &MockSigningService{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
