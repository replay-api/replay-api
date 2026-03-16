package wallet_usecases

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_vo "github.com/replay-api/replay-api/pkg/domain/squad/value-objects"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_in "github.com/replay-api/replay-api/pkg/domain/wallet/ports/in"
	wallet_out "github.com/replay-api/replay-api/pkg/domain/wallet/ports/out"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// VaultService implements TeamVaultCommand and TeamVaultQuery
type VaultService struct {
	vaultRepo     wallet_out.TeamVaultRepository
	proposalRepo  wallet_out.VaultProposalRepository
	activityRepo  wallet_out.VaultActivityRepository
	inventoryRepo wallet_out.InventoryItemRepository
	walletRepo    wallet_out.WalletRepository
	squadReader   SquadReaderForVault
}

// SquadReaderForVault is a subset of squad reader needed by the vault service
type SquadReaderForVault interface {
	GetByID(ctx context.Context, id uuid.UUID) (*squad_entities.Squad, error)
}

// NewVaultService creates a new VaultService
func NewVaultService(
	vaultRepo wallet_out.TeamVaultRepository,
	proposalRepo wallet_out.VaultProposalRepository,
	activityRepo wallet_out.VaultActivityRepository,
	inventoryRepo wallet_out.InventoryItemRepository,
	walletRepo wallet_out.WalletRepository,
	squadReader SquadReaderForVault,
) *VaultService {
	return &VaultService{
		vaultRepo:     vaultRepo,
		proposalRepo:  proposalRepo,
		activityRepo:  activityRepo,
		inventoryRepo: inventoryRepo,
		walletRepo:    walletRepo,
		squadReader:   squadReader,
	}
}

// ---------- Commands ----------

// CreateVault creates a new team vault for a squad
func (s *VaultService) CreateVault(ctx context.Context, cmd wallet_in.CreateVaultCommand) (*wallet_entities.TeamVault, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	// Check if vault already exists for this squad
	exists, err := s.vaultRepo.ExistsBySquadID(ctx, cmd.SquadID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing vault: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("vault already exists for squad %s", cmd.SquadID)
	}

	// Verify user is an Owner or Admin of the squad
	role, err := s.getSquadMemberRole(ctx, cmd.SquadID, cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("authorization failed: %w", err)
	}
	if role != squad_vo.SquadMembershipTypeOwner && role != squad_vo.SquadMembershipTypeAdmin {
		return nil, fmt.Errorf("only squad Owner or Admin can create a vault")
	}

	resourceOwner := shared.GetResourceOwner(ctx)
	resourceOwner.GroupID = cmd.SquadID
	vault := wallet_entities.NewTeamVault(cmd.SquadID, cmd.Name, cmd.Description, resourceOwner)

	if err := s.vaultRepo.Save(ctx, vault); err != nil {
		return nil, fmt.Errorf("failed to save vault: %w", err)
	}

	// Log activity
	s.logActivity(ctx, vault.ID, cmd.SquadID, cmd.UserID, "",
		wallet_entities.ActivityVaultCreated,
		fmt.Sprintf("Vault '%s' created", cmd.Name),
		&vault.ID, nil, resourceOwner)

	slog.InfoContext(ctx, "Team vault created",
		"vault_id", vault.ID,
		"squad_id", cmd.SquadID,
		"created_by", cmd.UserID)

	return vault, nil
}

// DepositToVault deposits funds from a user's personal wallet into the team vault
func (s *VaultService) DepositToVault(ctx context.Context, cmd wallet_in.VaultDepositCommand) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	// Verify user is a squad member
	_, err := s.getSquadMemberRole(ctx, cmd.SquadID, cmd.UserID)
	if err != nil {
		return fmt.Errorf("authorization failed: %w", err)
	}

	vault, err := s.vaultRepo.FindBySquadID(ctx, cmd.SquadID)
	if err != nil {
		return fmt.Errorf("vault not found: %w", err)
	}

	currency, _ := wallet_vo.ParseCurrency(cmd.Currency)
	amount := wallet_vo.NewAmount(cmd.Amount)

	if err := vault.Deposit(currency, amount); err != nil {
		return fmt.Errorf("deposit failed: %w", err)
	}

	if err := s.vaultRepo.Update(ctx, vault); err != nil {
		return fmt.Errorf("failed to update vault: %w", err)
	}

	resourceOwner := shared.GetResourceOwner(ctx)
	s.logActivity(ctx, vault.ID, cmd.SquadID, cmd.UserID, "",
		wallet_entities.ActivityDepositReceived,
		fmt.Sprintf("Deposited %s %s", amount.String(), currency),
		nil,
		map[string]interface{}{"amount": amount.Dollars(), "currency": string(currency)},
		resourceOwner)

	return nil
}

// ProposeTransaction creates a new withdrawal/transfer proposal
func (s *VaultService) ProposeTransaction(ctx context.Context, cmd wallet_in.ProposeTransactionCommand) (*wallet_entities.VaultProposal, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	role, err := s.getSquadMemberRole(ctx, cmd.SquadID, cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("authorization failed: %w", err)
	}

	vault, err := s.vaultRepo.FindBySquadID(ctx, cmd.SquadID)
	if err != nil {
		return nil, fmt.Errorf("vault not found: %w", err)
	}

	if vault.IsLocked {
		return nil, fmt.Errorf("vault is locked: %s", vault.LockReason)
	}

	amount := wallet_vo.NewAmount(cmd.Amount)
	currency, _ := wallet_vo.ParseCurrency(cmd.Currency)

	// Check vault has sufficient balance
	balance := vault.GetBalance(currency)
	if balance.Cents() < amount.Cents() {
		return nil, fmt.Errorf("insufficient vault balance: have %s, need %s", balance.String(), amount.String())
	}

	// Determine approval requirements from policy
	tier := vault.Settings.ApprovalPolicy.GetTierForAmount(amount)
	onChain := tier.OnChainRequired || amount.Cents() >= vault.Settings.OnChainThreshold.Cents()
	expiresAt := time.Now().Add(vault.Settings.ApprovalPolicy.ProposalTTL)

	resourceOwner := shared.GetResourceOwner(ctx)
	resourceOwner.GroupID = cmd.SquadID

	proposal := wallet_entities.NewVaultProposal(
		vault.ID, cmd.SquadID, cmd.UserID, role,
		cmd.Type, cmd.Title, cmd.Description,
		tier.RequiredApprovals, expiresAt, onChain,
		resourceOwner,
	)
	proposal.Amount = &amount
	proposal.Currency = &currency
	proposal.Destination = cmd.Destination

	// Auto-approve if tier allows and proposer has sufficient role
	if tier.AutoApprove && tier.IsRoleAllowed(role) {
		if err := proposal.AddApproval(cmd.UserID, role, "Auto-approved by policy", nil); err != nil {
			slog.WarnContext(ctx, "Failed to auto-approve", "error", err)
		}
		if proposal.HasQuorum() {
			proposal.MarkApproved()
			// Execute immediately
			if execErr := s.executeProposal(ctx, vault, proposal); execErr != nil {
				slog.ErrorContext(ctx, "Auto-execute failed", "error", execErr)
				proposal.MarkFailed()
			}
		}
	}

	if err := proposal.Validate(); err != nil {
		return nil, fmt.Errorf("invalid proposal: %w", err)
	}

	if err := s.proposalRepo.Save(ctx, proposal); err != nil {
		return nil, fmt.Errorf("failed to save proposal: %w", err)
	}

	vault.AddPendingProposal(proposal.ID)
	if err := s.vaultRepo.Update(ctx, vault); err != nil {
		slog.WarnContext(ctx, "Failed to update vault pending proposals", "error", err)
	}

	s.logActivity(ctx, vault.ID, cmd.SquadID, cmd.UserID, "",
		wallet_entities.ActivityProposalCreated,
		fmt.Sprintf("Proposed %s: %s (%s %s)", cmd.Type, cmd.Title, amount.String(), currency),
		&proposal.ID,
		map[string]interface{}{"type": string(cmd.Type), "amount": amount.Dollars(), "required_approvals": tier.RequiredApprovals},
		resourceOwner)

	return proposal, nil
}

// ApproveProposal records an approval vote on a proposal
func (s *VaultService) ApproveProposal(ctx context.Context, cmd wallet_in.ApproveProposalCommand) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	role, err := s.getSquadMemberRole(ctx, cmd.SquadID, cmd.UserID)
	if err != nil {
		return fmt.Errorf("authorization failed: %w", err)
	}

	proposal, err := s.proposalRepo.FindByID(ctx, cmd.ProposalID)
	if err != nil {
		return fmt.Errorf("proposal not found: %w", err)
	}

	if !proposal.CanUserApprove(cmd.UserID, role) {
		return fmt.Errorf("user cannot approve this proposal")
	}

	if err := proposal.AddApproval(cmd.UserID, role, cmd.Reason, cmd.SignatureHash); err != nil {
		return err
	}

	// Check if quorum is met
	if proposal.HasQuorum() {
		proposal.MarkApproved()

		vault, vaultErr := s.vaultRepo.FindByID(ctx, proposal.VaultID)
		if vaultErr != nil {
			return fmt.Errorf("failed to load vault for execution: %w", vaultErr)
		}

		if execErr := s.executeProposal(ctx, vault, proposal); execErr != nil {
			slog.ErrorContext(ctx, "Proposal execution failed", "error", execErr, "proposal_id", cmd.ProposalID)
			proposal.MarkFailed()
		}
	}

	if err := s.proposalRepo.Update(ctx, proposal); err != nil {
		return fmt.Errorf("failed to update proposal: %w", err)
	}

	resourceOwner := shared.GetResourceOwner(ctx)
	s.logActivity(ctx, proposal.VaultID, cmd.SquadID, cmd.UserID, "",
		wallet_entities.ActivityApprovalSubmitted,
		fmt.Sprintf("Approved proposal: %s", proposal.Title),
		&cmd.ProposalID,
		map[string]interface{}{"current_approvals": len(proposal.Approvals), "required": proposal.RequiredApprovals},
		resourceOwner)

	return nil
}

// RejectProposal records a rejection vote on a proposal
func (s *VaultService) RejectProposal(ctx context.Context, cmd wallet_in.RejectProposalCommand) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	role, err := s.getSquadMemberRole(ctx, cmd.SquadID, cmd.UserID)
	if err != nil {
		return fmt.Errorf("authorization failed: %w", err)
	}

	proposal, err := s.proposalRepo.FindByID(ctx, cmd.ProposalID)
	if err != nil {
		return fmt.Errorf("proposal not found: %w", err)
	}

	if err := proposal.AddRejection(cmd.UserID, role, cmd.Reason); err != nil {
		return err
	}

	if err := s.proposalRepo.Update(ctx, proposal); err != nil {
		return fmt.Errorf("failed to update proposal: %w", err)
	}

	if proposal.Status == wallet_entities.ProposalStatusRejected {
		vault, _ := s.vaultRepo.FindByID(ctx, proposal.VaultID)
		if vault != nil {
			vault.RemovePendingProposal(proposal.ID)
			_ = s.vaultRepo.Update(ctx, vault)
		}
	}

	resourceOwner := shared.GetResourceOwner(ctx)
	s.logActivity(ctx, proposal.VaultID, cmd.SquadID, cmd.UserID, "",
		wallet_entities.ActivityRejectionSubmitted,
		fmt.Sprintf("Rejected proposal: %s — %s", proposal.Title, cmd.Reason),
		&cmd.ProposalID, nil, resourceOwner)

	return nil
}

// CancelProposal cancels a pending proposal (only proposer or owner)
func (s *VaultService) CancelProposal(ctx context.Context, cmd wallet_in.CancelProposalCommand) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	role, err := s.getSquadMemberRole(ctx, cmd.SquadID, cmd.UserID)
	if err != nil {
		return fmt.Errorf("authorization failed: %w", err)
	}

	proposal, err := s.proposalRepo.FindByID(ctx, cmd.ProposalID)
	if err != nil {
		return fmt.Errorf("proposal not found: %w", err)
	}

	if proposal.Status != wallet_entities.ProposalStatusPending {
		return fmt.Errorf("can only cancel pending proposals, current status: %s", proposal.Status)
	}

	// Only proposer or owner can cancel
	if cmd.UserID != proposal.ProposerID && role != squad_vo.SquadMembershipTypeOwner {
		return fmt.Errorf("only the proposer or squad owner can cancel a proposal")
	}

	proposal.MarkCancelled()

	if err := s.proposalRepo.Update(ctx, proposal); err != nil {
		return fmt.Errorf("failed to update proposal: %w", err)
	}

	vault, _ := s.vaultRepo.FindByID(ctx, proposal.VaultID)
	if vault != nil {
		vault.RemovePendingProposal(proposal.ID)
		_ = s.vaultRepo.Update(ctx, vault)
	}

	resourceOwner := shared.GetResourceOwner(ctx)
	s.logActivity(ctx, proposal.VaultID, cmd.SquadID, cmd.UserID, "",
		wallet_entities.ActivityProposalCancelled,
		fmt.Sprintf("Cancelled proposal: %s", proposal.Title),
		&cmd.ProposalID, nil, resourceOwner)

	return nil
}

// UpdateVaultSettings creates a proposal to change vault settings
func (s *VaultService) UpdateVaultSettings(ctx context.Context, cmd wallet_in.UpdateVaultSettingsCommand) (*wallet_entities.VaultProposal, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	role, err := s.getSquadMemberRole(ctx, cmd.SquadID, cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("authorization failed: %w", err)
	}
	if role != squad_vo.SquadMembershipTypeOwner {
		return nil, fmt.Errorf("only squad Owner can change vault settings")
	}

	vault, err := s.vaultRepo.FindBySquadID(ctx, cmd.SquadID)
	if err != nil {
		return nil, fmt.Errorf("vault not found: %w", err)
	}

	// Build proposed settings (merge with existing)
	proposedSettings := vault.Settings
	if cmd.ApprovalPolicy != nil {
		proposedSettings.ApprovalPolicy = *cmd.ApprovalPolicy
	}
	if cmd.OnChainThreshold != nil {
		proposedSettings.OnChainThreshold = wallet_vo.NewAmount(*cmd.OnChainThreshold)
	}
	if cmd.DailyWithdrawLimit != nil {
		proposedSettings.DailyWithdrawLimit = wallet_vo.NewAmount(*cmd.DailyWithdrawLimit)
	}

	resourceOwner := shared.GetResourceOwner(ctx)
	resourceOwner.GroupID = cmd.SquadID

	proposal := wallet_entities.NewVaultProposal(
		vault.ID, cmd.SquadID, cmd.UserID, role,
		wallet_entities.ProposalTypeSettingsChange,
		"Update Vault Settings",
		"Proposed changes to vault configuration",
		1, // Owner auto-approves settings
		time.Now().Add(vault.Settings.ApprovalPolicy.ProposalTTL),
		false,
		resourceOwner,
	)
	proposal.ProposedSettings = &proposedSettings

	// Owner auto-approves settings changes
	_ = proposal.AddApproval(cmd.UserID, role, "Owner-initiated settings change", nil)
	proposal.MarkApproved()

	// Execute immediately: apply new settings
	vault.Settings = proposedSettings
	vault.Version++
	vault.UpdatedAt = time.Now()

	if err := s.proposalRepo.Save(ctx, proposal); err != nil {
		return nil, fmt.Errorf("failed to save settings proposal: %w", err)
	}
	if err := s.vaultRepo.Update(ctx, vault); err != nil {
		return nil, fmt.Errorf("failed to update vault settings: %w", err)
	}

	txID := proposal.ID
	proposal.MarkExecuted(&txID, "")

	_ = s.proposalRepo.Update(ctx, proposal)

	s.logActivity(ctx, vault.ID, cmd.SquadID, cmd.UserID, "",
		wallet_entities.ActivitySettingsChanged,
		"Vault settings updated",
		&proposal.ID, nil, resourceOwner)

	return proposal, nil
}

// DepositItem deposits an inventory item from a user's personal inventory into the team vault
func (s *VaultService) DepositItem(ctx context.Context, cmd wallet_in.DepositItemCommand) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	_, err := s.getSquadMemberRole(ctx, cmd.SquadID, cmd.UserID)
	if err != nil {
		return fmt.Errorf("authorization failed: %w", err)
	}

	vault, err := s.vaultRepo.FindBySquadID(ctx, cmd.SquadID)
	if err != nil {
		return fmt.Errorf("vault not found: %w", err)
	}

	item, err := s.inventoryRepo.FindByID(ctx, cmd.ItemID)
	if err != nil {
		return fmt.Errorf("item not found: %w", err)
	}

	// Verify the user owns the item
	if item.OwnerType != wallet_entities.InventoryOwnerUser || item.OwnerID != cmd.UserID {
		return fmt.Errorf("user does not own this item")
	}

	if !item.Transferable {
		return fmt.Errorf("item is not transferable")
	}

	// Transfer ownership to team vault
	if err := item.TransferOwnership(wallet_entities.InventoryOwnerTeam, vault.ID); err != nil {
		return fmt.Errorf("transfer failed: %w", err)
	}

	if err := s.inventoryRepo.Update(ctx, item); err != nil {
		return fmt.Errorf("failed to update item ownership: %w", err)
	}

	resourceOwner := shared.GetResourceOwner(ctx)
	s.logActivity(ctx, vault.ID, cmd.SquadID, cmd.UserID, "",
		wallet_entities.ActivityItemDeposited,
		fmt.Sprintf("Deposited item: %s", item.Name),
		&cmd.ItemID,
		map[string]interface{}{"item_name": item.Name, "item_type": string(item.ItemType)},
		resourceOwner)

	return nil
}

// ProposeItemTransfer creates a proposal to transfer items out of the vault
func (s *VaultService) ProposeItemTransfer(ctx context.Context, cmd wallet_in.ProposeItemTransferCommand) (*wallet_entities.VaultProposal, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	role, err := s.getSquadMemberRole(ctx, cmd.SquadID, cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("authorization failed: %w", err)
	}

	vault, err := s.vaultRepo.FindBySquadID(ctx, cmd.SquadID)
	if err != nil {
		return nil, fmt.Errorf("vault not found: %w", err)
	}

	// Verify all items exist in the vault
	for _, itemID := range cmd.InventoryItemIDs {
		item, err := s.inventoryRepo.FindByID(ctx, itemID)
		if err != nil {
			return nil, fmt.Errorf("item %s not found: %w", itemID, err)
		}
		if item.OwnerType != wallet_entities.InventoryOwnerTeam || item.OwnerID != vault.ID {
			return nil, fmt.Errorf("item %s is not in this vault", itemID)
		}
	}

	// Use default 2-approval tier for item transfers
	requiredApprovals := 2
	expiresAt := time.Now().Add(vault.Settings.ApprovalPolicy.ProposalTTL)

	resourceOwner := shared.GetResourceOwner(ctx)
	resourceOwner.GroupID = cmd.SquadID

	proposal := wallet_entities.NewVaultProposal(
		vault.ID, cmd.SquadID, cmd.UserID, role,
		wallet_entities.ProposalTypeItemTransfer,
		cmd.Title, cmd.Description,
		requiredApprovals, expiresAt, false,
		resourceOwner,
	)
	proposal.InventoryItemIDs = cmd.InventoryItemIDs
	proposal.Destination = cmd.DestinationUserID.String()

	if err := proposal.Validate(); err != nil {
		return nil, fmt.Errorf("invalid proposal: %w", err)
	}

	if err := s.proposalRepo.Save(ctx, proposal); err != nil {
		return nil, fmt.Errorf("failed to save proposal: %w", err)
	}

	vault.AddPendingProposal(proposal.ID)
	_ = s.vaultRepo.Update(ctx, vault)

	s.logActivity(ctx, vault.ID, cmd.SquadID, cmd.UserID, "",
		wallet_entities.ActivityProposalCreated,
		fmt.Sprintf("Proposed item transfer: %s (%d items)", cmd.Title, len(cmd.InventoryItemIDs)),
		&proposal.ID, nil, resourceOwner)

	return proposal, nil
}

// ---------- Queries ----------

// GetVaultBySquadID returns the team vault for a squad
func (s *VaultService) GetVaultBySquadID(ctx context.Context, squadID uuid.UUID) (*wallet_entities.TeamVault, error) {
	return s.vaultRepo.FindBySquadID(ctx, squadID)
}

// GetVaultBalance returns the balance details for a team vault
func (s *VaultService) GetVaultBalance(ctx context.Context, squadID uuid.UUID) (*wallet_in.VaultBalanceResult, error) {
	vault, err := s.vaultRepo.FindBySquadID(ctx, squadID)
	if err != nil {
		return nil, fmt.Errorf("vault not found: %w", err)
	}

	balances := make(map[string]string)
	for currency, amount := range vault.Balances {
		balances[string(currency)] = amount.String()
	}

	pending, _ := s.proposalRepo.CountByVaultIDAndStatus(ctx, vault.ID, wallet_entities.ProposalStatusPending)

	return &wallet_in.VaultBalanceResult{
		VaultID:          vault.ID,
		SquadID:          vault.SquadID,
		Name:             vault.Name,
		Balances:         balances,
		TotalDeposited:   vault.TotalDeposited.String(),
		TotalWithdrawn:   vault.TotalWithdrawn.String(),
		PendingProposals: int(pending),
		IsLocked:         vault.IsLocked,
		LockReason:       vault.LockReason,
	}, nil
}

// GetProposals returns paginated proposals for a vault
func (s *VaultService) GetProposals(ctx context.Context, squadID uuid.UUID, filters wallet_in.ProposalFilters) (*wallet_in.ProposalsResult, error) {
	vault, err := s.vaultRepo.FindBySquadID(ctx, squadID)
	if err != nil {
		return nil, fmt.Errorf("vault not found: %w", err)
	}

	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}

	proposals, total, err := s.proposalRepo.FindByVaultID(ctx, vault.ID, filters.Limit, filters.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposals: %w", err)
	}

	dtos := make([]wallet_in.VaultProposalDTO, len(proposals))
	for i, p := range proposals {
		dtos[i] = s.proposalToDTO(p)
	}

	return &wallet_in.ProposalsResult{
		Proposals:  dtos,
		TotalCount: total,
		Limit:      filters.Limit,
		Offset:     filters.Offset,
	}, nil
}

// GetProposalByID returns a specific proposal
func (s *VaultService) GetProposalByID(ctx context.Context, proposalID uuid.UUID) (*wallet_entities.VaultProposal, error) {
	return s.proposalRepo.FindByID(ctx, proposalID)
}

// GetVaultActivity returns paginated activity history for a vault
func (s *VaultService) GetVaultActivity(ctx context.Context, squadID uuid.UUID, filters wallet_in.ActivityFilters) (*wallet_in.VaultActivityResult, error) {
	vault, err := s.vaultRepo.FindBySquadID(ctx, squadID)
	if err != nil {
		return nil, fmt.Errorf("vault not found: %w", err)
	}

	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}

	activities, total, err := s.activityRepo.FindByVaultID(ctx, vault.ID, filters.Limit, filters.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}

	dtos := make([]wallet_in.VaultActivityDTO, len(activities))
	for i, a := range activities {
		dtos[i] = wallet_in.VaultActivityDTO{
			ID:              a.ID,
			ActorID:         a.ActorID,
			ActorName:       a.ActorName,
			ActivityType:    string(a.ActivityType),
			Description:     a.Description,
			RelatedEntityID: a.RelatedEntityID,
			Details:         a.Details,
			Timestamp:       a.Timestamp,
		}
	}

	return &wallet_in.VaultActivityResult{
		Activities: dtos,
		TotalCount: total,
		Limit:      filters.Limit,
		Offset:     filters.Offset,
	}, nil
}

// GetVaultAnalytics returns analytics for a vault within a time range
func (s *VaultService) GetVaultAnalytics(ctx context.Context, squadID uuid.UUID, timeRange wallet_in.VaultAnalyticsTimeRange) (*wallet_in.VaultAnalytics, error) {
	vault, err := s.vaultRepo.FindBySquadID(ctx, squadID)
	if err != nil {
		return nil, fmt.Errorf("vault not found: %w", err)
	}

	// Basic analytics from vault data
	return &wallet_in.VaultAnalytics{
		VaultID:          vault.ID,
		SquadID:          vault.SquadID,
		TimeRange:        timeRange,
		TotalIncome:      vault.TotalDeposited.String(),
		TotalExpenses:    vault.TotalWithdrawn.String(),
		NetFlow:          wallet_vo.NewAmountFromCents(vault.TotalDeposited.Cents() - vault.TotalWithdrawn.Cents()).String(),
		TransactionCount: 0,
		ProposalCount:    int64(len(vault.PendingProposals)),
		ApprovalRate:     0,
		AvgApprovalTime:  0,
		TopContributors:  []wallet_in.ContributorSummary{},
		IncomeByType:     map[string]string{},
		ExpenseByType:    map[string]string{},
		InventoryStats: wallet_in.InventoryStats{
			TotalItems:    0,
			NFTCount:      0,
			TotalValue:    "$0.00",
			ItemsByRarity: map[string]int64{},
			ItemsByType:   map[string]int64{},
		},
	}, nil
}

// GetVaultInventory returns paginated inventory items for a vault
func (s *VaultService) GetVaultInventory(ctx context.Context, squadID uuid.UUID, filters wallet_in.InventoryFilters) (*wallet_in.VaultInventoryResult, error) {
	vault, err := s.vaultRepo.FindBySquadID(ctx, squadID)
	if err != nil {
		return nil, fmt.Errorf("vault not found: %w", err)
	}

	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}

	items, total, err := s.inventoryRepo.FindByOwner(ctx, wallet_entities.InventoryOwnerTeam, vault.ID, filters.Limit, filters.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	dtos := make([]wallet_in.InventoryItemDTO, len(items))
	for i, item := range items {
		dtos[i] = s.inventoryItemToDTO(item)
	}

	return &wallet_in.VaultInventoryResult{
		Items:      dtos,
		TotalCount: total,
		Limit:      filters.Limit,
		Offset:     filters.Offset,
	}, nil
}

// ---------- Internal helpers ----------

// executeProposal executes an approved proposal
func (s *VaultService) executeProposal(ctx context.Context, vault *wallet_entities.TeamVault, proposal *wallet_entities.VaultProposal) error {
	proposal.MarkExecuting()

	switch proposal.Type {
	case wallet_entities.ProposalTypeWithdrawal, wallet_entities.ProposalTypeTransfer:
		if proposal.Amount == nil || proposal.Currency == nil {
			return fmt.Errorf("amount and currency required for execution")
		}
		if err := vault.Withdraw(*proposal.Currency, *proposal.Amount); err != nil {
			return err
		}
		if err := s.vaultRepo.Update(ctx, vault); err != nil {
			return fmt.Errorf("failed to update vault after withdrawal: %w", err)
		}

	case wallet_entities.ProposalTypeItemTransfer:
		destUserID, err := uuid.Parse(proposal.Destination)
		if err != nil {
			return fmt.Errorf("invalid destination user ID: %w", err)
		}
		for _, itemID := range proposal.InventoryItemIDs {
			if err := s.inventoryRepo.TransferOwnership(ctx, itemID, wallet_entities.InventoryOwnerUser, destUserID); err != nil {
				return fmt.Errorf("failed to transfer item %s: %w", itemID, err)
			}
		}

	case wallet_entities.ProposalTypeSettingsChange:
		// Already handled in UpdateVaultSettings
	}

	txID := proposal.ID
	proposal.MarkExecuted(&txID, "")

	vault.RemovePendingProposal(proposal.ID)
	_ = s.vaultRepo.Update(ctx, vault)

	resourceOwner := shared.GetResourceOwner(ctx)
	s.logActivity(ctx, vault.ID, vault.SquadID, proposal.ProposerID, "",
		wallet_entities.ActivityProposalExecuted,
		fmt.Sprintf("Executed proposal: %s", proposal.Title),
		&proposal.ID, nil, resourceOwner)

	return nil
}

// getSquadMemberRole resolves the squad and returns the user's membership role
func (s *VaultService) getSquadMemberRole(ctx context.Context, squadID, userID uuid.UUID) (squad_vo.SquadMembershipType, error) {
	squad, err := s.squadReader.GetByID(ctx, squadID)
	if err != nil {
		return "", fmt.Errorf("squad not found: %w", err)
	}

	for _, m := range squad.Membership {
		if m.UserID == userID {
			// Check latest status is active
			for _, status := range m.Status {
				if status == squad_vo.SquadMembershipStatusActive {
					return m.Type, nil
				}
			}
		}
	}

	return "", fmt.Errorf("user %s is not an active member of squad %s", userID, squadID)
}

// logActivity appends a vault activity log entry (fire-and-forget)
func (s *VaultService) logActivity(
	ctx context.Context,
	vaultID, squadID, actorID uuid.UUID,
	actorName string,
	activityType wallet_entities.VaultActivityType,
	description string,
	relatedEntityID *uuid.UUID,
	details map[string]interface{},
	resourceOwner shared.ResourceOwner,
) {
	activity := wallet_entities.NewVaultActivity(
		vaultID, squadID, actorID, actorName,
		activityType, description, relatedEntityID, details,
		resourceOwner,
	)
	if err := s.activityRepo.Append(ctx, activity); err != nil {
		slog.WarnContext(ctx, "Failed to log vault activity", "error", err, "type", activityType)
	}
}

// proposalToDTO converts a proposal entity to a DTO
func (s *VaultService) proposalToDTO(p wallet_entities.VaultProposal) wallet_in.VaultProposalDTO {
	dto := wallet_in.VaultProposalDTO{
		ID:                p.ID,
		VaultID:           p.VaultID,
		ProposerID:        p.ProposerID,
		Type:              p.Type,
		Title:             p.Title,
		Description:       p.Description,
		Destination:       p.Destination,
		InventoryItemIDs:  p.InventoryItemIDs,
		RequiredApprovals: p.RequiredApprovals,
		CurrentApprovals:  len(p.Approvals),
		Status:            p.Status,
		OnChain:           p.OnChain,
		TxHash:            p.TxHash,
		ExpiresAt:         p.ExpiresAt,
		CreatedAt:         p.CreatedAt,
		ExecutedAt:        p.ExecutedAt,
	}

	if p.Amount != nil {
		amtStr := p.Amount.String()
		dto.Amount = &amtStr
	}
	dto.Currency = p.Currency

	for _, a := range p.Approvals {
		dto.Approvals = append(dto.Approvals, wallet_in.VaultApprovalDTO{
			UserID:    a.UserID,
			Role:      string(a.Role),
			Decision:  string(a.Decision),
			Reason:    a.Reason,
			Timestamp: a.Timestamp,
		})
	}
	for _, r := range p.Rejections {
		dto.Rejections = append(dto.Rejections, wallet_in.VaultApprovalDTO{
			UserID:    r.UserID,
			Role:      string(r.Role),
			Decision:  string(r.Decision),
			Reason:    r.Reason,
			Timestamp: r.Timestamp,
		})
	}

	return dto
}

// inventoryItemToDTO converts an inventory item entity to a DTO
func (s *VaultService) inventoryItemToDTO(item *wallet_entities.InventoryItem) wallet_in.InventoryItemDTO {
	if item == nil {
		return wallet_in.InventoryItemDTO{}
	}

	dto := wallet_in.InventoryItemDTO{
		ID:           item.ID,
		ItemType:     item.ItemType,
		Name:         item.Name,
		Description:  item.Description,
		ImageURI:     item.ImageURI,
		Rarity:       item.Rarity,
		GameID:       item.GameID,
		Quantity:     item.Quantity,
		Tradeable:    item.Tradeable,
		Transferable: item.Transferable,
		AcquiredAt:   item.AcquiredAt,
		ExpiresAt:    item.ExpiresAt,
		Status:       item.Status,
	}

	if item.EstimatedValue != nil {
		v := item.EstimatedValue.String()
		dto.EstimatedValue = &v
	}

	if item.NFTData != nil {
		dto.NFTData = &wallet_in.NFTDataDTO{
			ChainID:         int(item.NFTData.ChainID),
			ContractAddress: item.NFTData.ContractAddress,
			TokenID:         item.NFTData.TokenID,
			Standard:        string(item.NFTData.Standard),
			MetadataURI:     item.NFTData.MetadataURI,
		}
	}

	return dto
}
