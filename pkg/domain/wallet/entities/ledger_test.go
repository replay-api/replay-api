package wallet_entities

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
)

// =============================================================================
// Test Strategy: TDD-driven, functionality-focused tests for double-entry ledger
//
// Business Context: The ledger is the financial backbone of LeetGaming PRO.
// It handles all monetary transactions including:
// - Player deposits and withdrawals
// - Match entry fees and prize distributions
// - Platform fees and subscriptions
//
// CRITICAL REQUIREMENTS:
// 1. Double-entry: Debits MUST equal Credits for every transaction
// 2. Immutability: Entries cannot be modified, only reversed
// 3. Thread safety: Concurrent access must be safe
// 4. Audit trail: Every transaction must be traceable
// =============================================================================

// TestAccountType_Constants verifies account type constants are unique
func TestAccountType_Constants(t *testing.T) {
	types := []AccountType{
		AccountTypeAsset,
		AccountTypeLiability,
		AccountTypeEquity,
		AccountTypeRevenue,
		AccountTypeExpense,
	}

	seen := make(map[AccountType]bool)
	for _, at := range types {
		if at == "" {
			t.Error("AccountType should not be empty")
		}
		if seen[at] {
			t.Errorf("Duplicate AccountType: %s", at)
		}
		seen[at] = true
	}

	// Verify we have all 5 standard accounting types
	if len(types) != 5 {
		t.Errorf("Expected 5 account types (ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE), got %d", len(types))
	}
}

// TestTransactionType_Constants verifies transaction types for esports platform
func TestTransactionType_Constants(t *testing.T) {
	// Core esports transaction types
	esportsTxTypes := []struct {
		txType      TransactionType
		description string
	}{
		{TxTypeDeposit, "Player deposits funds"},
		{TxTypeWithdrawal, "Player withdraws winnings"},
		{TxTypePrizeDistribution, "Match winner receives prize"},
		{TxTypeEntryFee, "Player pays to enter match"},
		{TxTypePlatformFee, "Platform takes commission"},
		{TxTypeRefund, "Cancelled match refund"},
	}

	for _, tt := range esportsTxTypes {
		if tt.txType == "" {
			t.Errorf("TransactionType for '%s' should not be empty", tt.description)
		}
	}
}

// TestNewLedgerEntry_ValidEntry verifies ledger entry creation
func TestNewLedgerEntry_ValidEntry(t *testing.T) {
	txID := uuid.New()
	accountID := uuid.New()
	userID := uuid.New()
	amount := big.NewFloat(100.00)

	entry := NewLedgerEntry(
		txID,
		accountID,
		AccountTypeAsset,
		EntryTypeDebit,
		AssetTypeFiat,
		amount,
		"Player deposit $100",
		"deposit-123",
		userID,
	)

	if entry == nil {
		t.Fatal("Expected non-nil ledger entry")
	}
	if entry.ID == uuid.Nil {
		t.Error("Entry ID should be generated")
	}
	if entry.TransactionID != txID {
		t.Error("TransactionID mismatch")
	}
	if entry.JournalID != txID {
		t.Error("JournalID should equal TransactionID")
	}
	if entry.AccountID != accountID {
		t.Error("AccountID mismatch")
	}
	if entry.EntryType != EntryTypeDebit {
		t.Errorf("EntryType = %s, want DEBIT", entry.EntryType)
	}
	if entry.IsReversed {
		t.Error("New entry should not be reversed")
	}
	if entry.IdempotencyKey != "deposit-123" {
		t.Error("IdempotencyKey mismatch")
	}
}

// TestLedgerEntry_Validate tests validation logic
func TestLedgerEntry_Validate(t *testing.T) {
	tests := []struct {
		name        string
		modifyEntry func(*LedgerEntry)
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "valid entry passes validation",
			modifyEntry: func(e *LedgerEntry) {},
			shouldError: false,
		},
		{
			name: "missing transaction_id fails",
			modifyEntry: func(e *LedgerEntry) {
				e.TransactionID = uuid.Nil
				e.JournalID = uuid.Nil
			},
			shouldError: true,
			errorMsg:    "transaction_id",
		},
		{
			name: "missing account_id fails",
			modifyEntry: func(e *LedgerEntry) {
				e.AccountID = uuid.Nil
			},
			shouldError: true,
			errorMsg:    "account_id",
		},
		{
			name: "zero amount fails",
			modifyEntry: func(e *LedgerEntry) {
				e.Amount = big.NewFloat(0)
			},
			shouldError: true,
			errorMsg:    "amount",
		},
		{
			name: "negative amount fails",
			modifyEntry: func(e *LedgerEntry) {
				e.Amount = big.NewFloat(-50)
			},
			shouldError: true,
			errorMsg:    "amount",
		},
		{
			name: "missing description fails",
			modifyEntry: func(e *LedgerEntry) {
				e.Description = ""
			},
			shouldError: true,
			errorMsg:    "description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := createValidLedgerEntry()
			tt.modifyEntry(entry)

			err := entry.Validate()

			if tt.shouldError && err == nil {
				t.Errorf("Expected error containing '%s'", tt.errorMsg)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestLedgerEntry_Reverse tests entry reversal for refunds/corrections
func TestLedgerEntry_Reverse(t *testing.T) {
	original := createValidLedgerEntry()
	original.EntryType = EntryTypeDebit
	original.BalanceAfter = big.NewFloat(100)
	reverserID := uuid.New()

	reversal := original.Reverse("Match cancelled", reverserID)

	// Verify reversal creates opposite entry type
	if reversal.EntryType != EntryTypeCredit {
		t.Errorf("Reversal entry type = %s, want CREDIT (opposite of DEBIT)", reversal.EntryType)
	}

	// Verify amounts match
	if reversal.Amount.Cmp(original.Amount) != 0 {
		t.Error("Reversal amount should match original")
	}

	// Verify original is marked as reversed
	if !original.IsReversed {
		t.Error("Original entry should be marked as reversed")
	}
	if original.ReversedBy == nil || *original.ReversedBy != reverserID {
		t.Error("Original should track who reversed it")
	}

	// Verify description includes reason
	if reversal.Description == "" {
		t.Error("Reversal should have description with reason")
	}

	// Verify idempotency key is derived
	expectedKey := fmt.Sprintf("reversal-%s", original.IdempotencyKey)
	if reversal.IdempotencyKey != expectedKey {
		t.Errorf("Idempotency key = %s, want %s", reversal.IdempotencyKey, expectedKey)
	}
}

// TestJournalEntry_AddDebit tests debit entry addition
func TestJournalEntry_AddDebit(t *testing.T) {
	journal := createTestJournalEntry()
	accountID := uuid.New()
	amount := big.NewFloat(50.00)

	err := journal.AddDebit(accountID, "1001", amount, "Cash received")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(journal.Entries) != 1 {
		t.Errorf("Entries count = %d, want 1", len(journal.Entries))
	}
	if journal.TotalDebit.Cmp(amount) != 0 {
		t.Error("TotalDebit should equal added amount")
	}
	if journal.Entries[0].EntryType != EntryTypeDebit {
		t.Error("Entry should be DEBIT type")
	}
}

// TestJournalEntry_AddCredit tests credit entry addition
func TestJournalEntry_AddCredit(t *testing.T) {
	journal := createTestJournalEntry()
	accountID := uuid.New()
	amount := big.NewFloat(50.00)

	err := journal.AddCredit(accountID, "2001", amount, "User wallet credited")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(journal.Entries) != 1 {
		t.Errorf("Entries count = %d, want 1", len(journal.Entries))
	}
	if journal.TotalCredit.Cmp(amount) != 0 {
		t.Error("TotalCredit should equal added amount")
	}
	if journal.Entries[0].EntryType != EntryTypeCredit {
		t.Error("Entry should be CREDIT type")
	}
}

// TestJournalEntry_Validate_BalancedEntries tests double-entry balancing
// CRITICAL: This is the core accounting invariant - debits must equal credits
func TestJournalEntry_Validate_BalancedEntries(t *testing.T) {
	journal := createTestJournalEntry()

	// Add balanced entries: $100 debit = $100 credit
	cashAccount := uuid.New()
	userAccount := uuid.New()
	amount := big.NewFloat(100.00)

	_ = journal.AddDebit(cashAccount, "1001", amount, "Cash received")
	_ = journal.AddCredit(userAccount, "2001", amount, "User wallet credited")

	err := journal.Validate()

	if err != nil {
		t.Errorf("Balanced journal should be valid: %v", err)
	}
}

// TestJournalEntry_Validate_UnbalancedEntries tests rejection of unbalanced journals
func TestJournalEntry_Validate_UnbalancedEntries(t *testing.T) {
	journal := createTestJournalEntry()

	// Add unbalanced entries: $100 debit ≠ $50 credit
	cashAccount := uuid.New()
	userAccount := uuid.New()

	_ = journal.AddDebit(cashAccount, "1001", big.NewFloat(100.00), "Cash received")
	_ = journal.AddCredit(userAccount, "2001", big.NewFloat(50.00), "Partial credit")

	err := journal.Validate()

	if err == nil {
		t.Error("Unbalanced journal should fail validation - debits must equal credits")
	}
}

// TestJournalEntry_ComputeHash tests hash chain integrity
// Business rule: Each journal links to the previous via hash for audit trail
func TestJournalEntry_ComputeHash(t *testing.T) {
	journal := createTestJournalEntry()
	previousHash := "abc123previoushash"

	journal.ComputeHash(previousHash)

	if journal.Hash == "" {
		t.Error("Hash should be computed")
	}
	if journal.PreviousHash != previousHash {
		t.Errorf("PreviousHash = %s, want %s", journal.PreviousHash, previousHash)
	}

	// Hash should be deterministic for same inputs
	firstHash := journal.Hash
	journal.ComputeHash(previousHash)
	if journal.Hash != firstHash {
		t.Error("Hash should be deterministic")
	}
}

// TestJournalEntry_MarkApproved tests approval workflow
func TestJournalEntry_MarkApproved(t *testing.T) {
	journal := createBalancedJournalEntry() // Must have balanced entries
	journal.Status = JournalStatusPending
	approverID := uuid.New()

	err := journal.MarkApproved(approverID)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if journal.Status != JournalStatusApproved {
		t.Errorf("Status = %s, want APPROVED", journal.Status)
	}
	if journal.ApprovedBy == nil || *journal.ApprovedBy != approverID {
		t.Error("ApprovedBy should be set")
	}
	if journal.ApprovedAt == nil {
		t.Error("ApprovedAt should be set")
	}
}

// TestJournalEntry_MarkPosted tests posting workflow
func TestJournalEntry_MarkPosted(t *testing.T) {
	journal := createTestJournalEntry()
	journal.Status = JournalStatusApproved

	err := journal.MarkPosted()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if journal.Status != JournalStatusPosted {
		t.Errorf("Status = %s, want POSTED", journal.Status)
	}
	if journal.PostedAt == nil {
		t.Error("PostedAt should be set")
	}
}

// TestJournalEntry_StatusTransitions tests valid state machine transitions
func TestJournalEntry_StatusTransitions(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  JournalStatus
		action         string
		expectedStatus JournalStatus
		shouldError    bool
	}{
		{
			name:           "Draft can be approved",
			initialStatus:  JournalStatusDraft,
			action:         "approve",
			expectedStatus: JournalStatusApproved,
			shouldError:    false,
		},
		{
			name:           "Pending can be approved",
			initialStatus:  JournalStatusPending,
			action:         "approve",
			expectedStatus: JournalStatusApproved,
			shouldError:    false,
		},
		{
			name:           "Approved can be posted",
			initialStatus:  JournalStatusApproved,
			action:         "post",
			expectedStatus: JournalStatusPosted,
			shouldError:    false,
		},
		{
			name:           "Posted cannot be approved again",
			initialStatus:  JournalStatusPosted,
			action:         "approve",
			expectedStatus: JournalStatusPosted,
			shouldError:    true,
		},
		{
			name:           "Draft cannot be posted directly",
			initialStatus:  JournalStatusDraft,
			action:         "post",
			expectedStatus: JournalStatusDraft,
			shouldError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			journal := createBalancedJournalEntry() // Need balanced entries for approval
			journal.Status = tt.initialStatus
			var err error

			switch tt.action {
			case "approve":
				err = journal.MarkApproved(uuid.New())
			case "post":
				err = journal.MarkPosted()
			}

			if tt.shouldError && err == nil {
				t.Errorf("Expected error for %s", tt.action)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if journal.Status != tt.expectedStatus {
				t.Errorf("Status = %s, want %s", journal.Status, tt.expectedStatus)
			}
		})
	}
}

// =============================================================================
// Esports-Specific Transaction Scenarios
// =============================================================================

// TestScenario_PlayerDeposit simulates a player adding funds
func TestScenario_PlayerDeposit(t *testing.T) {
	journal := NewJournalEntry(
		TxTypeDeposit,
		"DEP-123456",
		"Player deposit via credit card",
		"USD",
		uuid.New(),
		testResourceOwner(),
	)

	cashAccount := uuid.New()
	userAccount := uuid.New()
	depositAmount := big.NewFloat(50.00)

	// Double-entry: Cash (Asset) increases, User Liability increases
	_ = journal.AddDebit(cashAccount, "1001", depositAmount, "Cash from payment processor")
	_ = journal.AddCredit(userAccount, "2001", depositAmount, "User wallet balance")

	if err := journal.Validate(); err != nil {
		t.Errorf("Deposit transaction should be valid: %v", err)
	}
	if journal.TransactionType != TxTypeDeposit {
		t.Errorf("TransactionType = %s, want DEPOSIT", journal.TransactionType)
	}
}

// TestScenario_MatchEntryFee simulates a player paying to enter a match
func TestScenario_MatchEntryFee(t *testing.T) {
	journal := NewJournalEntry(
		TxTypeEntryFee,
		"ENTRY-789",
		"CS2 Competitive Match Entry",
		"USD",
		uuid.New(),
		testResourceOwner(),
	)

	userAccount := uuid.New()
	prizePoolAccount := uuid.New()
	entryFee := big.NewFloat(10.00)

	// Double-entry: User balance decreases, Prize pool increases
	_ = journal.AddDebit(userAccount, "2001", entryFee, "Entry fee from player")
	_ = journal.AddCredit(prizePoolAccount, "2100", entryFee, "Added to prize pool")

	if err := journal.Validate(); err != nil {
		t.Errorf("Entry fee transaction should be valid: %v", err)
	}
}

// TestScenario_PrizeDistribution simulates distributing winnings
func TestScenario_PrizeDistribution(t *testing.T) {
	journal := NewJournalEntry(
		TxTypePrizeDistribution,
		"PRIZE-456",
		"Match winner prize payout",
		"USD",
		uuid.New(),
		testResourceOwner(),
	)

	prizePoolAccount := uuid.New()
	winnerAccount := uuid.New()
	platformFeeAccount := uuid.New()

	totalPrize := big.NewFloat(100.00)
	platformFee := big.NewFloat(10.00) // 10% platform fee
	winnerPrize := big.NewFloat(90.00)

	// Prize pool decreases
	_ = journal.AddDebit(prizePoolAccount, "2100", totalPrize, "Prize pool distribution")
	// Winner receives prize (minus fee)
	_ = journal.AddCredit(winnerAccount, "2001", winnerPrize, "Winner receives prize")
	// Platform takes fee
	_ = journal.AddCredit(platformFeeAccount, "4001", platformFee, "Platform commission")

	if err := journal.Validate(); err != nil {
		t.Errorf("Prize distribution should be valid: %v", err)
	}
}

// =============================================================================
// Sequential Entry Tests  
// Note: JournalEntry is intentionally NOT thread-safe at the entity level.
// Thread safety is enforced at the LedgerService layer which uses mutexes.
// This follows the principle that entities are value objects, not concurrent data structures.
// =============================================================================

// TestJournalEntry_MultipleEntries tests adding multiple entries sequentially
func TestJournalEntry_MultipleEntries(t *testing.T) {
	journal := createTestJournalEntry()

	const numEntries = 20

	// Add entries sequentially (as designed)
	for i := 0; i < numEntries; i++ {
		accountID := uuid.New()
		amount := big.NewFloat(10.00)

		if i%2 == 0 {
			_ = journal.AddDebit(accountID, fmt.Sprintf("D%d", i), amount, "Debit entry")
		} else {
			_ = journal.AddCredit(accountID, fmt.Sprintf("C%d", i), amount, "Credit entry")
		}
	}

	// All entries should be added
	if len(journal.Entries) != numEntries {
		t.Errorf("Entry count = %d, want %d", len(journal.Entries), numEntries)
	}

	// Total debit should equal total credit (balanced)
	if journal.TotalDebit.Cmp(journal.TotalCredit) != 0 {
		t.Error("Debits should equal credits for balanced journal")
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func createValidLedgerEntry() *LedgerEntry {
	return NewLedgerEntry(
		uuid.New(),
		uuid.New(),
		AccountTypeAsset,
		EntryTypeDebit,
		AssetTypeFiat,
		big.NewFloat(100.00),
		"Test entry",
		"test-key-123",
		uuid.New(),
	)
}

func createTestJournalEntry() *JournalEntry {
	return NewJournalEntry(
		TxTypeDeposit,
		"TEST-123",
		"Test journal entry",
		"USD",
		uuid.New(),
		testResourceOwner(),
	)
}

func createBalancedJournalEntry() *JournalEntry {
	journal := NewJournalEntry(
		TxTypeDeposit,
		"TEST-BALANCED-123",
		"Balanced test journal entry",
		"USD",
		uuid.New(),
		testResourceOwner(),
	)
	// Add balanced entries
	_ = journal.AddDebit(uuid.New(), "1001", big.NewFloat(100.00), "Debit entry")
	_ = journal.AddCredit(uuid.New(), "2001", big.NewFloat(100.00), "Credit entry")
	return journal
}

func testResourceOwner() common.ResourceOwner {
	return common.ResourceOwner{
		UserID:   uuid.New(),
		TenantID: uuid.New(),
	}
}

