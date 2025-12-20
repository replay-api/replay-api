package custody_out

import (
	"context"

	"math/big"

	custody_out "github.com/replay-api/replay-api/pkg/domain/custody/ports/out"
	"github.com/stretchr/testify/mock"
)

// MockEVMClient is a mock implementation of EVMClient
type MockEVMClient struct {
	mock.Mock
}

// GetCode provides a mock function
func (_m *MockEVMClient) GetCode(ctx context.Context, address string) ([]byte, error) {
	ret := _m.Called(ctx, address)

	var r0 []byte
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) ([]byte, error)); ok {
		return rf(ctx, address)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) []byte); ok {
		r0 = rf(ctx, address)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetStorageAt provides a mock function
func (_m *MockEVMClient) GetStorageAt(ctx context.Context, address string, slot string) ([]byte, error) {
	ret := _m.Called(ctx, address, slot)

	var r0 []byte
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string, string) ([]byte, error)); ok {
		return rf(ctx, address, slot)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string, string) []byte); ok {
		r0 = rf(ctx, address, slot)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// BuildUserOperation provides a mock function
func (_m *MockEVMClient) BuildUserOperation(ctx context.Context, req *custody_out.UserOpRequest) (*custody_out.UserOperation, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_out.UserOperation
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.UserOpRequest) (*custody_out.UserOperation, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.UserOpRequest) *custody_out.UserOperation); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.UserOperation)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// EstimateUserOperationGas provides a mock function
func (_m *MockEVMClient) EstimateUserOperationGas(ctx context.Context, userOp *custody_out.UserOperation) (*custody_out.UserOpGasEstimate, error) {
	ret := _m.Called(ctx, userOp)

	var r0 *custody_out.UserOpGasEstimate
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.UserOperation) (*custody_out.UserOpGasEstimate, error)); ok {
		return rf(ctx, userOp)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.UserOperation) *custody_out.UserOpGasEstimate); ok {
		r0 = rf(ctx, userOp)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.UserOpGasEstimate)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// SubmitUserOperation provides a mock function
func (_m *MockEVMClient) SubmitUserOperation(ctx context.Context, userOp *custody_out.UserOperation) (*custody_out.UserOpResult, error) {
	ret := _m.Called(ctx, userOp)

	var r0 *custody_out.UserOpResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.UserOperation) (*custody_out.UserOpResult, error)); ok {
		return rf(ctx, userOp)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.UserOperation) *custody_out.UserOpResult); ok {
		r0 = rf(ctx, userOp)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.UserOpResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetUserOperationReceipt provides a mock function
func (_m *MockEVMClient) GetUserOperationReceipt(ctx context.Context, userOpHash string) (*custody_out.UserOpReceipt, error) {
	ret := _m.Called(ctx, userOpHash)

	var r0 *custody_out.UserOpReceipt
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*custody_out.UserOpReceipt, error)); ok {
		return rf(ctx, userOpHash)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) *custody_out.UserOpReceipt); ok {
		r0 = rf(ctx, userOpHash)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.UserOpReceipt)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// BuildERC20Transfer provides a mock function
func (_m *MockEVMClient) BuildERC20Transfer(ctx context.Context, req *custody_out.ERC20TransferRequest) (*custody_out.UnsignedTransaction, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_out.UnsignedTransaction
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.ERC20TransferRequest) (*custody_out.UnsignedTransaction, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.ERC20TransferRequest) *custody_out.UnsignedTransaction); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.UnsignedTransaction)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// BuildERC20Approve provides a mock function
func (_m *MockEVMClient) BuildERC20Approve(ctx context.Context, req *custody_out.ERC20ApproveRequest) (*custody_out.UnsignedTransaction, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_out.UnsignedTransaction
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.ERC20ApproveRequest) (*custody_out.UnsignedTransaction, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.ERC20ApproveRequest) *custody_out.UnsignedTransaction); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.UnsignedTransaction)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetERC20Allowance provides a mock function
func (_m *MockEVMClient) GetERC20Allowance(ctx context.Context, token string, owner string, spender string) (*big.Int, error) {
	ret := _m.Called(ctx, token, owner, spender)

	var r0 *big.Int
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) (*big.Int, error)); ok {
		return rf(ctx, token, owner, spender)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) *big.Int); ok {
		r0 = rf(ctx, token, owner, spender)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*big.Int)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockEVMClient creates a new instance of MockEVMClient
func NewMockEVMClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockEVMClient {
	mock := &MockEVMClient{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
