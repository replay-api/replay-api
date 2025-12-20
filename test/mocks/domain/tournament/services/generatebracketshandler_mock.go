package tournament_services

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockGenerateBracketsHandler is a mock implementation of GenerateBracketsHandler
type MockGenerateBracketsHandler struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockGenerateBracketsHandler) Exec(ctx context.Context, tournamentID uuid.UUID) error {
	ret := _m.Called(ctx, tournamentID)

	return ret.Error(0)
}

// NewMockGenerateBracketsHandler creates a new instance of MockGenerateBracketsHandler
func NewMockGenerateBracketsHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockGenerateBracketsHandler {
	mock := &MockGenerateBracketsHandler{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
