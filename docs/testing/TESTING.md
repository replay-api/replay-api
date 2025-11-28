# Replay API - Testing Guide

> Testing strategy and procedures for the LeetGaming Replay API

## Overview

The Replay API uses a **production-grade testing strategy** with real services. No mocks are used for integration and E2E tests.

```
┌─────────────────────────────────────────────────────────────────────────┐
│                       TESTING PYRAMID                                    │
│                                                                          │
│                         ┌─────────┐                                     │
│                         │  E2E    │  ← Real MongoDB + Hardhat EVM       │
│                         │  Tests  │                                     │
│                        ─┴─────────┴─                                    │
│                      ┌───────────────┐                                  │
│                      │  Integration  │  ← Real MongoDB (testcontainers) │
│                      │    Tests      │                                  │
│                     ─┴───────────────┴─                                 │
│                   ┌─────────────────────┐                               │
│                   │     Unit Tests      │  ← Mocked dependencies only   │
│                   │                     │                               │
│                  ─┴─────────────────────┴─                              │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Unit Tests

Run unit tests only:

```bash
make test-unit
```

**Coverage Target**: 80%

Unit tests mock external dependencies using `testify/mock`:

```go
func TestCreateLobbyUseCase(t *testing.T) {
    // Arrange
    mockWriter := new(MockLobbyWriter)
    mockWriter.On("Create", mock.Anything, mock.Anything).Return(&lobby, nil)
    
    useCase := NewCreateLobbyUseCase(mockWriter)
    
    // Act
    result, err := useCase.Execute(ctx, command)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    mockWriter.AssertExpectations(t)
}
```

---

## Integration Tests

Integration tests use **real MongoDB** via Docker:

```bash
make test-integration
```

This starts a MongoDB container and runs tests against it.

**Test Setup:**

```go
func TestMain(m *testing.M) {
    // Start MongoDB container
    container, _ := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "mongo:7.0",
            ExposedPorts: []string{"27017/tcp"},
        },
        Started: true,
    })
    
    // Run tests
    code := m.Run()
    
    // Cleanup
    container.Terminate(ctx)
    os.Exit(code)
}
```

---

## E2E Tests

End-to-end tests use **real services** with no mocks:

```bash
# Start test infrastructure
make -f Makefile.test test-setup

# Run E2E tests
make -f Makefile.test test-e2e
```

### Test Infrastructure

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    E2E TEST INFRASTRUCTURE                               │
│                                                                          │
│  docker-compose.test.yml                                                │
│                                                                          │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────────┐ │
│  │   MongoDB       │  │    Hardhat      │  │      Replay API         │ │
│  │   (Port 27018)  │  │  (Port 8545)    │  │     (Port 8080)         │ │
│  │                 │  │                 │  │                         │ │
│  │  • Test data    │  │  • Mock USDC    │  │  • Full API             │ │
│  │  • Clean state  │  │  • Mock USDT    │  │  • Real services        │ │
│  │                 │  │  • Mock NFT     │  │                         │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────┘
```

### Hardhat Smart Contracts

For testing crypto operations, we deploy mock ERC-20 and ERC-721 contracts:

```solidity
// MockUSDC.sol
contract MockUSDC is ERC20 {
    function decimals() public pure override returns (uint8) {
        return 6; // Matches real USDC
    }
    
    function faucet(address to, uint256 amount) external {
        _mint(to, amount);
    }
}
```

---

## Test Categories

### Wallet Tests

```bash
make test-wallet
```

- Deposit with double-entry ledger validation
- Withdrawal with balance checks
- Reconciliation (wallet matches ledger)
- Idempotency (duplicate detection)
- Entry fee with insufficient balance rollback
- Prize winning with daily limit enforcement
- Transaction history with pagination

### Matchmaking Tests

```bash
make test-matchmaking
```

- Lobby creation
- Player join/leave
- Ready check flow
- Match start
- Prize pool accumulation
- Prize distribution

### Tournament Tests

```bash
make test-tournament
```

- Tournament creation
- Player registration
- Bracket generation
- Match progression
- Prize distribution

---

## Running Tests

### All Tests

```bash
make test
```

### With Coverage

```bash
make test-coverage
```

Coverage report: `coverage/coverage.html`

### CI Pipeline

```bash
make -f Makefile.test test-ci
```

This runs:
1. Start test infrastructure
2. Wait for services
3. Run all tests with coverage
4. Generate coverage report
5. Cleanup

---

## Test Data

### Seed Data

The `e2e/db-init/01-seed-data.js` script creates initial test data:

```javascript
db.users.insertOne({
    _id: UUID("550e8400-e29b-41d4-a716-446655440000"),
    steam_id: "76561198012345678",
    email: "test@leetgaming.pro",
    created_at: new Date()
});

db.wallets.insertOne({
    _id: UUID("660e8400-e29b-41d4-a716-446655440001"),
    user_id: UUID("550e8400-e29b-41d4-a716-446655440000"),
    balances: { "USD": { cents: 100000 } },
    is_locked: false
});
```

### Test Users

| User ID | Email | Balance |
|---------|-------|---------|
| `550e8400-...` | test@leetgaming.pro | $1,000.00 |
| `551e8400-...` | player2@test.com | $500.00 |
| `552e8400-...` | player3@test.com | $250.00 |

---

## Benchmarks

Run performance benchmarks:

```bash
make bench
```

**Example Output:**

```
BenchmarkDeposit-8         500 ops    2847 ns/op    1456 B/op    23 allocs/op
BenchmarkWithdraw-8        450 ops    3124 ns/op    1678 B/op    26 allocs/op
BenchmarkGetBalance-8     2000 ops     890 ns/op     456 B/op    12 allocs/op
```

---

## Test Utilities

### Assert Helpers

```go
// assert_wallet.go
func AssertWalletBalance(t *testing.T, wallet *UserWallet, currency Currency, expected Amount) {
    actual := wallet.GetBalance(currency)
    assert.True(t, actual.Equals(expected), 
        "expected balance %s, got %s", expected.String(), actual.String())
}

func AssertLedgerBalanced(t *testing.T, entries []LedgerEntry) {
    var debits, credits Amount
    for _, e := range entries {
        if e.EntryType == EntryTypeDebit {
            debits = debits.Add(e.Amount)
        } else {
            credits = credits.Add(e.Amount)
        }
    }
    assert.True(t, debits.Equals(credits), "ledger not balanced")
}
```

### Test Fixtures

```go
// fixtures.go
func CreateTestWallet(t *testing.T, db *mongo.Database) *UserWallet {
    wallet, _ := NewUserWallet(testResourceOwner, testEVMAddress)
    wallet.Deposit(CurrencyUSD, NewAmount(100000)) // $1000
    
    _, err := db.Collection("wallets").InsertOne(ctx, wallet)
    require.NoError(t, err)
    
    return wallet
}
```

---

## Makefile Targets

```makefile
# Unit tests only
test-unit:
    go test -v -tags=unit ./...

# Integration tests with real MongoDB
test-integration:
    docker-compose -f docker-compose.test.yml up -d mongo
    go test -v -tags=integration ./...
    docker-compose -f docker-compose.test.yml down

# E2E tests with full infrastructure
test-e2e:
    docker-compose -f docker-compose.test.yml up -d
    ./scripts/wait-for-services.sh
    go test -v -tags=e2e ./test/...
    docker-compose -f docker-compose.test.yml down

# Coverage report
test-coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage/coverage.html

# Benchmarks
bench:
    go test -bench=. -benchmem ./...
```

---

## CI Configuration

GitHub Actions workflow:

```yaml
# .github/workflows/test.yml
jobs:
  test:
    runs-on: ubuntu-latest
    services:
      mongo:
        image: mongo:7.0
        ports:
          - 27017:27017
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - run: make test-coverage
      - uses: codecov/codecov-action@v4
        with:
          files: coverage.out
```

---

**Last Updated**: November 2025
