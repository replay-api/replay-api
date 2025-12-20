package replay_in

import (
	"context"

	"github.com/google/uuid"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	"github.com/stretchr/testify/mock"
)

// MockShareTokenReader is a mock implementation of ShareTokenReader
type MockShareTokenReader struct {
	mock.Mock
}

// FindByToken provides a mock function
func (_m *MockShareTokenReader) FindByToken(ctx context.Context, tokenID uuid.UUID) (*replay_entity.ShareToken, error) {
	ret := _m.Called(ctx, tokenID)

	var r0 *replay_entity.ShareToken
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*replay_entity.ShareToken, error)); ok {
		return rf(ctx, tokenID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *replay_entity.ShareToken); ok {
		r0 = rf(ctx, tokenID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*replay_entity.ShareToken)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockShareTokenReader creates a new instance of MockShareTokenReader
func NewMockShareTokenReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockShareTokenReader {
	mock := &MockShareTokenReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
