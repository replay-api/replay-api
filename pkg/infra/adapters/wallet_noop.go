package adapters

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	billing_out "github.com/replay-api/replay-api/pkg/domain/billing/ports/out"
)

// NoopWalletAdapter provides a no-op implementation for wallet operations
// Used when the wallet system is not fully implemented
type NoopWalletAdapter struct{}

// NewNoopWalletAdapter creates a new NoopWalletAdapter
func NewNoopWalletAdapter() *NoopWalletAdapter {
	return &NoopWalletAdapter{}
}

// GetBalance returns a placeholder balance for testing
func (a *NoopWalletAdapter) GetBalance(ctx context.Context, userID uuid.UUID, currency string) (float64, error) {
	slog.InfoContext(ctx, "NoopWalletAdapter.GetBalance called", "user_id", userID, "currency", currency)
	// Return a high balance for testing purposes
	return 100000.00, nil
}

// GetByID returns nil (wallet not found)
func (a *NoopWalletAdapter) GetByID(ctx context.Context, walletID uuid.UUID) (interface{}, error) {
	slog.InfoContext(ctx, "NoopWalletAdapter.GetByID called", "wallet_id", walletID)
	return map[string]interface{}{
		"id":       walletID,
		"balance":  100000.00,
		"currency": "USD",
	}, nil
}

// Debit logs the debit operation but doesn't do anything
func (a *NoopWalletAdapter) Debit(ctx context.Context, walletID uuid.UUID, amount float64, reference string) error {
	slog.InfoContext(ctx, "NoopWalletAdapter.Debit called",
		"wallet_id", walletID,
		"amount", amount,
		"reference", reference,
	)
	return nil
}

// Credit logs the credit operation but doesn't do anything
func (a *NoopWalletAdapter) Credit(ctx context.Context, walletID uuid.UUID, amount float64, reference string) error {
	slog.InfoContext(ctx, "NoopWalletAdapter.Credit called",
		"wallet_id", walletID,
		"amount", amount,
		"reference", reference,
	)
	return nil
}

// Ensure NoopWalletAdapter implements both interfaces
var _ billing_out.WalletReader = (*NoopWalletAdapter)(nil)
var _ billing_out.WalletDebiter = (*NoopWalletAdapter)(nil)

