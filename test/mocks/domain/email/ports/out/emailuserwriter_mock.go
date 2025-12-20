package email_out

import (
	"context"

	email_entities "github.com/replay-api/replay-api/pkg/domain/email/entities"
	"github.com/stretchr/testify/mock"
)

// MockEmailUserWriter is a mock implementation of EmailUserWriter
type MockEmailUserWriter struct {
	mock.Mock
}

// Create provides a mock function
func (_m *MockEmailUserWriter) Create(ctx context.Context, user *email_entities.EmailUser) (*email_entities.EmailUser, error) {
	ret := _m.Called(ctx, user)

	var r0 *email_entities.EmailUser
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *email_entities.EmailUser) (*email_entities.EmailUser, error)); ok {
		return rf(ctx, user)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *email_entities.EmailUser) *email_entities.EmailUser); ok {
		r0 = rf(ctx, user)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*email_entities.EmailUser)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockEmailUserWriter creates a new instance of MockEmailUserWriter
func NewMockEmailUserWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockEmailUserWriter {
	mock := &MockEmailUserWriter{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
