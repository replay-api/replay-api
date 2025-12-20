# Test Mocks

This directory contains shared mock implementations for testing purposes.

## Philosophy

This project follows a **"NO MOCKS"** philosophy for integration and E2E tests. However, mocks may be used in **unit tests** when testing pure business logic that doesn't require external dependencies.

## When to Use Mocks

✅ **Use mocks for:**
- Unit tests of pure business logic
- Testing error handling scenarios
- Isolating domain logic from infrastructure
- Testing edge cases that are hard to reproduce with real dependencies

❌ **Do NOT use mocks for:**
- Integration tests
- E2E tests
- Testing repository implementations
- Testing external service integrations

## Structure

```
test/mocks/
├── domain/          # Domain layer mocks
│   ├── billing/     # Billing domain mocks
│   ├── matchmaking/ # Matchmaking domain mocks
│   └── ...
└── infra/           # Infrastructure layer mocks (rarely used)
```

## Usage Example

```go
// In a unit test file
import (
    "github.com/replay-api/replay-api/test/mocks/domain/billing"
)

func TestUseCase(t *testing.T) {
    mockHandler := billing.NewMockBillableOperationHandler()
    // ... use mock in test
}
```

## Generating Mocks

### Option 1: Generate with mockery (automatic)

If you need to generate mocks automatically from interfaces, use `mockery`:

```bash
make mocks
```

This will generate mocks in `test/mocks/` based on interface definitions.

### Option 2: Create mocks manually (without mockery)

You can also create mocks manually using `testify/mock` directly. This is useful when:

- The interface has few methods (1-5)
- You need custom behavior
- You don't want to install mockery
- You're writing mocks specific to a test

Basic example:

```go
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) FindByID(ctx context.Context, id uuid.UUID) (*Entity, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*Entity), args.Error(1)
}
```

For more examples and details, see [MOCKING_GUIDE.md](../../docs/development/MOCKING_GUIDE.md).

## Best Practices

1. **Keep mocks simple**: Mocks should be thin wrappers around interfaces
2. **Use testify/mock**: Standardize on `github.com/stretchr/testify/mock`
3. **Document behavior**: If a mock has special behavior, document it
4. **Share common mocks**: Put reusable mocks in this directory
5. **Test-specific mocks**: Keep test-specific mocks in the test file
