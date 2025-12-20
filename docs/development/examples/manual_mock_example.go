package examples

// This file contains examples of how to create mocks manually without using mockery.
// These examples are for reference only - they should not be compiled.

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// EXAMPLE 1: Simple Mock (single return)
// ============================================================================

// Interface we want to mock
type SimpleRepository interface {
	Delete(ctx context.Context, id uuid.UUID) error
}

// Manual mock of the interface
type MockSimpleRepository struct {
	mock.Mock
}

func (m *MockSimpleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Usage in test:
// mockRepo := new(MockSimpleRepository)
// mockRepo.On("Delete", mock.Anything, uuid.New()).Return(nil)

// ============================================================================
// EXAMPLE 2: Mock with Return Value
// ============================================================================

type EntityRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Entity, error)
}

type Entity struct {
	ID   uuid.UUID
	Name string
}

type MockEntityRepository struct {
	mock.Mock
}

func (m *MockEntityRepository) FindByID(ctx context.Context, id uuid.UUID) (*Entity, error) {
	args := m.Called(ctx, id)
	
	// Check if first return value is nil
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	
	// Type assert the first return value
	return args.Get(0).(*Entity), args.Error(1)
}

// Usage in test:
// mockRepo := new(MockEntityRepository)
// expectedEntity := &Entity{ID: uuid.New(), Name: "Test"}
// mockRepo.On("FindByID", mock.Anything, id).Return(expectedEntity, nil)

// ============================================================================
// EXAMPLE 3: Mock with Multiple Return Values
// ============================================================================

type CommandHandler interface {
	Exec(ctx context.Context, cmd Command) (*Result, *Metadata, error)
}

type Command struct {
	ID string
}

type Result struct {
	Success bool
}

type Metadata struct {
	Timestamp int64
}

type MockCommandHandler struct {
	mock.Mock
}

func (m *MockCommandHandler) Exec(ctx context.Context, cmd Command) (*Result, *Metadata, error) {
	args := m.Called(ctx, cmd)
	
	var result *Result
	var metadata *Metadata
	
	// First return value (index 0)
	if args.Get(0) != nil {
		result = args.Get(0).(*Result)
	}
	
	// Second return value (index 1)
	if args.Get(1) != nil {
		metadata = args.Get(1).(*Metadata)
	}
	
	// Error is always the last (index 2 for third parameter)
	return result, metadata, args.Error(2)
}

// Usage in test:
// mockHandler := new(MockCommandHandler)
// mockHandler.On("Exec", mock.Anything, mock.Anything).
//     Return(&Result{Success: true}, &Metadata{Timestamp: 123}, nil)

// ============================================================================
// EXAMPLE 4: Mock with Slice as Return Value
// ============================================================================

type ListRepository interface {
	FindAll(ctx context.Context) ([]*Entity, error)
}

type MockListRepository struct {
	mock.Mock
}

func (m *MockListRepository) FindAll(ctx context.Context) ([]*Entity, error) {
	args := m.Called(ctx)
	
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	
	return args.Get(0).([]*Entity), args.Error(1)
}

// Usage in test:
// mockRepo := new(MockListRepository)
// entities := []*Entity{
//     {ID: uuid.New(), Name: "Entity1"},
//     {ID: uuid.New(), Name: "Entity2"},
// }
// mockRepo.On("FindAll", mock.Anything).Return(entities, nil)

// ============================================================================
// EXAMPLE 5: Complete Mock (Multiple Methods)
// ============================================================================

type FullRepository interface {
	Save(ctx context.Context, entity *Entity) error
	FindByID(ctx context.Context, id uuid.UUID) (*Entity, error)
	FindAll(ctx context.Context) ([]*Entity, error)
	Update(ctx context.Context, entity *Entity) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type MockFullRepository struct {
	mock.Mock
}

func (m *MockFullRepository) Save(ctx context.Context, entity *Entity) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockFullRepository) FindByID(ctx context.Context, id uuid.UUID) (*Entity, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Entity), args.Error(1)
}

func (m *MockFullRepository) FindAll(ctx context.Context) ([]*Entity, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Entity), args.Error(1)
}

func (m *MockFullRepository) Update(ctx context.Context, entity *Entity) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockFullRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// ============================================================================
// EXAMPLE 6: Mock with Custom Behavior
// ============================================================================

type MockSmartRepository struct {
	mock.Mock
	// You can add fields to control behavior
	shouldFail bool
}

func (m *MockSmartRepository) FindByID(ctx context.Context, id uuid.UUID) (*Entity, error) {
	// You can add custom logic before calling Called
	if m.shouldFail {
		return nil, errors.New("custom error")
	}
	
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Entity), args.Error(1)
}

// ============================================================================
// EXAMPLE 7: Complete Usage in Test
// ============================================================================

/*
func TestUseCase(t *testing.T) {
	// 1. Create the mock
	mockRepo := new(MockFullRepository)
	
	// 2. Set expectations
	userID := uuid.New()
	expectedEntity := &Entity{
		ID:   userID,
		Name: "Test Entity",
	}
	
	// Simulate that FindByID returns the entity
	mockRepo.On("FindByID", mock.Anything, userID).
		Return(expectedEntity, nil)
	
	// Simulate that Save works without error
	mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*Entity")).
		Return(nil)
	
	// 3. Use the mock in the use case
	useCase := NewUseCase(mockRepo)
	result, err := useCase.Process(ctx, userID)
	
	// 4. Verify results
	assert.NoError(t, err)
	assert.NotNil(t, result)
	
	// 5. Verify all expectations were met
	mockRepo.AssertExpectations(t)
	
	// 6. Verify specific calls (optional)
	mockRepo.AssertCalled(t, "FindByID", mock.Anything, userID)
	mockRepo.AssertNumberOfCalls(t, "FindByID", 1)
}
*/

// ============================================================================
// TIPS AND BEST PRACTICES
// ============================================================================

/*
1. ALWAYS check if args.Get(0) != nil before doing type assertion
   to avoid panics when the mock returns nil.

2. For methods that return only error, use args.Error(0) directly.

3. For methods that return (value, error), check args.Get(0) before
   doing type assertion.

4. For methods with multiple returns, the indices are:
   - args.Get(0) = first return value
   - args.Get(1) = second return value
   - ...
   - args.Error(N) = error (where N is the index of the error)

5. Use mock.Anything to accept any value in the parameter.

6. Use mock.AnythingOfType("type") to accept any value of the specified type.

7. Always call AssertExpectations(t) at the end of the test to ensure
   all expectations were met.

8. For simple mocks (1-3 methods), creating manually is faster.
   For complex interfaces (10+ methods), consider using mockery.
*/
