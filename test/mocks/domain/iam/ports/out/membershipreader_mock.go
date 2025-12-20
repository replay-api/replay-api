package iam_out

import (
	"github.com/stretchr/testify/mock"
)

// MockMembershipReader is a mock implementation of MembershipReader
type MockMembershipReader struct {
	mock.Mock
}

// NewMockMembershipReader creates a new instance of MockMembershipReader
func NewMockMembershipReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMembershipReader {
	mock := &MockMembershipReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
