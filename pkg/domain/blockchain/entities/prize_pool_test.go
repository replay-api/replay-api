package blockchain_entities

import (
	"testing"
	"time"

	"github.com/google/uuid"
	blockchain_vo "github.com/replay-api/replay-api/pkg/domain/blockchain/value-objects"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

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

	err := _ = pool.AddParticipant(addr, contribution)

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

// TestOnChainPrizePool_AddParticipant_Duplicate verifies duplicate rejection
func TestOnChainPrizePool_AddParticipant_Duplicate(t *testing.T) {
	pool := createTestPrizePool()
	txHash, _ := blockchain_vo.NewTxHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	pool.MarkCreated(txHash, wallet_vo.NewAmountFromCents(0))

	addr, _ := wallet_vo.NewEVMAddress("0x1111111111111111111111111111111111111111")
	contribution := wallet_vo.NewAmountFromCents(1000)

	_ = _ = pool.AddParticipant(addr, contribution)
	err := _ = pool.AddParticipant(addr, contribution)

	if err == nil {
		t.Error("Expected error for duplicate participant")
	}
}

// TestOnChainPrizePool_AddParticipant_WrongStatus verifies status check
func TestOnChainPrizePool_AddParticipant_WrongStatus(t *testing.T) {
	pool := createTestPrizePool()
	// Pool is NotCreated status

	addr, _ := wallet_vo.NewEVMAddress("0x1111111111111111111111111111111111111111")
	contribution := wallet_vo.NewAmountFromCents(1000)

	err := _ = pool.AddParticipant(addr, contribution)

	if err == nil {
		t.Error("Expected error for wrong status")
	}
}

// TestOnChainPrizePool_Lock verifies pool locking
func TestOnChainPrizePool_Lock(t *testing.T) {
	pool := createPoolWithParticipants(2)
	lockTxHash, _ := blockchain_vo.NewTxHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")

	err := _ = pool.Lock(lockTxHash)

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

	err := _ = pool.Lock(lockTxHash)

	if err == nil {
		t.Error("Expected error for not enough participants")
	}
}

// TestOnChainPrizePool_StartEscrow verifies escrow period start
func TestOnChainPrizePool_StartEscrow(t *testing.T) {
	pool := createPoolWithParticipants(2)
	lockTxHash, _ := blockchain_vo.NewTxHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")
	_ = _ = pool.Lock(lockTxHash)

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
