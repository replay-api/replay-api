package replay_in

import (
	"github.com/stretchr/testify/mock"
)

// MockMatchReader is a mock implementation of MatchReader
type MockMatchReader struct {
	mock.Mock
}

// NewMockMatchReader creates a new instance of MockMatchReader
func NewMockMatchReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMatchReader {
	mock := &MockMatchReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
