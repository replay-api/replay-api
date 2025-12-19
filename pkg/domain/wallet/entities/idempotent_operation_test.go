package wallet_entities

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// OperationStatus Constants Tests
// =============================================================================

func TestOperationStatus_Constants(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(OperationStatus("Processing"), OperationStatusProcessing)
	assert.Equal(OperationStatus("Completed"), OperationStatusCompleted)
	assert.Equal(OperationStatus("Failed"), OperationStatusFailed)
}

// =============================================================================
// NewIdempotentOperation Tests
// =============================================================================

func TestNewIdempotentOperation_CreatesValidOperation(t *testing.T) {
	assert := assert.New(t)
	key := "deposit-123-abc"
	opType := "Deposit"
	payload := map[string]interface{}{"amount": 100.00, "currency": "USD"}

	op := NewIdempotentOperation(key, opType, payload)

	assert.NotNil(op)
	assert.Equal(key, op.Key)
	assert.Equal(opType, op.OperationType)
	assert.Equal(OperationStatusProcessing, op.Status)
	assert.Equal(payload, op.RequestPayload)
	assert.Nil(op.ResponsePayload)
	assert.Nil(op.ResultID)
	assert.Empty(op.ErrorMessage)
	assert.Equal(1, op.AttemptCount)
	assert.NotNil(op.LastAttemptAt)
}

func TestNewIdempotentOperation_SetsExpiry24Hours(t *testing.T) {
	assert := assert.New(t)
	before := time.Now().UTC()

	op := NewIdempotentOperation("key-1", "Test", nil)

	// Expires should be approximately 24 hours from now
	expectedExpiry := before.Add(24 * time.Hour)
	assert.WithinDuration(expectedExpiry, op.ExpiresAt, time.Minute)
}

func TestNewIdempotentOperation_SetsTimestamps(t *testing.T) {
	assert := assert.New(t)
	before := time.Now().UTC()

	op := NewIdempotentOperation("key-2", "Test", nil)

	assert.WithinDuration(before, op.CreatedAt, time.Second)
	assert.NotNil(op.LastAttemptAt)
	assert.WithinDuration(before, *op.LastAttemptAt, time.Second)
	assert.Nil(op.CompletedAt)
}

// =============================================================================
// MarkCompleted Tests
// =============================================================================

func TestMarkCompleted_SetsCompletedStatus(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Deposit", nil)
	resultID := uuid.New()
	response := map[string]string{"status": "success"}

	op.MarkCompleted(resultID, response)

	assert.Equal(OperationStatusCompleted, op.Status)
	assert.NotNil(op.ResultID)
	assert.Equal(resultID, *op.ResultID)
	assert.Equal(response, op.ResponsePayload)
	assert.NotNil(op.CompletedAt)
}

func TestMarkCompleted_SetsCompletionTime(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Deposit", nil)
	before := time.Now().UTC()

	op.MarkCompleted(uuid.New(), nil)

	assert.WithinDuration(before, *op.CompletedAt, time.Second)
}

// =============================================================================
// MarkFailed Tests
// =============================================================================

func TestMarkFailed_SetsFailedStatus(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Deposit", nil)

	op.MarkFailed("Insufficient funds")

	assert.Equal(OperationStatusFailed, op.Status)
	assert.Equal("Insufficient funds", op.ErrorMessage)
}

func TestMarkFailed_PreservesOtherFields(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key-preserve", "Withdrawal", map[string]int{"amount": 500})
	originalKey := op.Key
	originalPayload := op.RequestPayload

	op.MarkFailed("Bank rejected")

	assert.Equal(originalKey, op.Key)
	assert.Equal(originalPayload, op.RequestPayload)
	assert.Equal(OperationStatusFailed, op.Status)
}

// =============================================================================
// IncrementAttempt Tests
// =============================================================================

func TestIncrementAttempt_IncreasesCounter(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Test", nil)
	assert.Equal(1, op.AttemptCount)

	op.IncrementAttempt()

	assert.Equal(2, op.AttemptCount)

	op.IncrementAttempt()
	op.IncrementAttempt()

	assert.Equal(4, op.AttemptCount)
}

func TestIncrementAttempt_UpdatesLastAttemptAt(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Test", nil)
	originalAttemptTime := *op.LastAttemptAt

	time.Sleep(time.Millisecond * 10)
	op.IncrementAttempt()

	assert.True(op.LastAttemptAt.After(originalAttemptTime))
}

// =============================================================================
// IsProcessing Tests
// =============================================================================

func TestIsProcessing_ReturnsTrueWhenProcessing(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Test", nil)

	assert.True(op.IsProcessing())
}

func TestIsProcessing_ReturnsFalseWhenCompleted(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Test", nil)
	op.MarkCompleted(uuid.New(), nil)

	assert.False(op.IsProcessing())
}

func TestIsProcessing_ReturnsFalseWhenFailed(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Test", nil)
	op.MarkFailed("Error")

	assert.False(op.IsProcessing())
}

// =============================================================================
// IsCompleted Tests
// =============================================================================

func TestIsCompleted_ReturnsTrueWhenCompleted(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Test", nil)
	op.MarkCompleted(uuid.New(), nil)

	assert.True(op.IsCompleted())
}

func TestIsCompleted_ReturnsFalseWhenProcessing(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Test", nil)

	assert.False(op.IsCompleted())
}

func TestIsCompleted_ReturnsFalseWhenFailed(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Test", nil)
	op.MarkFailed("Error")

	assert.False(op.IsCompleted())
}

// =============================================================================
// IsFailed Tests
// =============================================================================

func TestIsFailed_ReturnsTrueWhenFailed(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Test", nil)
	op.MarkFailed("Error")

	assert.True(op.IsFailed())
}

func TestIsFailed_ReturnsFalseWhenProcessing(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Test", nil)

	assert.False(op.IsFailed())
}

func TestIsFailed_ReturnsFalseWhenCompleted(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Test", nil)
	op.MarkCompleted(uuid.New(), nil)

	assert.False(op.IsFailed())
}

// =============================================================================
// CanRetry Tests
// =============================================================================

func TestCanRetry_ReturnsTrueWhenFailedAndUnderLimit(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Test", nil)
	op.MarkFailed("Temporary error")

	assert.True(op.CanRetry(3)) // Under 3 attempts
}

func TestCanRetry_ReturnsFalseWhenAtLimit(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Test", nil)
	op.IncrementAttempt() // 2
	op.IncrementAttempt() // 3
	op.MarkFailed("Error")

	assert.False(op.CanRetry(3)) // At 3 attempts
}

func TestCanRetry_ReturnsFalseWhenOverLimit(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Test", nil)
	op.IncrementAttempt() // 2
	op.IncrementAttempt() // 3
	op.IncrementAttempt() // 4
	op.MarkFailed("Error")

	assert.False(op.CanRetry(3)) // Over 3 attempts
}

func TestCanRetry_ReturnsFalseWhenCompleted(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Test", nil)
	op.MarkCompleted(uuid.New(), nil)

	assert.False(op.CanRetry(10)) // Completed, can't retry
}

func TestCanRetry_ReturnsFalseWhenProcessing(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Test", nil)

	assert.False(op.CanRetry(10)) // Still processing, not failed
}

// =============================================================================
// IsStale Tests
// =============================================================================

func TestIsStale_ReturnsFalseWhenNotProcessing(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Test", nil)
	op.MarkCompleted(uuid.New(), nil)

	assert.False(op.IsStale(time.Second))
}

func TestIsStale_ReturnsFalseWhenRecentlyStarted(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Test", nil)

	assert.False(op.IsStale(time.Hour))
}

func TestIsStale_ReturnsTrueWhenProcessingTooLong(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Test", nil)

	// Simulate old creation time
	oldTime := time.Now().Add(-time.Hour * 2)
	op.CreatedAt = oldTime
	op.LastAttemptAt = &oldTime

	assert.True(op.IsStale(time.Hour))
}

func TestIsStale_UsesLastAttemptAtIfAvailable(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Test", nil)

	// Created long ago
	op.CreatedAt = time.Now().Add(-time.Hour * 2)

	// But last attempt was recent
	recent := time.Now().Add(-time.Minute)
	op.LastAttemptAt = &recent

	assert.False(op.IsStale(time.Hour))
}

func TestIsStale_UsesCreatedAtIfNoLastAttempt(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Test", nil)

	// Created long ago with no last attempt
	op.CreatedAt = time.Now().Add(-time.Hour * 2)
	op.LastAttemptAt = nil

	assert.True(op.IsStale(time.Hour))
}

// =============================================================================
// GetElapsedTime Tests
// =============================================================================

func TestGetElapsedTime_ReturnsCompletionDuration(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Test", nil)
	createdAt := time.Now().Add(-time.Minute * 5)
	completedAt := time.Now().Add(-time.Minute * 2)
	op.CreatedAt = createdAt
	op.CompletedAt = &completedAt

	elapsed := op.GetElapsedTime()

	expected := completedAt.Sub(createdAt)
	assert.Equal(expected, elapsed)
}

func TestGetElapsedTime_ReturnsSinceCreatedIfNotCompleted(t *testing.T) {
	assert := assert.New(t)
	op := NewIdempotentOperation("key", "Test", nil)
	op.CreatedAt = time.Now().Add(-time.Minute * 5)

	elapsed := op.GetElapsedTime()

	// Should be approximately 5 minutes
	assert.InDelta(5*time.Minute, elapsed, float64(time.Second*10))
}

// =============================================================================
// Business Scenario Tests - E-Sports Platform Specific
// =============================================================================

func TestScenario_DepositIdempotency(t *testing.T) {
	assert := assert.New(t)

	// Simulate duplicate deposit request detection
	idempotencyKey := "deposit-user-123-tx-abc"
	payload := map[string]interface{}{
		"user_id":  "user-123",
		"amount":   100.00,
		"currency": "USD",
		"method":   "bank_transfer",
	}

	// First request: create operation
	op := NewIdempotentOperation(idempotencyKey, "Deposit", payload)
	assert.True(op.IsProcessing())

	// Simulate successful processing
	resultID := uuid.New()
	op.MarkCompleted(resultID, map[string]string{"status": "success", "tx_id": resultID.String()})

	assert.True(op.IsCompleted())
	assert.Equal(resultID, *op.ResultID)

	// Second request with same key would find this completed operation
	// and return the cached result instead of processing again
}

func TestScenario_WithdrawalRetry(t *testing.T) {
	assert := assert.New(t)

	// Withdrawal fails initially due to network issue
	op := NewIdempotentOperation("withdrawal-456", "Withdrawal", nil)

	// First attempt fails
	op.MarkFailed("Network timeout connecting to payment provider")
	assert.True(op.IsFailed())
	assert.True(op.CanRetry(3))

	// Retry
	op.IncrementAttempt()
	op.Status = OperationStatusProcessing // Reset for retry
	assert.Equal(2, op.AttemptCount)

	// Second attempt also fails
	op.MarkFailed("Payment provider returned 503")
	assert.True(op.CanRetry(3))

	// Third attempt
	op.IncrementAttempt()
	op.Status = OperationStatusProcessing
	assert.Equal(3, op.AttemptCount)

	// Third attempt succeeds
	op.MarkCompleted(uuid.New(), nil)
	assert.True(op.IsCompleted())
	assert.False(op.CanRetry(3))
}

func TestScenario_PrizeDistributionWithMaxRetries(t *testing.T) {
	assert := assert.New(t)

	op := NewIdempotentOperation("prize-match-789", "PrizeDistribution", map[string]interface{}{
		"match_id": "match-789",
		"winners":  []string{"player-1", "player-2"},
		"prize":    500.00,
	})

	maxRetries := 3

	// Exhaust all retries
	for i := 0; i < maxRetries; i++ {
		op.MarkFailed("Blockchain network congestion")
		if i < maxRetries-1 {
			assert.True(op.CanRetry(maxRetries))
			op.IncrementAttempt()
			op.Status = OperationStatusProcessing
		}
	}

	// Can't retry anymore
	assert.False(op.CanRetry(maxRetries))
	assert.Equal(maxRetries, op.AttemptCount)
	assert.True(op.IsFailed())
}

func TestScenario_StaleOperationDetection(t *testing.T) {
	assert := assert.New(t)

	// Simulate an operation that started but never completed (process crashed)
	op := NewIdempotentOperation("stuck-op", "EntryFee", nil)

	// Manually set old timestamps to simulate stale operation
	staleTime := time.Now().Add(-time.Hour * 6)
	op.CreatedAt = staleTime
	op.LastAttemptAt = &staleTime

	// Should be detected as stale (threshold: 5 minutes)
	assert.True(op.IsStale(5 * time.Minute))

	// Cleanup process would reset and retry
	op.IncrementAttempt()
	op.Status = OperationStatusProcessing

	// No longer stale after retry
	assert.False(op.IsStale(5 * time.Minute))
}

