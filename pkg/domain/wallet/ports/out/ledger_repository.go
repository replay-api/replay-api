package wallet_out

import (
	"context"
	"time"

	"github.com/google/uuid"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

// LedgerRepository manages persistence of immutable ledger entries
type LedgerRepository interface {
	// CreateTransaction atomically creates all entries in a transaction
	// All entries must succeed or all fail (atomicity)
	CreateTransaction(ctx context.Context, entries []*wallet_entities.LedgerEntry) error

	// CreateEntry creates a single ledger entry
	CreateEntry(ctx context.Context, entry *wallet_entities.LedgerEntry) error

	// FindByID retrieves a ledger entry by ID
	FindByID(ctx context.Context, id uuid.UUID) (*wallet_entities.LedgerEntry, error)

	// FindByTransactionID retrieves all entries for a transaction
	// Returns both debit and credit entries (should always be 2+)
	FindByTransactionID(ctx context.Context, txID uuid.UUID) ([]*wallet_entities.LedgerEntry, error)

	// FindByAccountID retrieves all entries for an account (wallet)
	// Used for balance calculation and history
	FindByAccountID(ctx context.Context, accountID uuid.UUID, limit int, offset int) ([]*wallet_entities.LedgerEntry, error)

	// FindByAccountAndCurrency retrieves entries for specific currency
	FindByAccountAndCurrency(ctx context.Context, accountID uuid.UUID, currency wallet_vo.Currency) ([]*wallet_entities.LedgerEntry, error)

	// FindByIdempotencyKey checks if an entry with this key exists
	FindByIdempotencyKey(ctx context.Context, key string) (*wallet_entities.LedgerEntry, error)

	// ExistsByIdempotencyKey checks if an idempotency key has been used
	ExistsByIdempotencyKey(ctx context.Context, key string) bool

	// FindByDateRange retrieves entries within a date range
	// Used for reconciliation and reporting
	FindByDateRange(ctx context.Context, accountID uuid.UUID, from time.Time, to time.Time) ([]*wallet_entities.LedgerEntry, error)

	// CalculateBalance calculates the current balance for an account
	// Balance = SUM(debits) - SUM(credits) for asset accounts
	CalculateBalance(ctx context.Context, accountID uuid.UUID, currency wallet_vo.Currency) (wallet_vo.Amount, error)

	// GetAccountHistory retrieves transaction history with pagination
	GetAccountHistory(ctx context.Context, accountID uuid.UUID, filters HistoryFilters) ([]*wallet_entities.LedgerEntry, int64, error)

	// FindPendingApprovals retrieves entries pending manual review
	FindPendingApprovals(ctx context.Context, limit int) ([]*wallet_entities.LedgerEntry, error)

	// UpdateApprovalStatus updates the approval status (for manual review)
	UpdateApprovalStatus(ctx context.Context, entryID uuid.UUID, status wallet_entities.ApprovalStatus, approverID uuid.UUID) error

	// MarkAsReversed marks an entry as reversed
	MarkAsReversed(ctx context.Context, entryID uuid.UUID, reversalEntryID uuid.UUID) error

	// GetDailyTransactionCount gets count of transactions for an account in last 24h
	// Used for fraud detection
	GetDailyTransactionCount(ctx context.Context, accountID uuid.UUID) (int64, error)

	// GetDailyTransactionVolume gets total volume for an account in last 24h
	GetDailyTransactionVolume(ctx context.Context, accountID uuid.UUID, currency wallet_vo.Currency) (wallet_vo.Amount, error)

	// FindByUserAndDateRange for tax reporting
	FindByUserAndDateRange(ctx context.Context, userID uuid.UUID, from time.Time, to time.Time) ([]*wallet_entities.LedgerEntry, error)
}

// HistoryFilters defines filters for transaction history queries
type HistoryFilters struct {
	Currency      *wallet_vo.Currency                  `json:"currency,omitempty"`
	AssetType     *wallet_entities.AssetType           `json:"asset_type,omitempty"`
	EntryType     *wallet_entities.EntryType           `json:"entry_type,omitempty"`
	OperationType *string                              `json:"operation_type,omitempty"`
	FromDate      *time.Time                           `json:"from_date,omitempty"`
	ToDate        *time.Time                           `json:"to_date,omitempty"`
	MinAmount     *wallet_vo.Amount                    `json:"min_amount,omitempty"`
	MaxAmount     *wallet_vo.Amount                    `json:"max_amount,omitempty"`
	Limit         int                                  `json:"limit"`
	Offset        int                                  `json:"offset"`
	SortBy        string                               `json:"sort_by"` // "created_at", "amount"
	SortOrder     string                               `json:"sort_order"` // "asc", "desc"
}

// IdempotencyRepository manages idempotent operation tracking
type IdempotencyRepository interface {
	// Create creates a new idempotent operation record
	Create(ctx context.Context, op *wallet_entities.IdempotentOperation) error

	// FindByKey retrieves an idempotent operation by key
	FindByKey(ctx context.Context, key string) (*wallet_entities.IdempotentOperation, error)

	// Update updates an existing idempotent operation
	Update(ctx context.Context, op *wallet_entities.IdempotentOperation) error

	// Delete removes an idempotent operation (for cleanup)
	Delete(ctx context.Context, key string) error

	// FindStaleOperations finds operations stuck in "Processing" state
	// Used for cleanup/monitoring
	FindStaleOperations(ctx context.Context, threshold time.Duration) ([]*wallet_entities.IdempotentOperation, error)

	// CleanupExpired removes expired operations (called by cron job)
	// MongoDB TTL index should handle this automatically, but this is a fallback
	CleanupExpired(ctx context.Context) (int64, error)
}
