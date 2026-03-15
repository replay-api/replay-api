package wallet_vo

import (
	"fmt"
	"time"

	squad_vo "github.com/replay-api/replay-api/pkg/domain/squad/value-objects"
)

// ApprovalPolicy defines the multisig approval rules for a team vault.
// Tiers are evaluated from lowest MaxAmount to highest; the first matching tier applies.
type ApprovalPolicy struct {
	Tiers          []PolicyTier  `json:"tiers" bson:"tiers"`
	ProposalTTL    time.Duration `json:"proposal_ttl" bson:"proposal_ttl"`       // How long proposals stay open
	CooldownPeriod time.Duration `json:"cooldown_period" bson:"cooldown_period"` // Min time between proposals by same user
}

// PolicyTier defines approval requirements for a spending range
type PolicyTier struct {
	MaxAmount         Amount                            `json:"max_amount" bson:"max_amount"`                   // Up to this amount (inclusive)
	RequiredApprovals int                               `json:"required_approvals" bson:"required_approvals"`   // How many approvals needed
	AllowedRoles      []squad_vo.SquadMembershipType    `json:"allowed_roles" bson:"allowed_roles"`             // Roles that can approve
	AutoApprove       bool                              `json:"auto_approve" bson:"auto_approve"`               // Skip approval for this tier
	OnChainRequired   bool                              `json:"on_chain_required" bson:"on_chain_required"`     // Settle on-chain
}

// DefaultApprovalPolicy returns a sensible default policy for team vaults.
// Tier 1: Owner auto-approves up to $50
// Tier 2: <$500 requires 2 approvals from Admin or Owner
// Tier 3: ≥$500 requires Owner + Admin + on-chain settlement
func DefaultApprovalPolicy() ApprovalPolicy {
	return ApprovalPolicy{
		Tiers: []PolicyTier{
			{
				MaxAmount:         NewAmount(50),
				RequiredApprovals: 1,
				AllowedRoles:      []squad_vo.SquadMembershipType{squad_vo.SquadMembershipTypeOwner},
				AutoApprove:       true,
				OnChainRequired:   false,
			},
			{
				MaxAmount:         NewAmount(500),
				RequiredApprovals: 2,
				AllowedRoles:      []squad_vo.SquadMembershipType{squad_vo.SquadMembershipTypeOwner, squad_vo.SquadMembershipTypeAdmin},
				AutoApprove:       false,
				OnChainRequired:   false,
			},
			{
				MaxAmount:         NewAmount(100_000_000), // Effectively unlimited
				RequiredApprovals: 3,
				AllowedRoles:      []squad_vo.SquadMembershipType{squad_vo.SquadMembershipTypeOwner, squad_vo.SquadMembershipTypeAdmin},
				AutoApprove:       false,
				OnChainRequired:   true,
			},
		},
		ProposalTTL:    72 * time.Hour, // 3 days
		CooldownPeriod: 5 * time.Minute,
	}
}

// GetTierForAmount returns the applicable policy tier for a given amount
func (p ApprovalPolicy) GetTierForAmount(amount Amount) PolicyTier {
	for _, tier := range p.Tiers {
		if amount.Cents() <= tier.MaxAmount.Cents() {
			return tier
		}
	}
	// Fallback: use the last (highest) tier
	if len(p.Tiers) > 0 {
		return p.Tiers[len(p.Tiers)-1]
	}
	// Safety fallback — should never happen with a valid policy
	return PolicyTier{
		RequiredApprovals: 3,
		AutoApprove:       false,
		OnChainRequired:   true,
	}
}

// IsRoleAllowed checks if a role can approve at a given tier
func (t PolicyTier) IsRoleAllowed(role squad_vo.SquadMembershipType) bool {
	for _, r := range t.AllowedRoles {
		if r == role {
			return true
		}
	}
	return false
}

// Validate ensures the approval policy is well-formed
func (p ApprovalPolicy) Validate() error {
	if len(p.Tiers) == 0 {
		return fmt.Errorf("approval policy must have at least one tier")
	}
	if p.ProposalTTL <= 0 {
		return fmt.Errorf("proposal TTL must be positive")
	}

	for i, tier := range p.Tiers {
		if tier.RequiredApprovals <= 0 {
			return fmt.Errorf("tier %d: required_approvals must be positive", i)
		}
		if len(tier.AllowedRoles) == 0 {
			return fmt.Errorf("tier %d: must have at least one allowed role", i)
		}
		if i > 0 && tier.MaxAmount.Cents() <= p.Tiers[i-1].MaxAmount.Cents() {
			return fmt.Errorf("tier %d: max_amount must be greater than previous tier", i)
		}
	}
	return nil
}
