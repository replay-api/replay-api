package replay_out

import (
	"github.com/stretchr/testify/mock"
)

// MockBadgeReader is a mock implementation of BadgeReader
type MockBadgeReader struct {
	mock.Mock
}

// NewMockBadgeReader creates a new instance of MockBadgeReader
func NewMockBadgeReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockBadgeReader {
	mock := &MockBadgeReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
