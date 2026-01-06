package blockchain_entities

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	blockchain_vo "github.com/replay-api/replay-api/pkg/domain/blockchain/value-objects"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

// OnChainPrizePoolStatus represents the on-chain prize pool state
type OnChainPrizePoolStatus string

const (
	PoolStatusNotCreated   OnChainPrizePoolStatus = "NotCreated"
	PoolStatusAccumulating OnChainPrizePoolStatus = "Accumulating"
	PoolStatusLocked       OnChainPrizePoolStatus = "Locked"
	PoolStatusInEscrow     OnChainPrizePoolStatus = "InEscrow"
	PoolStatusDistributed  OnChainPrizePoolStatus = "Distributed"
	PoolStatusCancelled    OnChainPrizePoolStatus = "Cancelled"
)

// OnChainPrizePool represents a prize pool on the blockchain.
// Thread-safe: all mutating operations are protected by internal mutex.
type OnChainPrizePool struct {
	shared.BaseEntity
	mu sync.RWMutex `json:"-" bson:"-"` // Protects concurrent access to mutable fields

	// Identifiers
	MatchID      uuid.UUID               `json:"match_id" bson:"match_id"`
	OnChainID    [32]byte                `json:"on_chain_id" bson:"on_chain_id"` // bytes32 in contract
	ChainID      blockchain_vo.ChainID   `json:"chain_id" bson:"chain_id"`
	ContractAddr wallet_vo.EVMAddress    `json:"contract_address" bson:"contract_address"`

	// Token info
	TokenAddress wallet_vo.EVMAddress   `json:"token_address" bson:"token_address"`
	Currency     wallet_vo.Currency     `json:"currency" bson:"currency"`

	// Amounts
	TotalAmount          wallet_vo.Amount `json:"total_amount" bson:"total_amount"`
	PlatformContribution wallet_vo.Amount `json:"platform_contribution" bson:"platform_contribution"`
	EntryFeePerPlayer    wallet_vo.Amount `json:"entry_fee_per_player" bson:"entry_fee_per_player"`
	PlatformFeePercent   uint16           `json:"platform_fee_percent" bson:"platform_fee_percent"` // basis points

	// Participants - Contributions map is the source of truth for O(1) duplicate checking
	Participants  []wallet_vo.EVMAddress       `json:"participants" bson:"participants"`
	Contributions map[string]wallet_vo.Amount  `json:"contributions" bson:"contributions"` // address -> amount

	// Status tracking
	Status         OnChainPrizePoolStatus `json:"status" bson:"status"`
	CreatedAt      time.Time              `json:"created_at_chain" bson:"created_at_chain"`
	LockedAt       *time.Time             `json:"locked_at,omitempty" bson:"locked_at,omitempty"`
	EscrowEndTime  *time.Time             `json:"escrow_end_time,omitempty" bson:"escrow_end_time,omitempty"`
	DistributedAt  *time.Time             `json:"distributed_at,omitempty" bson:"distributed_at,omitempty"`

	// Transaction references
	CreateTxHash     *blockchain_vo.TxHash `json:"create_tx_hash,omitempty" bson:"create_tx_hash,omitempty"`
	LockTxHash       *blockchain_vo.TxHash `json:"lock_tx_hash,omitempty" bson:"lock_tx_hash,omitempty"`
	DistributeTxHash *blockchain_vo.TxHash `json:"distribute_tx_hash,omitempty" bson:"distribute_tx_hash,omitempty"`

	// Distribution results
	Winners       []PrizeWinner `json:"winners,omitempty" bson:"winners,omitempty"`
	PlatformFee   wallet_vo.Amount `json:"platform_fee,omitempty" bson:"platform_fee,omitempty"`

	// Sync state
	LastSyncBlock uint64    `json:"last_sync_block" bson:"last_sync_block"`
	LastSyncTime  time.Time `json:"last_sync_time" bson:"last_sync_time"`
	IsSynced      bool      `json:"is_synced" bson:"is_synced"`
}

// PrizeWinner represents a winner in the prize distribution
type PrizeWinner struct {
	Address     wallet_vo.EVMAddress `json:"address" bson:"address"`
	Rank        uint8                `json:"rank" bson:"rank"`
	Amount      wallet_vo.Amount     `json:"amount" bson:"amount"`
	ShareBPS    uint16               `json:"share_bps" bson:"share_bps"` // basis points
	IsMVP       bool                 `json:"is_mvp" bson:"is_mvp"`
	WithdrawnAt *time.Time           `json:"withdrawn_at,omitempty" bson:"withdrawn_at,omitempty"`
}

// NewOnChainPrizePool creates a new on-chain prize pool record
func NewOnChainPrizePool(
	resourceOwner shared.ResourceOwner,
	matchID uuid.UUID,
	chainID blockchain_vo.ChainID,
	contractAddr wallet_vo.EVMAddress,
	tokenAddr wallet_vo.EVMAddress,
	currency wallet_vo.Currency,
	entryFee wallet_vo.Amount,
	platformFeePercent uint16,
) *OnChainPrizePool {
	baseEntity := shared.NewPrivateEntity(resourceOwner)

	// Generate on-chain ID from match ID
	var onChainID [32]byte
	copy(onChainID[:], matchID[:])

	return &OnChainPrizePool{
		BaseEntity:         baseEntity,
		MatchID:            matchID,
		OnChainID:          onChainID,
		ChainID:            chainID,
		ContractAddr:       contractAddr,
		TokenAddress:       tokenAddr,
		Currency:           currency,
		EntryFeePerPlayer:  entryFee,
		PlatformFeePercent: platformFeePercent,
		Status:             PoolStatusNotCreated,
		Participants:       []wallet_vo.EVMAddress{},
		Contributions:      make(map[string]wallet_vo.Amount),
		TotalAmount:        wallet_vo.NewAmount(0),
		LastSyncTime:       time.Now(),
	}
}

// MarkCreated marks the pool as created on-chain.
// Thread-safe: uses internal mutex.
func (p *OnChainPrizePool) MarkCreated(txHash blockchain_vo.TxHash, platformContribution wallet_vo.Amount) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.Status = PoolStatusAccumulating
	p.CreateTxHash = &txHash
	p.PlatformContribution = platformContribution
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
}

// AddParticipant adds a participant to the pool.
// Thread-safe: uses internal mutex for concurrent access protection.
// Uses O(1) map lookup for duplicate detection.
func (p *OnChainPrizePool) AddParticipant(addr wallet_vo.EVMAddress, contribution wallet_vo.Amount) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.Status != PoolStatusAccumulating {
		return fmt.Errorf("pool is not accepting participants: status=%s", p.Status)
	}

	// O(1) duplicate check using Contributions map as source of truth
	addrKey := addr.String()
	if _, exists := p.Contributions[addrKey]; exists {
		return fmt.Errorf("address already participating: %s", addrKey)
	}

	p.Participants = append(p.Participants, addr)
	p.Contributions[addrKey] = contribution
	p.TotalAmount = p.TotalAmount.Add(contribution)
	p.UpdatedAt = time.Now()

	return nil
}

// Lock locks the prize pool, preventing further participant additions.
// Requires minimum 2 participants for a valid competition.
// Thread-safe: uses internal mutex.
func (p *OnChainPrizePool) Lock(txHash blockchain_vo.TxHash) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.Status != PoolStatusAccumulating {
		return fmt.Errorf("cannot lock pool: status=%s", p.Status)
	}
	if len(p.Participants) < 2 {
		return fmt.Errorf("not enough participants: %d (minimum: 2)", len(p.Participants))
	}

	now := time.Now()
	p.Status = PoolStatusLocked
	p.LockTxHash = &txHash
	p.LockedAt = &now
	p.TotalAmount = p.TotalAmount.Add(p.PlatformContribution)
	p.UpdatedAt = now

	return nil
}

// StartEscrow moves pool to escrow period.
// The escrow period allows dispute resolution before prize distribution.
// Thread-safe: uses internal mutex.
func (p *OnChainPrizePool) StartEscrow(escrowDuration time.Duration) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.Status != PoolStatusLocked {
		return fmt.Errorf("cannot start escrow: status=%s", p.Status)
	}

	now := time.Now()
	endTime := now.Add(escrowDuration)
	p.Status = PoolStatusInEscrow
	p.EscrowEndTime = &endTime
	p.UpdatedAt = now

	return nil
}

// Distribute marks the pool as distributed and records winners.
// Must be called only after escrow period completes.
// Thread-safe: uses internal mutex.
func (p *OnChainPrizePool) Distribute(txHash blockchain_vo.TxHash, winners []PrizeWinner, platformFee wallet_vo.Amount) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.Status != PoolStatusInEscrow {
		return fmt.Errorf("cannot distribute: status=%s", p.Status)
	}

	now := time.Now()
	p.Status = PoolStatusDistributed
	p.DistributeTxHash = &txHash
	p.DistributedAt = &now
	p.Winners = winners
	p.PlatformFee = platformFee
	p.UpdatedAt = now

	return nil
}

// Cancel cancels the pool and allows participant refunds.
// Only allowed during Accumulating or Locked status.
// Thread-safe: uses internal mutex.
func (p *OnChainPrizePool) Cancel() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.Status != PoolStatusAccumulating && p.Status != PoolStatusLocked {
		return fmt.Errorf("cannot cancel: status=%s (only Accumulating or Locked can be cancelled)", p.Status)
	}

	p.Status = PoolStatusCancelled
	p.UpdatedAt = time.Now()

	return nil
}

// UpdateSyncState updates the blockchain sync tracking fields.
// Called by the sync worker after processing blockchain events.
// Thread-safe: uses internal mutex.
func (p *OnChainPrizePool) UpdateSyncState(blockNumber uint64, synced bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.LastSyncBlock = blockNumber
	p.LastSyncTime = time.Now()
	p.IsSynced = synced
	p.UpdatedAt = time.Now()
}

// GetOnChainIDHex returns the on-chain ID as hex string (0x-prefixed).
func (p *OnChainPrizePool) GetOnChainIDHex() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return fmt.Sprintf("0x%x", p.OnChainID)
}

// IsEscrowComplete checks if escrow period is complete.
// Returns false if pool is not in escrow status.
func (p *OnChainPrizePool) IsEscrowComplete() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.Status != PoolStatusInEscrow || p.EscrowEndTime == nil {
		return false
	}
	return time.Now().After(*p.EscrowEndTime)
}

// GetDistributableAmount returns amount available for distribution (after platform fee).
// Platform fee is deducted as basis points (1000 = 10%).
func (p *OnChainPrizePool) GetDistributableAmount() wallet_vo.Amount {
	p.mu.RLock()
	defer p.mu.RUnlock()

	feeAmount := wallet_vo.NewAmountFromCents(p.TotalAmount.Cents() * int64(p.PlatformFeePercent) / 10000)
	return p.TotalAmount.Subtract(feeAmount)
}

// ParticipantCount returns the number of participants (thread-safe).
func (p *OnChainPrizePool) ParticipantCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.Participants)
}

// GetStatus returns the current pool status (thread-safe).
func (p *OnChainPrizePool) GetStatus() OnChainPrizePoolStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Status
}
