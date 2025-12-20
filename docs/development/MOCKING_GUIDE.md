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

### Option 1: Generate with mockery (automatic)

Use `mockery` to generate mocks from interfaces:

```bash
make mocks
```

This generates mocks in `test/mocks/` based on interface definitions in `pkg/domain/`.

### Option 2: Create Mocks Manually (without mockery)

You can create mocks manually using `testify/mock` directly. This is a valid alternative to mockery, especially when you need full control over the mock implementation.

#### Basic Structure

A manual mock consists of:
1. A struct that embeds `mock.Mock`
2. Methods that implement the desired interface
3. Each method calls `m.Called(...)` and returns the appropriate values

#### Example: Repository Mock

```go
package wallet_usecases_test

import (
    "context"
    "github.com/google/uuid"
    "github.com/stretchr/testify/mock"
    wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
    wallet_out "github.com/replay-api/replay-api/pkg/domain/wallet/ports/out"
)

// MockWalletRepository is a mock implementation of wallet_out.WalletRepository
type MockWalletRepository struct {
    mock.Mock
}

// Save implements wallet_out.WalletRepository.Save
func (m *MockWalletRepository) Save(ctx context.Context, wallet *wallet_entities.UserWallet) error {
    args := m.Called(ctx, wallet)
    return args.Error(0)
}

// FindByID implements wallet_out.WalletRepository.FindByID
func (m *MockWalletRepository) FindByID(ctx context.Context, id uuid.UUID) (*wallet_entities.UserWallet, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*wallet_entities.UserWallet), args.Error(1)
}

// FindByUserID implements wallet_out.WalletRepository.FindByUserID
func (m *MockWalletRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*wallet_entities.UserWallet, error) {
    args := m.Called(ctx, userID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*wallet_entities.UserWallet), args.Error(1)
}
```

#### Example: Mock with Multiple Return Values

For methods that return multiple values:

```go
// MockBillableOperationHandler implements billing_in.BillableOperationCommandHandler
type MockBillableOperationHandler struct {
    mock.Mock
}

func (m *MockBillableOperationHandler) Exec(ctx context.Context, command billing_in.BillableOperationCommand) (*billing_entities.BillableEntry, *billing_entities.Subscription, error) {
    args := m.Called(ctx, command)
    
    var entry *billing_entities.BillableEntry
    var sub *billing_entities.Subscription
    
    // First return value (index 0)
    if args.Get(0) != nil {
        entry = args.Get(0).(*billing_entities.BillableEntry)
    }
    
    // Second return value (index 1)
    if args.Get(1) != nil {
        sub = args.Get(1).(*billing_entities.Subscription)
    }
    
    // Error (last parameter)
    return entry, sub, args.Error(2)
}
```

#### Using the Mock in Tests

```go
func TestUseCase(t *testing.T) {
    // Create the mock
    mockRepo := new(MockWalletRepository)
    
    // Set expectations
    expectedWallet := &wallet_entities.UserWallet{
        // ... wallet fields
    }
    mockRepo.On("FindByUserID", mock.Anything, userID).
        Return(expectedWallet, nil)
    
    // Use in use case
    useCase := NewUseCase(mockRepo)
    result, err := useCase.Exec(ctx, query)
    
    // Verify results
    assert.NoError(t, err)
    assert.NotNil(t, result)
    
    // Verify all expectations were met
    mockRepo.AssertExpectations(t)
}
```

#### Advantages of Manual Mocks

✅ **No external dependencies**: No need to install mockery  
✅ **Full control**: You define exactly how the mock behaves  
✅ **Simpler**: Less boilerplate code for small interfaces  
✅ **More readable**: You see exactly what the mock does  

#### When to Use Manual Mocks vs mockery

- **Use manual mocks** when:
  - The interface has few methods (1-5)
  - You need custom behavior
  - You don't want to install mockery
  - You're writing mocks specific to a test

- **Use mockery** when:
  - The interface has many methods
  - You want to keep mocks automatically synchronized
  - You need to generate mocks for multiple interfaces
  - The interface changes frequently

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
