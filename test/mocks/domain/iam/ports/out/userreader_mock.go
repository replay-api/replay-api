package iam_out

import (
	"context"

	common "github.com/replay-api/replay-api/pkg/domain"
	iam_entity "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	"github.com/stretchr/testify/mock"
)

// MockUserReader is a mock implementation of UserReader
type MockUserReader struct {
	mock.Mock
}

// Search provides a mock function
func (_m *MockUserReader) Search(ctx context.Context, s common.Search) ([]iam_entity.User, error) {
	ret := _m.Called(ctx, s)

	var r0 []iam_entity.User
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, common.Search) ([]iam_entity.User, error)); ok {
		return rf(ctx, s)
	}

	if rf, ok := ret.Get(0).(func(context.Context, common.Search) []iam_entity.User); ok {
		r0 = rf(ctx, s)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]iam_entity.User)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockUserReader creates a new instance of MockUserReader
func NewMockUserReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockUserReader {
	mock := &MockUserReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
