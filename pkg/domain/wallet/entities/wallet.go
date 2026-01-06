package wallet_entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

// UserWallet is an aggregate root representing a user's blockchain wallet
type UserWallet struct {
	shared.BaseEntity
	EVMAddress          wallet_vo.EVMAddress                    `json:"evm_address" bson:"evm_address"`
	Balances            map[wallet_vo.Currency]wallet_vo.Amount `json:"balances" bson:"balances"`
	PendingTransactions []uuid.UUID                             `json:"pending_transactions" bson:"pending_transactions"`
	TotalDeposited      wallet_vo.Amount                        `json:"total_deposited" bson:"total_deposited"`
	TotalWithdrawn      wallet_vo.Amount                        `json:"total_withdrawn" bson:"total_withdrawn"`
	TotalPrizesWon      wallet_vo.Amount                        `json:"total_prizes_won" bson:"total_prizes_won"`
	DailyPrizeWinnings  wallet_vo.Amount                        `json:"daily_prize_winnings" bson:"daily_prize_winnings"` // Resets daily for anti-fraud
	LastPrizeWinDate    time.Time                               `json:"last_prize_win_date" bson:"last_prize_win_date"`
	IsLocked            bool                                    `json:"is_locked" bson:"is_locked"` // For fraud prevention
	LockReason          string                                  `json:"lock_reason,omitempty" bson:"lock_reason,omitempty"`
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

// Deposit adds funds to the wallet (invariant: balance never negative)
func (w *UserWallet) Deposit(currency wallet_vo.Currency, amount wallet_vo.Amount) error {
	if amount.IsNegative() || amount.IsZero() {
		return fmt.Errorf("deposit amount must be positive, got: %s", amount.String())
	}

	if w.IsLocked {
		return fmt.Errorf("wallet is locked: %s", w.LockReason)
	}

	currentBalance := w.GetBalance(currency)
	newBalance := currentBalance.Add(amount)

	w.Balances[currency] = newBalance
	w.TotalDeposited = w.TotalDeposited.Add(amount)
	w.UpdatedAt = time.Now()

	return nil
}

// Withdraw removes funds from the wallet (invariant: balance never negative)
func (w *UserWallet) Withdraw(currency wallet_vo.Currency, amount wallet_vo.Amount) error {
	if amount.IsNegative() || amount.IsZero() {
		return fmt.Errorf("withdrawal amount must be positive, got: %s", amount.String())
	}

	if w.IsLocked {
		return fmt.Errorf("wallet is locked: %s", w.LockReason)
	}

	currentBalance := w.GetBalance(currency)

	if currentBalance.LessThan(amount) {
		return fmt.Errorf("insufficient balance: have %s, need %s", currentBalance.String(), amount.String())
	}

	newBalance := currentBalance.Subtract(amount)
	w.Balances[currency] = newBalance
	w.TotalWithdrawn = w.TotalWithdrawn.Add(amount)
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
func (w *UserWallet) AddPrize(currency wallet_vo.Currency, amount wallet_vo.Amount, maxDailyWinnings wallet_vo.Amount) error {
	if amount.IsNegative() || amount.IsZero() {
		return fmt.Errorf("prize amount must be positive, got: %s", amount.String())
	}

	if w.IsLocked {
		return fmt.Errorf("wallet is locked: %s", w.LockReason)
	}

	// Reset daily counter if it's a new day
	now := time.Now()
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
	w.UpdatedAt = now

	return nil
}

// Lock locks the wallet for fraud investigation
func (w *UserWallet) Lock(reason string) {
	w.IsLocked = true
	w.LockReason = reason
	w.UpdatedAt = time.Now()
}

// Unlock unlocks the wallet
func (w *UserWallet) Unlock() {
	w.IsLocked = false
	w.LockReason = ""
	w.UpdatedAt = time.Now()
}

// AddPendingTransaction adds a transaction ID to pending list
func (w *UserWallet) AddPendingTransaction(txID uuid.UUID) {
	w.PendingTransactions = append(w.PendingTransactions, txID)
	w.UpdatedAt = time.Now()
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

// Helper function to check if two times are on the same day
func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// Validate ensures wallet invariants are maintained
func (w *UserWallet) Validate() error {
	if !w.EVMAddress.IsValid() {
		return fmt.Errorf("invalid EVM address")
	}

	for currency, balance := range w.Balances {
		if balance.IsNegative() {
			return fmt.Errorf("negative balance for currency %s: %s", currency, balance.String())
		}
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
