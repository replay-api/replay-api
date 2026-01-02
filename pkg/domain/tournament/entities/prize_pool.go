package tournament_entities

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
)

// PrizePoolType defines the type of prize pool
type PrizePoolType string

const (
	PrizePoolTypeFixed       PrizePoolType = "FIXED"       // Pre-defined prize pool
	PrizePoolTypeCrowdfunded PrizePoolType = "CROWDFUNDED" // Contributions from entries
	PrizePoolTypeSponsored   PrizePoolType = "SPONSORED"   // Sponsor-funded
	PrizePoolTypeHybrid      PrizePoolType = "HYBRID"      // Mixed funding sources
)

// PrizePoolStatus tracks the prize pool state
type PrizePoolStatus string

const (
	PrizePoolStatusPending     PrizePoolStatus = "PENDING"
	PrizePoolStatusFunded      PrizePoolStatus = "FUNDED"
	PrizePoolStatusDistributing PrizePoolStatus = "DISTRIBUTING"
	PrizePoolStatusCompleted   PrizePoolStatus = "COMPLETED"
	PrizePoolStatusDisputed    PrizePoolStatus = "DISPUTED"
	PrizePoolStatusRefunded    PrizePoolStatus = "REFUNDED"
)

// PaymentMethod defines how prizes are paid
type PaymentMethod string

const (
	PaymentMethodPlatformWallet PaymentMethod = "PLATFORM_WALLET"
	PaymentMethodCrypto         PaymentMethod = "CRYPTO"
	PaymentMethodBankTransfer   PaymentMethod = "BANK_TRANSFER"
	PaymentMethodPayPal         PaymentMethod = "PAYPAL"
)

// PrizePool represents a tournament's prize pool with blockchain verification
type PrizePool struct {
	common.BaseEntity
	TournamentID   uuid.UUID         `json:"tournament_id" bson:"tournament_id"`
	Type           PrizePoolType     `json:"type" bson:"type"`
	Status         PrizePoolStatus   `json:"status" bson:"status"`

	// Financial totals
	TotalAmount    *big.Float        `json:"total_amount" bson:"total_amount"`
	Currency       string            `json:"currency" bson:"currency"`
	EntryFeeAmount *big.Float        `json:"entry_fee_amount" bson:"entry_fee_amount"`
	SponsorAmount  *big.Float        `json:"sponsor_amount" bson:"sponsor_amount"`
	
	// Platform fee configuration
	PlatformFeePercent float64       `json:"platform_fee_percent" bson:"platform_fee_percent"`
	PlatformFeeAmount  *big.Float    `json:"platform_fee_amount" bson:"platform_fee_amount"`
	
	// Distribution configuration
	Distribution   []PrizeDistribution `json:"distribution" bson:"distribution"`
	
	// Funding sources
	Contributions  []PrizeContribution `json:"contributions" bson:"contributions"`
	
	// Payouts
	Payouts        []PrizePayout       `json:"payouts" bson:"payouts"`
	
	// Blockchain verification
	BlockchainVerification *BlockchainVerification `json:"blockchain_verification,omitempty" bson:"blockchain_verification,omitempty"`
	
	// Integrity
	Hash           string              `json:"hash" bson:"hash"`
	MerkleRoot     string              `json:"merkle_root" bson:"merkle_root"`
	
	// Timing
	FundingDeadline *time.Time         `json:"funding_deadline,omitempty" bson:"funding_deadline,omitempty"`
	DistributedAt   *time.Time         `json:"distributed_at,omitempty" bson:"distributed_at,omitempty"`
}

// PrizeDistribution defines how prizes are split
type PrizeDistribution struct {
	Position       int        `json:"position" bson:"position"`         // 1st, 2nd, 3rd, etc.
	Percentage     float64    `json:"percentage" bson:"percentage"`     // Percentage of pool
	FixedAmount    *big.Float `json:"fixed_amount" bson:"fixed_amount"` // Or fixed amount
	Description    string     `json:"description" bson:"description"`   // "1st Place", "MVP", etc.
}

// PrizeContribution represents a funding source
type PrizeContribution struct {
	ID             uuid.UUID       `json:"id" bson:"id"`
	Type           string          `json:"type" bson:"type"`           // "entry_fee", "sponsor", "donation"
	SourceUserID   *uuid.UUID      `json:"source_user_id,omitempty" bson:"source_user_id,omitempty"`
	SponsorName    string          `json:"sponsor_name,omitempty" bson:"sponsor_name,omitempty"`
	Amount         *big.Float      `json:"amount" bson:"amount"`
	Currency       string          `json:"currency" bson:"currency"`
	TransactionID  uuid.UUID       `json:"transaction_id" bson:"transaction_id"`
	ExternalRef    string          `json:"external_ref,omitempty" bson:"external_ref,omitempty"`
	ContributedAt  time.Time       `json:"contributed_at" bson:"contributed_at"`
	Status         string          `json:"status" bson:"status"`
	Hash           string          `json:"hash" bson:"hash"`
}

// PrizePayout represents a payment to a winner
type PrizePayout struct {
	ID               uuid.UUID       `json:"id" bson:"id"`
	RecipientUserID  uuid.UUID       `json:"recipient_user_id" bson:"recipient_user_id"`
	RecipientTeamID  *uuid.UUID      `json:"recipient_team_id,omitempty" bson:"recipient_team_id,omitempty"`
	Position         int             `json:"position" bson:"position"`
	GrossAmount      *big.Float      `json:"gross_amount" bson:"gross_amount"`
	NetAmount        *big.Float      `json:"net_amount" bson:"net_amount"`
	PlatformFee      *big.Float      `json:"platform_fee" bson:"platform_fee"`
	TaxWithheld      *big.Float      `json:"tax_withheld" bson:"tax_withheld"`
	Currency         string          `json:"currency" bson:"currency"`
	PaymentMethod    PaymentMethod   `json:"payment_method" bson:"payment_method"`
	WalletAddress    string          `json:"wallet_address,omitempty" bson:"wallet_address,omitempty"`
	Status           PayoutStatus    `json:"status" bson:"status"`
	TransactionHash  string          `json:"transaction_hash,omitempty" bson:"transaction_hash,omitempty"`
	BlockConfirmations int           `json:"block_confirmations" bson:"block_confirmations"`
	ScheduledAt      time.Time       `json:"scheduled_at" bson:"scheduled_at"`
	ProcessedAt      *time.Time      `json:"processed_at,omitempty" bson:"processed_at,omitempty"`
	CompletedAt      *time.Time      `json:"completed_at,omitempty" bson:"completed_at,omitempty"`
	FailureReason    string          `json:"failure_reason,omitempty" bson:"failure_reason,omitempty"`
	Hash             string          `json:"hash" bson:"hash"`
}

// PayoutStatus tracks payout state
type PayoutStatus string

const (
	PayoutStatusPending    PayoutStatus = "PENDING"
	PayoutStatusApproved   PayoutStatus = "APPROVED"
	PayoutStatusProcessing PayoutStatus = "PROCESSING"
	PayoutStatusCompleted  PayoutStatus = "COMPLETED"
	PayoutStatusFailed     PayoutStatus = "FAILED"
	PayoutStatusCanceled   PayoutStatus = "CANCELED"
)

// BlockchainVerification stores blockchain proof data
type BlockchainVerification struct {
	ChainID          string    `json:"chain_id" bson:"chain_id"`               // e.g., "ethereum", "polygon", "solana"
	ContractAddress  string    `json:"contract_address" bson:"contract_address"`
	FundingTxHash    string    `json:"funding_tx_hash" bson:"funding_tx_hash"`
	DistributionTxHash string  `json:"distribution_tx_hash,omitempty" bson:"distribution_tx_hash,omitempty"`
	MerkleRoot       string    `json:"merkle_root" bson:"merkle_root"`
	BlockNumber      uint64    `json:"block_number" bson:"block_number"`
	Timestamp        time.Time `json:"timestamp" bson:"timestamp"`
	Verified         bool      `json:"verified" bson:"verified"`
	VerificationProof string   `json:"verification_proof" bson:"verification_proof"`
}

// NewPrizePool creates a new prize pool
func NewPrizePool(tournamentID uuid.UUID, poolType PrizePoolType, currency string, platformFeePercent float64, rxn common.ResourceOwner) *PrizePool {
	return &PrizePool{
		BaseEntity:         common.NewEntity(rxn),
		TournamentID:       tournamentID,
		Type:               poolType,
		Status:             PrizePoolStatusPending,
		TotalAmount:        big.NewFloat(0),
		Currency:           currency,
		EntryFeeAmount:     big.NewFloat(0),
		SponsorAmount:      big.NewFloat(0),
		PlatformFeePercent: platformFeePercent,
		PlatformFeeAmount:  big.NewFloat(0),
		Distribution:       make([]PrizeDistribution, 0),
		Contributions:      make([]PrizeContribution, 0),
		Payouts:            make([]PrizePayout, 0),
	}
}

// SetDistribution configures how the prize pool is split
func (p *PrizePool) SetDistribution(distribution []PrizeDistribution) error {
	var totalPercent float64
	for _, d := range distribution {
		totalPercent += d.Percentage
	}

	// Allow for rounding errors
	if totalPercent > 100.01 || totalPercent < 99.99 && totalPercent != 0 {
		return fmt.Errorf("distribution percentages must sum to 100%%, got %.2f%%", totalPercent)
	}

	p.Distribution = distribution
	return nil
}

// AddContribution adds a funding contribution
func (p *PrizePool) AddContribution(contribution PrizeContribution) error {
	if contribution.Amount == nil || contribution.Amount.Cmp(big.NewFloat(0)) <= 0 {
		return errors.New("contribution amount must be positive")
	}

	// Compute hash for contribution integrity
	contribution.Hash = contribution.computeHash()
	p.Contributions = append(p.Contributions, contribution)

	// Update totals
	p.TotalAmount = new(big.Float).Add(p.TotalAmount, contribution.Amount)
	switch contribution.Type {
	case "entry_fee":
		p.EntryFeeAmount = new(big.Float).Add(p.EntryFeeAmount, contribution.Amount)
	case "sponsor":
		p.SponsorAmount = new(big.Float).Add(p.SponsorAmount, contribution.Amount)
	}

	// Recalculate platform fee
	p.calculatePlatformFee()

	// Update merkle root
	p.updateMerkleRoot()

	return nil
}

// calculatePlatformFee calculates the platform's fee
func (p *PrizePool) calculatePlatformFee() {
	feeMultiplier := big.NewFloat(p.PlatformFeePercent / 100)
	p.PlatformFeeAmount = new(big.Float).Mul(p.TotalAmount, feeMultiplier)
}

// GetDistributableAmount returns total minus platform fee
func (p *PrizePool) GetDistributableAmount() *big.Float {
	return new(big.Float).Sub(p.TotalAmount, p.PlatformFeeAmount)
}

// CalculatePayouts generates payouts based on distribution and results
func (p *PrizePool) CalculatePayouts(results []TournamentResult) ([]PrizePayout, error) {
	if len(p.Distribution) == 0 {
		return nil, errors.New("distribution not configured")
	}

	distributable := p.GetDistributableAmount()
	payouts := make([]PrizePayout, 0)

	for _, dist := range p.Distribution {
		// Find winner for this position
		var winner *TournamentResult
		for i := range results {
			if results[i].Position == dist.Position {
				winner = &results[i]
				break
			}
		}

		if winner == nil {
			continue
		}

		// Calculate payout amount
		var grossAmount *big.Float
		if dist.FixedAmount != nil && dist.FixedAmount.Cmp(big.NewFloat(0)) > 0 {
			grossAmount = dist.FixedAmount
		} else {
			percentMultiplier := big.NewFloat(dist.Percentage / 100)
			grossAmount = new(big.Float).Mul(distributable, percentMultiplier)
		}

		// Calculate net amount (after any additional withholdings)
		// For simplicity, net = gross (tax withholding would be handled separately)
		netAmount := new(big.Float).Set(grossAmount)
		platformFee := big.NewFloat(0)
		taxWithheld := big.NewFloat(0)

		payout := PrizePayout{
			ID:              uuid.New(),
			RecipientUserID: winner.UserID,
			RecipientTeamID: winner.TeamID,
			Position:        dist.Position,
			GrossAmount:     grossAmount,
			NetAmount:       netAmount,
			PlatformFee:     platformFee,
			TaxWithheld:     taxWithheld,
			Currency:        p.Currency,
			PaymentMethod:   PaymentMethodPlatformWallet,
			Status:          PayoutStatusPending,
			ScheduledAt:     time.Now().UTC(),
		}

		payout.Hash = payout.computeHash()
		payouts = append(payouts, payout)
	}

	return payouts, nil
}

// SetPayouts sets the calculated payouts
func (p *PrizePool) SetPayouts(payouts []PrizePayout) {
	p.Payouts = payouts
	p.updateMerkleRoot()
}

// MarkAsFunded marks the prize pool as fully funded
func (p *PrizePool) MarkAsFunded() {
	p.Status = PrizePoolStatusFunded
	p.ComputeHash()
}

// StartDistribution begins the payout process
func (p *PrizePool) StartDistribution() error {
	if p.Status != PrizePoolStatusFunded {
		return fmt.Errorf("prize pool must be funded before distribution, current status: %s", p.Status)
	}

	p.Status = PrizePoolStatusDistributing
	return nil
}

// MarkPayoutCompleted marks a specific payout as completed
func (p *PrizePool) MarkPayoutCompleted(payoutID uuid.UUID, txHash string, confirmations int) error {
	for i := range p.Payouts {
		if p.Payouts[i].ID == payoutID {
			now := time.Now().UTC()
			p.Payouts[i].Status = PayoutStatusCompleted
			p.Payouts[i].TransactionHash = txHash
			p.Payouts[i].BlockConfirmations = confirmations
			p.Payouts[i].CompletedAt = &now
			p.Payouts[i].Hash = p.Payouts[i].computeHash()
			
			// Check if all payouts are complete
			p.checkDistributionComplete()
			return nil
		}
	}

	return errors.New("payout not found")
}

// checkDistributionComplete checks if all payouts are done
func (p *PrizePool) checkDistributionComplete() {
	for _, payout := range p.Payouts {
		if payout.Status != PayoutStatusCompleted && payout.Status != PayoutStatusCanceled {
			return
		}
	}

	now := time.Now().UTC()
	p.Status = PrizePoolStatusCompleted
	p.DistributedAt = &now
	p.ComputeHash()
}

// SetBlockchainVerification sets blockchain verification data
func (p *PrizePool) SetBlockchainVerification(verification BlockchainVerification) {
	p.BlockchainVerification = &verification
}

// ComputeHash generates the integrity hash for the prize pool
func (p *PrizePool) ComputeHash() string {
	data := struct {
		ID             uuid.UUID
		TournamentID   uuid.UUID
		TotalAmount    string
		Currency       string
		Status         PrizePoolStatus
		MerkleRoot     string
		DistributedAt  *time.Time
	}{
		ID:            p.ID,
		TournamentID:  p.TournamentID,
		TotalAmount:   p.TotalAmount.Text('f', 8),
		Currency:      p.Currency,
		Status:        p.Status,
		MerkleRoot:    p.MerkleRoot,
		DistributedAt: p.DistributedAt,
	}

	jsonBytes, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonBytes)
	p.Hash = hex.EncodeToString(hash[:])
	return p.Hash
}

// updateMerkleRoot builds a merkle tree from all transactions
func (p *PrizePool) updateMerkleRoot() {
	var leaves []string

	// Add contribution hashes
	for _, c := range p.Contributions {
		leaves = append(leaves, c.Hash)
	}

	// Add payout hashes
	for _, payout := range p.Payouts {
		leaves = append(leaves, payout.Hash)
	}

	if len(leaves) == 0 {
		p.MerkleRoot = ""
		return
	}

	p.MerkleRoot = computeMerkleRoot(leaves)
}

// VerifyIntegrity verifies the prize pool's data integrity
func (p *PrizePool) VerifyIntegrity() bool {
	// Verify each contribution hash
	for _, c := range p.Contributions {
		if c.Hash != c.computeHash() {
			return false
		}
	}

	// Verify each payout hash
	for _, payout := range p.Payouts {
		if payout.Hash != payout.computeHash() {
			return false
		}
	}

	// Verify merkle root
	expectedMerkle := p.MerkleRoot
	p.updateMerkleRoot()
	if p.MerkleRoot != expectedMerkle {
		return false
	}

	// Verify overall hash
	expectedHash := p.Hash
	p.ComputeHash()
	return p.Hash == expectedHash
}

// TournamentResult represents a participant's result
type TournamentResult struct {
	Position int
	UserID   uuid.UUID
	TeamID   *uuid.UUID
	Score    int
}

// computeHash generates hash for a contribution
func (c *PrizeContribution) computeHash() string {
	data := struct {
		ID            uuid.UUID
		Type          string
		Amount        string
		Currency      string
		TransactionID uuid.UUID
		ContributedAt time.Time
	}{
		ID:            c.ID,
		Type:          c.Type,
		Amount:        c.Amount.Text('f', 8),
		Currency:      c.Currency,
		TransactionID: c.TransactionID,
		ContributedAt: c.ContributedAt,
	}

	jsonBytes, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonBytes)
	return hex.EncodeToString(hash[:])
}

// computeHash generates hash for a payout
func (p *PrizePayout) computeHash() string {
	data := struct {
		ID              uuid.UUID
		RecipientUserID uuid.UUID
		Position        int
		GrossAmount     string
		NetAmount       string
		Currency        string
		TransactionHash string
		ScheduledAt     time.Time
	}{
		ID:              p.ID,
		RecipientUserID: p.RecipientUserID,
		Position:        p.Position,
		GrossAmount:     p.GrossAmount.Text('f', 8),
		NetAmount:       p.NetAmount.Text('f', 8),
		Currency:        p.Currency,
		TransactionHash: p.TransactionHash,
		ScheduledAt:     p.ScheduledAt,
	}

	jsonBytes, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonBytes)
	return hex.EncodeToString(hash[:])
}

// computeMerkleRoot builds merkle root from leaf hashes
func computeMerkleRoot(leaves []string) string {
	if len(leaves) == 0 {
		return ""
	}

	if len(leaves) == 1 {
		return leaves[0]
	}

	// Pad to even number
	if len(leaves)%2 != 0 {
		leaves = append(leaves, leaves[len(leaves)-1])
	}

	var parents []string
	for i := 0; i < len(leaves); i += 2 {
		combined := leaves[i] + leaves[i+1]
		hash := sha256.Sum256([]byte(combined))
		parents = append(parents, hex.EncodeToString(hash[:]))
	}

	return computeMerkleRoot(parents)
}

// GetVerificationProof generates a merkle proof for a specific transaction
func (p *PrizePool) GetVerificationProof(txHash string) ([]string, error) {
	var leaves []string
	targetIndex := -1

	// Collect all leaves
	for i, c := range p.Contributions {
		leaves = append(leaves, c.Hash)
		if c.Hash == txHash {
			targetIndex = len(leaves) - 1
		}
		_ = i
	}

	for i, payout := range p.Payouts {
		leaves = append(leaves, payout.Hash)
		if payout.Hash == txHash {
			targetIndex = len(leaves) - 1
		}
		_ = i
	}

	if targetIndex == -1 {
		return nil, errors.New("transaction not found in prize pool")
	}

	// Build merkle proof
	return buildMerkleProof(leaves, targetIndex), nil
}

// buildMerkleProof creates proof path from leaf to root
func buildMerkleProof(leaves []string, targetIndex int) []string {
	if len(leaves) <= 1 {
		return nil
	}

	proof := make([]string, 0)

	// Pad to even
	if len(leaves)%2 != 0 {
		leaves = append(leaves, leaves[len(leaves)-1])
	}

	// Get sibling
	siblingIndex := targetIndex ^ 1 // XOR to get sibling
	proof = append(proof, leaves[siblingIndex])

	// Build parent level
	var parents []string
	for i := 0; i < len(leaves); i += 2 {
		combined := leaves[i] + leaves[i+1]
		hash := sha256.Sum256([]byte(combined))
		parents = append(parents, hex.EncodeToString(hash[:]))
	}

	// Recursively build proof
	parentIndex := targetIndex / 2
	parentProof := buildMerkleProof(parents, parentIndex)
	proof = append(proof, parentProof...)

	return proof
}

