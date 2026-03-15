package wallet_in

import (
	"context"
	"time"

	"github.com/google/uuid"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

// TeamVaultQuery defines query operations for team vault
type TeamVaultQuery interface {
	GetVaultBySquadID(ctx context.Context, squadID uuid.UUID) (*wallet_entities.TeamVault, error)
	GetVaultBalance(ctx context.Context, squadID uuid.UUID) (*VaultBalanceResult, error)
	GetProposals(ctx context.Context, squadID uuid.UUID, filters ProposalFilters) (*ProposalsResult, error)
	GetProposalByID(ctx context.Context, proposalID uuid.UUID) (*wallet_entities.VaultProposal, error)
	GetVaultActivity(ctx context.Context, squadID uuid.UUID, filters ActivityFilters) (*VaultActivityResult, error)
	GetVaultAnalytics(ctx context.Context, squadID uuid.UUID, timeRange VaultAnalyticsTimeRange) (*VaultAnalytics, error)
	GetVaultInventory(ctx context.Context, squadID uuid.UUID, filters InventoryFilters) (*VaultInventoryResult, error)
}

// VaultBalanceResult represents the balance of a team vault
type VaultBalanceResult struct {
	VaultID          uuid.UUID         `json:"vault_id"`
	SquadID          uuid.UUID         `json:"squad_id"`
	Name             string            `json:"name"`
	Balances         map[string]string `json:"balances"`
	TotalDeposited   string            `json:"total_deposited"`
	TotalWithdrawn   string            `json:"total_withdrawn"`
	PendingProposals int               `json:"pending_proposals"`
	IsLocked         bool              `json:"is_locked"`
	LockReason       string            `json:"lock_reason,omitempty"`
}

// ProposalFilters defines filters for listing proposals
type ProposalFilters struct {
	Status *wallet_entities.ProposalStatus `json:"status,omitempty"`
	Type   *wallet_entities.ProposalType   `json:"type,omitempty"`
	Limit  int                            `json:"limit"`
	Offset int                            `json:"offset"`
}

// ProposalsResult represents a paginated list of proposals
type ProposalsResult struct {
	Proposals  []VaultProposalDTO `json:"proposals"`
	TotalCount int64              `json:"total_count"`
	Limit      int                `json:"limit"`
	Offset     int                `json:"offset"`
}

// VaultProposalDTO represents a proposal in API responses
type VaultProposalDTO struct {
	ID                uuid.UUID                            `json:"id"`
	VaultID           uuid.UUID                            `json:"vault_id"`
	ProposerID        uuid.UUID                            `json:"proposer_id"`
	ProposerName      string                               `json:"proposer_name,omitempty"`
	Type              wallet_entities.ProposalType          `json:"type"`
	Title             string                               `json:"title"`
	Description       string                               `json:"description"`
	Amount            *string                              `json:"amount,omitempty"`
	Currency          *wallet_vo.Currency                   `json:"currency,omitempty"`
	Destination       string                               `json:"destination,omitempty"`
	InventoryItemIDs  []uuid.UUID                          `json:"inventory_item_ids,omitempty"`
	RequiredApprovals int                                  `json:"required_approvals"`
	CurrentApprovals  int                                  `json:"current_approvals"`
	Approvals         []VaultApprovalDTO                   `json:"approvals"`
	Rejections        []VaultApprovalDTO                   `json:"rejections"`
	Status            wallet_entities.ProposalStatus        `json:"status"`
	OnChain           bool                                 `json:"on_chain"`
	TxHash            string                               `json:"tx_hash,omitempty"`
	ExpiresAt         time.Time                            `json:"expires_at"`
	CreatedAt         time.Time                            `json:"created_at"`
	ExecutedAt        *time.Time                           `json:"executed_at,omitempty"`
}

// VaultApprovalDTO represents an approval vote in API responses
type VaultApprovalDTO struct {
	UserID    uuid.UUID `json:"user_id"`
	UserName  string    `json:"user_name,omitempty"`
	Role      string    `json:"role"`
	Decision  string    `json:"decision"`
	Reason    string    `json:"reason,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ActivityFilters defines filters for activity history
type ActivityFilters struct {
	ActivityType *wallet_entities.VaultActivityType `json:"activity_type,omitempty"`
	ActorID      *uuid.UUID                         `json:"actor_id,omitempty"`
	FromDate     *time.Time                         `json:"from_date,omitempty"`
	ToDate       *time.Time                         `json:"to_date,omitempty"`
	Limit        int                                `json:"limit"`
	Offset       int                                `json:"offset"`
}

// VaultActivityResult represents a paginated list of vault activities
type VaultActivityResult struct {
	Activities []VaultActivityDTO `json:"activities"`
	TotalCount int64              `json:"total_count"`
	Limit      int                `json:"limit"`
	Offset     int                `json:"offset"`
}

// VaultActivityDTO represents a vault activity in API responses
type VaultActivityDTO struct {
	ID              uuid.UUID              `json:"id"`
	ActorID         uuid.UUID              `json:"actor_id"`
	ActorName       string                 `json:"actor_name"`
	ActivityType    string                 `json:"activity_type"`
	Description     string                 `json:"description"`
	RelatedEntityID *uuid.UUID             `json:"related_entity_id,omitempty"`
	Details         map[string]interface{} `json:"details,omitempty"`
	Timestamp       time.Time              `json:"timestamp"`
}

// InventoryFilters defines filters for inventory queries
type InventoryFilters struct {
	ItemType *wallet_entities.InventoryItemType `json:"item_type,omitempty"`
	Rarity   *wallet_entities.ItemRarity        `json:"rarity,omitempty"`
	GameID   *string                            `json:"game_id,omitempty"`
	Status   *wallet_entities.InventoryItemStatus `json:"status,omitempty"`
	Limit    int                                `json:"limit"`
	Offset   int                                `json:"offset"`
}

// VaultInventoryResult represents a paginated list of inventory items
type VaultInventoryResult struct {
	Items      []InventoryItemDTO `json:"items"`
	TotalCount int64              `json:"total_count"`
	Limit      int                `json:"limit"`
	Offset     int                `json:"offset"`
}

// InventoryItemDTO represents an inventory item in API responses
type InventoryItemDTO struct {
	ID             uuid.UUID                         `json:"id"`
	ItemType       wallet_entities.InventoryItemType  `json:"item_type"`
	Name           string                            `json:"name"`
	Description    string                            `json:"description"`
	ImageURI       string                            `json:"image_uri"`
	Rarity         wallet_entities.ItemRarity         `json:"rarity"`
	GameID         *string                           `json:"game_id,omitempty"`
	Quantity       int64                             `json:"quantity"`
	Tradeable      bool                              `json:"tradeable"`
	Transferable   bool                              `json:"transferable"`
	NFTData        *NFTDataDTO                       `json:"nft_data,omitempty"`
	EstimatedValue *string                           `json:"estimated_value,omitempty"`
	AcquiredAt     time.Time                         `json:"acquired_at"`
	ExpiresAt      *time.Time                        `json:"expires_at,omitempty"`
	Status         wallet_entities.InventoryItemStatus `json:"status"`
}

// NFTDataDTO represents NFT metadata in API responses
type NFTDataDTO struct {
	ChainID         int    `json:"chain_id"`
	ContractAddress string `json:"contract_address"`
	TokenID         string `json:"token_id"`
	Standard        string `json:"standard"`
	MetadataURI     string `json:"metadata_uri"`
}

// VaultAnalytics represents analytics data for a team vault
type VaultAnalytics struct {
	VaultID         uuid.UUID             `json:"vault_id"`
	SquadID         uuid.UUID             `json:"squad_id"`
	TimeRange       VaultAnalyticsTimeRange `json:"time_range"`
	TotalIncome     string                `json:"total_income"`
	TotalExpenses   string                `json:"total_expenses"`
	NetFlow         string                `json:"net_flow"`
	TransactionCount int64               `json:"transaction_count"`
	ProposalCount   int64                 `json:"proposal_count"`
	ApprovalRate    float64               `json:"approval_rate"`        // Percentage of approved proposals
	AvgApprovalTime float64               `json:"avg_approval_time_hrs"` // Average hours to reach quorum
	TopContributors []ContributorSummary  `json:"top_contributors"`
	IncomeByType    map[string]string     `json:"income_by_type"`
	ExpenseByType   map[string]string     `json:"expense_by_type"`
	InventoryStats  InventoryStats        `json:"inventory_stats"`
}

// ContributorSummary represents a member's contribution to the vault
type ContributorSummary struct {
	UserID       uuid.UUID `json:"user_id"`
	UserName     string    `json:"user_name,omitempty"`
	TotalDeposit string    `json:"total_deposit"`
	TxCount      int64     `json:"transaction_count"`
}

// InventoryStats summarizes the vault's inventory
type InventoryStats struct {
	TotalItems     int64  `json:"total_items"`
	NFTCount       int64  `json:"nft_count"`
	TotalValue     string `json:"total_estimated_value"`
	ItemsByRarity  map[string]int64 `json:"items_by_rarity"`
	ItemsByType    map[string]int64 `json:"items_by_type"`
}
