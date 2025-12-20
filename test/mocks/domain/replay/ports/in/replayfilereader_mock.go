package replay_in

import (
	"github.com/stretchr/testify/mock"
)

// MockReplayFileReader is a mock implementation of ReplayFileReader
type MockReplayFileReader struct {
	mock.Mock
}

// NewMockReplayFileReader creates a new instance of MockReplayFileReader
func NewMockReplayFileReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockReplayFileReader {
	mock := &MockReplayFileReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
