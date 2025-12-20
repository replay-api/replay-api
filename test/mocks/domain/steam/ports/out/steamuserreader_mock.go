package steam_out

import (
	"github.com/stretchr/testify/mock"
)

// MockSteamUserReader is a mock implementation of SteamUserReader
type MockSteamUserReader struct {
	mock.Mock
}

// NewMockSteamUserReader creates a new instance of MockSteamUserReader
func NewMockSteamUserReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockSteamUserReader {
	mock := &MockSteamUserReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
