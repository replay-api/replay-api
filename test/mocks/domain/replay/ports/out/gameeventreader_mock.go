package replay_out

import (
	"github.com/stretchr/testify/mock"
)

// MockGameEventReader is a mock implementation of GameEventReader
type MockGameEventReader struct {
	mock.Mock
}

// NewMockGameEventReader creates a new instance of MockGameEventReader
func NewMockGameEventReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockGameEventReader {
	mock := &MockGameEventReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
