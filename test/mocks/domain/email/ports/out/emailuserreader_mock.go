package email_out

import (
	"context"

	common "github.com/replay-api/replay-api/pkg/domain"
	email_entities "github.com/replay-api/replay-api/pkg/domain/email/entities"
	"github.com/stretchr/testify/mock"
)

// MockEmailUserReader is a mock implementation of EmailUserReader
type MockEmailUserReader struct {
	mock.Mock
}

// Search provides a mock function
func (_m *MockEmailUserReader) Search(ctx context.Context, s common.Search) ([]email_entities.EmailUser, error) {
	ret := _m.Called(ctx, s)

	var r0 []email_entities.EmailUser
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, common.Search) ([]email_entities.EmailUser, error)); ok {
		return rf(ctx, s)
	}

	if rf, ok := ret.Get(0).(func(context.Context, common.Search) []email_entities.EmailUser); ok {
		r0 = rf(ctx, s)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]email_entities.EmailUser)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockEmailUserReader creates a new instance of MockEmailUserReader
func NewMockEmailUserReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockEmailUserReader {
	mock := &MockEmailUserReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
