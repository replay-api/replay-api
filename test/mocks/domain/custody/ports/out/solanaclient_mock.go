package custody_out

import (
	"context"

	custody_out "github.com/replay-api/replay-api/pkg/domain/custody/ports/out"
	"github.com/stretchr/testify/mock"
)

// MockSolanaClient is a mock implementation of SolanaClient
type MockSolanaClient struct {
	mock.Mock
}

// GetAccountInfo provides a mock function
func (_m *MockSolanaClient) GetAccountInfo(ctx context.Context, address string) (*custody_out.SolanaAccountInfo, error) {
	ret := _m.Called(ctx, address)

	var r0 *custody_out.SolanaAccountInfo
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*custody_out.SolanaAccountInfo, error)); ok {
		return rf(ctx, address)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) *custody_out.SolanaAccountInfo); ok {
		r0 = rf(ctx, address)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.SolanaAccountInfo)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetTokenAccounts provides a mock function
func (_m *MockSolanaClient) GetTokenAccounts(ctx context.Context, owner string) ([]*custody_out.SolanaTokenAccount, error) {
	ret := _m.Called(ctx, owner)

	var r0 []*custody_out.SolanaTokenAccount
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) ([]*custody_out.SolanaTokenAccount, error)); ok {
		return rf(ctx, owner)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) []*custody_out.SolanaTokenAccount); ok {
		r0 = rf(ctx, owner)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*custody_out.SolanaTokenAccount)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// CreateAssociatedTokenAccount provides a mock function
func (_m *MockSolanaClient) CreateAssociatedTokenAccount(ctx context.Context, owner string, mint string) (*custody_out.UnsignedTransaction, error) {
	ret := _m.Called(ctx, owner, mint)

	var r0 *custody_out.UnsignedTransaction
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*custody_out.UnsignedTransaction, error)); ok {
		return rf(ctx, owner, mint)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string, string) *custody_out.UnsignedTransaction); ok {
		r0 = rf(ctx, owner, mint)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.UnsignedTransaction)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// BuildProgramInstruction provides a mock function
func (_m *MockSolanaClient) BuildProgramInstruction(ctx context.Context, req *custody_out.ProgramInstructionRequest) (*custody_out.UnsignedTransaction, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_out.UnsignedTransaction
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.ProgramInstructionRequest) (*custody_out.UnsignedTransaction, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.ProgramInstructionRequest) *custody_out.UnsignedTransaction); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.UnsignedTransaction)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetProgramAccounts provides a mock function
func (_m *MockSolanaClient) GetProgramAccounts(ctx context.Context, programID string, filter *custody_out.ProgramAccountFilter) ([]*custody_out.SolanaAccountInfo, error) {
	ret := _m.Called(ctx, programID, filter)

	var r0 []*custody_out.SolanaAccountInfo
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string, *custody_out.ProgramAccountFilter) ([]*custody_out.SolanaAccountInfo, error)); ok {
		return rf(ctx, programID, filter)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string, *custody_out.ProgramAccountFilter) []*custody_out.SolanaAccountInfo); ok {
		r0 = rf(ctx, programID, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*custody_out.SolanaAccountInfo)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// BuildSPLTransfer provides a mock function
func (_m *MockSolanaClient) BuildSPLTransfer(ctx context.Context, req *custody_out.SPLTransferRequest) (*custody_out.UnsignedTransaction, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_out.UnsignedTransaction
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.SPLTransferRequest) (*custody_out.UnsignedTransaction, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.SPLTransferRequest) *custody_out.UnsignedTransaction); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.UnsignedTransaction)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetRecentPriorityFees provides a mock function
func (_m *MockSolanaClient) GetRecentPriorityFees(ctx context.Context) (*custody_out.PriorityFeeInfo, error) {
	ret := _m.Called(ctx)

	var r0 *custody_out.PriorityFeeInfo
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context) (*custody_out.PriorityFeeInfo, error)); ok {
		return rf(ctx)
	}

	if rf, ok := ret.Get(0).(func(context.Context) *custody_out.PriorityFeeInfo); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.PriorityFeeInfo)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockSolanaClient creates a new instance of MockSolanaClient
func NewMockSolanaClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockSolanaClient {
	mock := &MockSolanaClient{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
