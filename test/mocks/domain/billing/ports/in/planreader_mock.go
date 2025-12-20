package billing_in

import (
	"github.com/stretchr/testify/mock"
)

// MockPlanReader is a mock implementation of PlanReader
type MockPlanReader struct {
	mock.Mock
}

// NewMockPlanReader creates a new instance of MockPlanReader
func NewMockPlanReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPlanReader {
	mock := &MockPlanReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
