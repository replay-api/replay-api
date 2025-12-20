package blockchain_ports

import (
	"context"

	blockchain_out "github.com/replay-api/replay-api/pkg/domain/blockchain/ports/out"
	blockchain_vo "github.com/replay-api/replay-api/pkg/domain/blockchain/value-objects"
	"github.com/stretchr/testify/mock"
)

// MockMultiChainClient is a mock implementation of MultiChainClient
type MockMultiChainClient struct {
	mock.Mock
}

// GetClient provides a mock function
func (_m *MockMultiChainClient) GetClient(chainID blockchain_vo.ChainID) (blockchain_out.ChainClient, error) {
	ret := _m.Called(chainID)

	var r0 blockchain_out.ChainClient
	var r1 error

	if rf, ok := ret.Get(0).(func(blockchain_vo.ChainID) (blockchain_out.ChainClient, error)); ok {
		return rf(chainID)
	}

	if rf, ok := ret.Get(0).(func(blockchain_vo.ChainID) blockchain_out.ChainClient); ok {
		r0 = rf(chainID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(blockchain_out.ChainClient)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetPrimaryClient provides a mock function
func (_m *MockMultiChainClient) GetPrimaryClient() blockchain_out.ChainClient {
	ret := _m.Called()

	var r0 blockchain_out.ChainClient
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(blockchain_out.ChainClient)
	}
	return r0
}

// HealthCheck provides a mock function
func (_m *MockMultiChainClient) HealthCheck(ctx context.Context) map[blockchain_vo.ChainID]bool {
	ret := _m.Called(ctx)

	var r0 map[blockchain_vo.ChainID]bool
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(map[blockchain_vo.ChainID]bool)
	}
	return r0
}

// GetOptimalChain provides a mock function
func (_m *MockMultiChainClient) GetOptimalChain(ctx context.Context) blockchain_vo.ChainID {
	ret := _m.Called(ctx)

	var r0 blockchain_vo.ChainID
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(blockchain_vo.ChainID)
	}
	return r0
}

// NewMockMultiChainClient creates a new instance of MockMultiChainClient
func NewMockMultiChainClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMultiChainClient {
	mock := &MockMultiChainClient{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
