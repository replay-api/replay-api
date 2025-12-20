package iam_out

import (
	"context"

	common "github.com/replay-api/replay-api/pkg/domain"
	iam_entity "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	"github.com/stretchr/testify/mock"
)

// MockGroupReader is a mock implementation of GroupReader
type MockGroupReader struct {
	mock.Mock
}

// Search provides a mock function
func (_m *MockGroupReader) Search(ctx context.Context, s common.Search) ([]iam_entity.Group, error) {
	ret := _m.Called(ctx, s)

	var r0 []iam_entity.Group
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, common.Search) ([]iam_entity.Group, error)); ok {
		return rf(ctx, s)
	}

	if rf, ok := ret.Get(0).(func(context.Context, common.Search) []iam_entity.Group); ok {
		r0 = rf(ctx, s)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]iam_entity.Group)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockGroupReader creates a new instance of MockGroupReader
func NewMockGroupReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockGroupReader {
	mock := &MockGroupReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
