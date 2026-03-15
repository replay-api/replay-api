package wallet_entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// TeamVault represents a shared wallet belonging to a squad/team.
// It is NOT a replacement for individual user wallets — it is additional,
// team-owned storage for funds and inventory items.
//
// Invariants:
//   - Balance must never be negative for any currency
//   - IsLocked vaults reject all mutations
//   - Version is used for optimistic concurrency control
//   - Only active squad members can interact with the vault
type TeamVault struct {
	shared.BaseEntity `bson:"baseentity"`

	// Identity
	SquadID       uuid.UUID `json:"squad_id" bson:"squad_id"`
	Name          string    `json:"name" bson:"name"`
	Description   string    `json:"description" bson:"description"`
	SmartWalletID *uuid.UUID `json:"smart_wallet_id,omitempty" bson:"smart_wallet_id,omitempty"` // Links to custody SmartWallet for on-chain

	// Financial
	Balances         map[wallet_vo.Currency]wallet_vo.Amount `json:"balances" bson:"balances"`
	TotalDeposited   wallet_vo.Amount                        `json:"total_deposited" bson:"total_deposited"`
	TotalWithdrawn   wallet_vo.Amount                        `json:"total_withdrawn" bson:"total_withdrawn"`
	PendingProposals []uuid.UUID                             `json:"pending_proposals" bson:"pending_proposals"`

	// Configuration
	Settings TeamVaultSettings `json:"settings" bson:"settings"`

	// Status
	IsLocked   bool   `json:"is_locked" bson:"is_locked"`
	LockReason string `json:"lock_reason,omitempty" bson:"lock_reason,omitempty"`

	// Optimistic locking
	Version int `json:"version" bson:"version"`
}

// TeamVaultSettings holds configuration for a team vault
type TeamVaultSettings struct {
	ApprovalPolicy       wallet_vo.ApprovalPolicy `json:"approval_policy" bson:"approval_policy"`
	OnChainThreshold     wallet_vo.Amount         `json:"on_chain_threshold" bson:"on_chain_threshold"`         // Amount above which on-chain settlement is required
	DailyWithdrawLimit   wallet_vo.Amount         `json:"daily_withdraw_limit" bson:"daily_withdraw_limit"`
	WhitelistedAddresses []string                 `json:"whitelisted_addresses" bson:"whitelisted_addresses"`
}

// NewTeamVault creates a new team vault for a squad
func NewTeamVault(squadID uuid.UUID, name, description string, resourceOwner shared.ResourceOwner) *TeamVault {
	baseEntity := shared.NewRestrictedEntity(resourceOwner) // Restricted to squad members

	vault := &TeamVault{
		BaseEntity:       baseEntity,
		SquadID:          squadID,
		Name:             name,
		Description:      description,
		Balances:         make(map[wallet_vo.Currency]wallet_vo.Amount),
		TotalDeposited:   wallet_vo.NewAmount(0),
		TotalWithdrawn:   wallet_vo.NewAmount(0),
		PendingProposals: []uuid.UUID{},
		Settings: TeamVaultSettings{
			ApprovalPolicy:       wallet_vo.DefaultApprovalPolicy(),
			OnChainThreshold:     wallet_vo.NewAmount(500),
			DailyWithdrawLimit:   wallet_vo.NewAmount(10000),
			WhitelistedAddresses: []string{},
		},
		IsLocked: false,
		Version:  1,
	}

	// Initialize balances for supported currencies
	vault.Balances[wallet_vo.CurrencyUSD] = wallet_vo.NewAmount(0)
	vault.Balances[wallet_vo.CurrencyUSDC] = wallet_vo.NewAmount(0)
	vault.Balances[wallet_vo.CurrencyUSDT] = wallet_vo.NewAmount(0)

	return vault
}

// GetID returns the vault ID
func (v TeamVault) GetID() uuid.UUID {
	return v.ID
}

// Deposit adds funds to the vault
func (v *TeamVault) Deposit(currency wallet_vo.Currency, amount wallet_vo.Amount) error {
	if amount.Cents() <= 0 {
		return fmt.Errorf("deposit amount must be positive")
	}
	if v.IsLocked {
		return fmt.Errorf("vault is locked: %s", v.LockReason)
	}

	current := v.Balances[currency]
	v.Balances[currency] = wallet_vo.NewAmountFromCents(current.Cents() + amount.Cents())
	v.TotalDeposited = wallet_vo.NewAmountFromCents(v.TotalDeposited.Cents() + amount.Cents())
	v.Version++
	v.UpdatedAt = time.Now()
	return nil
}

// Withdraw deducts funds from the vault (called after proposal approval)
func (v *TeamVault) Withdraw(currency wallet_vo.Currency, amount wallet_vo.Amount) error {
	if amount.Cents() <= 0 {
		return fmt.Errorf("withdrawal amount must be positive")
	}
	if v.IsLocked {
		return fmt.Errorf("vault is locked: %s", v.LockReason)
	}

	current := v.Balances[currency]
	if current.Cents() < amount.Cents() {
		return fmt.Errorf("insufficient vault balance: have %s, need %s", current.String(), amount.String())
	}

	v.Balances[currency] = wallet_vo.NewAmountFromCents(current.Cents() - amount.Cents())
	v.TotalWithdrawn = wallet_vo.NewAmountFromCents(v.TotalWithdrawn.Cents() + amount.Cents())
	v.Version++
	v.UpdatedAt = time.Now()
	return nil
}

// GetBalance returns the balance for a given currency
func (v *TeamVault) GetBalance(currency wallet_vo.Currency) wallet_vo.Amount {
	if bal, ok := v.Balances[currency]; ok {
		return bal
	}
	return wallet_vo.NewAmount(0)
}

// AddPendingProposal tracks a pending proposal
func (v *TeamVault) AddPendingProposal(proposalID uuid.UUID) {
	v.PendingProposals = append(v.PendingProposals, proposalID)
	v.UpdatedAt = time.Now()
}

// RemovePendingProposal removes a resolved proposal from tracking
func (v *TeamVault) RemovePendingProposal(proposalID uuid.UUID) {
	for i, id := range v.PendingProposals {
		if id == proposalID {
			v.PendingProposals = append(v.PendingProposals[:i], v.PendingProposals[i+1:]...)
			break
		}
	}
	v.UpdatedAt = time.Now()
}

// Lock prevents any mutations on the vault
func (v *TeamVault) Lock(reason string) {
	v.IsLocked = true
	v.LockReason = reason
	v.Version++
	v.UpdatedAt = time.Now()
}

// Unlock allows mutations again
func (v *TeamVault) Unlock() {
	v.IsLocked = false
	v.LockReason = ""
	v.Version++
	v.UpdatedAt = time.Now()
}

// Validate ensures vault invariants
func (v *TeamVault) Validate() error {
	if v.SquadID == uuid.Nil {
		return fmt.Errorf("squad_id is required")
	}
	if v.Name == "" {
		return fmt.Errorf("vault name is required")
	}
	for currency, balance := range v.Balances {
		if balance.Cents() < 0 {
			return fmt.Errorf("negative balance for %s: %s", currency, balance.String())
		}
	}
	return v.Settings.ApprovalPolicy.Validate()
}
