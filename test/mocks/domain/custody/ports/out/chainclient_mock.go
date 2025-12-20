package custody_out

import (
	"context"

	custody_out "github.com/replay-api/replay-api/pkg/domain/custody/ports/out"
	custody_vo "github.com/replay-api/replay-api/pkg/domain/custody/value-objects"
	"github.com/stretchr/testify/mock"
)

// MockChainClient is a mock implementation of ChainClient
type MockChainClient struct {
	mock.Mock
}

// GetChainID provides a mock function
func (_m *MockChainClient) GetChainID() custody_vo.ChainID {
	ret := _m.Called()

	var r0 custody_vo.ChainID
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(custody_vo.ChainID)
	}
	return r0
}

// GetChainInfo provides a mock function
func (_m *MockChainClient) GetChainInfo(ctx context.Context) (*custody_out.ChainInfo, error) {
	ret := _m.Called(ctx)

	var r0 *custody_out.ChainInfo
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context) (*custody_out.ChainInfo, error)); ok {
		return rf(ctx)
	}

	if rf, ok := ret.Get(0).(func(context.Context) *custody_out.ChainInfo); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.ChainInfo)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetLatestBlock provides a mock function
func (_m *MockChainClient) GetLatestBlock(ctx context.Context) (*custody_out.BlockInfo, error) {
	ret := _m.Called(ctx)

	var r0 *custody_out.BlockInfo
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context) (*custody_out.BlockInfo, error)); ok {
		return rf(ctx)
	}

	if rf, ok := ret.Get(0).(func(context.Context) *custody_out.BlockInfo); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.BlockInfo)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetBalance provides a mock function
func (_m *MockChainClient) GetBalance(ctx context.Context, address string) (*custody_out.Balance, error) {
	ret := _m.Called(ctx, address)

	var r0 *custody_out.Balance
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*custody_out.Balance, error)); ok {
		return rf(ctx, address)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) *custody_out.Balance); ok {
		r0 = rf(ctx, address)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.Balance)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetTokenBalance provides a mock function
func (_m *MockChainClient) GetTokenBalance(ctx context.Context, address string, tokenAddress string) (*custody_out.TokenBalance, error) {
	ret := _m.Called(ctx, address, tokenAddress)

	var r0 *custody_out.TokenBalance
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*custody_out.TokenBalance, error)); ok {
		return rf(ctx, address, tokenAddress)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string, string) *custody_out.TokenBalance); ok {
		r0 = rf(ctx, address, tokenAddress)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.TokenBalance)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetNonce provides a mock function
func (_m *MockChainClient) GetNonce(ctx context.Context, address string) (uint64, error) {
	ret := _m.Called(ctx, address)

	var r0 uint64
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (uint64, error)); ok {
		return rf(ctx, address)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) uint64); ok {
		r0 = rf(ctx, address)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(uint64)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// BuildTransaction provides a mock function
func (_m *MockChainClient) BuildTransaction(ctx context.Context, req *custody_out.TxBuildRequest) (*custody_out.UnsignedTransaction, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_out.UnsignedTransaction
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.TxBuildRequest) (*custody_out.UnsignedTransaction, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.TxBuildRequest) *custody_out.UnsignedTransaction); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.UnsignedTransaction)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// EstimateGas provides a mock function
func (_m *MockChainClient) EstimateGas(ctx context.Context, tx *custody_out.UnsignedTransaction) (*custody_out.GasEstimate, error) {
	ret := _m.Called(ctx, tx)

	var r0 *custody_out.GasEstimate
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.UnsignedTransaction) (*custody_out.GasEstimate, error)); ok {
		return rf(ctx, tx)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.UnsignedTransaction) *custody_out.GasEstimate); ok {
		r0 = rf(ctx, tx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.GasEstimate)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetGasPrice provides a mock function
func (_m *MockChainClient) GetGasPrice(ctx context.Context) (*custody_out.GasPriceInfo, error) {
	ret := _m.Called(ctx)

	var r0 *custody_out.GasPriceInfo
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context) (*custody_out.GasPriceInfo, error)); ok {
		return rf(ctx)
	}

	if rf, ok := ret.Get(0).(func(context.Context) *custody_out.GasPriceInfo); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.GasPriceInfo)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// SubmitTransaction provides a mock function
func (_m *MockChainClient) SubmitTransaction(ctx context.Context, signedTx []byte) (*custody_out.TxSubmitResult, error) {
	ret := _m.Called(ctx, signedTx)

	var r0 *custody_out.TxSubmitResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, []byte) (*custody_out.TxSubmitResult, error)); ok {
		return rf(ctx, signedTx)
	}

	if rf, ok := ret.Get(0).(func(context.Context, []byte) *custody_out.TxSubmitResult); ok {
		r0 = rf(ctx, signedTx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.TxSubmitResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetTransaction provides a mock function
func (_m *MockChainClient) GetTransaction(ctx context.Context, txHash string) (*custody_out.TransactionInfo, error) {
	ret := _m.Called(ctx, txHash)

	var r0 *custody_out.TransactionInfo
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*custody_out.TransactionInfo, error)); ok {
		return rf(ctx, txHash)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) *custody_out.TransactionInfo); ok {
		r0 = rf(ctx, txHash)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.TransactionInfo)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// WaitForConfirmation provides a mock function
func (_m *MockChainClient) WaitForConfirmation(ctx context.Context, txHash string, confirmations uint64) (*custody_out.TransactionReceipt, error) {
	ret := _m.Called(ctx, txHash, confirmations)

	var r0 *custody_out.TransactionReceipt
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string, uint64) (*custody_out.TransactionReceipt, error)); ok {
		return rf(ctx, txHash, confirmations)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string, uint64) *custody_out.TransactionReceipt); ok {
		r0 = rf(ctx, txHash, confirmations)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_out.TransactionReceipt)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// CallContract provides a mock function
func (_m *MockChainClient) CallContract(ctx context.Context, req *custody_out.ContractCallRequest) ([]byte, error) {
	ret := _m.Called(ctx, req)

	var r0 []byte
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.ContractCallRequest) ([]byte, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.ContractCallRequest) []byte); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetContractLogs provides a mock function
func (_m *MockChainClient) GetContractLogs(ctx context.Context, filter *custody_out.LogFilter) ([]*custody_out.ContractLog, error) {
	ret := _m.Called(ctx, filter)

	var r0 []*custody_out.ContractLog
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.LogFilter) ([]*custody_out.ContractLog, error)); ok {
		return rf(ctx, filter)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_out.LogFilter) []*custody_out.ContractLog); ok {
		r0 = rf(ctx, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*custody_out.ContractLog)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// HealthCheck provides a mock function
func (_m *MockChainClient) HealthCheck(ctx context.Context) error {
	ret := _m.Called(ctx)

	return ret.Error(0)
}

// NewMockChainClient creates a new instance of MockChainClient
func NewMockChainClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockChainClient {
	mock := &MockChainClient{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
