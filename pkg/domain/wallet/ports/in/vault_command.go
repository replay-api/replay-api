package wallet_in

import (
	"context"
	"time"

	"github.com/google/uuid"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

// TeamVaultCommand defines command operations for team vault management
type TeamVaultCommand interface {
	CreateVault(ctx context.Context, cmd CreateVaultCommand) (*wallet_entities.TeamVault, error)
	DepositToVault(ctx context.Context, cmd VaultDepositCommand) error
	ProposeTransaction(ctx context.Context, cmd ProposeTransactionCommand) (*wallet_entities.VaultProposal, error)
	ApproveProposal(ctx context.Context, cmd ApproveProposalCommand) error
	RejectProposal(ctx context.Context, cmd RejectProposalCommand) error
	CancelProposal(ctx context.Context, cmd CancelProposalCommand) error
	UpdateVaultSettings(ctx context.Context, cmd UpdateVaultSettingsCommand) (*wallet_entities.VaultProposal, error)
	DepositItem(ctx context.Context, cmd DepositItemCommand) error
	ProposeItemTransfer(ctx context.Context, cmd ProposeItemTransferCommand) (*wallet_entities.VaultProposal, error)
}

// CreateVaultCommand request to create a new team vault
type CreateVaultCommand struct {
	SquadID     uuid.UUID
	Name        string
	Description string
	UserID      uuid.UUID // Creating user (must be Owner/Admin)
}

func (c *CreateVaultCommand) Validate() error {
	if c.SquadID == uuid.Nil {
		return &ValidationError{Field: "squad_id", Message: "squad_id is required"}
	}
	if c.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	if c.Name == "" {
		return &ValidationError{Field: "name", Message: "vault name is required"}
	}
	return nil
}

// VaultDepositCommand request to deposit funds into a team vault from a personal wallet
type VaultDepositCommand struct {
	SquadID        uuid.UUID
	UserID         uuid.UUID
	Currency       string
	Amount         float64
	IdempotencyKey string
}

func (c *VaultDepositCommand) Validate() error {
	if c.SquadID == uuid.Nil {
		return &ValidationError{Field: "squad_id", Message: "squad_id is required"}
	}
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

// ProposeTransactionCommand request to create a withdrawal/transfer proposal
type ProposeTransactionCommand struct {
	SquadID     uuid.UUID
	UserID      uuid.UUID
	Type        wallet_entities.ProposalType
	Title       string
	Description string
	Amount      float64
	Currency    string
	Destination string // Wallet address or user ID for transfers
}

func (c *ProposeTransactionCommand) Validate() error {
	if c.SquadID == uuid.Nil {
		return &ValidationError{Field: "squad_id", Message: "squad_id is required"}
	}
	if c.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	if c.Title == "" {
		return &ValidationError{Field: "title", Message: "title is required"}
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

// ApproveProposalCommand request to approve a vault proposal
type ApproveProposalCommand struct {
	SquadID       uuid.UUID
	ProposalID    uuid.UUID
	UserID        uuid.UUID
	Reason        string
	SignatureHash *string // For on-chain signed approvals
}

func (c *ApproveProposalCommand) Validate() error {
	if c.SquadID == uuid.Nil {
		return &ValidationError{Field: "squad_id", Message: "squad_id is required"}
	}
	if c.ProposalID == uuid.Nil {
		return &ValidationError{Field: "proposal_id", Message: "proposal_id is required"}
	}
	if c.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	return nil
}

// RejectProposalCommand request to reject a vault proposal
type RejectProposalCommand struct {
	SquadID    uuid.UUID
	ProposalID uuid.UUID
	UserID     uuid.UUID
	Reason     string
}

func (c *RejectProposalCommand) Validate() error {
	if c.SquadID == uuid.Nil {
		return &ValidationError{Field: "squad_id", Message: "squad_id is required"}
	}
	if c.ProposalID == uuid.Nil {
		return &ValidationError{Field: "proposal_id", Message: "proposal_id is required"}
	}
	if c.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	if c.Reason == "" {
		return &ValidationError{Field: "reason", Message: "reason is required for rejections"}
	}
	return nil
}

// CancelProposalCommand request to cancel a vault proposal (only proposer or owner)
type CancelProposalCommand struct {
	SquadID    uuid.UUID
	ProposalID uuid.UUID
	UserID     uuid.UUID
}

func (c *CancelProposalCommand) Validate() error {
	if c.SquadID == uuid.Nil {
		return &ValidationError{Field: "squad_id", Message: "squad_id is required"}
	}
	if c.ProposalID == uuid.Nil {
		return &ValidationError{Field: "proposal_id", Message: "proposal_id is required"}
	}
	if c.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	return nil
}

// UpdateVaultSettingsCommand request to change vault settings (creates a proposal)
type UpdateVaultSettingsCommand struct {
	SquadID            uuid.UUID
	UserID             uuid.UUID
	ApprovalPolicy     *wallet_vo.ApprovalPolicy `json:"approval_policy,omitempty"`
	OnChainThreshold   *float64                  `json:"on_chain_threshold,omitempty"`
	DailyWithdrawLimit *float64                  `json:"daily_withdraw_limit,omitempty"`
}

func (c *UpdateVaultSettingsCommand) Validate() error {
	if c.SquadID == uuid.Nil {
		return &ValidationError{Field: "squad_id", Message: "squad_id is required"}
	}
	if c.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	if c.ApprovalPolicy != nil {
		if err := c.ApprovalPolicy.Validate(); err != nil {
			return &ValidationError{Field: "approval_policy", Message: err.Error()}
		}
	}
	return nil
}

// DepositItemCommand request to deposit an inventory item into the team vault
type DepositItemCommand struct {
	SquadID uuid.UUID
	UserID  uuid.UUID
	ItemID  uuid.UUID
}

func (c *DepositItemCommand) Validate() error {
	if c.SquadID == uuid.Nil {
		return &ValidationError{Field: "squad_id", Message: "squad_id is required"}
	}
	if c.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	if c.ItemID == uuid.Nil {
		return &ValidationError{Field: "item_id", Message: "item_id is required"}
	}
	return nil
}

// ProposeItemTransferCommand request to propose transferring items out of the vault
type ProposeItemTransferCommand struct {
	SquadID          uuid.UUID
	UserID           uuid.UUID
	Title            string
	Description      string
	InventoryItemIDs []uuid.UUID
	DestinationUserID uuid.UUID // User to transfer items to
}

func (c *ProposeItemTransferCommand) Validate() error {
	if c.SquadID == uuid.Nil {
		return &ValidationError{Field: "squad_id", Message: "squad_id is required"}
	}
	if c.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	if c.Title == "" {
		return &ValidationError{Field: "title", Message: "title is required"}
	}
	if len(c.InventoryItemIDs) == 0 {
		return &ValidationError{Field: "inventory_item_ids", Message: "at least one item is required"}
	}
	if c.DestinationUserID == uuid.Nil {
		return &ValidationError{Field: "destination_user_id", Message: "destination_user_id is required"}
	}
	return nil
}

// VaultAnalyticsTimeRange defines the time range for analytics queries
type VaultAnalyticsTimeRange struct {
	From time.Time
	To   time.Time
}
