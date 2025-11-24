// Package wallet_in defines inbound command interfaces for wallet operations
package wallet_in

import (
	"context"

	"github.com/google/uuid"
wallet_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/wallet/entities"
)// WalletCommand defines operations for wallet management
type WalletCommand interface {
	CreateWallet(ctx context.Context, cmd CreateWalletCommand) (*wallet_entities.UserWallet, error)
	Deposit(ctx context.Context, cmd DepositCommand) error
	Withdraw(ctx context.Context, cmd WithdrawCommand) error
	DeductEntryFee(ctx context.Context, cmd DeductEntryFeeCommand) error
	AddPrize(ctx context.Context, cmd AddPrizeCommand) error
	Refund(ctx context.Context, cmd RefundCommand) error
}

// CreateWalletCommand request to create a new wallet
type CreateWalletCommand struct {
	UserID     uuid.UUID
	EVMAddress string
}

// DepositCommand request to deposit funds
type DepositCommand struct {
	UserID   uuid.UUID
	Currency string
	Amount   float64
	TxHash   string
}

// WithdrawCommand request to withdraw funds
type WithdrawCommand struct {
	UserID   uuid.UUID
	Currency string
	Amount   float64
	ToAddress string
}

// DeductEntryFeeCommand request to deduct matchmaking entry fee
type DeductEntryFeeCommand struct {
	UserID   uuid.UUID
	Currency string
	Amount   float64
}

// AddPrizeCommand request to add prize winnings
type AddPrizeCommand struct {
	UserID   uuid.UUID
	Currency string
	Amount   float64
}

// RefundCommand request to refund amount
type RefundCommand struct {
	UserID   uuid.UUID
	Currency string
	Amount   float64
	Reason   string
}
