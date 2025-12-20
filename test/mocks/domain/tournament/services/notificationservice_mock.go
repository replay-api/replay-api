package tournament_services

import (
	"context"

	"math/big"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockNotificationService is a mock implementation of NotificationService
type MockNotificationService struct {
	mock.Mock
}

// SendPrizeNotification provides a mock function
func (_m *MockNotificationService) SendPrizeNotification(ctx context.Context, userID uuid.UUID, amount *big.Float, currency string, position int, tournamentName string) error {
	ret := _m.Called(ctx, userID, amount, currency, position, tournamentName)

	return ret.Error(0)
}

// NewMockNotificationService creates a new instance of MockNotificationService
func NewMockNotificationService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockNotificationService {
	mock := &MockNotificationService{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
