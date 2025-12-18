package tournament_services

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
)

// PrizeDistributionService handles prize pool management and distribution
type PrizeDistributionService struct {
	prizePoolRepo   PrizePoolRepository
	walletService   WalletService
	blockchainGateway BlockchainGateway
	auditTrail      billing_in.AuditTrailCommand
	notifier        NotificationService
}

// PrizePoolRepository defines persistence operations
type PrizePoolRepository interface {
	Create(ctx context.Context, pool *tournament_entities.PrizePool) error
	Update(ctx context.Context, pool *tournament_entities.PrizePool) error
	GetByID(ctx context.Context, id uuid.UUID) (*tournament_entities.PrizePool, error)
	GetByTournamentID(ctx context.Context, tournamentID uuid.UUID) (*tournament_entities.PrizePool, error)
}

// WalletService handles wallet operations
type WalletService interface {
	GetBalance(ctx context.Context, userID uuid.UUID, currency string) (*big.Float, error)
	Hold(ctx context.Context, userID uuid.UUID, amount *big.Float, currency string, reference uuid.UUID, description string) error
	Release(ctx context.Context, userID uuid.UUID, amount *big.Float, currency string, reference uuid.UUID) error
	Transfer(ctx context.Context, fromUserID, toUserID uuid.UUID, amount *big.Float, currency string, reference uuid.UUID, description string) error
}

// BlockchainGateway handles blockchain interactions
type BlockchainGateway interface {
	GetSupportedChains() []string
	VerifyTransaction(ctx context.Context, chainID, txHash string) (*BlockchainTxInfo, error)
	RecordMerkleRoot(ctx context.Context, chainID, contractAddress, merkleRoot string) (string, error)
	ExecuteDistribution(ctx context.Context, chainID, contractAddress string, payouts []BlockchainPayout) (string, error)
	GetTransactionConfirmations(ctx context.Context, chainID, txHash string) (int, error)
}

// BlockchainTxInfo contains transaction verification data
type BlockchainTxInfo struct {
	TxHash        string
	BlockNumber   uint64
	Confirmations int
	Status        string
	Timestamp     time.Time
}

// BlockchainPayout represents a payout for blockchain execution
type BlockchainPayout struct {
	RecipientAddress string
	Amount           *big.Float
	Reference        string
}

// NotificationService sends notifications
type NotificationService interface {
	SendPrizeNotification(ctx context.Context, userID uuid.UUID, amount *big.Float, currency string, position int, tournamentName string) error
}

// NewPrizeDistributionService creates a new prize distribution service
func NewPrizeDistributionService(
	prizePoolRepo PrizePoolRepository,
	walletService WalletService,
	blockchainGateway BlockchainGateway,
	auditTrail billing_in.AuditTrailCommand,
	notifier NotificationService,
) *PrizeDistributionService {
	return &PrizeDistributionService{
		prizePoolRepo:     prizePoolRepo,
		walletService:     walletService,
		blockchainGateway: blockchainGateway,
		auditTrail:        auditTrail,
		notifier:          notifier,
	}
}

// CreatePrizePool creates a new prize pool for a tournament
func (s *PrizeDistributionService) CreatePrizePool(ctx context.Context, req CreatePrizePoolRequest) (*tournament_entities.PrizePool, error) {
	resourceOwner := common.GetResourceOwner(ctx)

	pool := tournament_entities.NewPrizePool(
		req.TournamentID,
		req.Type,
		req.Currency,
		req.PlatformFeePercent,
		resourceOwner,
	)

	if err := pool.SetDistribution(req.Distribution); err != nil {
		return nil, fmt.Errorf("invalid distribution: %w", err)
	}

	if req.FundingDeadline != nil {
		pool.FundingDeadline = req.FundingDeadline
	}

	if err := s.prizePoolRepo.Create(ctx, pool); err != nil {
		return nil, fmt.Errorf("failed to create prize pool: %w", err)
	}

	slog.InfoContext(ctx, "Prize pool created",
		"pool_id", pool.ID,
		"tournament_id", req.TournamentID,
		"type", req.Type,
	)

	return pool, nil
}

// AddContribution adds a contribution to the prize pool
func (s *PrizeDistributionService) AddContribution(ctx context.Context, poolID uuid.UUID, contribution tournament_entities.PrizeContribution) error {
	resourceOwner := common.GetResourceOwner(ctx)

	pool, err := s.prizePoolRepo.GetByID(ctx, poolID)
	if err != nil {
		return fmt.Errorf("prize pool not found: %w", err)
	}

	if pool.Status != tournament_entities.PrizePoolStatusPending {
		return fmt.Errorf("cannot add contribution to %s prize pool", pool.Status)
	}

	// Hold funds if contribution is from a user
	if contribution.SourceUserID != nil {
		err = s.walletService.Hold(ctx, *contribution.SourceUserID, contribution.Amount, contribution.Currency, contribution.TransactionID, "Prize pool contribution")
		if err != nil {
			return fmt.Errorf("failed to hold funds: %w", err)
		}
	}

	// Add contribution
	contribution.ID = uuid.New()
	contribution.ContributedAt = time.Now().UTC()
	contribution.Status = "confirmed"

	if err := pool.AddContribution(contribution); err != nil {
		// Release hold if contribution fails
		if contribution.SourceUserID != nil {
			_ = s.walletService.Release(ctx, *contribution.SourceUserID, contribution.Amount, contribution.Currency, contribution.TransactionID)
		}
		return fmt.Errorf("failed to add contribution: %w", err)
	}

	if err := s.prizePoolRepo.Update(ctx, pool); err != nil {
		return fmt.Errorf("failed to update prize pool: %w", err)
	}

	// Audit trail
	if s.auditTrail != nil {
		_ = s.auditTrail.RecordFinancialEvent(ctx, billing_in.RecordFinancialEventRequest{
			EventType:     billing_entities.AuditEventEntryFee,
			UserID:        resourceOwner.UserID,
			TargetType:    "prize_pool",
			TargetID:      poolID,
			Amount:        floatFromBig(contribution.Amount),
			Currency:      contribution.Currency,
			TransactionID: contribution.TransactionID,
			Description:   fmt.Sprintf("Prize pool contribution: %s", contribution.Type),
		})
	}

	slog.InfoContext(ctx, "Contribution added to prize pool",
		"pool_id", poolID,
		"contribution_id", contribution.ID,
		"amount", contribution.Amount.Text('f', 2),
		"type", contribution.Type,
	)

	return nil
}

// FundFromEntryFees processes entry fee contributions
func (s *PrizeDistributionService) FundFromEntryFees(ctx context.Context, poolID uuid.UUID, userID uuid.UUID, entryFee *big.Float, transactionID uuid.UUID) error {
	contribution := tournament_entities.PrizeContribution{
		Type:          "entry_fee",
		SourceUserID:  &userID,
		Amount:        entryFee,
		Currency:      "USD", // Default, should be from pool
		TransactionID: transactionID,
	}

	return s.AddContribution(ctx, poolID, contribution)
}

// FinalizePool marks the prize pool as funded and locks it
func (s *PrizeDistributionService) FinalizePool(ctx context.Context, poolID uuid.UUID) error {
	pool, err := s.prizePoolRepo.GetByID(ctx, poolID)
	if err != nil {
		return fmt.Errorf("prize pool not found: %w", err)
	}

	if pool.Status != tournament_entities.PrizePoolStatusPending {
		return fmt.Errorf("prize pool already finalized")
	}

	if pool.TotalAmount.Cmp(big.NewFloat(0)) <= 0 {
		return fmt.Errorf("prize pool has no contributions")
	}

	pool.MarkAsFunded()

	// Record on blockchain if gateway available
	if s.blockchainGateway != nil && pool.BlockchainVerification != nil {
		txHash, err := s.blockchainGateway.RecordMerkleRoot(
			ctx,
			pool.BlockchainVerification.ChainID,
			pool.BlockchainVerification.ContractAddress,
			pool.MerkleRoot,
		)
		if err != nil {
			slog.WarnContext(ctx, "Failed to record merkle root on blockchain", "error", err)
		} else {
			pool.BlockchainVerification.FundingTxHash = txHash
			pool.BlockchainVerification.Verified = true
			pool.BlockchainVerification.Timestamp = time.Now().UTC()
		}
	}

	if err := s.prizePoolRepo.Update(ctx, pool); err != nil {
		return fmt.Errorf("failed to finalize prize pool: %w", err)
	}

	slog.InfoContext(ctx, "Prize pool finalized",
		"pool_id", poolID,
		"total_amount", pool.TotalAmount.Text('f', 2),
		"merkle_root", pool.MerkleRoot,
	)

	return nil
}

// CalculateAndSetPayouts calculates payouts based on tournament results
func (s *PrizeDistributionService) CalculateAndSetPayouts(ctx context.Context, poolID uuid.UUID, results []tournament_entities.TournamentResult) error {
	pool, err := s.prizePoolRepo.GetByID(ctx, poolID)
	if err != nil {
		return fmt.Errorf("prize pool not found: %w", err)
	}

	if pool.Status != tournament_entities.PrizePoolStatusFunded {
		return fmt.Errorf("prize pool must be funded before calculating payouts")
	}

	payouts, err := pool.CalculatePayouts(results)
	if err != nil {
		return fmt.Errorf("failed to calculate payouts: %w", err)
	}

	pool.SetPayouts(payouts)

	if err := s.prizePoolRepo.Update(ctx, pool); err != nil {
		return fmt.Errorf("failed to save payouts: %w", err)
	}

	slog.InfoContext(ctx, "Payouts calculated",
		"pool_id", poolID,
		"payout_count", len(payouts),
	)

	return nil
}

// DistributePrizes executes all pending payouts
func (s *PrizeDistributionService) DistributePrizes(ctx context.Context, poolID uuid.UUID) error {
	resourceOwner := common.GetResourceOwner(ctx)

	pool, err := s.prizePoolRepo.GetByID(ctx, poolID)
	if err != nil {
		return fmt.Errorf("prize pool not found: %w", err)
	}

	if err := pool.StartDistribution(); err != nil {
		return err
	}

	// Process each payout
	for _, payout := range pool.Payouts {
		if payout.Status != tournament_entities.PayoutStatusPending {
			continue
		}

		// Process based on payment method
		switch payout.PaymentMethod {
		case tournament_entities.PaymentMethodPlatformWallet:
			err = s.processPlatformWalletPayout(ctx, pool, &payout)
		case tournament_entities.PaymentMethodCrypto:
			err = s.processCryptoPayout(ctx, pool, &payout)
		default:
			err = s.processPlatformWalletPayout(ctx, pool, &payout)
		}

		if err != nil {
			slog.ErrorContext(ctx, "Failed to process payout",
				"payout_id", payout.ID,
				"error", err,
			)
			payout.Status = tournament_entities.PayoutStatusFailed
			payout.FailureReason = err.Error()
		}

		// Audit trail
		if s.auditTrail != nil {
			_ = s.auditTrail.RecordFinancialEvent(ctx, billing_in.RecordFinancialEventRequest{
				EventType:     billing_entities.AuditEventPrizeDistribution,
				UserID:        resourceOwner.UserID,
				TargetType:    "payout",
				TargetID:      payout.ID,
				Amount:        floatFromBig(payout.NetAmount),
				Currency:      payout.Currency,
				TransactionID: payout.ID,
				Description:   fmt.Sprintf("Prize payout for position %d", payout.Position),
				Metadata: map[string]interface{}{
					"recipient_user_id": payout.RecipientUserID,
					"position":          payout.Position,
					"gross_amount":      payout.GrossAmount.Text('f', 2),
					"net_amount":        payout.NetAmount.Text('f', 2),
				},
			})
		}

		// Send notification
		if s.notifier != nil && payout.Status == tournament_entities.PayoutStatusCompleted {
			_ = s.notifier.SendPrizeNotification(ctx, payout.RecipientUserID, payout.NetAmount, payout.Currency, payout.Position, "")
		}
	}

	if err := s.prizePoolRepo.Update(ctx, pool); err != nil {
		return fmt.Errorf("failed to update prize pool: %w", err)
	}

	slog.InfoContext(ctx, "Prize distribution completed",
		"pool_id", poolID,
		"status", pool.Status,
	)

	return nil
}

// processPlatformWalletPayout handles internal wallet payouts
func (s *PrizeDistributionService) processPlatformWalletPayout(ctx context.Context, pool *tournament_entities.PrizePool, payout *tournament_entities.PrizePayout) error {
	// Transfer from prize pool escrow to winner
	err := s.walletService.Transfer(
		ctx,
		uuid.Nil, // System escrow
		payout.RecipientUserID,
		payout.NetAmount,
		payout.Currency,
		payout.ID,
		fmt.Sprintf("Tournament prize - Position %d", payout.Position),
	)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	payout.Status = tournament_entities.PayoutStatusCompleted
	payout.ProcessedAt = &now
	payout.CompletedAt = &now

	// Update pool
	pool.MarkPayoutCompleted(payout.ID, "", 0)

	return nil
}

// processCryptoPayout handles blockchain-based payouts
func (s *PrizeDistributionService) processCryptoPayout(ctx context.Context, pool *tournament_entities.PrizePool, payout *tournament_entities.PrizePayout) error {
	if s.blockchainGateway == nil {
		return fmt.Errorf("blockchain gateway not configured")
	}

	if pool.BlockchainVerification == nil {
		return fmt.Errorf("blockchain verification not configured for pool")
	}

	if payout.WalletAddress == "" {
		return fmt.Errorf("recipient wallet address not set")
	}

	// Execute blockchain transfer
	blockchainPayouts := []BlockchainPayout{
		{
			RecipientAddress: payout.WalletAddress,
			Amount:           payout.NetAmount,
			Reference:        payout.ID.String(),
		},
	}

	txHash, err := s.blockchainGateway.ExecuteDistribution(
		ctx,
		pool.BlockchainVerification.ChainID,
		pool.BlockchainVerification.ContractAddress,
		blockchainPayouts,
	)
	if err != nil {
		return fmt.Errorf("blockchain execution failed: %w", err)
	}

	now := time.Now().UTC()
	payout.Status = tournament_entities.PayoutStatusProcessing
	payout.TransactionHash = txHash
	payout.ProcessedAt = &now

	// Start confirmation monitoring (in production, this would be async)
	confirmations, err := s.blockchainGateway.GetTransactionConfirmations(ctx, pool.BlockchainVerification.ChainID, txHash)
	if err != nil {
		slog.WarnContext(ctx, "Failed to get confirmations", "error", err, "tx_hash", txHash)
	} else if confirmations >= 12 { // Minimum confirmations for finality
		payout.BlockConfirmations = confirmations
		payout.Status = tournament_entities.PayoutStatusCompleted
		payout.CompletedAt = &now
		pool.MarkPayoutCompleted(payout.ID, txHash, confirmations)
	}

	return nil
}

// GetPrizePoolStatus retrieves the current status of a prize pool
func (s *PrizeDistributionService) GetPrizePoolStatus(ctx context.Context, poolID uuid.UUID) (*PrizePoolStatusResponse, error) {
	pool, err := s.prizePoolRepo.GetByID(ctx, poolID)
	if err != nil {
		return nil, fmt.Errorf("prize pool not found: %w", err)
	}

	// Verify integrity
	integrityValid := pool.VerifyIntegrity()

	response := &PrizePoolStatusResponse{
		ID:              pool.ID,
		TournamentID:    pool.TournamentID,
		Status:          pool.Status,
		Type:            pool.Type,
		TotalAmount:     pool.TotalAmount.Text('f', 2),
		Currency:        pool.Currency,
		PlatformFee:     pool.PlatformFeeAmount.Text('f', 2),
		Distributable:   pool.GetDistributableAmount().Text('f', 2),
		ContributionCount: len(pool.Contributions),
		PayoutCount:     len(pool.Payouts),
		MerkleRoot:      pool.MerkleRoot,
		IntegrityValid:  integrityValid,
		Distribution:    pool.Distribution,
	}

	if pool.BlockchainVerification != nil {
		response.BlockchainVerified = pool.BlockchainVerification.Verified
		response.BlockchainChain = pool.BlockchainVerification.ChainID
		response.BlockchainTxHash = pool.BlockchainVerification.FundingTxHash
	}

	// Calculate payout statistics
	var completedPayouts, failedPayouts int
	var distributedAmount big.Float
	for _, payout := range pool.Payouts {
		if payout.Status == tournament_entities.PayoutStatusCompleted {
			completedPayouts++
			distributedAmount.Add(&distributedAmount, payout.NetAmount)
		} else if payout.Status == tournament_entities.PayoutStatusFailed {
			failedPayouts++
		}
	}

	response.CompletedPayouts = completedPayouts
	response.FailedPayouts = failedPayouts
	response.DistributedAmount = distributedAmount.Text('f', 2)

	return response, nil
}

// VerifyPayoutProof generates a merkle proof for a specific payout
func (s *PrizeDistributionService) VerifyPayoutProof(ctx context.Context, poolID, payoutID uuid.UUID) (*PayoutProofResponse, error) {
	pool, err := s.prizePoolRepo.GetByID(ctx, poolID)
	if err != nil {
		return nil, fmt.Errorf("prize pool not found: %w", err)
	}

	// Find payout
	var targetPayout *tournament_entities.PrizePayout
	for i := range pool.Payouts {
		if pool.Payouts[i].ID == payoutID {
			targetPayout = &pool.Payouts[i]
			break
		}
	}

	if targetPayout == nil {
		return nil, fmt.Errorf("payout not found")
	}

	// Get merkle proof
	proof, err := pool.GetVerificationProof(targetPayout.Hash)
	if err != nil {
		return nil, fmt.Errorf("failed to generate proof: %w", err)
	}

	return &PayoutProofResponse{
		PayoutID:         payoutID,
		RecipientUserID:  targetPayout.RecipientUserID,
		Amount:           targetPayout.NetAmount.Text('f', 8),
		Currency:         targetPayout.Currency,
		PayoutHash:       targetPayout.Hash,
		MerkleRoot:       pool.MerkleRoot,
		MerkleProof:      proof,
		TransactionHash:  targetPayout.TransactionHash,
		Confirmations:    targetPayout.BlockConfirmations,
		VerifiedOnChain:  pool.BlockchainVerification != nil && pool.BlockchainVerification.Verified,
	}, nil
}

// Request/Response types

// CreatePrizePoolRequest contains data for creating a prize pool
type CreatePrizePoolRequest struct {
	TournamentID       uuid.UUID
	Type               tournament_entities.PrizePoolType
	Currency           string
	PlatformFeePercent float64
	Distribution       []tournament_entities.PrizeDistribution
	FundingDeadline    *time.Time
}

// PrizePoolStatusResponse contains prize pool status
type PrizePoolStatusResponse struct {
	ID                 uuid.UUID                         `json:"id"`
	TournamentID       uuid.UUID                         `json:"tournament_id"`
	Status             tournament_entities.PrizePoolStatus `json:"status"`
	Type               tournament_entities.PrizePoolType  `json:"type"`
	TotalAmount        string                            `json:"total_amount"`
	Currency           string                            `json:"currency"`
	PlatformFee        string                            `json:"platform_fee"`
	Distributable      string                            `json:"distributable_amount"`
	ContributionCount  int                               `json:"contribution_count"`
	PayoutCount        int                               `json:"payout_count"`
	CompletedPayouts   int                               `json:"completed_payouts"`
	FailedPayouts      int                               `json:"failed_payouts"`
	DistributedAmount  string                            `json:"distributed_amount"`
	MerkleRoot         string                            `json:"merkle_root"`
	IntegrityValid     bool                              `json:"integrity_valid"`
	BlockchainVerified bool                              `json:"blockchain_verified"`
	BlockchainChain    string                            `json:"blockchain_chain,omitempty"`
	BlockchainTxHash   string                            `json:"blockchain_tx_hash,omitempty"`
	Distribution       []tournament_entities.PrizeDistribution `json:"distribution"`
}

// PayoutProofResponse contains merkle proof for a payout
type PayoutProofResponse struct {
	PayoutID        uuid.UUID  `json:"payout_id"`
	RecipientUserID uuid.UUID  `json:"recipient_user_id"`
	Amount          string     `json:"amount"`
	Currency        string     `json:"currency"`
	PayoutHash      string     `json:"payout_hash"`
	MerkleRoot      string     `json:"merkle_root"`
	MerkleProof     []string   `json:"merkle_proof"`
	TransactionHash string     `json:"transaction_hash,omitempty"`
	Confirmations   int        `json:"confirmations"`
	VerifiedOnChain bool       `json:"verified_on_chain"`
}

// Helper to convert big.Float to float64
func floatFromBig(f *big.Float) float64 {
	if f == nil {
		return 0
	}
	result, _ := f.Float64()
	return result
}

