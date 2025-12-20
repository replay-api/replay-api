package blockchain_ports

import (
	"context"

	big "math/big"

	blockchain_out "github.com/replay-api/replay-api/pkg/domain/blockchain/ports/out"
	blockchain_vo "github.com/replay-api/replay-api/pkg/domain/blockchain/value-objects"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	"github.com/stretchr/testify/mock"
)

// MockChainClient is a mock implementation of ChainClient
type MockChainClient struct {
	mock.Mock
}

// Connect provides a mock function
func (_m *MockChainClient) Connect(ctx context.Context) error {
	ret := _m.Called(ctx)

	return ret.Error(0)
}

// Disconnect provides a mock function
func (_m *MockChainClient) Disconnect() error {
	ret := _m.Called()

	return ret.Error(0)
}

// IsConnected provides a mock function
func (_m *MockChainClient) IsConnected() bool {
	ret := _m.Called()

	var r0 bool
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(bool)
	}
	return r0
}

// GetChainID provides a mock function
func (_m *MockChainClient) GetChainID() blockchain_vo.ChainID {
	ret := _m.Called()

	var r0 blockchain_vo.ChainID
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(blockchain_vo.ChainID)
	}
	return r0
}

// GetLatestBlockNumber provides a mock function
func (_m *MockChainClient) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	ret := _m.Called(ctx)

	var r0 uint64
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context) (uint64, error)); ok {
		return rf(ctx)
	}

	if rf, ok := ret.Get(0).(func(context.Context) uint64); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(uint64)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetBlockByNumber provides a mock function
func (_m *MockChainClient) GetBlockByNumber(ctx context.Context, blockNumber uint64) (*blockchain_out.BlockInfo, error) {
	ret := _m.Called(ctx, blockNumber)

	var r0 *blockchain_out.BlockInfo
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uint64) (*blockchain_out.BlockInfo, error)); ok {
		return rf(ctx, blockNumber)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uint64) *blockchain_out.BlockInfo); ok {
		r0 = rf(ctx, blockNumber)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*blockchain_out.BlockInfo)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetTransaction provides a mock function
func (_m *MockChainClient) GetTransaction(ctx context.Context, txHash blockchain_vo.TxHash) (*blockchain_out.TransactionInfo, error) {
	ret := _m.Called(ctx, txHash)

	var r0 *blockchain_out.TransactionInfo
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, blockchain_vo.TxHash) (*blockchain_out.TransactionInfo, error)); ok {
		return rf(ctx, txHash)
	}

	if rf, ok := ret.Get(0).(func(context.Context, blockchain_vo.TxHash) *blockchain_out.TransactionInfo); ok {
		r0 = rf(ctx, txHash)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*blockchain_out.TransactionInfo)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetTransactionReceipt provides a mock function
func (_m *MockChainClient) GetTransactionReceipt(ctx context.Context, txHash blockchain_vo.TxHash) (*blockchain_out.TransactionReceipt, error) {
	ret := _m.Called(ctx, txHash)

	var r0 *blockchain_out.TransactionReceipt
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, blockchain_vo.TxHash) (*blockchain_out.TransactionReceipt, error)); ok {
		return rf(ctx, txHash)
	}

	if rf, ok := ret.Get(0).(func(context.Context, blockchain_vo.TxHash) *blockchain_out.TransactionReceipt); ok {
		r0 = rf(ctx, txHash)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*blockchain_out.TransactionReceipt)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// SendRawTransaction provides a mock function
func (_m *MockChainClient) SendRawTransaction(ctx context.Context, signedTx []byte) (blockchain_vo.TxHash, error) {
	ret := _m.Called(ctx, signedTx)

	var r0 blockchain_vo.TxHash
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, []byte) (blockchain_vo.TxHash, error)); ok {
		return rf(ctx, signedTx)
	}

	if rf, ok := ret.Get(0).(func(context.Context, []byte) blockchain_vo.TxHash); ok {
		r0 = rf(ctx, signedTx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(blockchain_vo.TxHash)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// EstimateGas provides a mock function
func (_m *MockChainClient) EstimateGas(ctx context.Context, from wallet_vo.EVMAddress, to wallet_vo.EVMAddress, data []byte, value *big.Int) (uint64, error) {
	ret := _m.Called(ctx, from, to, data, value)

	var r0 uint64
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, wallet_vo.EVMAddress, wallet_vo.EVMAddress, []byte, *big.Int) (uint64, error)); ok {
		return rf(ctx, from, to, data, value)
	}

	if rf, ok := ret.Get(0).(func(context.Context, wallet_vo.EVMAddress, wallet_vo.EVMAddress, []byte, *big.Int) uint64); ok {
		r0 = rf(ctx, from, to, data, value)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(uint64)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetGasPrice provides a mock function
func (_m *MockChainClient) GetGasPrice(ctx context.Context) (*big.Int, error) {
	ret := _m.Called(ctx)

	var r0 *big.Int
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context) (*big.Int, error)); ok {
		return rf(ctx)
	}

	if rf, ok := ret.Get(0).(func(context.Context) *big.Int); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*big.Int)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetBalance provides a mock function
func (_m *MockChainClient) GetBalance(ctx context.Context, address wallet_vo.EVMAddress) (*big.Int, error) {
	ret := _m.Called(ctx, address)

	var r0 *big.Int
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, wallet_vo.EVMAddress) (*big.Int, error)); ok {
		return rf(ctx, address)
	}

	if rf, ok := ret.Get(0).(func(context.Context, wallet_vo.EVMAddress) *big.Int); ok {
		r0 = rf(ctx, address)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*big.Int)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetNonce provides a mock function
func (_m *MockChainClient) GetNonce(ctx context.Context, address wallet_vo.EVMAddress) (uint64, error) {
	ret := _m.Called(ctx, address)

	var r0 uint64
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, wallet_vo.EVMAddress) (uint64, error)); ok {
		return rf(ctx, address)
	}

	if rf, ok := ret.Get(0).(func(context.Context, wallet_vo.EVMAddress) uint64); ok {
		r0 = rf(ctx, address)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(uint64)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetTokenBalance provides a mock function
func (_m *MockChainClient) GetTokenBalance(ctx context.Context, tokenAddr wallet_vo.EVMAddress, accountAddr wallet_vo.EVMAddress) (*big.Int, error) {
	ret := _m.Called(ctx, tokenAddr, accountAddr)

	var r0 *big.Int
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, wallet_vo.EVMAddress, wallet_vo.EVMAddress) (*big.Int, error)); ok {
		return rf(ctx, tokenAddr, accountAddr)
	}

	if rf, ok := ret.Get(0).(func(context.Context, wallet_vo.EVMAddress, wallet_vo.EVMAddress) *big.Int); ok {
		r0 = rf(ctx, tokenAddr, accountAddr)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*big.Int)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetTokenDecimals provides a mock function
func (_m *MockChainClient) GetTokenDecimals(ctx context.Context, tokenAddr wallet_vo.EVMAddress) (uint8, error) {
	ret := _m.Called(ctx, tokenAddr)

	var r0 uint8
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, wallet_vo.EVMAddress) (uint8, error)); ok {
		return rf(ctx, tokenAddr)
	}

	if rf, ok := ret.Get(0).(func(context.Context, wallet_vo.EVMAddress) uint8); ok {
		r0 = rf(ctx, tokenAddr)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(uint8)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// SubscribeToEvents provides a mock function
func (_m *MockChainClient) SubscribeToEvents(ctx context.Context, contractAddr wallet_vo.EVMAddress, topics [][]byte) (interface{}, error) {
	ret := _m.Called(ctx, contractAddr, topics)

	var r0 interface{}
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, wallet_vo.EVMAddress, [][]byte) (interface{}, error)); ok {
		return rf(ctx, contractAddr, topics)
	}

	if rf, ok := ret.Get(0).(func(context.Context, wallet_vo.EVMAddress, [][]byte) interface{}); ok {
		r0 = rf(ctx, contractAddr, topics)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetLogs provides a mock function
func (_m *MockChainClient) GetLogs(ctx context.Context, filter blockchain_out.LogFilter) ([]blockchain_out.ContractEvent, error) {
	ret := _m.Called(ctx, filter)

	var r0 []blockchain_out.ContractEvent
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, blockchain_out.LogFilter) ([]blockchain_out.ContractEvent, error)); ok {
		return rf(ctx, filter)
	}

	if rf, ok := ret.Get(0).(func(context.Context, blockchain_out.LogFilter) []blockchain_out.ContractEvent); ok {
		r0 = rf(ctx, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]blockchain_out.ContractEvent)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
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
