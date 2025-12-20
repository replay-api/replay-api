package replay_in

import (
	"github.com/stretchr/testify/mock"
)

// MockRoundReader is a mock implementation of RoundReader
type MockRoundReader struct {
	mock.Mock
}

// NewMockRoundReader creates a new instance of MockRoundReader
func NewMockRoundReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockRoundReader {
	mock := &MockRoundReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
