package replay_out

import (
	"github.com/stretchr/testify/mock"
)

// MockMatchMetadataReader is a mock implementation of MatchMetadataReader
type MockMatchMetadataReader struct {
	mock.Mock
}

// NewMockMatchMetadataReader creates a new instance of MockMatchMetadataReader
func NewMockMatchMetadataReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMatchMetadataReader {
	mock := &MockMatchMetadataReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
