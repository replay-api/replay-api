package replay_in

import (
	"github.com/stretchr/testify/mock"
)

// MockPlayerMetadataReader is a mock implementation of PlayerMetadataReader
type MockPlayerMetadataReader struct {
	mock.Mock
}

// NewMockPlayerMetadataReader creates a new instance of MockPlayerMetadataReader
func NewMockPlayerMetadataReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPlayerMetadataReader {
	mock := &MockPlayerMetadataReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
