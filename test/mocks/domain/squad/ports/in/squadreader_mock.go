package squad_in

import (
	"github.com/stretchr/testify/mock"
)

// MockSquadReader is a mock implementation of SquadReader
type MockSquadReader struct {
	mock.Mock
}

// NewMockSquadReader creates a new instance of MockSquadReader
func NewMockSquadReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockSquadReader {
	mock := &MockSquadReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
