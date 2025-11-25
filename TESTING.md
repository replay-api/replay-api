# Testing Principles

## Core Principle: NO MOCKS!

**NEVER use mocks in this codebase.** This is a fundamental architectural decision.

### Why No Mocks?

1. **Real Behavior**: Tests should verify actual system behavior, not mocked approximations
2. **Integration Confidence**: Real dependencies catch real issues that mocks hide
3. **Refactoring Safety**: Tests remain valid when internal implementations change
4. **Maintainability**: No mock setup/teardown code to maintain
5. **True Testing**: If a test needs a database, use a real database (containerized, in-memory, or test instance)

### Testing Strategy

#### Unit Tests
- Test pure business logic without external dependencies
- Use real value objects and entities
- Test domain validation and calculations

#### Integration Tests
- Use real database instances (Docker containers, test databases)
- Use real HTTP clients for external APIs
- Use real message queues, caches, etc.

#### Test Infrastructure
- **Docker Compose**: Spin up real dependencies for integration tests
- **Test Databases**: Use separate test MongoDB instances
- **Test Isolation**: Clean up data between tests, not dependencies

### Examples

✅ **CORRECT - Real Database**
```go
func TestUserRepository(t *testing.T) {
    // Use real MongoDB test instance
    client := setupTestMongoDB(t)
    defer cleanupTestMongoDB(t, client)

    repo := NewUserRepository(client)
    // Test with real database operations
}
```

❌ **WRONG - Mocked Database**
```go
func TestUserRepository(t *testing.T) {
    mockDB := &MockDatabase{}  // NEVER DO THIS!
    mockDB.On("Find", ...).Return(...)

    repo := NewUserRepository(mockDB)
}
```

### Running Tests

**Unit Tests** (no external dependencies):
```bash
go test ./pkg/domain/...
```

**Integration Tests** (require infrastructure):

1. **Start MongoDB for tests:**
```bash
# Using Docker (recommended for local development)
docker run -d --name replay-api-test-mongo -p 37019:27017 mongo:7

# Or use the existing MongoDB instance
export MONGO_URI=mongodb://host.docker.internal:37019
```

2. **Run tests:**
```bash
# Run all tests
go test -v ./...

# Run tests in Docker (automatically uses host.docker.internal)
docker run --rm \
  -v "$(pwd)":/app \
  -w /app \
  --add-host host.docker.internal:host-gateway \
  golang:1.23 \
  go test -v ./...
```

3. **Cleanup:**
```bash
docker stop replay-api-test-mongo
docker rm replay-api-test-mongo
```

**MongoDB Configuration:**

Tests read MongoDB connection from environment variables:
- `MONGO_URI`: MongoDB connection string (default: `mongodb://host.docker.internal:37019`)
- Tests automatically append the database name (`/replay`)

**Local vs Docker Testing:**
- **Local tests**: Use `MONGO_URI=mongodb://127.0.0.1:37019` for local MongoDB
- **Docker tests**: Use `host.docker.internal:37019` (default, works with `--add-host` flag)

### Test Failures

If integration tests fail due to missing infrastructure:
1. **Start the required services** (MongoDB, Redis, etc.)
2. **Do NOT mock the dependency**
3. **Fix the test infrastructure setup**

### DRY Principle for Tests

- Reuse test infrastructure setup functions
- Share test database cleanup utilities
- Create test data builders (real objects, not mocks)
- Use table-driven tests for multiple scenarios

---

**Remember: Real dependencies, real confidence. NO MOCKS!**
