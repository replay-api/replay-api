package google_out

import (
	"context"

	google_entity "github.com/replay-api/replay-api/pkg/domain/google/entities"
	"github.com/stretchr/testify/mock"
)

// MockGoogleUserWriter is a mock implementation of GoogleUserWriter
type MockGoogleUserWriter struct {
	mock.Mock
}

// Create provides a mock function
func (_m *MockGoogleUserWriter) Create(ctx context.Context, user *google_entity.GoogleUser) (*google_entity.GoogleUser, error) {
	ret := _m.Called(ctx, user)

	var r0 *google_entity.GoogleUser
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *google_entity.GoogleUser) (*google_entity.GoogleUser, error)); ok {
		return rf(ctx, user)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *google_entity.GoogleUser) *google_entity.GoogleUser); ok {
		r0 = rf(ctx, user)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*google_entity.GoogleUser)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockGoogleUserWriter creates a new instance of MockGoogleUserWriter
func NewMockGoogleUserWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockGoogleUserWriter {
	mock := &MockGoogleUserWriter{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
