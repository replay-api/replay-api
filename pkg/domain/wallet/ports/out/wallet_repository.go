// Package wallet_out defines outbound repository interfaces for wallet
package wallet_out

import (
	"context"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
)

// WalletRepository defines persistence operations for wallets
type WalletRepository interface {
	shared.Searchable[wallet_entities.UserWallet]

	Save(ctx context.Context, wallet *wallet_entities.UserWallet) error
	FindByID(ctx context.Context, id uuid.UUID) (*wallet_entities.UserWallet, error)
	Update(ctx context.Context, wallet *wallet_entities.UserWallet) error
	Delete(ctx context.Context, id uuid.UUID) error
}
