package google_out

import (
	"context"

	common "github.com/replay-api/replay-api/pkg/domain"
	google_entity "github.com/replay-api/replay-api/pkg/domain/google/entities"
	"github.com/stretchr/testify/mock"
)

// MockGoogleUserReader is a mock implementation of GoogleUserReader
type MockGoogleUserReader struct {
	mock.Mock
}

// Search provides a mock function
func (_m *MockGoogleUserReader) Search(ctx context.Context, s common.Search) ([]google_entity.GoogleUser, error) {
	ret := _m.Called(ctx, s)

	var r0 []google_entity.GoogleUser
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, common.Search) ([]google_entity.GoogleUser, error)); ok {
		return rf(ctx, s)
	}

	if rf, ok := ret.Get(0).(func(context.Context, common.Search) []google_entity.GoogleUser); ok {
		r0 = rf(ctx, s)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]google_entity.GoogleUser)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockGoogleUserReader creates a new instance of MockGoogleUserReader
func NewMockGoogleUserReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockGoogleUserReader {
	mock := &MockGoogleUserReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
