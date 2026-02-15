// Package wallet_out defines outbound repository interfaces for wallet
package wallet_out

import (
	"context"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
)

// WalletRepository defines persistence operations for wallets
// SECURITY: All operations MUST enforce RLS (Resource-Level Security)
// through the context's ResourceOwner. Never expose cross-user access.
type WalletRepository interface {
	shared.Searchable[wallet_entities.UserWallet]

	Save(ctx context.Context, wallet *wallet_entities.UserWallet) error
	FindByID(ctx context.Context, id uuid.UUID) (*wallet_entities.UserWallet, error)

	// Update persists wallet changes. Implementations SHOULD check the wallet's
	// Version field for optimistic concurrency control and return an error
	// if the version has been incremented by another process.
	Update(ctx context.Context, wallet *wallet_entities.UserWallet) error
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByUserID retrieves a wallet by user ID (RLS-scoped)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*wallet_entities.UserWallet, error)

	// FindByEVMAddress retrieves a wallet by EVM address
	FindByEVMAddress(ctx context.Context, address string) (*wallet_entities.UserWallet, error)

	// ExistsByUserID checks if a wallet exists for a user (efficient existence check)
	ExistsByUserID(ctx context.Context, userID uuid.UUID) (bool, error)
}
