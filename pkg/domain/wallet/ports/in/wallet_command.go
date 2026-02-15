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
	UserID         uuid.UUID
	Currency       string
	Amount         float64
	TxHash         string
	ChainID        int    // Source chain for crypto deposits (0 = off-chain/fiat)
	PaymentMethod  string // crypto, credit_card, pix, bank_transfer
	IdempotencyKey string // Client-provided key for duplicate prevention
	SourceIP       string // Populated by controller from HTTP request
	UserAgent      string // Populated by controller from HTTP request
}

// Validate validates the command parameters
func (c *DepositCommand) Validate() error {
	if c.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	if c.Amount <= 0 {
		return &ValidationError{Field: "amount", Message: "amount must be positive"}
	}
	if c.Amount > wallet_entities.MaxSingleTransactionAmount {
		return &ValidationError{Field: "amount", Message: "amount exceeds maximum transaction limit"}
	}
	if c.Currency == "" {
		return &ValidationError{Field: "currency", Message: "currency is required"}
	}
	parsedCurrency, err := wallet_vo.ParseCurrency(c.Currency)
	if err != nil {
		return &ValidationError{Field: "currency", Message: "unsupported currency"}
	}
	// Crypto deposits require chain ID and tx hash
	if parsedCurrency.IsStablecoin() && c.PaymentMethod == "crypto" {
		if c.ChainID == 0 {
			return &ValidationError{Field: "chain_id", Message: "chain_id is required for crypto deposits"}
		}
		if _, err := wallet_vo.ParseChainID(c.ChainID); err != nil {
			return &ValidationError{Field: "chain_id", Message: err.Error()}
		}
	}
	return nil
}

// WithdrawCommand request to withdraw funds
type WithdrawCommand struct {
	UserID         uuid.UUID
	Currency       string
	Amount         float64
	ToAddress      string
	ChainID        int    // Target chain for crypto withdrawals (0 = off-chain/fiat)
	PaymentMethod  string // crypto, bank_transfer, pix
	IdempotencyKey string // Client-provided key for duplicate prevention
	SourceIP       string // Populated by controller from HTTP request
	UserAgent      string // Populated by controller from HTTP request
}

// Validate validates the command parameters
func (c *WithdrawCommand) Validate() error {
	if c.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	if c.Amount <= 0 {
		return &ValidationError{Field: "amount", Message: "amount must be positive"}
	}
	if c.Amount > wallet_entities.MaxSingleTransactionAmount {
		return &ValidationError{Field: "amount", Message: "amount exceeds maximum transaction limit"}
	}
	if c.Currency == "" {
		return &ValidationError{Field: "currency", Message: "currency is required"}
	}
	parsedCurrency, err := wallet_vo.ParseCurrency(c.Currency)
	if err != nil {
		return &ValidationError{Field: "currency", Message: "unsupported currency"}
	}
	if c.ToAddress == "" {
		return &ValidationError{Field: "to_address", Message: "to_address is required"}
	}
	// Crypto withdrawals require chain ID
	if parsedCurrency.IsStablecoin() && c.PaymentMethod == "crypto" {
		if c.ChainID == 0 {
			return &ValidationError{Field: "chain_id", Message: "chain_id is required for crypto withdrawals"}
		}
		if _, err := wallet_vo.ParseChainID(c.ChainID); err != nil {
			return &ValidationError{Field: "chain_id", Message: err.Error()}
		}
	}
	return nil
}

// DeductEntryFeeCommand request to deduct matchmaking entry fee
type DeductEntryFeeCommand struct {
	UserID       uuid.UUID
	Currency     string
	Amount       float64
	MatchID      *uuid.UUID // The match this entry fee is for
	TournamentID *uuid.UUID // The tournament this entry fee is for
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
	if _, err := wallet_vo.ParseCurrency(c.Currency); err != nil {
		return &ValidationError{Field: "currency", Message: "unsupported currency"}
	}
	return nil
}

// AddPrizeCommand request to add prize winnings
type AddPrizeCommand struct {
	UserID       uuid.UUID
	Currency     string
	Amount       float64
	MatchID      *uuid.UUID // The match this prize is from
	TournamentID *uuid.UUID // The tournament this prize is from
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
	if _, err := wallet_vo.ParseCurrency(c.Currency); err != nil {
		return &ValidationError{Field: "currency", Message: "unsupported currency"}
	}
	return nil
}

// RefundCommand request to refund amount
type RefundCommand struct {
	UserID         uuid.UUID
	Currency       string
	Amount         float64
	Reason         string
	OriginalTxID   *uuid.UUID // Original transaction being refunded (optional but recommended)
	IdempotencyKey string     // Client-provided key for duplicate prevention
	SourceIP       string     // Populated by controller from HTTP request
	UserAgent      string     // Populated by controller from HTTP request
}

// Validate validates the command parameters
func (c *RefundCommand) Validate() error {
	if c.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	if c.Amount <= 0 {
		return &ValidationError{Field: "amount", Message: "amount must be positive"}
	}
	if c.Amount > wallet_entities.MaxSingleTransactionAmount {
		return &ValidationError{Field: "amount", Message: "amount exceeds maximum transaction limit"}
	}
	if c.Currency == "" {
		return &ValidationError{Field: "currency", Message: "currency is required"}
	}
	if _, err := wallet_vo.ParseCurrency(c.Currency); err != nil {
		return &ValidationError{Field: "currency", Message: "unsupported currency"}
	}
	if c.Reason == "" {
		return &ValidationError{Field: "reason", Message: "reason is required for refunds"}
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
	if _, err := wallet_vo.ParseCurrency(c.Currency); err != nil {
		return &ValidationError{Field: "currency", Message: "unsupported currency"}
	}
	if c.Description == "" {
		return &ValidationError{Field: "description", Message: "description is required for debit operations"}
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
	if _, err := wallet_vo.ParseCurrency(c.Currency); err != nil {
		return &ValidationError{Field: "currency", Message: "unsupported currency"}
	}
	if c.Description == "" {
		return &ValidationError{Field: "description", Message: "description is required for credit operations"}
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
