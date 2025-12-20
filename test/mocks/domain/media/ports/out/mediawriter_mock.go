package media_out

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockMediaWriter is a mock implementation of MediaWriter
type MockMediaWriter struct {
	mock.Mock
}

// Create provides a mock function
func (_m *MockMediaWriter) Create(ctx context.Context, media []byte, name string, extension string) (string, error) {
	ret := _m.Called(ctx, media, name, extension)

	var r0 string
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, []byte, string, string) (string, error)); ok {
		return rf(ctx, media, name, extension)
	}

	if rf, ok := ret.Get(0).(func(context.Context, []byte, string, string) string); ok {
		r0 = rf(ctx, media, name, extension)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(string)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockMediaWriter creates a new instance of MockMediaWriter
func NewMockMediaWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMediaWriter {
	mock := &MockMediaWriter{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
