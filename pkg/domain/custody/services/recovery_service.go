package custody_services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	custody_entities "github.com/replay-api/replay-api/pkg/domain/custody/entities"
	custody_in "github.com/replay-api/replay-api/pkg/domain/custody/ports/in"
	custody_out "github.com/replay-api/replay-api/pkg/domain/custody/ports/out"
	custody_vo "github.com/replay-api/replay-api/pkg/domain/custody/value-objects"
)

// RecoveryServiceImpl implements the RecoveryService interface
type RecoveryServiceImpl struct {
	walletRepo   custody_out.SmartWalletRepository
	mpcProvider  custody_out.MPCProvider
	solanaClient custody_out.SolanaClient
	evmClients   map[custody_vo.ChainID]custody_out.EVMClient
	config       *RecoveryConfig
}

// RecoveryConfig contains recovery service configuration
type RecoveryConfig struct {
	MinGuardians       uint8
	MaxGuardians       uint8
	DefaultThreshold   uint8
	MinRecoveryDelay   time.Duration
	MaxRecoveryDelay   time.Duration
	ApprovalExpiry     time.Duration
}

// NewRecoveryService creates a new recovery service
func NewRecoveryService(
	walletRepo custody_out.SmartWalletRepository,
	mpcProvider custody_out.MPCProvider,
	solanaClient custody_out.SolanaClient,
	evmClients map[custody_vo.ChainID]custody_out.EVMClient,
	config *RecoveryConfig,
) *RecoveryServiceImpl {
	if config == nil {
		config = &RecoveryConfig{
			MinGuardians:     2,
			MaxGuardians:     7,
			DefaultThreshold: 2,
			MinRecoveryDelay: 24 * time.Hour,
			MaxRecoveryDelay: 30 * 24 * time.Hour,
			ApprovalExpiry:   7 * 24 * time.Hour,
		}
	}
	return &RecoveryServiceImpl{
		walletRepo:   walletRepo,
		mpcProvider:  mpcProvider,
		solanaClient: solanaClient,
		evmClients:   evmClients,
		config:       config,
	}
}

// AddGuardian adds a guardian to a wallet
func (s *RecoveryServiceImpl) AddGuardian(
	ctx context.Context,
	req *custody_in.AddGuardianRequest,
) (*custody_in.GuardianResult, error) {
	wallet, err := s.walletRepo.GetByID(ctx, req.WalletID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	// Check guardian limits
	guardians, err := s.walletRepo.GetGuardians(ctx, wallet.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get guardians: %w", err)
	}

	if len(guardians) >= int(s.config.MaxGuardians) {
		return nil, fmt.Errorf("maximum guardians (%d) reached", s.config.MaxGuardians)
	}

	// Validate guardian based on type
	var guardianAddress string
	switch req.GuardianType {
	case custody_entities.GuardianTypeWallet:
		if req.Address == "" {
			return nil, fmt.Errorf("wallet address required for wallet guardian")
		}
		guardianAddress = req.Address
	case custody_entities.GuardianTypeEmail:
		if req.Email == "" {
			return nil, fmt.Errorf("email required for email guardian")
		}
		// Generate deterministic address from email hash
		emailHash := sha256.Sum256([]byte(req.Email))
		guardianAddress = "0x" + hex.EncodeToString(emailHash[:20])
	case custody_entities.GuardianTypePhone:
		if req.Phone == "" {
			return nil, fmt.Errorf("phone required for phone guardian")
		}
		// Generate deterministic address from phone hash
		phoneHash := sha256.Sum256([]byte(req.Phone))
		guardianAddress = "0x" + hex.EncodeToString(phoneHash[:20])
	case custody_entities.GuardianTypeHardware:
		if req.Address == "" {
			return nil, fmt.Errorf("hardware key address required")
		}
		guardianAddress = req.Address
	case custody_entities.GuardianTypeInstitution:
		if req.Address == "" {
			return nil, fmt.Errorf("institution address required")
		}
		guardianAddress = req.Address
	default:
		return nil, fmt.Errorf("invalid guardian type")
	}

	// Check for duplicate
	existing, _ := s.walletRepo.GetGuardianByAddress(ctx, wallet.ID, guardianAddress)
	if existing != nil {
		return nil, fmt.Errorf("guardian already exists")
	}

	// Create guardian entity
	metadata := make(map[string]interface{})
	for k, v := range req.Metadata {
		metadata[k] = v
	}

	guardian := &custody_entities.Guardian{
		ID:           uuid.New(),
		WalletID:     wallet.ID,
		GuardianType: req.GuardianType,
		Address:      guardianAddress,
		Email:        req.Email,
		Phone:        req.Phone,
		Label:        req.Label,
		Weight:       req.Weight,
		IsActive:     true,
		AddedAt:      time.Now(),
		Metadata:     metadata,
	}

	if guardian.Weight == 0 {
		guardian.Weight = 1 // Default weight
	}

	// Add guardian to database
	if err := s.walletRepo.AddGuardian(ctx, wallet.ID, guardian); err != nil {
		return nil, fmt.Errorf("failed to add guardian: %w", err)
	}

	// Register guardian on-chain if wallet is deployed
	var txHash *string
	for chainID, deployed := range wallet.AAConfig.IsDeployed {
		if deployed {
			hash, err := s.registerGuardianOnChain(ctx, wallet, guardian, chainID)
			if err != nil {
				// Log warning but don't fail
				continue
			}
			txHash = &hash
			break // Only need one chain for now
		}
	}

	return &custody_in.GuardianResult{
		GuardianID:   guardian.ID,
		WalletID:     wallet.ID,
		GuardianType: guardian.GuardianType,
		Address:      guardian.Address,
		Label:        guardian.Label,
		Weight:       guardian.Weight,
		TxHash:       txHash,
		CreatedAt:    guardian.AddedAt,
	}, nil
}

// RemoveGuardian removes a guardian from a wallet
func (s *RecoveryServiceImpl) RemoveGuardian(
	ctx context.Context,
	walletID uuid.UUID,
	guardianID uuid.UUID,
) error {
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return fmt.Errorf("wallet not found: %w", err)
	}

	guardians, err := s.walletRepo.GetGuardians(ctx, walletID)
	if err != nil {
		return fmt.Errorf("failed to get guardians: %w", err)
	}

	// Ensure we don't go below minimum
	if len(guardians) <= int(s.config.MinGuardians) {
		return fmt.Errorf("cannot remove guardian: minimum guardians (%d) required", s.config.MinGuardians)
	}

	// Ensure threshold is still achievable
	if len(guardians)-1 < int(wallet.RecoveryConfig.GuardianThreshold) {
		return fmt.Errorf("cannot remove guardian: would make threshold unreachable")
	}

	// Remove from database
	if err := s.walletRepo.RemoveGuardian(ctx, walletID, guardianID); err != nil {
		return fmt.Errorf("failed to remove guardian: %w", err)
	}

	// Remove from on-chain (best effort, log errors but don't fail)
	for chainID, deployed := range wallet.AAConfig.IsDeployed {
		if deployed {
			if err := s.removeGuardianOnChain(ctx, wallet, guardianID, chainID); err != nil {
				slog.WarnContext(ctx, "Failed to remove guardian on-chain (best effort)", "chainID", chainID, "error", err)
			}
		}
	}

	return nil
}

// GetGuardians returns all guardians for a wallet
func (s *RecoveryServiceImpl) GetGuardians(
	ctx context.Context,
	walletID uuid.UUID,
) ([]*custody_entities.Guardian, error) {
	return s.walletRepo.GetGuardians(ctx, walletID)
}

// SetGuardianThreshold sets the guardian threshold for recovery
func (s *RecoveryServiceImpl) SetGuardianThreshold(
	ctx context.Context,
	walletID uuid.UUID,
	threshold uint8,
) error {
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return fmt.Errorf("wallet not found: %w", err)
	}

	guardians, err := s.walletRepo.GetGuardians(ctx, walletID)
	if err != nil {
		return fmt.Errorf("failed to get guardians: %w", err)
	}

	if threshold == 0 {
		return fmt.Errorf("threshold must be at least 1")
	}

	if int(threshold) > len(guardians) {
		return fmt.Errorf("threshold cannot exceed guardian count")
	}

	wallet.RecoveryConfig.GuardianThreshold = threshold
	wallet.UpdatedAt = time.Now()

	return s.walletRepo.Update(ctx, wallet)
}

// InitiateRecovery starts the recovery process
func (s *RecoveryServiceImpl) InitiateRecovery(
	ctx context.Context,
	req *custody_in.InitiateRecoveryRequest,
) (*custody_in.RecoveryInitResult, error) {
	wallet, err := s.walletRepo.GetByID(ctx, req.WalletID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	// Verify initiator is a guardian
	guardians, err := s.walletRepo.GetGuardians(ctx, wallet.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get guardians: %w", err)
	}

	var initiatorGuardian *custody_entities.Guardian
	for _, g := range guardians {
		if g.ID == req.InitiatorID && g.IsActive {
			initiatorGuardian = g
			break
		}
	}

	if initiatorGuardian == nil {
		return nil, fmt.Errorf("initiator is not an active guardian")
	}

	// Check for existing pending recovery
	if wallet.PendingRecovery != nil && !wallet.PendingRecovery.Executed {
		return nil, fmt.Errorf("recovery already pending")
	}

	// Calculate execution time
	executableAt := time.Now().Add(wallet.RecoveryConfig.RecoveryDelay)

	// Create pending recovery
	recoveryID := uuid.New()
	pendingRecovery := &custody_entities.PendingRecovery{
		ID:               recoveryID,
		NewOwnerKey:      req.NewOwnerKey,
		NewEVMAddress:    req.NewEVMAddress,
		NewSolanaAddress: req.NewSolanaAddress,
		InitiatedAt:      time.Now(),
		ExecutableAt:     executableAt,
		InitiatedBy:      req.InitiatorID,
		ApprovalCount:    1, // Initiator counts as first approval
		Approvers:        []uuid.UUID{req.InitiatorID},
		Executed:         false,
		Reason:           req.Reason,
	}

	// Freeze wallet during recovery
	wallet.IsFrozen = true
	wallet.PendingRecovery = pendingRecovery
	wallet.UpdatedAt = time.Now()

	if err := s.walletRepo.Update(ctx, wallet); err != nil {
		return nil, fmt.Errorf("failed to initiate recovery: %w", err)
	}

	// Initiate on-chain recovery (best effort)
	var txHash *string
	for chainID, deployed := range wallet.AAConfig.IsDeployed {
		if deployed {
			hash, err := s.initiateRecoveryOnChain(ctx, wallet, pendingRecovery, chainID)
			if err == nil {
				txHash = &hash
				break
			}
		}
	}

	return &custody_in.RecoveryInitResult{
		RecoveryID:        recoveryID,
		WalletID:          wallet.ID,
		NewOwnerKey:       req.NewOwnerKey,
		ExecutableAt:      executableAt,
		RequiredApprovals: wallet.RecoveryConfig.GuardianThreshold,
		CurrentApprovals:  1,
		TxHash:            txHash,
		InitiatedAt:       pendingRecovery.InitiatedAt,
	}, nil
}

// ApproveRecovery approves a pending recovery
func (s *RecoveryServiceImpl) ApproveRecovery(
	ctx context.Context,
	req *custody_in.ApproveRecoveryRequest,
) (*custody_in.RecoveryApprovalResult, error) {
	wallet, err := s.walletRepo.GetByID(ctx, req.WalletID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	if wallet.PendingRecovery == nil {
		return nil, fmt.Errorf("no recovery pending")
	}

	if wallet.PendingRecovery.Executed {
		return nil, fmt.Errorf("recovery already executed")
	}

	// Verify guardian
	guardians, err := s.walletRepo.GetGuardians(ctx, wallet.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get guardians: %w", err)
	}

	var guardian *custody_entities.Guardian
	for _, g := range guardians {
		if g.ID == req.GuardianID && g.IsActive {
			guardian = g
			break
		}
	}

	if guardian == nil {
		return nil, fmt.Errorf("not an active guardian")
	}

	// Check if already approved
	for _, approver := range wallet.PendingRecovery.Approvers {
		if approver == req.GuardianID {
			return nil, fmt.Errorf("already approved")
		}
	}

	// Add approval
	wallet.PendingRecovery.Approvers = append(wallet.PendingRecovery.Approvers, req.GuardianID)
	wallet.PendingRecovery.ApprovalCount++
	wallet.UpdatedAt = time.Now()

	if err := s.walletRepo.Update(ctx, wallet); err != nil {
		return nil, fmt.Errorf("failed to record approval: %w", err)
	}

	// Approve on-chain (best effort)
	var txHash *string
	for chainID, deployed := range wallet.AAConfig.IsDeployed {
		if deployed {
			hash, err := s.approveRecoveryOnChain(ctx, wallet, guardian, chainID)
			if err == nil {
				txHash = &hash
				break
			}
		}
	}

	isReady := wallet.PendingRecovery.ApprovalCount >= wallet.RecoveryConfig.GuardianThreshold &&
		time.Now().After(wallet.PendingRecovery.ExecutableAt)

	return &custody_in.RecoveryApprovalResult{
		WalletID:          wallet.ID,
		GuardianID:        guardian.ID,
		ApprovalCount:     wallet.PendingRecovery.ApprovalCount,
		RequiredApprovals: wallet.RecoveryConfig.GuardianThreshold,
		IsReady:           isReady,
		TxHash:            txHash,
		ApprovedAt:        time.Now(),
	}, nil
}

// ExecuteRecovery executes a pending recovery
func (s *RecoveryServiceImpl) ExecuteRecovery(
	ctx context.Context,
	walletID uuid.UUID,
) (*custody_in.RecoveryExecutionResult, error) {
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	if wallet.PendingRecovery == nil {
		return nil, fmt.Errorf("no recovery pending")
	}

	if wallet.PendingRecovery.Executed {
		return nil, fmt.Errorf("recovery already executed")
	}

	// Check threshold
	if wallet.PendingRecovery.ApprovalCount < wallet.RecoveryConfig.GuardianThreshold {
		return nil, fmt.Errorf("insufficient approvals: %d/%d",
			wallet.PendingRecovery.ApprovalCount,
			wallet.RecoveryConfig.GuardianThreshold)
	}

	// Check delay
	if time.Now().Before(wallet.PendingRecovery.ExecutableAt) {
		return nil, fmt.Errorf("recovery delay not met: %v remaining",
			time.Until(wallet.PendingRecovery.ExecutableAt))
	}

	// Store old values
	oldOwnerKey := wallet.PublicKey
	oldEVMAddress := wallet.Addresses[wallet.PrimaryChain]
	oldSolanaAddress := wallet.Addresses[custody_vo.ChainSolanaMainnet]

	// Execute on-chain recovery for all deployed chains
	txHashes := make(map[custody_vo.ChainID]string)
	for chainID, deployed := range wallet.AAConfig.IsDeployed {
		if deployed {
			hash, err := s.executeRecoveryOnChain(ctx, wallet, chainID)
			if err != nil {
				return nil, fmt.Errorf("failed to execute recovery on %s: %w", chainID, err)
			}
			txHashes[chainID] = hash
		}
	}

	// Update wallet with new owner
	wallet.PublicKey = hex.EncodeToString(wallet.PendingRecovery.NewOwnerKey)
	if wallet.PendingRecovery.NewEVMAddress != "" {
		for chainID := range wallet.Addresses {
			if chainID.IsEVM() {
				wallet.Addresses[chainID] = wallet.PendingRecovery.NewEVMAddress
			}
		}
	}
	if wallet.PendingRecovery.NewSolanaAddress != "" {
		for chainID := range wallet.Addresses {
			if chainID.IsSolana() {
				wallet.Addresses[chainID] = wallet.PendingRecovery.NewSolanaAddress
			}
		}
	}

	// Mark recovery as executed
	wallet.PendingRecovery.Executed = true
	wallet.IsFrozen = false
	wallet.UpdatedAt = time.Now()

	if err := s.walletRepo.Update(ctx, wallet); err != nil {
		return nil, fmt.Errorf("failed to update wallet: %w", err)
	}

	return &custody_in.RecoveryExecutionResult{
		WalletID:         wallet.ID,
		OldOwnerKey:      []byte(oldOwnerKey),
		NewOwnerKey:      wallet.PendingRecovery.NewOwnerKey,
		OldEVMAddress:    oldEVMAddress,
		NewEVMAddress:    wallet.PendingRecovery.NewEVMAddress,
		OldSolanaAddress: oldSolanaAddress,
		NewSolanaAddress: wallet.PendingRecovery.NewSolanaAddress,
		TxHashes:         txHashes,
		ExecutedAt:       time.Now(),
	}, nil
}

// CancelRecovery cancels a pending recovery (owner only)
func (s *RecoveryServiceImpl) CancelRecovery(
	ctx context.Context,
	walletID uuid.UUID,
) error {
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return fmt.Errorf("wallet not found: %w", err)
	}

	if wallet.PendingRecovery == nil {
		return fmt.Errorf("no recovery pending")
	}

	// Cancel on-chain (best effort, log errors but don't fail)
	for chainID, deployed := range wallet.AAConfig.IsDeployed {
		if deployed {
			if err := s.cancelRecoveryOnChain(ctx, wallet, chainID); err != nil {
				slog.WarnContext(ctx, "Failed to cancel recovery on-chain (best effort)", "chainID", chainID, "error", err)
			}
		}
	}

	// Clear pending recovery
	wallet.PendingRecovery = nil
	wallet.IsFrozen = false
	wallet.UpdatedAt = time.Now()

	return s.walletRepo.Update(ctx, wallet)
}

// GetRecoveryStatus returns the current recovery status
func (s *RecoveryServiceImpl) GetRecoveryStatus(
	ctx context.Context,
	walletID uuid.UUID,
) (*custody_in.RecoveryStatus, error) {
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	guardians, err := s.walletRepo.GetGuardians(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get guardians: %w", err)
	}

	status := &custody_in.RecoveryStatus{
		WalletID:          walletID,
		HasPendingRecovery: wallet.PendingRecovery != nil && !wallet.PendingRecovery.Executed,
		RequiredApprovals: wallet.RecoveryConfig.GuardianThreshold,
	}

	if wallet.PendingRecovery != nil && !wallet.PendingRecovery.Executed {
		status.NewOwnerKey = wallet.PendingRecovery.NewOwnerKey
		status.NewEVMAddress = wallet.PendingRecovery.NewEVMAddress
		status.NewSolanaAddress = wallet.PendingRecovery.NewSolanaAddress
		status.InitiatedAt = &wallet.PendingRecovery.InitiatedAt
		status.ExecutableAt = &wallet.PendingRecovery.ExecutableAt
		status.CurrentApprovals = wallet.PendingRecovery.ApprovalCount

		remaining := time.Until(wallet.PendingRecovery.ExecutableAt)
		if remaining > 0 {
			status.TimeRemaining = &remaining
		}

		status.CanExecute = status.CurrentApprovals >= status.RequiredApprovals &&
			time.Now().After(wallet.PendingRecovery.ExecutableAt)
		status.CanCancel = true

		// Build approvers list
		approverMap := make(map[uuid.UUID]bool)
		for _, id := range wallet.PendingRecovery.Approvers {
			approverMap[id] = true
		}

		for _, g := range guardians {
			info := custody_in.ApproverInfo{
				GuardianID:   g.ID,
				GuardianType: g.GuardianType,
				Label:        g.Label,
				Approved:     approverMap[g.ID],
			}
			status.Approvers = append(status.Approvers, info)
		}
	}

	return status, nil
}

// GetPendingRecoveries returns all pending recoveries for a guardian
func (s *RecoveryServiceImpl) GetPendingRecoveries(
	ctx context.Context,
	guardianAddress string,
) ([]*custody_in.PendingRecoveryInfo, error) {
	wallets, err := s.walletRepo.GetPendingRecoveries(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending recoveries: %w", err)
	}

	var result []*custody_in.PendingRecoveryInfo

	for _, wallet := range wallets {
		if wallet.PendingRecovery == nil || wallet.PendingRecovery.Executed {
			continue
		}

		guardians, _ := s.walletRepo.GetGuardians(ctx, wallet.ID)

		var myStatus string = "not_guardian"
		for _, g := range guardians {
			if g.Address == guardianAddress && g.IsActive {
				myStatus = "pending"
				for _, approver := range wallet.PendingRecovery.Approvers {
					if approver == g.ID {
						myStatus = "approved"
						break
					}
				}
				break
			}
		}

		if myStatus != "not_guardian" {
			info := &custody_in.PendingRecoveryInfo{
				WalletID:          wallet.ID,
				WalletLabel:       wallet.Label,
				UserID:            wallet.UserID,
				NewOwnerKey:       wallet.PendingRecovery.NewOwnerKey,
				InitiatedAt:       wallet.PendingRecovery.InitiatedAt,
				ExecutableAt:      wallet.PendingRecovery.ExecutableAt,
				RequiredApprovals: wallet.RecoveryConfig.GuardianThreshold,
				CurrentApprovals:  wallet.PendingRecovery.ApprovalCount,
				MyApprovalStatus:  myStatus,
				Reason:            wallet.PendingRecovery.Reason,
			}
			result = append(result, info)
		}
	}

	return result, nil
}

// SetRecoveryDelay sets the recovery delay for a wallet
func (s *RecoveryServiceImpl) SetRecoveryDelay(
	ctx context.Context,
	walletID uuid.UUID,
	delay time.Duration,
) error {
	if delay < s.config.MinRecoveryDelay {
		return fmt.Errorf("delay must be at least %v", s.config.MinRecoveryDelay)
	}
	if delay > s.config.MaxRecoveryDelay {
		return fmt.Errorf("delay cannot exceed %v", s.config.MaxRecoveryDelay)
	}

	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return fmt.Errorf("wallet not found: %w", err)
	}

	wallet.RecoveryConfig.RecoveryDelay = delay
	wallet.UpdatedAt = time.Now()

	return s.walletRepo.Update(ctx, wallet)
}

// GetRecoveryConfig returns the recovery configuration
func (s *RecoveryServiceImpl) GetRecoveryConfig(
	ctx context.Context,
	walletID uuid.UUID,
) (*custody_in.RecoveryConfig, error) {
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	guardians, err := s.walletRepo.GetGuardians(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get guardians: %w", err)
	}

	config := &custody_in.RecoveryConfig{
		WalletID:          walletID,
		RecoveryDelay:     wallet.RecoveryConfig.RecoveryDelay,
		GuardianThreshold: wallet.RecoveryConfig.GuardianThreshold,
		TotalGuardians:    uint8(len(guardians)),
	}

	for _, g := range guardians {
		config.Guardians = append(config.Guardians, custody_in.GuardianInfo{
			GuardianID:   g.ID,
			GuardianType: g.GuardianType,
			Address:      g.Address,
			Label:        g.Label,
			Weight:       g.Weight,
			IsActive:     g.IsActive,
			AddedAt:      g.AddedAt,
		})
	}

	return config, nil
}

// On-chain interaction helpers

func (s *RecoveryServiceImpl) registerGuardianOnChain(
	ctx context.Context,
	wallet *custody_entities.SmartWallet,
	guardian *custody_entities.Guardian,
	chainID custody_vo.ChainID,
) (string, error) {
	// Build and submit addGuardian transaction
	// This is chain-specific implementation
	if chainID.IsSolana() {
		return s.registerGuardianSolana(ctx, wallet, guardian)
	}
	return s.registerGuardianEVM(ctx, wallet, guardian, chainID)
}

func (s *RecoveryServiceImpl) registerGuardianSolana(
	ctx context.Context,
	wallet *custody_entities.SmartWallet,
	guardian *custody_entities.Guardian,
) (string, error) {
	// Build add_guardian instruction for Solana program
	// Implementation would use s.solanaClient
	return "", nil
}

func (s *RecoveryServiceImpl) registerGuardianEVM(
	ctx context.Context,
	wallet *custody_entities.SmartWallet,
	guardian *custody_entities.Guardian,
	chainID custody_vo.ChainID,
) (string, error) {
	// Build addGuardian UserOperation for ERC-4337 wallet
	// Implementation would use s.evmClients[chainID]
	return "", nil
}

func (s *RecoveryServiceImpl) removeGuardianOnChain(
	ctx context.Context,
	wallet *custody_entities.SmartWallet,
	guardianID uuid.UUID,
	chainID custody_vo.ChainID,
) error {
	// Build and submit removeGuardian transaction
	return nil
}

func (s *RecoveryServiceImpl) initiateRecoveryOnChain(
	ctx context.Context,
	wallet *custody_entities.SmartWallet,
	recovery *custody_entities.PendingRecovery,
	chainID custody_vo.ChainID,
) (string, error) {
	// Build and submit initiateRecovery transaction
	return "", nil
}

func (s *RecoveryServiceImpl) approveRecoveryOnChain(
	ctx context.Context,
	wallet *custody_entities.SmartWallet,
	guardian *custody_entities.Guardian,
	chainID custody_vo.ChainID,
) (string, error) {
	// Build and submit approveRecovery transaction
	return "", nil
}

func (s *RecoveryServiceImpl) executeRecoveryOnChain(
	ctx context.Context,
	wallet *custody_entities.SmartWallet,
	chainID custody_vo.ChainID,
) (string, error) {
	// Build and submit executeRecovery transaction
	return "", nil
}

func (s *RecoveryServiceImpl) cancelRecoveryOnChain(
	ctx context.Context,
	wallet *custody_entities.SmartWallet,
	chainID custody_vo.ChainID,
) error {
	// Build and submit cancelRecovery transaction
	return nil
}
