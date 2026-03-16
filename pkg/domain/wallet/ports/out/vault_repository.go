package wallet_out

import (
	"context"

	"github.com/google/uuid"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// TeamVaultRepository defines persistence operations for team vaults
type TeamVaultRepository interface {
	shared.Searchable[wallet_entities.TeamVault]

	Save(ctx context.Context, vault *wallet_entities.TeamVault) error
	FindByID(ctx context.Context, id uuid.UUID) (*wallet_entities.TeamVault, error)
	Update(ctx context.Context, vault *wallet_entities.TeamVault) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindBySquadID(ctx context.Context, squadID uuid.UUID) (*wallet_entities.TeamVault, error)
	ExistsBySquadID(ctx context.Context, squadID uuid.UUID) (bool, error)
}

// VaultProposalRepository defines persistence operations for vault proposals
type VaultProposalRepository interface {
	shared.Searchable[wallet_entities.VaultProposal]

	Save(ctx context.Context, proposal *wallet_entities.VaultProposal) error
	FindByID(ctx context.Context, id uuid.UUID) (*wallet_entities.VaultProposal, error)
	Update(ctx context.Context, proposal *wallet_entities.VaultProposal) error
	FindByVaultID(ctx context.Context, vaultID uuid.UUID, limit, offset int) ([]wallet_entities.VaultProposal, int64, error)
	FindPendingByVaultID(ctx context.Context, vaultID uuid.UUID) ([]wallet_entities.VaultProposal, error)
	FindExpired(ctx context.Context) ([]wallet_entities.VaultProposal, error)
	CountByVaultIDAndStatus(ctx context.Context, vaultID uuid.UUID, status wallet_entities.ProposalStatus) (int64, error)
}

// VaultActivityRepository defines persistence operations for vault activity logs
type VaultActivityRepository interface {
	Append(ctx context.Context, activity *wallet_entities.VaultActivity) error
	FindByVaultID(ctx context.Context, vaultID uuid.UUID, limit, offset int) ([]wallet_entities.VaultActivity, int64, error)
	FindByActorID(ctx context.Context, vaultID uuid.UUID, actorID uuid.UUID, limit, offset int) ([]wallet_entities.VaultActivity, int64, error)
	FindByType(ctx context.Context, vaultID uuid.UUID, activityType wallet_entities.VaultActivityType, limit, offset int) ([]wallet_entities.VaultActivity, int64, error)
}

// InventoryItemRepository defines persistence operations for inventory items
type InventoryItemRepository interface {
	shared.Searchable[wallet_entities.InventoryItem]

	Save(ctx context.Context, item *wallet_entities.InventoryItem) error
	FindByID(ctx context.Context, id uuid.UUID) (*wallet_entities.InventoryItem, error)
	Update(ctx context.Context, item *wallet_entities.InventoryItem) error
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByOwner returns items owned by a specific user or team vault
	FindByOwner(ctx context.Context, ownerType wallet_entities.InventoryOwnerType, ownerID uuid.UUID, limit, offset int) ([]*wallet_entities.InventoryItem, int64, error)

	// FindByNFTContract finds an item by its on-chain NFT identifiers
	FindByNFTContract(ctx context.Context, chainID int, contractAddress, tokenID string) (*wallet_entities.InventoryItem, error)

	// TransferOwnership atomically transfers item ownership
	TransferOwnership(ctx context.Context, itemID uuid.UUID, newOwnerType wallet_entities.InventoryOwnerType, newOwnerID uuid.UUID) error
}
