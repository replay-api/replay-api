package billing_in

import (
	"github.com/stretchr/testify/mock"
)

// MockSubscriptionReader is a mock implementation of SubscriptionReader
type MockSubscriptionReader struct {
	mock.Mock
}

// NewMockSubscriptionReader creates a new instance of MockSubscriptionReader
func NewMockSubscriptionReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockSubscriptionReader {
	mock := &MockSubscriptionReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
