package email_out

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockVHashWriter is a mock implementation of VHashWriter
type MockVHashWriter struct {
	mock.Mock
}

// CreateVHash provides a mock function
func (_m *MockVHashWriter) CreateVHash(ctx context.Context, email string) string {
	ret := _m.Called(ctx, email)

	var r0 string
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(string)
	}
	return r0
}

// NewMockVHashWriter creates a new instance of MockVHashWriter
func NewMockVHashWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockVHashWriter {
	mock := &MockVHashWriter{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
