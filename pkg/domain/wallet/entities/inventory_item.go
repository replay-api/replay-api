package wallet_entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// InventoryOwnerType distinguishes personal vs team inventory ownership
type InventoryOwnerType string

const (
	InventoryOwnerUser InventoryOwnerType = "USER"
	InventoryOwnerTeam InventoryOwnerType = "TEAM"
)

// InventoryItemType categorizes inventory items
type InventoryItemType string

const (
	ItemTypeGeneric    InventoryItemType = "GENERIC"
	ItemTypeNFT        InventoryItemType = "NFT"
	ItemTypeGameAsset  InventoryItemType = "GAME_ASSET"
	ItemTypeConsumable InventoryItemType = "CONSUMABLE"
	ItemTypeCosmetic   InventoryItemType = "COSMETIC"
	ItemTypeLootBox    InventoryItemType = "LOOT_BOX"
)

// ItemRarity defines rarity tiers for items
type ItemRarity string

const (
	RarityCommon    ItemRarity = "COMMON"
	RarityUncommon  ItemRarity = "UNCOMMON"
	RarityRare      ItemRarity = "RARE"
	RarityEpic      ItemRarity = "EPIC"
	RarityLegendary ItemRarity = "LEGENDARY"
)

// NFTStandard defines supported NFT token standards
type NFTStandard string

const (
	NFTStandardERC721  NFTStandard = "ERC-721"
	NFTStandardERC1155 NFTStandard = "ERC-1155"
)

// InventoryItemStatus tracks the lifecycle of an inventory item
type InventoryItemStatus string

const (
	ItemStatusActive          InventoryItemStatus = "ACTIVE"
	ItemStatusLocked          InventoryItemStatus = "LOCKED"
	ItemStatusPendingTransfer InventoryItemStatus = "PENDING_TRANSFER"
	ItemStatusTransferred     InventoryItemStatus = "TRANSFERRED"
	ItemStatusExpired         InventoryItemStatus = "EXPIRED"
	ItemStatusBurned          InventoryItemStatus = "BURNED"
)

// NFTMetadata holds on-chain NFT information
type NFTMetadata struct {
	ChainID         wallet_vo.ChainID `json:"chain_id" bson:"chain_id"`
	ContractAddress string            `json:"contract_address" bson:"contract_address"`
	TokenID         string            `json:"token_id" bson:"token_id"`
	Standard        NFTStandard       `json:"standard" bson:"standard"`
	MetadataURI     string            `json:"metadata_uri" bson:"metadata_uri"`
	LastSyncedAt    time.Time         `json:"last_synced_at" bson:"last_synced_at"`
}

// InventoryItem represents a digital item in a user's or team's inventory.
// Items are owner-agnostic — the same model serves both personal and team vaults.
//
// Invariants:
//   - Quantity must be positive for active items
//   - NFT items must have NFTData populated
//   - Locked items cannot be transferred or burned
type InventoryItem struct {
	shared.BaseEntity `bson:"baseentity"`

	// Ownership
	OwnerType InventoryOwnerType `json:"owner_type" bson:"owner_type"` // USER or TEAM
	OwnerID   uuid.UUID          `json:"owner_id" bson:"owner_id"`     // UserID or VaultID

	// Item details
	ItemType    InventoryItemType      `json:"item_type" bson:"item_type"`
	Name        string                 `json:"name" bson:"name"`
	Description string                 `json:"description" bson:"description"`
	ImageURI    string                 `json:"image_uri" bson:"image_uri"`
	Rarity      ItemRarity             `json:"rarity" bson:"rarity"`
	GameID      *string                `json:"game_id,omitempty" bson:"game_id,omitempty"` // Game-specific items
	Quantity    int64                  `json:"quantity" bson:"quantity"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"` // Game-specific attributes

	// NFT data (populated only for NFT items)
	NFTData *NFTMetadata `json:"nft_data,omitempty" bson:"nft_data,omitempty"`

	// Marketplace
	Tradeable    bool `json:"tradeable" bson:"tradeable"`
	Transferable bool `json:"transferable" bson:"transferable"`

	// Valuation (optional, for analytics)
	EstimatedValue *wallet_vo.Amount `json:"estimated_value,omitempty" bson:"estimated_value,omitempty"`

	// Lifecycle
	AcquiredAt time.Time            `json:"acquired_at" bson:"acquired_at"`
	ExpiresAt  *time.Time           `json:"expires_at,omitempty" bson:"expires_at,omitempty"`
	Status     InventoryItemStatus  `json:"status" bson:"status"`
}

// NewInventoryItem creates a new inventory item
func NewInventoryItem(
	ownerType InventoryOwnerType,
	ownerID uuid.UUID,
	itemType InventoryItemType,
	name, description, imageURI string,
	rarity ItemRarity,
	quantity int64,
	resourceOwner shared.ResourceOwner,
) *InventoryItem {
	visibility := shared.NewPrivateEntity(resourceOwner) // Private by default
	if ownerType == InventoryOwnerTeam {
		visibility = shared.NewRestrictedEntity(resourceOwner) // Restricted to team members
	}

	return &InventoryItem{
		BaseEntity:   visibility,
		OwnerType:    ownerType,
		OwnerID:      ownerID,
		ItemType:     itemType,
		Name:         name,
		Description:  description,
		ImageURI:     imageURI,
		Rarity:       rarity,
		Quantity:     quantity,
		Metadata:     make(map[string]interface{}),
		Tradeable:    true,
		Transferable: true,
		AcquiredAt:   time.Now(),
		Status:       ItemStatusActive,
	}
}

// GetID returns the item ID
func (i InventoryItem) GetID() uuid.UUID {
	return i.ID
}

// TransferOwnership transfers the item to a new owner
func (i *InventoryItem) TransferOwnership(newOwnerType InventoryOwnerType, newOwnerID uuid.UUID) error {
	if !i.Transferable {
		return fmt.Errorf("item is not transferable")
	}
	if i.Status == ItemStatusLocked {
		return fmt.Errorf("item is locked and cannot be transferred")
	}
	if i.Status != ItemStatusActive {
		return fmt.Errorf("item must be in active status to transfer, current: %s", i.Status)
	}

	i.OwnerType = newOwnerType
	i.OwnerID = newOwnerID
	i.UpdatedAt = time.Now()
	return nil
}

// Lock prevents the item from being transferred or burned
func (i *InventoryItem) Lock() {
	i.Status = ItemStatusLocked
	i.UpdatedAt = time.Now()
}

// Unlock allows transfers and burns again
func (i *InventoryItem) Unlock() {
	i.Status = ItemStatusActive
	i.UpdatedAt = time.Now()
}

// MarkPendingTransfer sets the item as pending a transfer
func (i *InventoryItem) MarkPendingTransfer() {
	i.Status = ItemStatusPendingTransfer
	i.UpdatedAt = time.Now()
}

// Burn permanently removes the item from circulation
func (i *InventoryItem) Burn() error {
	if i.Status == ItemStatusLocked {
		return fmt.Errorf("cannot burn a locked item")
	}
	i.Status = ItemStatusBurned
	i.Quantity = 0
	i.UpdatedAt = time.Now()
	return nil
}

// IsExpired checks if the item has expired
func (i *InventoryItem) IsExpired() bool {
	if i.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*i.ExpiresAt)
}

// Validate ensures the item is well-formed
func (i *InventoryItem) Validate() error {
	if i.OwnerID == uuid.Nil {
		return fmt.Errorf("owner_id is required")
	}
	if i.Name == "" {
		return fmt.Errorf("item name is required")
	}
	if i.Quantity <= 0 && i.Status == ItemStatusActive {
		return fmt.Errorf("quantity must be positive for active items")
	}
	if i.ItemType == ItemTypeNFT && i.NFTData == nil {
		return fmt.Errorf("nft_data is required for NFT items")
	}
	return nil
}
