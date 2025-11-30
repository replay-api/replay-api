// Package wallet_in defines inbound command interfaces for wallet operations
package wallet_in

import (
	"context"

	"github.com/google/uuid"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

// WalletCommand defines operations for wallet management
type WalletCommand interface {
	CreateWallet(ctx context.Context, cmd CreateWalletCommand) (*wallet_entities.UserWallet, error)
	Deposit(ctx context.Context, cmd DepositCommand) error
	Withdraw(ctx context.Context, cmd WithdrawCommand) error
	DeductEntryFee(ctx context.Context, cmd DeductEntryFeeCommand) error
	AddPrize(ctx context.Context, cmd AddPrizeCommand) error
	Refund(ctx context.Context, cmd RefundCommand) error
	DebitWallet(ctx context.Context, cmd DebitWalletCommand) (*wallet_entities.WalletTransaction, error)
	CreditWallet(ctx context.Context, cmd CreditWalletCommand) (*wallet_entities.WalletTransaction, error)
}

// CreateWalletCommand request to create a new wallet
type CreateWalletCommand struct {
	UserID     uuid.UUID
	EVMAddress string
}

// Validate validates the command parameters
func (c *CreateWalletCommand) Validate() error {
	if c.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	if c.EVMAddress == "" {
		return &ValidationError{Field: "evm_address", Message: "evm_address is required"}
	}
	return nil
}

// DepositCommand request to deposit funds
type DepositCommand struct {
	UserID   uuid.UUID
	Currency string
	Amount   float64
	TxHash   string
}

// Validate validates the command parameters
func (c *DepositCommand) Validate() error {
	if c.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	if c.Amount <= 0 {
		return &ValidationError{Field: "amount", Message: "amount must be positive"}
	}
	if c.Currency == "" {
		return &ValidationError{Field: "currency", Message: "currency is required"}
	}
	return nil
}

// WithdrawCommand request to withdraw funds
type WithdrawCommand struct {
	UserID    uuid.UUID
	Currency  string
	Amount    float64
	ToAddress string
}

// Validate validates the command parameters
func (c *WithdrawCommand) Validate() error {
	if c.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	if c.Amount <= 0 {
		return &ValidationError{Field: "amount", Message: "amount must be positive"}
	}
	if c.Currency == "" {
		return &ValidationError{Field: "currency", Message: "currency is required"}
	}
	if c.ToAddress == "" {
		return &ValidationError{Field: "to_address", Message: "to_address is required"}
	}
	return nil
}

// DeductEntryFeeCommand request to deduct matchmaking entry fee
type DeductEntryFeeCommand struct {
	UserID   uuid.UUID
	Currency string
	Amount   float64
}

// Validate validates the command parameters
func (c *DeductEntryFeeCommand) Validate() error {
	if c.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	if c.Amount <= 0 {
		return &ValidationError{Field: "amount", Message: "amount must be positive"}
	}
	if c.Currency == "" {
		return &ValidationError{Field: "currency", Message: "currency is required"}
	}
	return nil
}

// AddPrizeCommand request to add prize winnings
type AddPrizeCommand struct {
	UserID   uuid.UUID
	Currency string
	Amount   float64
}

// Validate validates the command parameters
func (c *AddPrizeCommand) Validate() error {
	if c.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	if c.Amount <= 0 {
		return &ValidationError{Field: "amount", Message: "amount must be positive"}
	}
	if c.Currency == "" {
		return &ValidationError{Field: "currency", Message: "currency is required"}
	}
	return nil
}

// RefundCommand request to refund amount
type RefundCommand struct {
	UserID   uuid.UUID
	Currency string
	Amount   float64
	Reason   string
}

// Validate validates the command parameters
func (c *RefundCommand) Validate() error {
	if c.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	if c.Amount <= 0 {
		return &ValidationError{Field: "amount", Message: "amount must be positive"}
	}
	if c.Currency == "" {
		return &ValidationError{Field: "currency", Message: "currency is required"}
	}
	return nil
}

// DebitWalletCommand request to debit wallet (generic debit operation)
type DebitWalletCommand struct {
	UserID      uuid.UUID
	Amount      wallet_vo.Amount
	Currency    string
	Description string
	Metadata    map[string]interface{}
}

// Validate validates the command parameters
func (c *DebitWalletCommand) Validate() error {
	if c.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	if c.Amount.IsZero() || c.Amount.IsNegative() {
		return &ValidationError{Field: "amount", Message: "amount must be positive"}
	}
	if c.Currency == "" {
		return &ValidationError{Field: "currency", Message: "currency is required"}
	}
	return nil
}

// CreditWalletCommand request to credit wallet (generic credit operation)
type CreditWalletCommand struct {
	UserID      uuid.UUID
	Amount      wallet_vo.Amount
	Currency    string
	Description string
	Metadata    map[string]interface{}
}

// Validate validates the command parameters
func (c *CreditWalletCommand) Validate() error {
	if c.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	if c.Amount.IsZero() || c.Amount.IsNegative() {
		return &ValidationError{Field: "amount", Message: "amount must be positive"}
	}
	if c.Currency == "" {
		return &ValidationError{Field: "currency", Message: "currency is required"}
	}
	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
