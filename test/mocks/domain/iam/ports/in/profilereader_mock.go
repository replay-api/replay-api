package iam_in

import (
	"github.com/stretchr/testify/mock"
)

// MockProfileReader is a mock implementation of ProfileReader
type MockProfileReader struct {
	mock.Mock
}

// NewMockProfileReader creates a new instance of MockProfileReader
func NewMockProfileReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockProfileReader {
	mock := &MockProfileReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
