package blockchain_entities

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	blockchain_vo "github.com/replay-api/replay-api/pkg/domain/blockchain/value-objects"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

// =============================================================================
// Test Strategy: TDD-driven, functionality-focused tests
// 
// Coverage targets:
// - All public methods have at least one positive and one negative test
// - Edge cases: zero amounts, max participants, concurrent access
// - State machine: all valid transitions tested
// - Thread safety: race condition tests for concurrent operations
// =============================================================================

// TestOnChainPrizePoolStatus_Constants verifies status constants
func TestOnChainPrizePoolStatus_Constants(t *testing.T) {
	statuses := []OnChainPrizePoolStatus{
		PoolStatusNotCreated,
		PoolStatusAccumulating,
		PoolStatusLocked,
		PoolStatusInEscrow,
		PoolStatusDistributed,
		PoolStatusCancelled,
	}

	seen := make(map[OnChainPrizePoolStatus]bool)
	for _, status := range statuses {
		if status == "" {
			t.Error("OnChainPrizePoolStatus should not be empty")
		}
		if seen[status] {
			t.Errorf("Duplicate OnChainPrizePoolStatus: %s", status)
		}
		seen[status] = true
	}
}

// TestNewOnChainPrizePool verifies prize pool creation
func TestNewOnChainPrizePool(t *testing.T) {
	owner := testResourceOwner()
	matchID := uuid.New()
	chainID := blockchain_vo.ChainIDPolygon
	contractAddr, _ := wallet_vo.NewEVMAddress("0x1234567890123456789012345678901234567890")
	tokenAddr, _ := wallet_vo.NewEVMAddress("0xabcdef1234567890abcdef1234567890abcdef12")
	currency := wallet_vo.CurrencyUSDC
	entryFee := wallet_vo.NewAmountFromCents(1000) // 1000 cents = $10.00
	platformFeePercent := uint16(1000)    // 10%

	pool := NewOnChainPrizePool(owner, matchID, chainID, contractAddr, tokenAddr, currency, entryFee, platformFeePercent)

	if pool == nil {
		t.Fatal("Expected non-nil prize pool")
	}
	if pool.MatchID != matchID {
		t.Error("MatchID mismatch")
	}
	if pool.ChainID != chainID {
		t.Errorf("ChainID = %d, want %d", pool.ChainID, chainID)
	}
	if pool.Status != PoolStatusNotCreated {
		t.Errorf("Status = %s, want %s", pool.Status, PoolStatusNotCreated)
	}
	if pool.PlatformFeePercent != 1000 {
		t.Errorf("PlatformFeePercent = %d, want 1000", pool.PlatformFeePercent)
	}
	if len(pool.Participants) != 0 {
		t.Error("Expected empty participants list")
	}
}

// TestOnChainPrizePool_MarkCreated verifies pool creation marking
func TestOnChainPrizePool_MarkCreated(t *testing.T) {
	pool := createTestPrizePool()
	txHash, _ := blockchain_vo.NewTxHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	contribution := wallet_vo.NewAmountFromCents(5000)

	pool.MarkCreated(txHash, contribution)

	if pool.Status != PoolStatusAccumulating {
		t.Errorf("Status = %s, want %s", pool.Status, PoolStatusAccumulating)
	}
	if pool.CreateTxHash == nil {
		t.Error("CreateTxHash should be set")
	}
	if pool.PlatformContribution.Cents() != 5000 {
		t.Errorf("PlatformContribution = %d, want 5000", pool.PlatformContribution.Cents())
	}
}

// TestOnChainPrizePool_AddParticipant verifies participant addition
func TestOnChainPrizePool_AddParticipant(t *testing.T) {
	pool := createTestPrizePool()
	txHash, _ := blockchain_vo.NewTxHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	pool.MarkCreated(txHash, wallet_vo.NewAmountFromCents(0))

	addr, _ := wallet_vo.NewEVMAddress("0x1111111111111111111111111111111111111111")
	contribution := wallet_vo.NewAmountFromCents(1000)

	err := pool.AddParticipant(addr, contribution)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(pool.Participants) != 1 {
		t.Errorf("Participants count = %d, want 1", len(pool.Participants))
	}
	if pool.TotalAmount.Cents() != 1000 {
		t.Errorf("TotalAmount = %d, want 1000", pool.TotalAmount.Cents())
	}
}

// TestOnChainPrizePool_AddParticipant_Duplicate verifies duplicate rejection.
// Business rule: Each address can only participate once per match.
// This prevents double-spending and ensures fair prize distribution.
func TestOnChainPrizePool_AddParticipant_Duplicate(t *testing.T) {
	pool := createTestPrizePool()
	txHash, _ := blockchain_vo.NewTxHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	pool.MarkCreated(txHash, wallet_vo.NewAmountFromCents(0))

	addr, _ := wallet_vo.NewEVMAddress("0x1111111111111111111111111111111111111111")
	contribution := wallet_vo.NewAmountFromCents(1000)

	// First participation should succeed
	err := pool.AddParticipant(addr, contribution)
	if err != nil {
		t.Fatalf("First AddParticipant should succeed: %v", err)
	}

	// Second attempt with same address should fail (duplicate detection)
	err = pool.AddParticipant(addr, contribution)
	if err == nil {
		t.Error("Expected error for duplicate participant - same address cannot join twice")
	}
}

// TestOnChainPrizePool_AddParticipant_WrongStatus verifies status check
func TestOnChainPrizePool_AddParticipant_WrongStatus(t *testing.T) {
	pool := createTestPrizePool()
	// Pool is NotCreated status

	addr, _ := wallet_vo.NewEVMAddress("0x1111111111111111111111111111111111111111")
	contribution := wallet_vo.NewAmountFromCents(1000)

	err := pool.AddParticipant(addr, contribution)

	if err == nil {
		t.Error("Expected error for wrong status")
	}
}

// TestOnChainPrizePool_Lock verifies pool locking
func TestOnChainPrizePool_Lock(t *testing.T) {
	pool := createPoolWithParticipants(2)
	lockTxHash, _ := blockchain_vo.NewTxHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")

	err := pool.Lock(lockTxHash)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if pool.Status != PoolStatusLocked {
		t.Errorf("Status = %s, want %s", pool.Status, PoolStatusLocked)
	}
	if pool.LockTxHash == nil {
		t.Error("LockTxHash should be set")
	}
	if pool.LockedAt == nil {
		t.Error("LockedAt should be set")
	}
}

// TestOnChainPrizePool_Lock_NotEnoughParticipants verifies minimum participants
func TestOnChainPrizePool_Lock_NotEnoughParticipants(t *testing.T) {
	pool := createPoolWithParticipants(1)
	lockTxHash, _ := blockchain_vo.NewTxHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")

	err := pool.Lock(lockTxHash)

	if err == nil {
		t.Error("Expected error for not enough participants")
	}
}

// TestOnChainPrizePool_StartEscrow verifies escrow period start
func TestOnChainPrizePool_StartEscrow(t *testing.T) {
	pool := createPoolWithParticipants(2)
	lockTxHash, _ := blockchain_vo.NewTxHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")
	_ = pool.Lock(lockTxHash)

	err := pool.StartEscrow(24 * time.Hour)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if pool.Status != PoolStatusInEscrow {
		t.Errorf("Status = %s, want %s", pool.Status, PoolStatusInEscrow)
	}
	if pool.EscrowEndTime == nil {
		t.Error("EscrowEndTime should be set")
	}
}

// TestOnChainPrizePool_StartEscrow_WrongStatus verifies status check
func TestOnChainPrizePool_StartEscrow_WrongStatus(t *testing.T) {
	pool := createPoolWithParticipants(2)
	// Pool is Accumulating, not Locked

	err := pool.StartEscrow(24 * time.Hour)

	if err == nil {
		t.Error("Expected error for wrong status")
	}
}

// TestOnChainPrizePool_Distribute verifies prize distribution
func TestOnChainPrizePool_Distribute(t *testing.T) {
	pool := createPoolInEscrow()
	distributeTxHash, _ := blockchain_vo.NewTxHash("0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321")
	winnerAddr, _ := wallet_vo.NewEVMAddress("0x1111111111111111111111111111111111111111")

	winners := []PrizeWinner{
		{
			Address:  winnerAddr,
			Rank:     1,
			Amount:   wallet_vo.NewAmountFromCents(1800), // 90% of pool (1800 cents)
			ShareBPS: 9000,
		},
	}
	platformFee := wallet_vo.NewAmountFromCents(200) // 10% (200 cents)

	err := pool.Distribute(distributeTxHash, winners, platformFee)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if pool.Status != PoolStatusDistributed {
		t.Errorf("Status = %s, want %s", pool.Status, PoolStatusDistributed)
	}
	if pool.DistributeTxHash == nil {
		t.Error("DistributeTxHash should be set")
	}
	if len(pool.Winners) != 1 {
		t.Errorf("Winners count = %d, want 1", len(pool.Winners))
	}
}

// TestOnChainPrizePool_Distribute_WrongStatus verifies status check
func TestOnChainPrizePool_Distribute_WrongStatus(t *testing.T) {
	pool := createPoolWithParticipants(2)
	distributeTxHash, _ := blockchain_vo.NewTxHash("0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321")

	err := pool.Distribute(distributeTxHash, []PrizeWinner{}, wallet_vo.NewAmountFromCents(0))

	if err == nil {
		t.Error("Expected error for wrong status")
	}
}

// TestOnChainPrizePool_Cancel verifies pool cancellation
func TestOnChainPrizePool_Cancel(t *testing.T) {
	pool := createPoolWithParticipants(2)

	err := pool.Cancel()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if pool.Status != PoolStatusCancelled {
		t.Errorf("Status = %s, want %s", pool.Status, PoolStatusCancelled)
	}
}

// TestOnChainPrizePool_Cancel_WrongStatus verifies status check
func TestOnChainPrizePool_Cancel_WrongStatus(t *testing.T) {
	pool := createPoolInEscrow()

	err := pool.Cancel()

	if err == nil {
		t.Error("Expected error for wrong status")
	}
}

// TestOnChainPrizePool_UpdateSyncState verifies sync tracking
func TestOnChainPrizePool_UpdateSyncState(t *testing.T) {
	pool := createTestPrizePool()

	pool.UpdateSyncState(12345, true)

	if pool.LastSyncBlock != 12345 {
		t.Errorf("LastSyncBlock = %d, want 12345", pool.LastSyncBlock)
	}
	if !pool.IsSynced {
		t.Error("IsSynced should be true")
	}
}

// TestOnChainPrizePool_GetOnChainIDHex verifies hex formatting
func TestOnChainPrizePool_GetOnChainIDHex(t *testing.T) {
	pool := createTestPrizePool()

	hexID := pool.GetOnChainIDHex()

	if hexID[:2] != "0x" {
		t.Errorf("GetOnChainIDHex should start with 0x, got %s", hexID)
	}
}

// TestOnChainPrizePool_IsEscrowComplete verifies escrow completion check
func TestOnChainPrizePool_IsEscrowComplete(t *testing.T) {
	pool := createPoolInEscrow()

	// Not complete immediately
	if pool.IsEscrowComplete() {
		t.Error("Escrow should not be complete immediately")
	}

	// Simulate time passing
	pastTime := time.Now().Add(-1 * time.Hour)
	pool.EscrowEndTime = &pastTime

	if !pool.IsEscrowComplete() {
		t.Error("Escrow should be complete after end time")
	}
}

// TestOnChainPrizePool_IsEscrowComplete_WrongStatus verifies status check
func TestOnChainPrizePool_IsEscrowComplete_WrongStatus(t *testing.T) {
	pool := createPoolWithParticipants(2)

	if pool.IsEscrowComplete() {
		t.Error("IsEscrowComplete should return false for non-escrow pool")
	}
}

// TestOnChainPrizePool_GetDistributableAmount verifies distribution calculation
func TestOnChainPrizePool_GetDistributableAmount(t *testing.T) {
	pool := createTestPrizePool()
	pool.TotalAmount = wallet_vo.NewAmountFromCents(10000) // 10000 cents = $100.00
	pool.PlatformFeePercent = 1000                          // 10%

	distributable := pool.GetDistributableAmount()

	// 10000 - (10000 * 1000 / 10000) = 10000 - 1000 = 9000 cents
	if distributable.Cents() != 9000 {
		t.Errorf("GetDistributableAmount = %d, want 9000", distributable.Cents())
	}
}

// TestPrizeWinner_Structure verifies winner structure
func TestPrizeWinner_Structure(t *testing.T) {
	addr, _ := wallet_vo.NewEVMAddress("0x1111111111111111111111111111111111111111")
	now := time.Now()

	winner := PrizeWinner{
		Address:     addr,
		Rank:        1,
		Amount:      wallet_vo.NewAmountFromCents(5000),
		ShareBPS:    5000,
		IsMVP:       true,
		WithdrawnAt: &now,
	}

	if winner.Rank != 1 {
		t.Errorf("Rank = %d, want 1", winner.Rank)
	}
	if winner.ShareBPS != 5000 {
		t.Errorf("ShareBPS = %d, want 5000", winner.ShareBPS)
	}
	if !winner.IsMVP {
		t.Error("IsMVP should be true")
	}
}

// Helper functions

func createTestPrizePool() *OnChainPrizePool {
	owner := testResourceOwner()
	matchID := uuid.New()
	chainID := blockchain_vo.ChainIDPolygon
	contractAddr, _ := wallet_vo.NewEVMAddress("0x1234567890123456789012345678901234567890")
	tokenAddr, _ := wallet_vo.NewEVMAddress("0xabcdef1234567890abcdef1234567890abcdef12")

	return NewOnChainPrizePool(owner, matchID, chainID, contractAddr, tokenAddr, wallet_vo.CurrencyUSDC, wallet_vo.NewAmountFromCents(1000), 1000)
}

func createPoolWithParticipants(count int) *OnChainPrizePool {
	pool := createTestPrizePool()
	txHash, _ := blockchain_vo.NewTxHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	pool.MarkCreated(txHash, wallet_vo.NewAmountFromCents(0))

	for i := 0; i < count; i++ {
		addrHex := "0x" + string(rune('1'+i)) + "111111111111111111111111111111111111111"
		addr, _ := wallet_vo.NewEVMAddress(addrHex[:42])
		_ = pool.AddParticipant(addr, wallet_vo.NewAmountFromCents(1000))
	}

	return pool
}

func createPoolInEscrow() *OnChainPrizePool {
	pool := createPoolWithParticipants(2)
	lockTxHash, _ := blockchain_vo.NewTxHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")
	_ = pool.Lock(lockTxHash)
	_ = pool.StartEscrow(24 * time.Hour)
	return pool
}

// =============================================================================
// Table-Driven Tests: State Transitions
// =============================================================================

// TestOnChainPrizePool_StateTransitions validates the pool state machine.
// State flow: NotCreated -> Accumulating -> Locked -> InEscrow -> Distributed
func TestOnChainPrizePool_StateTransitions(t *testing.T) {
	tests := []struct {
		name          string
		setupPool     func() *OnChainPrizePool
		action        string
		expectedState OnChainPrizePoolStatus
		shouldError   bool
	}{
		{
			name:          "MarkCreated transitions NotCreated to Accumulating",
			setupPool:     createTestPrizePool,
			action:        "MarkCreated",
			expectedState: PoolStatusAccumulating,
			shouldError:   false,
		},
		{
			name:          "Lock transitions Accumulating to Locked (with 2+ participants)",
			setupPool:     func() *OnChainPrizePool { return createPoolWithParticipants(2) },
			action:        "Lock",
			expectedState: PoolStatusLocked,
			shouldError:   false,
		},
		{
			name:          "Lock fails with only 1 participant",
			setupPool:     func() *OnChainPrizePool { return createPoolWithParticipants(1) },
			action:        "Lock",
			expectedState: PoolStatusAccumulating,
			shouldError:   true,
		},
		{
			name:          "Cancel works from Accumulating",
			setupPool:     func() *OnChainPrizePool { return createPoolWithParticipants(1) },
			action:        "Cancel",
			expectedState: PoolStatusCancelled,
			shouldError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := tt.setupPool()
			var err error

			switch tt.action {
			case "MarkCreated":
				txHash, _ := blockchain_vo.NewTxHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
				pool.MarkCreated(txHash, wallet_vo.NewAmountFromCents(0))
			case "Lock":
				txHash, _ := blockchain_vo.NewTxHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")
				err = pool.Lock(txHash)
			case "Cancel":
				err = pool.Cancel()
			}

			if tt.shouldError && err == nil {
				t.Errorf("Expected error for %s", tt.action)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.action, err)
			}
			if pool.GetStatus() != tt.expectedState {
				t.Errorf("Status = %s, want %s", pool.GetStatus(), tt.expectedState)
			}
		})
	}
}

// =============================================================================
// Edge Case Tests
// =============================================================================

// TestOnChainPrizePool_ZeroContribution verifies zero contribution handling.
func TestOnChainPrizePool_ZeroContribution(t *testing.T) {
	pool := createTestPrizePool()
	txHash, _ := blockchain_vo.NewTxHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	pool.MarkCreated(txHash, wallet_vo.NewAmountFromCents(0))

	addr, _ := wallet_vo.NewEVMAddress("0x1111111111111111111111111111111111111111")
	zeroContribution := wallet_vo.NewAmountFromCents(0)

	err := pool.AddParticipant(addr, zeroContribution)
	if err != nil {
		t.Errorf("Zero contribution should be allowed (for free tournaments): %v", err)
	}
	if pool.ParticipantCount() != 1 {
		t.Errorf("Participant count = %d, want 1", pool.ParticipantCount())
	}
}

// TestOnChainPrizePool_MaxParticipants verifies large participant handling.
func TestOnChainPrizePool_MaxParticipants(t *testing.T) {
	pool := createTestPrizePool()
	txHash, _ := blockchain_vo.NewTxHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	pool.MarkCreated(txHash, wallet_vo.NewAmountFromCents(0))

	const maxParticipants = 100 // Typical CS2 tournament size limit
	contribution := wallet_vo.NewAmountFromCents(100)

	for i := 0; i < maxParticipants; i++ {
		addrHex := "0x" + fmt.Sprintf("%040d", i)
		addr, _ := wallet_vo.NewEVMAddress(addrHex)
		if err := pool.AddParticipant(addr, contribution); err != nil {
			t.Fatalf("Failed to add participant %d: %v", i, err)
		}
	}

	if pool.ParticipantCount() != maxParticipants {
		t.Errorf("Participant count = %d, want %d", pool.ParticipantCount(), maxParticipants)
	}

	expectedTotal := int64(maxParticipants * 100)
	if pool.TotalAmount.Cents() != expectedTotal {
		t.Errorf("TotalAmount = %d, want %d", pool.TotalAmount.Cents(), expectedTotal)
	}
}

// =============================================================================
// Concurrency Tests: Thread Safety
// =============================================================================

// TestOnChainPrizePool_ConcurrentAddParticipants verifies thread safety.
// This test ensures multiple goroutines can safely add participants.
func TestOnChainPrizePool_ConcurrentAddParticipants(t *testing.T) {
	pool := createTestPrizePool()
	txHash, _ := blockchain_vo.NewTxHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	pool.MarkCreated(txHash, wallet_vo.NewAmountFromCents(0))

	const numGoroutines = 50
	var wg sync.WaitGroup
	successCount := 0
	var successMu sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			addrHex := "0x" + fmt.Sprintf("%040d", idx)
			addr, _ := wallet_vo.NewEVMAddress(addrHex)
			contribution := wallet_vo.NewAmountFromCents(100)

			if err := pool.AddParticipant(addr, contribution); err == nil {
				successMu.Lock()
				successCount++
				successMu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	if successCount != numGoroutines {
		t.Errorf("Success count = %d, want %d", successCount, numGoroutines)
	}
	if pool.ParticipantCount() != numGoroutines {
		t.Errorf("Participant count = %d, want %d", pool.ParticipantCount(), numGoroutines)
	}
}

// TestOnChainPrizePool_ConcurrentDuplicateAttempts verifies duplicate detection under concurrency.
// Business rule: Only one of concurrent duplicate attempts should succeed.
func TestOnChainPrizePool_ConcurrentDuplicateAttempts(t *testing.T) {
	pool := createTestPrizePool()
	txHash, _ := blockchain_vo.NewTxHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	pool.MarkCreated(txHash, wallet_vo.NewAmountFromCents(0))

	addr, _ := wallet_vo.NewEVMAddress("0x1111111111111111111111111111111111111111")
	contribution := wallet_vo.NewAmountFromCents(100)

	const numAttempts = 10
	var wg sync.WaitGroup
	successCount := 0
	var successMu sync.Mutex

	for i := 0; i < numAttempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := pool.AddParticipant(addr, contribution); err == nil {
				successMu.Lock()
				successCount++
				successMu.Unlock()
			}
		}()
	}

	wg.Wait()

	if successCount != 1 {
		t.Errorf("Only one duplicate attempt should succeed, got %d", successCount)
	}
	if pool.ParticipantCount() != 1 {
		t.Errorf("Participant count = %d, want 1", pool.ParticipantCount())
	}
}
