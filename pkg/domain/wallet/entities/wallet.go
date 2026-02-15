package wallet_entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// MaxPendingTransactions is the maximum number of concurrent pending transactions
// to prevent abuse or stuck transactions from accumulating
const MaxPendingTransactions = 50

// MaxSingleTransactionAmount is the maximum amount for a single transaction (in dollars)
// Transactions above this threshold require manual review
const MaxSingleTransactionAmount = 10000.00

// MinTransactionAmount is the minimum amount for a transaction (in dollars)
const MinTransactionAmount = 0.01

// MaxDailyPrizeWinnings is the default daily prize limit (in dollars)
const MaxDailyPrizeWinnings = 50.00

// UserWallet is an aggregate root representing a user's blockchain wallet
// Invariants:
//   - Balance must never be negative for any currency
//   - IsLocked wallets must reject all mutations
//   - Version must be checked for optimistic concurrency control
//   - PendingTransactions must not exceed MaxPendingTransactions
type UserWallet struct {
	shared.BaseEntity   `bson:"baseentity"`
	EVMAddress          wallet_vo.EVMAddress                    `json:"evm_address" bson:"evm_address"`
	Balances            map[wallet_vo.Currency]wallet_vo.Amount `json:"balances" bson:"balances"`
	PendingTransactions []uuid.UUID                             `json:"pending_transactions" bson:"pending_transactions"`
	TotalDeposited      wallet_vo.Amount                        `json:"total_deposited" bson:"total_deposited"`
	TotalWithdrawn      wallet_vo.Amount                        `json:"total_withdrawn" bson:"total_withdrawn"`
	TotalPrizesWon      wallet_vo.Amount                        `json:"total_prizes_won" bson:"total_prizes_won"`
	DailyPrizeWinnings  wallet_vo.Amount                        `json:"daily_prize_winnings" bson:"daily_prize_winnings"` // Resets daily for anti-fraud
	LastPrizeWinDate    time.Time                               `json:"last_prize_win_date" bson:"last_prize_win_date"`
	IsLocked            bool                                    `json:"is_locked" bson:"is_locked"`                       // For fraud prevention
	LockReason          string                                  `json:"lock_reason,omitempty" bson:"lock_reason,omitempty"`
	Version             int                                     `json:"version" bson:"version"` // Optimistic locking — prevents concurrent write conflicts
}

// NewUserWallet creates a new wallet with the given EVM address
func NewUserWallet(resourceOwner shared.ResourceOwner, evmAddress wallet_vo.EVMAddress) (*UserWallet, error) {
	if !evmAddress.IsValid() {
		return nil, fmt.Errorf("invalid EVM address: %s", evmAddress.String())
	}

	baseEntity := shared.NewPrivateEntity(resourceOwner) // Wallets are private to user

	wallet := &UserWallet{
		BaseEntity:          baseEntity,
		EVMAddress:          evmAddress,
		Balances:            make(map[wallet_vo.Currency]wallet_vo.Amount),
		PendingTransactions: []uuid.UUID{},
		TotalDeposited:      wallet_vo.NewAmount(0),
		TotalWithdrawn:      wallet_vo.NewAmount(0),
		TotalPrizesWon:      wallet_vo.NewAmount(0),
		DailyPrizeWinnings:  wallet_vo.NewAmount(0),
		LastPrizeWinDate:    time.Now(),
		IsLocked:            false,
		Version:             1,
	}

	// Initialize balances for supported currencies
	wallet.Balances[wallet_vo.CurrencyUSD] = wallet_vo.NewAmount(0)
	wallet.Balances[wallet_vo.CurrencyUSDC] = wallet_vo.NewAmount(0)
	wallet.Balances[wallet_vo.CurrencyUSDT] = wallet_vo.NewAmount(0)

	return wallet, nil
}

// GetBalance returns the balance for a specific currency
func (w *UserWallet) GetBalance(currency wallet_vo.Currency) wallet_vo.Amount {
	if balance, exists := w.Balances[currency]; exists {
		return balance
	}
	return wallet_vo.NewAmount(0)
}

// Deposit adds funds to the wallet
// Invariants enforced:
//   - Amount must be positive
//   - Wallet must not be locked
//   - Currency must be valid
//   - Amount must not exceed single transaction limit
func (w *UserWallet) Deposit(currency wallet_vo.Currency, amount wallet_vo.Amount) error {
	if amount.IsNegative() || amount.IsZero() {
		return fmt.Errorf("deposit amount must be positive, got: %s", amount.String())
	}

	if w.IsLocked {
		return fmt.Errorf("wallet is locked: %s", w.LockReason)
	}

	if !currency.IsValid() {
		return fmt.Errorf("unsupported currency: %s", currency)
	}

	if amount.ToFloat64() > MaxSingleTransactionAmount {
		return fmt.Errorf("deposit amount %s exceeds maximum single transaction limit of $%.2f",
			amount.String(), MaxSingleTransactionAmount)
	}

	currentBalance := w.GetBalance(currency)
	newBalance := currentBalance.Add(amount)

	w.Balances[currency] = newBalance
	w.TotalDeposited = w.TotalDeposited.Add(amount)
	w.Version++
	w.UpdatedAt = time.Now()

	return nil
}

// Withdraw removes funds from the wallet
// Invariants enforced:
//   - Amount must be positive
//   - Wallet must not be locked
//   - Currency must be valid
//   - Sufficient balance must exist
//   - Amount must not exceed single transaction limit
func (w *UserWallet) Withdraw(currency wallet_vo.Currency, amount wallet_vo.Amount) error {
	if amount.IsNegative() || amount.IsZero() {
		return fmt.Errorf("withdrawal amount must be positive, got: %s", amount.String())
	}

	if w.IsLocked {
		return fmt.Errorf("wallet is locked: %s", w.LockReason)
	}

	if !currency.IsValid() {
		return fmt.Errorf("unsupported currency: %s", currency)
	}

	if amount.ToFloat64() > MaxSingleTransactionAmount {
		return fmt.Errorf("withdrawal amount %s exceeds maximum single transaction limit of $%.2f",
			amount.String(), MaxSingleTransactionAmount)
	}

	currentBalance := w.GetBalance(currency)

	if currentBalance.LessThan(amount) {
		return fmt.Errorf("insufficient balance: have %s, need %s", currentBalance.String(), amount.String())
	}

	newBalance := currentBalance.Subtract(amount)
	w.Balances[currency] = newBalance
	w.TotalWithdrawn = w.TotalWithdrawn.Add(amount)
	w.Version++
	w.UpdatedAt = time.Now()

	return nil
}

// DeductEntryFee deducts matchmaking entry fee (with validation)
func (w *UserWallet) DeductEntryFee(currency wallet_vo.Currency, amount wallet_vo.Amount) error {
	if err := w.Withdraw(currency, amount); err != nil {
		return fmt.Errorf("failed to deduct entry fee: %w", err)
	}
	return nil
}

// AddPrize adds prize winnings with daily limit check (anti-fraud)
// Invariants enforced:
//   - Amount must be positive
//   - Wallet must not be locked
//   - Currency must be valid
//   - Daily prize winnings must not exceed maxDailyWinnings
func (w *UserWallet) AddPrize(currency wallet_vo.Currency, amount wallet_vo.Amount, maxDailyWinnings wallet_vo.Amount) error {
	if amount.IsNegative() || amount.IsZero() {
		return fmt.Errorf("prize amount must be positive, got: %s", amount.String())
	}

	if w.IsLocked {
		return fmt.Errorf("wallet is locked: %s", w.LockReason)
	}

	if !currency.IsValid() {
		return fmt.Errorf("unsupported currency: %s", currency)
	}

	// Reset daily counter if it's a new day (using UTC for consistency)
	now := time.Now().UTC()
	if !isSameDay(w.LastPrizeWinDate, now) {
		w.DailyPrizeWinnings = wallet_vo.NewAmount(0)
		w.LastPrizeWinDate = now
	}

	// Check daily limit (anti-fraud measure)
	newDailyTotal := w.DailyPrizeWinnings.Add(amount)
	if newDailyTotal.GreaterThan(maxDailyWinnings) {
		return fmt.Errorf("daily prize limit exceeded: current %s, attempting to add %s, limit %s",
			w.DailyPrizeWinnings.String(), amount.String(), maxDailyWinnings.String())
	}

	currentBalance := w.GetBalance(currency)
	newBalance := currentBalance.Add(amount)

	w.Balances[currency] = newBalance
	w.TotalPrizesWon = w.TotalPrizesWon.Add(amount)
	w.DailyPrizeWinnings = newDailyTotal
	w.Version++
	w.UpdatedAt = now

	return nil
}

// Lock locks the wallet for fraud investigation
func (w *UserWallet) Lock(reason string) {
	w.IsLocked = true
	w.LockReason = reason
	w.Version++
	w.UpdatedAt = time.Now()
}

// Unlock unlocks the wallet
func (w *UserWallet) Unlock() {
	w.IsLocked = false
	w.LockReason = ""
	w.Version++
	w.UpdatedAt = time.Now()
}

// AddPendingTransaction adds a transaction ID to pending list
func (w *UserWallet) AddPendingTransaction(txID uuid.UUID) error {
	if len(w.PendingTransactions) >= MaxPendingTransactions {
		return fmt.Errorf("too many pending transactions (%d), max allowed: %d",
			len(w.PendingTransactions), MaxPendingTransactions)
	}

	// Check for duplicate
	for _, id := range w.PendingTransactions {
		if id == txID {
			return nil // Already pending, idempotent
		}
	}

	w.PendingTransactions = append(w.PendingTransactions, txID)
	w.UpdatedAt = time.Now()
	return nil
}

// RemovePendingTransaction removes a transaction from pending list
func (w *UserWallet) RemovePendingTransaction(txID uuid.UUID) {
	filtered := make([]uuid.UUID, 0, len(w.PendingTransactions))
	for _, id := range w.PendingTransactions {
		if id != txID {
			filtered = append(filtered, id)
		}
	}
	w.PendingTransactions = filtered
	w.UpdatedAt = time.Now()
}

// Helper function to check if two times are on the same day (UTC)
func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.UTC().Date()
	y2, m2, d2 := t2.UTC().Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// Validate ensures wallet invariants are maintained
func (w *UserWallet) Validate() error {
	if !w.EVMAddress.IsValid() {
		return fmt.Errorf("invalid EVM address")
	}

	if w.Version < 1 {
		return fmt.Errorf("invalid wallet version: %d", w.Version)
	}

	for currency, balance := range w.Balances {
		if !currency.IsValid() {
			return fmt.Errorf("unsupported currency in balances: %s", currency)
		}
		if balance.IsNegative() {
			return fmt.Errorf("negative balance for currency %s: %s", currency, balance.String())
		}
	}

	if w.TotalDeposited.IsNegative() {
		return fmt.Errorf("negative total deposited: %s", w.TotalDeposited.String())
	}

	if w.TotalWithdrawn.IsNegative() {
		return fmt.Errorf("negative total withdrawn: %s", w.TotalWithdrawn.String())
	}

	if w.TotalPrizesWon.IsNegative() {
		return fmt.Errorf("negative total prizes won: %s", w.TotalPrizesWon.String())
	}

	if len(w.PendingTransactions) > MaxPendingTransactions {
		return fmt.Errorf("too many pending transactions: %d", len(w.PendingTransactions))
	}

	return nil
}

// TransactionStatus represents the status of a coordinated transaction
type TransactionStatus string

const (
	TransactionStatusPending    TransactionStatus = "Pending"
	TransactionStatusCompleted  TransactionStatus = "Completed"
	TransactionStatusFailed     TransactionStatus = "Failed"
	TransactionStatusRolledBack TransactionStatus = "RolledBack"
)

// WalletTransaction represents a coordinated wallet transaction
type WalletTransaction struct {
	ID           uuid.UUID              `json:"id"`
	WalletID     uuid.UUID              `json:"wallet_id"`
	Type         string                 `json:"type"` // Deposit, Withdrawal, EntryFee, Prize, Debit, Credit
	Status       TransactionStatus      `json:"status"`
	LedgerTxID   *uuid.UUID             `json:"ledger_tx_id,omitempty"`
	StartedAt    time.Time              `json:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
}
