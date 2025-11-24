// Package wallet_out defines outbound repository interfaces for wallet
package wallet_out

import (
"context"

"github.com/google/uuid"
wallet_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/wallet/entities"
)

// WalletRepository defines persistence operations for wallets
type WalletRepository interface {
	Save(ctx context.Context, wallet *wallet_entities.UserWallet) error
	FindByID(ctx context.Context, id uuid.UUID) (*wallet_entities.UserWallet, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*wallet_entities.UserWallet, error)
	FindByEVMAddress(ctx context.Context, address string) (*wallet_entities.UserWallet, error)
	Update(ctx context.Context, wallet *wallet_entities.UserWallet) error
	Delete(ctx context.Context, id uuid.UUID) error
}
