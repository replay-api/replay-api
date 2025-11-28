// Package wallet_in defines inbound query interfaces for wallet operations
package wallet_in

import (
	"context"
	"time"

	"github.com/google/uuid"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

// GetWalletBalanceQuery request to get wallet balance
type GetWalletBalanceQuery struct {
	UserID uuid.UUID
}

// Validate validates the query parameters
func (q *GetWalletBalanceQuery) Validate() error {
	if q.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	return nil
}

// WalletBalanceResult represents the result of a wallet balance query
type WalletBalanceResult struct {
	WalletID        uuid.UUID         `json:"wallet_id,omitempty"`
	UserID          uuid.UUID         `json:"user_id"`
	EVMAddress      string            `json:"evm_address,omitempty"`
	Balances        map[string]string `json:"balances"`
	TotalDeposited  string            `json:"total_deposited"`
	TotalWithdrawn  string            `json:"total_withdrawn"`
	TotalPrizesWon  string            `json:"total_prizes_won"`
	IsLocked        bool              `json:"is_locked"`
	LockReason      string            `json:"lock_reason,omitempty"`
	CreatedAt       *time.Time        `json:"created_at,omitempty"`
	UpdatedAt       *time.Time        `json:"updated_at,omitempty"`
}

// GetTransactionsQuery request to get wallet transactions
type GetTransactionsQuery struct {
	UserID   uuid.UUID
	WalletID uuid.UUID
	Filters  TransactionFilters
}

// Validate validates the query parameters
func (q *GetTransactionsQuery) Validate() error {
	if q.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	if q.Filters.Limit <= 0 {
		q.Filters.Limit = 50
	}
	if q.Filters.Limit > 100 {
		q.Filters.Limit = 100
	}
	return nil
}

// TransactionFilters defines filters for transaction history queries
type TransactionFilters struct {
	Currency      *wallet_vo.Currency           `json:"currency,omitempty"`
	AssetType     *wallet_entities.AssetType    `json:"asset_type,omitempty"`
	EntryType     *wallet_entities.EntryType    `json:"entry_type,omitempty"`
	OperationType *string                       `json:"operation_type,omitempty"`
	FromDate      *time.Time                    `json:"from_date,omitempty"`
	ToDate        *time.Time                    `json:"to_date,omitempty"`
	Limit         int                           `json:"limit"`
	Offset        int                           `json:"offset"`
	SortBy        string                        `json:"sort_by"`    // "created_at", "amount"
	SortOrder     string                        `json:"sort_order"` // "asc", "desc"
}

// TransactionsResult represents the result of a transactions query
type TransactionsResult struct {
	Transactions []TransactionDTO `json:"transactions"`
	TotalCount   int64            `json:"total_count"`
	Limit        int              `json:"limit"`
	Offset       int              `json:"offset"`
}

// TransactionDTO represents a transaction in API responses
type TransactionDTO struct {
	ID            uuid.UUID `json:"id"`
	TransactionID uuid.UUID `json:"transaction_id"`
	Type          string    `json:"type"`
	EntryType     string    `json:"entry_type"`
	AssetType     string    `json:"asset_type"`
	Currency      string    `json:"currency,omitempty"`
	Amount        string    `json:"amount"`
	BalanceAfter  string    `json:"balance_after"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
	IsReversed    bool      `json:"is_reversed"`
}

// WalletQuery defines query operations for wallet
type WalletQuery interface {
	GetBalance(ctx context.Context, query GetWalletBalanceQuery) (*WalletBalanceResult, error)
	GetTransactions(ctx context.Context, query GetTransactionsQuery) (*TransactionsResult, error)
}

// Note: ValidationError is defined in wallet_command.go (same package)
