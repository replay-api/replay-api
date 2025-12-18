package billing_in

import (
	"context"

	"github.com/google/uuid"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
)

// CreateWithdrawalCommand represents a request to create a withdrawal
type CreateWithdrawalCommand struct {
	UserID      uuid.UUID
	WalletID    uuid.UUID
	Amount      float64
	Currency    string
	Method      billing_entities.WithdrawalMethod
	BankDetails billing_entities.BankDetails
}

// WithdrawalCommand defines the use case interface for withdrawal operations
type WithdrawalCommand interface {
	// Create creates a new withdrawal request
	Create(ctx context.Context, cmd CreateWithdrawalCommand) (*billing_entities.Withdrawal, error)

	// Cancel cancels a pending withdrawal
	Cancel(ctx context.Context, withdrawalID uuid.UUID) (*billing_entities.Withdrawal, error)

	// GetByID retrieves a withdrawal by ID
	GetByID(ctx context.Context, withdrawalID uuid.UUID) (*billing_entities.Withdrawal, error)

	// GetByUserID retrieves all withdrawals for a user
	GetByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]billing_entities.Withdrawal, error)
}

// WithdrawalAdminCommand defines admin operations for withdrawals
type WithdrawalAdminCommand interface {
	// Approve approves a withdrawal for processing
	Approve(ctx context.Context, withdrawalID uuid.UUID, reviewerID uuid.UUID) (*billing_entities.Withdrawal, error)

	// Reject rejects a withdrawal
	Reject(ctx context.Context, withdrawalID uuid.UUID, reviewerID uuid.UUID, reason string) (*billing_entities.Withdrawal, error)

	// Process marks a withdrawal as being processed
	Process(ctx context.Context, withdrawalID uuid.UUID) (*billing_entities.Withdrawal, error)

	// Complete marks a withdrawal as completed
	Complete(ctx context.Context, withdrawalID uuid.UUID, providerRef string) (*billing_entities.Withdrawal, error)

	// Fail marks a withdrawal as failed
	Fail(ctx context.Context, withdrawalID uuid.UUID, reason string) (*billing_entities.Withdrawal, error)

	// GetPending retrieves all pending withdrawals
	GetPending(ctx context.Context, limit int, offset int) ([]billing_entities.Withdrawal, error)
}

