package squad_out

import (
	"github.com/stretchr/testify/mock"
)

// MockPlayerProfileReader is a mock implementation of PlayerProfileReader
type MockPlayerProfileReader struct {
	mock.Mock
}

// NewMockPlayerProfileReader creates a new instance of MockPlayerProfileReader
func NewMockPlayerProfileReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPlayerProfileReader {
	mock := &MockPlayerProfileReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
