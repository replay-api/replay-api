package replay_in

import (
	"github.com/stretchr/testify/mock"
)

// MockTeamReader is a mock implementation of TeamReader
type MockTeamReader struct {
	mock.Mock
}

// NewMockTeamReader creates a new instance of MockTeamReader
func NewMockTeamReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockTeamReader {
	mock := &MockTeamReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
