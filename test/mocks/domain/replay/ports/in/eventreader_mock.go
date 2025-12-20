package replay_in

import (
	"github.com/stretchr/testify/mock"
)

// MockEventReader is a mock implementation of EventReader
type MockEventReader struct {
	mock.Mock
}

// NewMockEventReader creates a new instance of MockEventReader
func NewMockEventReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockEventReader {
	mock := &MockEventReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
