package wallet_entities

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
)

// AccountType defines the type of ledger account
type AccountType string

const (
	AccountTypeAsset     AccountType = "ASSET"     // Cash, Bank, Crypto wallets
	AccountTypeLiability AccountType = "LIABILITY" // User balances, deposits
	AccountTypeEquity    AccountType = "EQUITY"    // Platform equity, retained earnings
	AccountTypeRevenue   AccountType = "REVENUE"   // Platform fees, subscriptions
	AccountTypeExpense   AccountType = "EXPENSE"   // Payouts, refunds
)

// EntryType defines debit or credit
type EntryType string

const (
	EntryTypeDebit  EntryType = "DEBIT"
	EntryTypeCredit EntryType = "CREDIT"
)

// TransactionType categorizes transactions for reporting
type TransactionType string

const (
	TxTypeDeposit           TransactionType = "DEPOSIT"
	TxTypeWithdrawal        TransactionType = "WITHDRAWAL"
	TxTypeTransfer          TransactionType = "TRANSFER"
	TxTypePrizeDistribution TransactionType = "PRIZE_DISTRIBUTION"
	TxTypeEntryFee          TransactionType = "ENTRY_FEE"
	TxTypePlatformFee       TransactionType = "PLATFORM_FEE"
	TxTypeSubscription      TransactionType = "SUBSCRIPTION"
	TxTypeRefund            TransactionType = "REFUND"
	TxTypeAdjustment        TransactionType = "ADJUSTMENT"
	TxTypeHold              TransactionType = "HOLD"
	TxTypeRelease           TransactionType = "RELEASE"
)

// AssetType defines the type of asset being transacted
type AssetType string

const (
	AssetTypeFiat       AssetType = "FIAT"
	AssetTypeCrypto     AssetType = "CRYPTO"
	AssetTypeNFT        AssetType = "NFT"
	AssetTypeGameCredit AssetType = "GAME_CREDIT"
)

// ApprovalStatus indicates whether the transaction was auto-approved or requires review
type ApprovalStatus string

const (
	ApprovalStatusAutoApproved     ApprovalStatus = "AUTO_APPROVED"
	ApprovalStatusPendingReview    ApprovalStatus = "PENDING_REVIEW"
	ApprovalStatusManuallyApproved ApprovalStatus = "MANUALLY_APPROVED"
	ApprovalStatusRejected         ApprovalStatus = "REJECTED"
)

// LedgerAccount represents a double-entry accounting account
type LedgerAccount struct {
	ID            uuid.UUID   `json:"id" bson:"_id"`
	Code          string      `json:"code" bson:"code"` // e.g., "1001", "2001"
	Name          string      `json:"name" bson:"name"`
	Type          AccountType `json:"type" bson:"type"`
	Currency      string      `json:"currency" bson:"currency"`
	Balance       *big.Float  `json:"balance" bson:"balance"`
	AvailableBalance *big.Float `json:"available_balance" bson:"available_balance"`
	HeldBalance   *big.Float  `json:"held_balance" bson:"held_balance"`
	ParentID      *uuid.UUID  `json:"parent_id,omitempty" bson:"parent_id,omitempty"`
	UserID        *uuid.UUID  `json:"user_id,omitempty" bson:"user_id,omitempty"` // For user-specific accounts
	IsActive      bool        `json:"is_active" bson:"is_active"`
	CreatedAt     time.Time   `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at" bson:"updated_at"`
	Version       int         `json:"version" bson:"version"` // Optimistic locking
}

// LedgerMetadata stores additional context about the transaction
type LedgerMetadata struct {
	OperationType      string                 `json:"operation_type" bson:"operation_type"` // "Deposit", "Withdrawal", "Transfer", "Prize"
	TxHash             string                 `json:"tx_hash,omitempty" bson:"tx_hash,omitempty"`
	PaymentID          *uuid.UUID             `json:"payment_id,omitempty" bson:"payment_id,omitempty"`
	MatchID            *uuid.UUID             `json:"match_id,omitempty" bson:"match_id,omitempty"`
	TournamentID       *uuid.UUID             `json:"tournament_id,omitempty" bson:"tournament_id,omitempty"`
	SourceIP           string                 `json:"source_ip,omitempty" bson:"source_ip,omitempty"`
	UserAgent          string                 `json:"user_agent,omitempty" bson:"user_agent,omitempty"`
	GeolocationCountry string                 `json:"geolocation_country,omitempty" bson:"geolocation_country,omitempty"`
	RiskScore          float64                `json:"risk_score" bson:"risk_score"`
	ApprovalStatus     ApprovalStatus         `json:"approval_status" bson:"approval_status"`
	ApproverID         *uuid.UUID             `json:"approver_id,omitempty" bson:"approver_id,omitempty"`
	Notes              string                 `json:"notes,omitempty" bson:"notes,omitempty"`
	CustomData         map[string]interface{} `json:"custom_data,omitempty" bson:"custom_data,omitempty"`
}

// LedgerEntry represents a single entry in a journal entry
type LedgerEntry struct {
	ID            uuid.UUID      `json:"id" bson:"_id"`
	JournalID     uuid.UUID      `json:"journal_id" bson:"journal_id"`
	TransactionID uuid.UUID      `json:"transaction_id" bson:"transaction_id"` // Alias for JournalID for backward compat
	AccountID     uuid.UUID      `json:"account_id" bson:"account_id"`
	AccountCode   string         `json:"account_code" bson:"account_code"`
	EntryType     EntryType      `json:"entry_type" bson:"entry_type"`
	AssetType     AssetType      `json:"asset_type" bson:"asset_type"`
	Amount        *big.Float     `json:"amount" bson:"amount"`
	Currency      string         `json:"currency" bson:"currency"`
	BalanceBefore *big.Float     `json:"balance_before" bson:"balance_before"`
	BalanceAfter  *big.Float     `json:"balance_after" bson:"balance_after"`
	Description   string         `json:"description" bson:"description"`
	Metadata      LedgerMetadata `json:"metadata" bson:"metadata"`
	CreatedAt     time.Time      `json:"created_at" bson:"created_at"`
	CreatedBy     uuid.UUID      `json:"created_by" bson:"created_by"`
	IsReversed      bool           `json:"is_reversed" bson:"is_reversed"`
	ReversedBy      *uuid.UUID     `json:"reversed_by,omitempty" bson:"reversed_by,omitempty"`
	IdempotencyKey  string         `json:"idempotency_key,omitempty" bson:"idempotency_key,omitempty"`
}

// NewLedgerEntry creates a new ledger entry with the given parameters
func NewLedgerEntry(
	transactionID uuid.UUID,
	accountID uuid.UUID,
	accountType AccountType,
	entryType EntryType,
	assetType AssetType,
	amount *big.Float,
	description string,
	idempotencyKey string,
	createdBy uuid.UUID,
) *LedgerEntry {
	now := time.Now().UTC()
	return &LedgerEntry{
		ID:             uuid.New(),
		JournalID:      transactionID,
		TransactionID:  transactionID,
		AccountID:      accountID,
		AccountCode:    string(accountType),
		EntryType:      entryType,
		AssetType:      assetType,
		Amount:         amount,
		Currency:       "USD", // Default currency
		BalanceBefore:  big.NewFloat(0),
		BalanceAfter:   big.NewFloat(0),
		Description:    description,
		Metadata:       LedgerMetadata{},
		CreatedAt:      now,
		CreatedBy:      createdBy,
		IsReversed:     false,
		IdempotencyKey: idempotencyKey,
	}
}

// Reverse creates a reversal entry for this ledger entry
func (l *LedgerEntry) Reverse(reason string, reversedBy uuid.UUID) *LedgerEntry {
	// Opposite entry type
	reverseType := EntryTypeCredit
	if l.EntryType == EntryTypeCredit {
		reverseType = EntryTypeDebit
	}

	reversal := &LedgerEntry{
		ID:             uuid.New(),
		JournalID:      uuid.New(), // New journal for reversal
		TransactionID:  uuid.New(),
		AccountID:      l.AccountID,
		AccountCode:    l.AccountCode,
		EntryType:      reverseType,
		AssetType:      l.AssetType,
		Amount:         l.Amount,
		Currency:       l.Currency,
		BalanceBefore:  l.BalanceAfter,
		BalanceAfter:   l.BalanceBefore,
		Description:    fmt.Sprintf("REVERSAL: %s - %s", l.Description, reason),
		Metadata:       l.Metadata,
		CreatedAt:      time.Now().UTC(),
		CreatedBy:      reversedBy,
		IsReversed:     false,
		IdempotencyKey: fmt.Sprintf("reversal-%s", l.IdempotencyKey),
	}

	// Mark original as reversed
	l.IsReversed = true
	l.ReversedBy = &reversedBy

	return reversal
}

// Validate ensures the ledger entry is valid
func (l *LedgerEntry) Validate() error {
	if l.TransactionID == uuid.Nil && l.JournalID == uuid.Nil {
		return errors.New("transaction_id or journal_id is required")
	}
	if l.AccountID == uuid.Nil {
		return errors.New("account_id is required")
	}
	if l.Amount == nil || l.Amount.Cmp(big.NewFloat(0)) <= 0 {
		return errors.New("amount must be positive")
	}
	if l.Description == "" {
		return errors.New("description is required")
	}
	return nil
}

// JournalEntry represents a complete double-entry transaction
// CRITICAL: Debits MUST equal Credits for every journal entry
type JournalEntry struct {
	ID              uuid.UUID           `json:"id" bson:"_id"`
	TransactionType TransactionType     `json:"transaction_type" bson:"transaction_type"`
	Reference       string              `json:"reference" bson:"reference"`
	ExternalRef     string              `json:"external_ref,omitempty" bson:"external_ref,omitempty"`
	Description     string              `json:"description" bson:"description"`
	Entries         []LedgerEntry       `json:"entries" bson:"entries"`
	TotalDebit      *big.Float          `json:"total_debit" bson:"total_debit"`
	TotalCredit     *big.Float          `json:"total_credit" bson:"total_credit"`
	Currency        string              `json:"currency" bson:"currency"`
	Status          JournalStatus       `json:"status" bson:"status"`
	Metadata        map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
	CreatedBy       uuid.UUID           `json:"created_by" bson:"created_by"`
	ApprovedBy      *uuid.UUID          `json:"approved_by,omitempty" bson:"approved_by,omitempty"`
	CreatedAt       time.Time           `json:"created_at" bson:"created_at"`
	ApprovedAt      *time.Time          `json:"approved_at,omitempty" bson:"approved_at,omitempty"`
	PostedAt        *time.Time          `json:"posted_at,omitempty" bson:"posted_at,omitempty"`
	Hash            string              `json:"hash" bson:"hash"`
	PreviousHash    string              `json:"previous_hash" bson:"previous_hash"`
	ResourceOwner   common.ResourceOwner `json:"-" bson:"resource_owner"`
}

// JournalStatus defines the status of a journal entry
type JournalStatus string

const (
	JournalStatusDraft    JournalStatus = "DRAFT"
	JournalStatusPending  JournalStatus = "PENDING"
	JournalStatusApproved JournalStatus = "APPROVED"
	JournalStatusPosted   JournalStatus = "POSTED"
	JournalStatusVoided   JournalStatus = "VOIDED"
	JournalStatusReversed JournalStatus = "REVERSED"
)

// LedgerService manages double-entry accounting
//
//nolint:unused // Some fields reserved for future in-memory caching
type LedgerService struct {
	mu           sync.Mutex                   //nolint:unused // Reserved for thread-safe operations
	accounts     map[uuid.UUID]*LedgerAccount //nolint:unused // Reserved for in-memory account cache
	journals     map[uuid.UUID]*JournalEntry  //nolint:unused // Reserved for in-memory journal cache
	lastHash     string                       //nolint:unused // Reserved for hash chain verification
	
	// System accounts (pre-defined)
	SystemCashAccount    uuid.UUID
	SystemRevenueAccount uuid.UUID
	SystemExpenseAccount uuid.UUID
	SystemHoldAccount    uuid.UUID
}

// NewJournalEntry creates a new journal entry
func NewJournalEntry(
	txType TransactionType,
	reference string,
	description string,
	currency string,
	createdBy uuid.UUID,
	rxn common.ResourceOwner,
) *JournalEntry {
	return &JournalEntry{
		ID:              uuid.New(),
		TransactionType: txType,
		Reference:       reference,
		Description:     description,
		Currency:        currency,
		Entries:         make([]LedgerEntry, 0),
		TotalDebit:      big.NewFloat(0),
		TotalCredit:     big.NewFloat(0),
		Status:          JournalStatusDraft,
		Metadata:        make(map[string]interface{}),
		CreatedBy:       createdBy,
		CreatedAt:       time.Now().UTC(),
		ResourceOwner:   rxn,
	}
}

// AddDebit adds a debit entry
func (j *JournalEntry) AddDebit(accountID uuid.UUID, accountCode string, amount *big.Float, description string) error {
	if amount.Cmp(big.NewFloat(0)) <= 0 {
		return errors.New("debit amount must be positive")
	}

	entry := LedgerEntry{
		ID:          uuid.New(),
		JournalID:   j.ID,
		AccountID:   accountID,
		AccountCode: accountCode,
		EntryType:   EntryTypeDebit,
		Amount:      amount,
		Currency:    j.Currency,
		Description: description,
		CreatedAt:   time.Now().UTC(),
	}

	j.Entries = append(j.Entries, entry)
	j.TotalDebit = new(big.Float).Add(j.TotalDebit, amount)

	return nil
}

// AddCredit adds a credit entry
func (j *JournalEntry) AddCredit(accountID uuid.UUID, accountCode string, amount *big.Float, description string) error {
	if amount.Cmp(big.NewFloat(0)) <= 0 {
		return errors.New("credit amount must be positive")
	}

	entry := LedgerEntry{
		ID:          uuid.New(),
		JournalID:   j.ID,
		AccountID:   accountID,
		AccountCode: accountCode,
		EntryType:   EntryTypeCredit,
		Amount:      amount,
		Currency:    j.Currency,
		Description: description,
		CreatedAt:   time.Now().UTC(),
	}

	j.Entries = append(j.Entries, entry)
	j.TotalCredit = new(big.Float).Add(j.TotalCredit, amount)

	return nil
}

// Validate ensures the journal entry follows double-entry rules
func (j *JournalEntry) Validate() error {
	if len(j.Entries) < 2 {
		return errors.New("journal entry must have at least 2 entries")
	}

	// Check debits equal credits (with small tolerance for floating point)
	tolerance := big.NewFloat(0.001)
	diff := new(big.Float).Sub(j.TotalDebit, j.TotalCredit)
	if diff.Abs(diff).Cmp(tolerance) > 0 {
		return fmt.Errorf("debits (%.2f) must equal credits (%.2f), difference: %.4f",
			floatFromBig(j.TotalDebit),
			floatFromBig(j.TotalCredit),
			floatFromBig(diff),
		)
	}

	// Check all entries have valid amounts
	for _, entry := range j.Entries {
		if entry.Amount.Cmp(big.NewFloat(0)) <= 0 {
			return fmt.Errorf("entry %s has invalid amount", entry.ID)
		}
		if entry.AccountID == uuid.Nil {
			return fmt.Errorf("entry %s has no account", entry.ID)
		}
	}

	return nil
}

// ComputeHash generates a SHA-256 hash for integrity
func (j *JournalEntry) ComputeHash(previousHash string) string {
	j.PreviousHash = previousHash

	data := struct {
		ID              uuid.UUID
		TransactionType TransactionType
		Reference       string
		TotalDebit      string
		TotalCredit     string
		Currency        string
		CreatedAt       time.Time
		PreviousHash    string
	}{
		ID:              j.ID,
		TransactionType: j.TransactionType,
		Reference:       j.Reference,
		TotalDebit:      j.TotalDebit.Text('f', 8),
		TotalCredit:     j.TotalCredit.Text('f', 8),
		Currency:        j.Currency,
		CreatedAt:       j.CreatedAt,
		PreviousHash:    previousHash,
	}

	jsonBytes, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonBytes)
	j.Hash = hex.EncodeToString(hash[:])

	return j.Hash
}

// MarkApproved marks the journal as approved
func (j *JournalEntry) MarkApproved(approverID uuid.UUID) error {
	if j.Status != JournalStatusPending && j.Status != JournalStatusDraft {
		return fmt.Errorf("cannot approve journal in status %s", j.Status)
	}

	if err := j.Validate(); err != nil {
		return err
	}

	now := time.Now().UTC()
	j.Status = JournalStatusApproved
	j.ApprovedBy = &approverID
	j.ApprovedAt = &now

	return nil
}

// MarkPosted marks the journal as posted (finalized)
func (j *JournalEntry) MarkPosted() error {
	if j.Status != JournalStatusApproved {
		return fmt.Errorf("cannot post journal in status %s", j.Status)
	}

	now := time.Now().UTC()
	j.Status = JournalStatusPosted
	j.PostedAt = &now

	return nil
}

// CreateReversal creates a reversal journal entry
func (j *JournalEntry) CreateReversal(reason string, createdBy uuid.UUID, rxn common.ResourceOwner) (*JournalEntry, error) {
	if j.Status != JournalStatusPosted {
		return nil, errors.New("can only reverse posted journals")
	}

	reversal := NewJournalEntry(
		j.TransactionType,
		"REV-"+j.Reference,
		fmt.Sprintf("Reversal of %s: %s", j.Reference, reason),
		j.Currency,
		createdBy,
		rxn,
	)

	// Swap debits and credits
	for _, entry := range j.Entries {
		if entry.EntryType == EntryTypeDebit {
			if err := reversal.AddCredit(entry.AccountID, entry.AccountCode, entry.Amount, "Reversal: "+entry.Description); err != nil {
				return nil, fmt.Errorf("failed to add credit entry for reversal: %w", err)
			}
		} else {
			if err := reversal.AddDebit(entry.AccountID, entry.AccountCode, entry.Amount, "Reversal: "+entry.Description); err != nil {
				return nil, fmt.Errorf("failed to add debit entry for reversal: %w", err)
			}
		}
	}

	reversal.Metadata["reversed_journal_id"] = j.ID.String()
	reversal.Metadata["reversal_reason"] = reason

	j.Status = JournalStatusReversed

	return reversal, nil
}

// LedgerWallet represents the ledger-integrated view of a user's wallet
// This provides the double-entry accounting perspective, complementing the domain UserWallet
type LedgerWallet struct {
	ID               uuid.UUID  `json:"id" bson:"_id"`
	UserID           uuid.UUID  `json:"user_id" bson:"user_id"`
	LedgerAccountID  uuid.UUID  `json:"ledger_account_id" bson:"ledger_account_id"`
	Currency         string     `json:"currency" bson:"currency"`
	Balance          *big.Float `json:"balance" bson:"balance"`
	AvailableBalance *big.Float `json:"available_balance" bson:"available_balance"`
	HeldBalance      *big.Float `json:"held_balance" bson:"held_balance"`
	TotalDeposits    *big.Float `json:"total_deposits" bson:"total_deposits"`
	TotalWithdrawals *big.Float `json:"total_withdrawals" bson:"total_withdrawals"`
	TotalWinnings    *big.Float `json:"total_winnings" bson:"total_winnings"`
	TotalLosses      *big.Float `json:"total_losses" bson:"total_losses"`
	TotalFees        *big.Float `json:"total_fees" bson:"total_fees"`
	CreatedAt        time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" bson:"updated_at"`
	Version          int        `json:"version" bson:"version"`
}

// NewLedgerWallet creates a new ledger-integrated wallet
func NewLedgerWallet(userID uuid.UUID, ledgerAccountID uuid.UUID, currency string) *LedgerWallet {
	now := time.Now().UTC()
	return &LedgerWallet{
		ID:               uuid.New(),
		UserID:           userID,
		LedgerAccountID:  ledgerAccountID,
		Currency:         currency,
		Balance:          big.NewFloat(0),
		AvailableBalance: big.NewFloat(0),
		HeldBalance:      big.NewFloat(0),
		TotalDeposits:    big.NewFloat(0),
		TotalWithdrawals: big.NewFloat(0),
		TotalWinnings:    big.NewFloat(0),
		TotalLosses:      big.NewFloat(0),
		TotalFees:        big.NewFloat(0),
		CreatedAt:        now,
		UpdatedAt:        now,
		Version:          1,
	}
}

// CanWithdraw checks if user can withdraw the specified amount
func (w *LedgerWallet) CanWithdraw(amount *big.Float) bool {
	return w.AvailableBalance.Cmp(amount) >= 0
}

// BalanceStatement represents a point-in-time balance statement
type BalanceStatement struct {
	WalletID         uuid.UUID  `json:"wallet_id"`
	UserID           uuid.UUID  `json:"user_id"`
	Currency         string     `json:"currency"`
	OpeningBalance   *big.Float `json:"opening_balance"`
	ClosingBalance   *big.Float `json:"closing_balance"`
	TotalDebits      *big.Float `json:"total_debits"`
	TotalCredits     *big.Float `json:"total_credits"`
	NetChange        *big.Float `json:"net_change"`
	PeriodStart      time.Time  `json:"period_start"`
	PeriodEnd        time.Time  `json:"period_end"`
	TransactionCount int        `json:"transaction_count"`
	GeneratedAt      time.Time  `json:"generated_at"`
}

// TrialBalance represents an accounting trial balance
type TrialBalance struct {
	AsOfDate     time.Time                `json:"as_of_date"`
	Accounts     []TrialBalanceAccount    `json:"accounts"`
	TotalDebits  *big.Float               `json:"total_debits"`
	TotalCredits *big.Float               `json:"total_credits"`
	IsBalanced   bool                     `json:"is_balanced"`
	Currency     string                   `json:"currency"`
	GeneratedAt  time.Time                `json:"generated_at"`
}

// TrialBalanceAccount represents an account in trial balance
type TrialBalanceAccount struct {
	AccountCode   string      `json:"account_code"`
	AccountName   string      `json:"account_name"`
	AccountType   AccountType `json:"account_type"`
	DebitBalance  *big.Float  `json:"debit_balance"`
	CreditBalance *big.Float  `json:"credit_balance"`
}

// TransactionBuilder helps build complex multi-leg transactions
type TransactionBuilder struct {
	journal     *JournalEntry
	accounts    map[uuid.UUID]*LedgerAccount
	validations []func() error
}

// NewTransactionBuilder creates a new transaction builder
func NewTransactionBuilder(
	txType TransactionType,
	reference string,
	description string,
	currency string,
	createdBy uuid.UUID,
	rxn common.ResourceOwner,
) *TransactionBuilder {
	return &TransactionBuilder{
		journal:     NewJournalEntry(txType, reference, description, currency, createdBy, rxn),
		accounts:    make(map[uuid.UUID]*LedgerAccount),
		validations: make([]func() error, 0),
	}
}

// Debit adds a debit entry
func (tb *TransactionBuilder) Debit(account *LedgerAccount, amount float64, description string) *TransactionBuilder {
	tb.accounts[account.ID] = account
	if err := tb.journal.AddDebit(account.ID, account.Code, big.NewFloat(amount), description); err != nil {
		tb.validations = append(tb.validations, func() error { return err })
	}
	return tb
}

// Credit adds a credit entry
func (tb *TransactionBuilder) Credit(account *LedgerAccount, amount float64, description string) *TransactionBuilder {
	tb.accounts[account.ID] = account
	if err := tb.journal.AddCredit(account.ID, account.Code, big.NewFloat(amount), description); err != nil {
		tb.validations = append(tb.validations, func() error { return err })
	}
	return tb
}

// WithMetadata adds metadata to the transaction
func (tb *TransactionBuilder) WithMetadata(key string, value interface{}) *TransactionBuilder {
	tb.journal.Metadata[key] = value
	return tb
}

// WithValidation adds a validation function
func (tb *TransactionBuilder) WithValidation(fn func() error) *TransactionBuilder {
	tb.validations = append(tb.validations, fn)
	return tb
}

// Build validates and returns the journal entry
func (tb *TransactionBuilder) Build() (*JournalEntry, error) {
	// Run custom validations
	for _, fn := range tb.validations {
		if err := fn(); err != nil {
			return nil, err
		}
	}

	// Validate journal entry
	if err := tb.journal.Validate(); err != nil {
		return nil, err
	}

	return tb.journal, nil
}

// Standard chart of accounts for gaming platform
var StandardChartOfAccounts = []struct {
	Code        string
	Name        string
	AccountType AccountType
}{
	// Assets (1xxx)
	{"1001", "Operating Cash Account", AccountTypeAsset},
	{"1002", "Stripe Settlement Account", AccountTypeAsset},
	{"1003", "Crypto Custody Account", AccountTypeAsset},
	{"1004", "Accounts Receivable", AccountTypeAsset},
	{"1005", "Held Funds Account", AccountTypeAsset},

	// Liabilities (2xxx)
	{"2001", "User Wallet Balances", AccountTypeLiability},
	{"2002", "Prize Pool Escrow", AccountTypeLiability},
	{"2003", "Pending Withdrawals", AccountTypeLiability},
	{"2004", "Subscription Credits", AccountTypeLiability},
	{"2005", "Held User Funds", AccountTypeLiability},

	// Equity (3xxx)
	{"3001", "Platform Equity", AccountTypeEquity},
	{"3002", "Retained Earnings", AccountTypeEquity},

	// Revenue (4xxx)
	{"4001", "Platform Fees Revenue", AccountTypeRevenue},
	{"4002", "Subscription Revenue", AccountTypeRevenue},
	{"4003", "Premium Feature Revenue", AccountTypeRevenue},
	{"4004", "Marketplace Commissions", AccountTypeRevenue},

	// Expenses (5xxx)
	{"5001", "Payment Processing Fees", AccountTypeExpense},
	{"5002", "Refunds Expense", AccountTypeExpense},
	{"5003", "Promotional Credits", AccountTypeExpense},
	{"5004", "Prize Pool Contributions", AccountTypeExpense},
}

// Helper function to convert big.Float to float64
func floatFromBig(f *big.Float) float64 {
	if f == nil {
		return 0
	}
	result, _ := f.Float64()
	return result
}

