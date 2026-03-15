package wallet_entities

import (
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// VaultActivityType categorizes vault events for the audit log
type VaultActivityType string

const (
	ActivityVaultCreated      VaultActivityType = "VAULT_CREATED"
	ActivityProposalCreated   VaultActivityType = "PROPOSAL_CREATED"
	ActivityApprovalSubmitted VaultActivityType = "APPROVAL_SUBMITTED"
	ActivityRejectionSubmitted VaultActivityType = "REJECTION_SUBMITTED"
	ActivityProposalExecuted  VaultActivityType = "PROPOSAL_EXECUTED"
	ActivityProposalCancelled VaultActivityType = "PROPOSAL_CANCELLED"
	ActivityProposalExpired   VaultActivityType = "PROPOSAL_EXPIRED"
	ActivityProposalFailed    VaultActivityType = "PROPOSAL_FAILED"
	ActivityDepositReceived   VaultActivityType = "DEPOSIT_RECEIVED"
	ActivityWithdrawalSent    VaultActivityType = "WITHDRAWAL_SENT"
	ActivityItemDeposited     VaultActivityType = "ITEM_DEPOSITED"
	ActivityItemTransferred   VaultActivityType = "ITEM_TRANSFERRED"
	ActivitySettingsChanged   VaultActivityType = "SETTINGS_CHANGED"
	ActivityVaultLocked       VaultActivityType = "VAULT_LOCKED"
	ActivityVaultUnlocked     VaultActivityType = "VAULT_UNLOCKED"
	ActivityMemberJoined      VaultActivityType = "MEMBER_JOINED"
	ActivityMemberLeft        VaultActivityType = "MEMBER_LEFT"
)

// VaultActivity represents an immutable, append-only audit log entry for a team vault.
// Follows the AuditTrailEntry pattern from the billing domain.
type VaultActivity struct {
	shared.BaseEntity `bson:"baseentity"`

	VaultID         uuid.UUID              `json:"vault_id" bson:"vault_id"`
	SquadID         uuid.UUID              `json:"squad_id" bson:"squad_id"`
	ActorID         uuid.UUID              `json:"actor_id" bson:"actor_id"`
	ActorName       string                 `json:"actor_name" bson:"actor_name"`
	ActivityType    VaultActivityType      `json:"activity_type" bson:"activity_type"`
	Description     string                 `json:"description" bson:"description"`
	RelatedEntityID *uuid.UUID             `json:"related_entity_id,omitempty" bson:"related_entity_id,omitempty"`
	Details         map[string]interface{} `json:"details,omitempty" bson:"details,omitempty"`
	Timestamp       time.Time              `json:"timestamp" bson:"timestamp"`
}

// NewVaultActivity creates a new vault activity log entry
func NewVaultActivity(
	vaultID, squadID, actorID uuid.UUID,
	actorName string,
	activityType VaultActivityType,
	description string,
	relatedEntityID *uuid.UUID,
	details map[string]interface{},
	resourceOwner shared.ResourceOwner,
) *VaultActivity {
	return &VaultActivity{
		BaseEntity:      shared.NewRestrictedEntity(resourceOwner),
		VaultID:         vaultID,
		SquadID:         squadID,
		ActorID:         actorID,
		ActorName:       actorName,
		ActivityType:    activityType,
		Description:     description,
		RelatedEntityID: relatedEntityID,
		Details:         details,
		Timestamp:       time.Now(),
	}
}

// GetID returns the activity ID
func (a VaultActivity) GetID() uuid.UUID {
	return a.ID
}
