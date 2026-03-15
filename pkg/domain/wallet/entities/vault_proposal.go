package wallet_entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	squad_vo "github.com/replay-api/replay-api/pkg/domain/squad/value-objects"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// ProposalType categorizes what a vault proposal is requesting
type ProposalType string

const (
	ProposalTypeWithdrawal      ProposalType = "WITHDRAWAL"
	ProposalTypeTransfer        ProposalType = "TRANSFER"
	ProposalTypeItemTransfer    ProposalType = "ITEM_TRANSFER"
	ProposalTypeSettingsChange  ProposalType = "SETTINGS_CHANGE"
)

// ProposalStatus tracks the lifecycle of a proposal
type ProposalStatus string

const (
	ProposalStatusPending   ProposalStatus = "PENDING"
	ProposalStatusApproved  ProposalStatus = "APPROVED"
	ProposalStatusExecuting ProposalStatus = "EXECUTING"
	ProposalStatusExecuted  ProposalStatus = "EXECUTED"
	ProposalStatusRejected  ProposalStatus = "REJECTED"
	ProposalStatusExpired   ProposalStatus = "EXPIRED"
	ProposalStatusCancelled ProposalStatus = "CANCELLED"
	ProposalStatusFailed    ProposalStatus = "FAILED"
)

// ApprovalDecision represents a signer's vote
type ApprovalDecision string

const (
	ApprovalDecisionApprove ApprovalDecision = "APPROVE"
	ApprovalDecisionReject  ApprovalDecision = "REJECT"
)

// VaultApproval records a single member's vote on a proposal
type VaultApproval struct {
	UserID        uuid.UUID                     `json:"user_id" bson:"user_id"`
	Role          squad_vo.SquadMembershipType   `json:"role" bson:"role"`
	Decision      ApprovalDecision              `json:"decision" bson:"decision"`
	Reason        string                        `json:"reason,omitempty" bson:"reason,omitempty"`
	SignatureHash *string                       `json:"signature_hash,omitempty" bson:"signature_hash,omitempty"` // For on-chain votes
	Timestamp     time.Time                     `json:"timestamp" bson:"timestamp"`
}

// VaultProposal represents a multisig transaction proposal for a team vault.
//
// Invariants:
//   - A user cannot vote twice on the same proposal
//   - Quorum check considers both RequiredApprovals count and role requirements
//   - Expired proposals cannot receive new votes
//   - Only the proposer or an Owner can cancel a pending proposal
type VaultProposal struct {
	shared.BaseEntity `bson:"baseentity"`

	// Core
	VaultID       uuid.UUID    `json:"vault_id" bson:"vault_id"`
	SquadID       uuid.UUID    `json:"squad_id" bson:"squad_id"`
	ProposerID    uuid.UUID    `json:"proposer_id" bson:"proposer_id"`
	ProposerRole  squad_vo.SquadMembershipType `json:"proposer_role" bson:"proposer_role"`
	Type          ProposalType   `json:"type" bson:"type"`
	Title         string         `json:"title" bson:"title"`
	Description   string         `json:"description" bson:"description"`

	// Financial (for withdrawal/transfer proposals)
	Amount   *wallet_vo.Amount   `json:"amount,omitempty" bson:"amount,omitempty"`
	Currency *wallet_vo.Currency `json:"currency,omitempty" bson:"currency,omitempty"`

	// Destination (for withdrawal/transfer)
	Destination   string `json:"destination,omitempty" bson:"destination,omitempty"` // Wallet address or user ID

	// Inventory (for item transfer proposals)
	InventoryItemIDs []uuid.UUID `json:"inventory_item_ids,omitempty" bson:"inventory_item_ids,omitempty"`

	// Settings change payload
	ProposedSettings *TeamVaultSettings `json:"proposed_settings,omitempty" bson:"proposed_settings,omitempty"`

	// Approval flow
	RequiredApprovals int             `json:"required_approvals" bson:"required_approvals"`
	Approvals         []VaultApproval `json:"approvals" bson:"approvals"`
	Rejections        []VaultApproval `json:"rejections" bson:"rejections"`
	Status            ProposalStatus  `json:"status" bson:"status"`

	// On-chain
	OnChain     bool       `json:"on_chain" bson:"on_chain"`
	ExecutedTxID *uuid.UUID `json:"executed_tx_id,omitempty" bson:"executed_tx_id,omitempty"`
	TxHash      string     `json:"tx_hash,omitempty" bson:"tx_hash,omitempty"`

	// Lifecycle
	ExpiresAt  time.Time  `json:"expires_at" bson:"expires_at"`
	ExecutedAt *time.Time `json:"executed_at,omitempty" bson:"executed_at,omitempty"`
}

// NewVaultProposal creates a new proposal with the given parameters
func NewVaultProposal(
	vaultID, squadID, proposerID uuid.UUID,
	proposerRole squad_vo.SquadMembershipType,
	proposalType ProposalType,
	title, description string,
	requiredApprovals int,
	expiresAt time.Time,
	onChain bool,
	resourceOwner shared.ResourceOwner,
) *VaultProposal {
	return &VaultProposal{
		BaseEntity:        shared.NewRestrictedEntity(resourceOwner),
		VaultID:           vaultID,
		SquadID:           squadID,
		ProposerID:        proposerID,
		ProposerRole:      proposerRole,
		Type:              proposalType,
		Title:             title,
		Description:       description,
		RequiredApprovals: requiredApprovals,
		Approvals:         []VaultApproval{},
		Rejections:        []VaultApproval{},
		Status:            ProposalStatusPending,
		OnChain:           onChain,
		ExpiresAt:         expiresAt,
	}
}

// GetID returns the proposal ID
func (p VaultProposal) GetID() uuid.UUID {
	return p.ID
}

// AddApproval records a member's approval vote
func (p *VaultProposal) AddApproval(userID uuid.UUID, role squad_vo.SquadMembershipType, reason string, signatureHash *string) error {
	if p.Status != ProposalStatusPending {
		return fmt.Errorf("cannot approve proposal in status %s", p.Status)
	}
	if p.IsExpired() {
		return fmt.Errorf("proposal has expired")
	}
	if p.HasUserVoted(userID) {
		return fmt.Errorf("user has already voted on this proposal")
	}

	p.Approvals = append(p.Approvals, VaultApproval{
		UserID:        userID,
		Role:          role,
		Decision:      ApprovalDecisionApprove,
		Reason:        reason,
		SignatureHash: signatureHash,
		Timestamp:     time.Now(),
	})

	p.UpdatedAt = time.Now()
	return nil
}

// AddRejection records a member's rejection vote
func (p *VaultProposal) AddRejection(userID uuid.UUID, role squad_vo.SquadMembershipType, reason string) error {
	if p.Status != ProposalStatusPending {
		return fmt.Errorf("cannot reject proposal in status %s", p.Status)
	}
	if p.IsExpired() {
		return fmt.Errorf("proposal has expired")
	}
	if p.HasUserVoted(userID) {
		return fmt.Errorf("user has already voted on this proposal")
	}

	p.Rejections = append(p.Rejections, VaultApproval{
		UserID:    userID,
		Role:      role,
		Decision:  ApprovalDecisionReject,
		Reason:    reason,
		Timestamp: time.Now(),
	})

	// If rejections exceed possible remaining approvers, auto-reject
	// (simplified: if rejections >= required, it's dead)
	if len(p.Rejections) >= p.RequiredApprovals {
		p.Status = ProposalStatusRejected
	}

	p.UpdatedAt = time.Now()
	return nil
}

// HasQuorum returns true if enough approvals have been received
func (p *VaultProposal) HasQuorum() bool {
	return len(p.Approvals) >= p.RequiredApprovals
}

// HasUserVoted checks if a user has already submitted a vote (approve or reject)
func (p *VaultProposal) HasUserVoted(userID uuid.UUID) bool {
	for _, a := range p.Approvals {
		if a.UserID == userID {
			return true
		}
	}
	for _, r := range p.Rejections {
		if r.UserID == userID {
			return true
		}
	}
	return false
}

// CanUserApprove checks if a user with the given role is allowed to approve
func (p *VaultProposal) CanUserApprove(userID uuid.UUID, role squad_vo.SquadMembershipType) bool {
	if p.Status != ProposalStatusPending {
		return false
	}
	if p.IsExpired() {
		return false
	}
	if p.HasUserVoted(userID) {
		return false
	}
	// Proposer cannot approve their own proposal (unless they're Owner with auto-approve)
	if userID == p.ProposerID {
		return false
	}
	// Owners and Admins can always approve
	return role == squad_vo.SquadMembershipTypeOwner || role == squad_vo.SquadMembershipTypeAdmin
}

// IsExpired checks if the proposal has passed its expiration time
func (p *VaultProposal) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

// MarkApproved transitions the proposal to APPROVED status
func (p *VaultProposal) MarkApproved() {
	p.Status = ProposalStatusApproved
	p.UpdatedAt = time.Now()
}

// MarkExecuting transitions the proposal to EXECUTING status
func (p *VaultProposal) MarkExecuting() {
	p.Status = ProposalStatusExecuting
	p.UpdatedAt = time.Now()
}

// MarkExecuted transitions the proposal to EXECUTED status
func (p *VaultProposal) MarkExecuted(txID *uuid.UUID, txHash string) {
	p.Status = ProposalStatusExecuted
	p.ExecutedTxID = txID
	p.TxHash = txHash
	now := time.Now()
	p.ExecutedAt = &now
	p.UpdatedAt = now
}

// MarkFailed transitions the proposal to FAILED status
func (p *VaultProposal) MarkFailed() {
	p.Status = ProposalStatusFailed
	p.UpdatedAt = time.Now()
}

// MarkExpired transitions the proposal to EXPIRED status
func (p *VaultProposal) MarkExpired() {
	p.Status = ProposalStatusExpired
	p.UpdatedAt = time.Now()
}

// MarkCancelled transitions the proposal to CANCELLED status
func (p *VaultProposal) MarkCancelled() {
	p.Status = ProposalStatusCancelled
	p.UpdatedAt = time.Now()
}

// Validate ensures the proposal is well-formed
func (p *VaultProposal) Validate() error {
	if p.VaultID == uuid.Nil {
		return fmt.Errorf("vault_id is required")
	}
	if p.SquadID == uuid.Nil {
		return fmt.Errorf("squad_id is required")
	}
	if p.ProposerID == uuid.Nil {
		return fmt.Errorf("proposer_id is required")
	}
	if p.Title == "" {
		return fmt.Errorf("title is required")
	}
	if p.RequiredApprovals <= 0 {
		return fmt.Errorf("required_approvals must be positive")
	}

	switch p.Type {
	case ProposalTypeWithdrawal, ProposalTypeTransfer:
		if p.Amount == nil || p.Amount.Cents() <= 0 {
			return fmt.Errorf("amount is required and must be positive for %s proposals", p.Type)
		}
		if p.Currency == nil {
			return fmt.Errorf("currency is required for %s proposals", p.Type)
		}
	case ProposalTypeItemTransfer:
		if len(p.InventoryItemIDs) == 0 {
			return fmt.Errorf("inventory_item_ids are required for item transfer proposals")
		}
	case ProposalTypeSettingsChange:
		if p.ProposedSettings == nil {
			return fmt.Errorf("proposed_settings is required for settings change proposals")
		}
	default:
		return fmt.Errorf("unknown proposal type: %s", p.Type)
	}

	return nil
}
