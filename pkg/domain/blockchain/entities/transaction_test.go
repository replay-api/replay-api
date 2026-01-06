package blockchain_entities

import (
	"testing"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	blockchain_vo "github.com/replay-api/replay-api/pkg/domain/blockchain/value-objects"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

// Helper to create a test resource owner
func testResourceOwner() shared.ResourceOwner {
	return shared.ResourceOwner{
		TenantID: uuid.New(),
		ClientID: uuid.New(),
		GroupID:  uuid.New(),
		UserID:   uuid.New(),
	}
}

// TestTransactionStatus_Constants verifies status constants
func TestTransactionStatus_Constants(t *testing.T) {
	statuses := []TransactionStatus{
		TxStatusPending,
		TxStatusConfirmed,
		TxStatusFailed,
		TxStatusReplaced,
	}

	seen := make(map[TransactionStatus]bool)
	for _, status := range statuses {
		if status == "" {
			t.Error("TransactionStatus should not be empty")
		}
		if seen[status] {
			t.Errorf("Duplicate TransactionStatus: %s", status)
		}
		seen[status] = true
	}
}

// TestTransactionType_Constants verifies type constants
func TestTransactionType_Constants(t *testing.T) {
	types := []TransactionType{
		TxTypeDeposit,
		TxTypeWithdrawal,
		TxTypeEntryFee,
		TxTypePrize,
		TxTypeRefund,
		TxTypePlatformFee,
		TxTypeBridge,
		TxTypeContractCall,
	}

	seen := make(map[TransactionType]bool)
	for _, txType := range types {
		if txType == "" {
			t.Error("TransactionType should not be empty")
		}
		if seen[txType] {
			t.Errorf("Duplicate TransactionType: %s", txType)
		}
		seen[txType] = true
	}
}

// TestNewBlockchainTransaction verifies transaction creation
func TestNewBlockchainTransaction(t *testing.T) {
	owner := testResourceOwner()
	chainID := blockchain_vo.ChainIDPolygon
	from, _ := wallet_vo.NewEVMAddress("0x1234567890123456789012345678901234567890")
	to, _ := wallet_vo.NewEVMAddress("0xabcdef1234567890abcdef1234567890abcdef12")
	currency := wallet_vo.CurrencyUSDC
	amount := wallet_vo.NewAmount(10000)

	tx := NewBlockchainTransaction(owner, chainID, TxTypeDeposit, from, to, currency, amount)

	if tx == nil {
		t.Fatal("Expected non-nil transaction")
	}
	if tx.ChainID != chainID {
		t.Errorf("ChainID = %d, want %d", tx.ChainID, chainID)
	}
	if tx.Type != TxTypeDeposit {
		t.Errorf("Type = %s, want %s", tx.Type, TxTypeDeposit)
	}
	if tx.Status != TxStatusPending {
		t.Errorf("Status = %s, want %s", tx.Status, TxStatusPending)
	}
	if tx.SubmittedAt == nil {
		t.Error("SubmittedAt should be set")
	}
	if tx.RequiredConfirms != 12 {
		t.Errorf("RequiredConfirms = %d, want 12", tx.RequiredConfirms)
	}
}

// TestBlockchainTransaction_SetTxHash verifies hash setting
func TestBlockchainTransaction_SetTxHash(t *testing.T) {
	owner := testResourceOwner()
	from, _ := wallet_vo.NewEVMAddress("0x1234567890123456789012345678901234567890")
	to, _ := wallet_vo.NewEVMAddress("0xabcdef1234567890abcdef1234567890abcdef12")

	tx := NewBlockchainTransaction(owner, blockchain_vo.ChainIDPolygon, TxTypeDeposit, from, to, wallet_vo.CurrencyUSDC, wallet_vo.NewAmount(100))
	hash, _ := blockchain_vo.NewTxHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")

	tx.SetTxHash(hash)

	if !tx.TxHash.Equals(hash) {
		t.Error("TxHash not set correctly")
	}
}

// TestBlockchainTransaction_Confirm verifies confirmation logic
func TestBlockchainTransaction_Confirm(t *testing.T) {
	owner := testResourceOwner()
	from, _ := wallet_vo.NewEVMAddress("0x1234567890123456789012345678901234567890")
	to, _ := wallet_vo.NewEVMAddress("0xabcdef1234567890abcdef1234567890abcdef12")

	tx := NewBlockchainTransaction(owner, blockchain_vo.ChainIDPolygon, TxTypeDeposit, from, to, wallet_vo.CurrencyUSDC, wallet_vo.NewAmount(100))
	blockHash, _ := blockchain_vo.NewTxHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")

	tx.Confirm(12345, blockHash, 21000)

	if tx.Status != TxStatusConfirmed {
		t.Errorf("Status = %s, want %s", tx.Status, TxStatusConfirmed)
	}
	if tx.BlockNumber != 12345 {
		t.Errorf("BlockNumber = %d, want 12345", tx.BlockNumber)
	}
	if tx.GasUsed != 21000 {
		t.Errorf("GasUsed = %d, want 21000", tx.GasUsed)
	}
	if tx.ConfirmedAt == nil {
		t.Error("ConfirmedAt should be set")
	}
}

// TestBlockchainTransaction_Fail verifies failure logic
func TestBlockchainTransaction_Fail(t *testing.T) {
	owner := testResourceOwner()
	from, _ := wallet_vo.NewEVMAddress("0x1234567890123456789012345678901234567890")
	to, _ := wallet_vo.NewEVMAddress("0xabcdef1234567890abcdef1234567890abcdef12")

	tx := NewBlockchainTransaction(owner, blockchain_vo.ChainIDPolygon, TxTypeDeposit, from, to, wallet_vo.CurrencyUSDC, wallet_vo.NewAmount(100))

	tx.Fail("out of gas")

	if tx.Status != TxStatusFailed {
		t.Errorf("Status = %s, want %s", tx.Status, TxStatusFailed)
	}
	if tx.ErrorMessage != "out of gas" {
		t.Errorf("ErrorMessage = %s, want 'out of gas'", tx.ErrorMessage)
	}
}

// TestBlockchainTransaction_IncrementRetry verifies retry counter
func TestBlockchainTransaction_IncrementRetry(t *testing.T) {
	owner := testResourceOwner()
	from, _ := wallet_vo.NewEVMAddress("0x1234567890123456789012345678901234567890")
	to, _ := wallet_vo.NewEVMAddress("0xabcdef1234567890abcdef1234567890abcdef12")

	tx := NewBlockchainTransaction(owner, blockchain_vo.ChainIDPolygon, TxTypeDeposit, from, to, wallet_vo.CurrencyUSDC, wallet_vo.NewAmount(100))

	if tx.RetryCount != 0 {
		t.Errorf("Initial RetryCount = %d, want 0", tx.RetryCount)
	}

	tx.IncrementRetry()
	if tx.RetryCount != 1 {
		t.Errorf("RetryCount after increment = %d, want 1", tx.RetryCount)
	}

	tx.IncrementRetry()
	if tx.RetryCount != 2 {
		t.Errorf("RetryCount after second increment = %d, want 2", tx.RetryCount)
	}
}

// TestBlockchainTransaction_UpdateConfirmations verifies confirmation tracking
func TestBlockchainTransaction_UpdateConfirmations(t *testing.T) {
	owner := testResourceOwner()
	from, _ := wallet_vo.NewEVMAddress("0x1234567890123456789012345678901234567890")
	to, _ := wallet_vo.NewEVMAddress("0xabcdef1234567890abcdef1234567890abcdef12")

	tx := NewBlockchainTransaction(owner, blockchain_vo.ChainIDPolygon, TxTypeDeposit, from, to, wallet_vo.CurrencyUSDC, wallet_vo.NewAmount(100))

	tx.UpdateConfirmations(5)
	if tx.Confirmations != 5 {
		t.Errorf("Confirmations = %d, want 5", tx.Confirmations)
	}
}

// TestBlockchainTransaction_IsConfirmed verifies confirmation check
func TestBlockchainTransaction_IsConfirmed(t *testing.T) {
	owner := testResourceOwner()
	from, _ := wallet_vo.NewEVMAddress("0x1234567890123456789012345678901234567890")
	to, _ := wallet_vo.NewEVMAddress("0xabcdef1234567890abcdef1234567890abcdef12")
	blockHash, _ := blockchain_vo.NewTxHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")

	tx := NewBlockchainTransaction(owner, blockchain_vo.ChainIDPolygon, TxTypeDeposit, from, to, wallet_vo.CurrencyUSDC, wallet_vo.NewAmount(100))

	// Not confirmed initially
	if tx.IsConfirmed() {
		t.Error("New transaction should not be confirmed")
	}

	// Confirm but not enough confirmations
	tx.Confirm(12345, blockHash, 21000)
	tx.UpdateConfirmations(5)
	if tx.IsConfirmed() {
		t.Error("Transaction with 5 confirmations should not be fully confirmed")
	}

	// Add enough confirmations
	tx.UpdateConfirmations(12)
	if !tx.IsConfirmed() {
		t.Error("Transaction with 12 confirmations should be confirmed")
	}
}

// TestBlockchainTransaction_IsPending verifies pending check
func TestBlockchainTransaction_IsPending(t *testing.T) {
	owner := testResourceOwner()
	from, _ := wallet_vo.NewEVMAddress("0x1234567890123456789012345678901234567890")
	to, _ := wallet_vo.NewEVMAddress("0xabcdef1234567890abcdef1234567890abcdef12")
	blockHash, _ := blockchain_vo.NewTxHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")

	tx := NewBlockchainTransaction(owner, blockchain_vo.ChainIDPolygon, TxTypeDeposit, from, to, wallet_vo.CurrencyUSDC, wallet_vo.NewAmount(100))

	if !tx.IsPending() {
		t.Error("New transaction should be pending")
	}

	tx.Confirm(12345, blockHash, 21000)
	if tx.IsPending() {
		t.Error("Confirmed transaction should not be pending")
	}
}

// TestBlockchainTransaction_SetRelatedEntities verifies related entity linking
func TestBlockchainTransaction_SetRelatedEntities(t *testing.T) {
	owner := testResourceOwner()
	from, _ := wallet_vo.NewEVMAddress("0x1234567890123456789012345678901234567890")
	to, _ := wallet_vo.NewEVMAddress("0xabcdef1234567890abcdef1234567890abcdef12")

	tx := NewBlockchainTransaction(owner, blockchain_vo.ChainIDPolygon, TxTypeDeposit, from, to, wallet_vo.CurrencyUSDC, wallet_vo.NewAmount(100))

	matchID := uuid.New()
	tournamentID := uuid.New()
	walletID := uuid.New()
	ledgerTxID := uuid.New()

	tx.SetRelatedMatch(matchID)
	tx.SetRelatedTournament(tournamentID)
	tx.SetWallet(walletID)
	tx.SetLedgerTransaction(ledgerTxID)

	if tx.MatchID == nil || *tx.MatchID != matchID {
		t.Error("MatchID not set correctly")
	}
	if tx.TournamentID == nil || *tx.TournamentID != tournamentID {
		t.Error("TournamentID not set correctly")
	}
	if tx.WalletID == nil || *tx.WalletID != walletID {
		t.Error("WalletID not set correctly")
	}
	if tx.LedgerTxID == nil || *tx.LedgerTxID != ledgerTxID {
		t.Error("LedgerTxID not set correctly")
	}
}

// TestBlockchainTransaction_UpdatedAtChanges verifies timestamp updates
func TestBlockchainTransaction_UpdatedAtChanges(t *testing.T) {
	owner := testResourceOwner()
	from, _ := wallet_vo.NewEVMAddress("0x1234567890123456789012345678901234567890")
	to, _ := wallet_vo.NewEVMAddress("0xabcdef1234567890abcdef1234567890abcdef12")

	tx := NewBlockchainTransaction(owner, blockchain_vo.ChainIDPolygon, TxTypeDeposit, from, to, wallet_vo.CurrencyUSDC, wallet_vo.NewAmount(100))
	initialUpdatedAt := tx.UpdatedAt

	time.Sleep(1 * time.Millisecond)
	tx.IncrementRetry()

	if !tx.UpdatedAt.After(initialUpdatedAt) {
		t.Error("UpdatedAt should change after IncrementRetry")
	}
}
