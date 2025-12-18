package billing_out

import (
	"context"

	"github.com/google/uuid"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
)

// WithdrawalRepository defines persistence operations for withdrawals
type WithdrawalRepository interface {
	// Create stores a new withdrawal
	Create(ctx context.Context, withdrawal *billing_entities.Withdrawal) (*billing_entities.Withdrawal, error)

	// Update updates an existing withdrawal
	Update(ctx context.Context, withdrawal *billing_entities.Withdrawal) (*billing_entities.Withdrawal, error)

	// GetByID retrieves a withdrawal by ID
	GetByID(ctx context.Context, id uuid.UUID) (*billing_entities.Withdrawal, error)

	// GetByUserID retrieves withdrawals for a user
	GetByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]billing_entities.Withdrawal, error)

	// GetByStatus retrieves withdrawals by status
	GetByStatus(ctx context.Context, status billing_entities.WithdrawalStatus, limit int, offset int) ([]billing_entities.Withdrawal, error)

	// GetPending retrieves all pending withdrawals
	GetPending(ctx context.Context, limit int, offset int) ([]billing_entities.Withdrawal, error)
}

// WalletReader retrieves wallet information for withdrawal validation
type WalletReader interface {
	// GetBalance retrieves the current wallet balance for a user
	GetBalance(ctx context.Context, userID uuid.UUID, currency string) (float64, error)

	// GetByID retrieves a wallet by ID
	GetByID(ctx context.Context, walletID uuid.UUID) (interface{}, error)
}

// WalletDebiter handles debiting wallet balances
type WalletDebiter interface {
	// Debit removes funds from a wallet for withdrawal
	Debit(ctx context.Context, walletID uuid.UUID, amount float64, reference string) error

	// Credit returns funds to a wallet (for failed/canceled withdrawals)
	Credit(ctx context.Context, walletID uuid.UUID, amount float64, reference string) error
}

// PaymentProvider handles external payment processing for withdrawals
type PaymentProvider interface {
	// ProcessWithdrawal sends money to the recipient
	ProcessWithdrawal(ctx context.Context, withdrawal *billing_entities.Withdrawal) (string, error)

	// GetWithdrawalStatus checks the status of a withdrawal with the provider
	GetWithdrawalStatus(ctx context.Context, providerRef string) (string, error)
}

