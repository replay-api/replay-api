# Mocking Guide

This guide explains when and how to use mocks in the LeetGaming Replay API project.

## Philosophy

**Primary Rule: Prefer Real Dependencies Over Mocks**

This project follows a philosophy of using real dependencies (databases, services, etc.) in tests whenever possible. Mocks should only be used when absolutely necessary for unit testing pure business logic.

## When to Use Mocks

### ✅ Appropriate Use Cases

1. **Pure Business Logic Testing**
   - Testing domain calculations
   - Testing validation logic
   - Testing state transitions
   - When external dependencies would add unnecessary complexity

2. **Error Scenarios**
   - Testing error handling paths
   - Simulating failure conditions
   - Testing retry logic

3. **Performance-Critical Unit Tests**
   - Fast unit tests that don't require I/O
   - Testing algorithms and data structures

### ❌ Inappropriate Use Cases

1. **Integration Tests**
   - Always use real databases, message queues, etc.
   - See `test/integration/` for examples

2. **Repository Tests**
   - Test repositories against real databases
   - Use test containers or in-memory databases

3. **Service Integration Tests**
   - Use real service instances or test doubles
   - Avoid mocking external APIs unless absolutely necessary

## Mock Structure

### Using testify/mock

The project standardizes on `github.com/stretchr/testify/mock` for mock implementations.

```go
// test/mocks/domain/billing/mock_billable_handler.go
package billing

import (
    "context"
    "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
    "github.com/stretchr/testify/mock"
)

type MockBillableOperationHandler struct {
    mock.Mock
}

func (m *MockBillableOperationHandler) Exec(ctx context.Context, cmd billing_in.BillableOperationCommand) (*billing_entities.BillableEntry, *billing_entities.Subscription, error) {
    args := m.Called(ctx, cmd)
    if args.Get(0) == nil {
        return nil, nil, args.Error(2)
    }
    return args.Get(0).(*billing_entities.BillableEntry), args.Get(1).(*billing_entities.Subscription), args.Error(2)
}
```

### Using Mocks in Tests

```go
func TestUseCase(t *testing.T) {
    // Create mock
    mockHandler := billing.NewMockBillableOperationHandler()
    
    // Set expectations
    mockHandler.On("Exec", mock.Anything, mock.Anything).
        Return(entry, subscription, nil)
    
    // Use in test
    useCase := NewUseCase(mockHandler)
    result, err := useCase.Exec(ctx, cmd)
    
    // Assert expectations
    assert.NoError(t, err)
    mockHandler.AssertExpectations(t)
}
```

## Generating Mocks

Use `mockery` to generate mocks from interfaces:

```bash
make mocks
```

This generates mocks in `test/mocks/` based on interface definitions in `pkg/domain/`.

## Best Practices

1. **Keep Mocks Simple**
   - Mocks should be thin wrappers
   - Avoid complex mock logic

2. **Use Interface-Based Mocks**
   - Mock interfaces, not concrete types
   - This maintains testability

3. **Document Mock Behavior**
   - If a mock has special behavior, document it
   - Use clear naming conventions

4. **Share Common Mocks**
   - Put reusable mocks in `test/mocks/`
   - Keep test-specific mocks in test files

5. **Verify Mock Calls**
   - Always call `AssertExpectations(t)`
   - Verify that expected methods were called

## Alternatives to Mocks

### Test Doubles

Instead of mocks, consider:

1. **Fake Implementations**
   - In-memory implementations
   - Example: In-memory repository

2. **Test Containers**
   - Real databases in containers
   - Example: MongoDB test container

3. **Stubs**
   - Simple implementations that return fixed values
   - Useful for testing error paths

## Examples

### ✅ Good: Using Mock for Business Logic

```go
func TestCalculateRating(t *testing.T) {
    mockRepo := matchmaking.NewMockRatingRepository()
    mockRepo.On("GetPlayerRating", playerID).Return(rating, nil)
    
    service := NewRatingService(mockRepo)
    result := service.CalculateNewRating(playerID, matchResult)
    
    assert.Equal(t, expectedRating, result)
}
```

### ❌ Bad: Mocking Database in Integration Test

```go
// DON'T DO THIS in integration tests
func TestCreateUser(t *testing.T) {
    mockRepo := iam.NewMockUserRepository() // ❌
    // ... use real database instead
}
```

### ✅ Good: Real Database in Integration Test

```go
//go:build integration
func TestCreateUser(t *testing.T) {
    client := setupTestMongoDB(t)
    defer cleanupTestMongoDB(t, client)
    
    repo := db.NewMongoUserRepository(client, "test_db")
    // ... test with real database
}
```

## Summary

- **Unit tests**: Mocks are acceptable for pure business logic
- **Integration tests**: Always use real dependencies
- **E2E tests**: Always use real dependencies
- **When in doubt**: Use real dependencies

For more information, see [TESTING.md](../../TESTING.md).
