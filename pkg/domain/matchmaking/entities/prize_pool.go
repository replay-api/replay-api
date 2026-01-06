package matchmaking_entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	matchmaking_vo "github.com/replay-api/replay-api/pkg/domain/matchmaking/value-objects"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// PrizePool is an aggregate root representing accumulated prize money for a match
type PrizePool struct {
	shared.BaseEntity
	MatchID              uuid.UUID                         `json:"match_id" bson:"match_id"`
	GameID               replay_common.GameIDKey           `json:"game_id" bson:"game_id"`
	Region               string                            `json:"region" bson:"region"`
	Currency             wallet_vo.Currency                `json:"currency" bson:"currency"`
	TotalAmount          wallet_vo.Amount                  `json:"total_amount" bson:"total_amount"`
	PlatformContribution wallet_vo.Amount                  `json:"platform_contribution" bson:"platform_contribution"` // $0.50 from platform
	PlayerContributions  map[uuid.UUID]wallet_vo.Amount    `json:"player_contributions" bson:"player_contributions"`
	DistributionRule     matchmaking_vo.DistributionRule   `json:"distribution_rule" bson:"distribution_rule"`
	Status               PrizePoolStatus                   `json:"status" bson:"status"`
	LockedAt             *time.Time                        `json:"locked_at,omitempty" bson:"locked_at,omitempty"`
	DistributedAt        *time.Time                             `json:"distributed_at,omitempty" bson:"distributed_at,omitempty"`
	Winners              []PrizeWinner                          `json:"winners,omitempty" bson:"winners,omitempty"`
	MVPPlayerID          *uuid.UUID                             `json:"mvp_player_id,omitempty" bson:"mvp_player_id,omitempty"`
	EscrowEndTime        *time.Time                             `json:"escrow_end_time,omitempty" bson:"escrow_end_time,omitempty"` // 72h hold for disputes
}

// PrizePoolStatus represents the lifecycle state of a prize pool
type PrizePoolStatus string

const (
	PrizePoolStatusAccumulating PrizePoolStatus = "accumulating" // Players are joining, pool is growing
	PrizePoolStatusLocked       PrizePoolStatus = "locked"       // Match started, no more contributions
	PrizePoolStatusInEscrow     PrizePoolStatus = "in_escrow"    // Match complete, waiting dispute period
	PrizePoolStatusDistributed  PrizePoolStatus = "distributed"  // Prizes paid out
	PrizePoolStatusCancelled    PrizePoolStatus = "cancelled"    // Match cancelled, refunds issued
)

// PrizeWinner represents a prize recipient
type PrizeWinner struct {
	PlayerID uuid.UUID        `json:"player_id" bson:"player_id"`
	Rank     int              `json:"rank" bson:"rank"` // 1 = winner, 2 = runner-up, etc.
	Amount   wallet_vo.Amount `json:"amount" bson:"amount"`
	PaidAt   *time.Time       `json:"paid_at,omitempty" bson:"paid_at,omitempty"`
}

// NewPrizePool creates a new prize pool
func NewPrizePool(
	resourceOwner shared.ResourceOwner,
	matchID uuid.UUID,
	gameID replay_common.GameIDKey,
	region string,
	currency wallet_vo.Currency,
	distributionRule matchmaking_vo.DistributionRule,
	platformContribution wallet_vo.Amount,
) *PrizePool {
	baseEntity := shared.NewUnrestrictedEntity(resourceOwner) // Prize pools are public

	return &PrizePool{
		BaseEntity:           baseEntity,
		MatchID:              matchID,
		GameID:               gameID,
		Region:               region,
		Currency:             currency,
		TotalAmount:          platformContribution, // Start with platform contribution
		PlatformContribution: platformContribution,
		PlayerContributions:  make(map[uuid.UUID]wallet_vo.Amount),
		DistributionRule:     distributionRule,
		Status:               PrizePoolStatusAccumulating,
	}
}

// AddPlayerContribution adds a player's entry fee to the pool (invariant: total = sum of contributions)
func (p *PrizePool) AddPlayerContribution(playerID uuid.UUID, amount wallet_vo.Amount) error {
	if p.Status != PrizePoolStatusAccumulating {
		return fmt.Errorf("cannot add contribution to prize pool in status: %s", p.Status)
	}

	if amount.IsNegative() || amount.IsZero() {
		return fmt.Errorf("contribution amount must be positive, got: %s", amount.String())
	}

	// Add to player's total contribution (supports multiple contributions from same player)
	currentContribution := p.PlayerContributions[playerID]
	p.PlayerContributions[playerID] = currentContribution.Add(amount)

	// Update total
	p.TotalAmount = p.TotalAmount.Add(amount)
	p.UpdatedAt = time.Now()

	// Validate invariant
	if err := p.validateTotalAmount(); err != nil {
		return fmt.Errorf("invariant violation after contribution: %w", err)
	}

	return nil
}

// Lock locks the prize pool when match starts (no more contributions allowed)
func (p *PrizePool) Lock() error {
	if p.Status != PrizePoolStatusAccumulating {
		return fmt.Errorf("can only lock prize pool in accumulating status, current: %s", p.Status)
	}

	now := time.Now()
	p.Status = PrizePoolStatusLocked
	p.LockedAt = &now
	p.UpdatedAt = now

	return nil
}

// EnterEscrow moves the pool to escrow after match completion (72h dispute period)
func (p *PrizePool) EnterEscrow(escrowPeriodHours int) error {
	if p.Status != PrizePoolStatusLocked {
		return fmt.Errorf("can only enter escrow from locked status, current: %s", p.Status)
	}

	now := time.Now()
	escrowEnd := now.Add(time.Duration(escrowPeriodHours) * time.Hour)

	p.Status = PrizePoolStatusInEscrow
	p.EscrowEndTime = &escrowEnd
	p.UpdatedAt = now

	return nil
}

// CalculateDistribution calculates prize amounts based on the distribution rule
func (p *PrizePool) CalculateDistribution(rankedPlayerIDs []uuid.UUID, mvpPlayerID *uuid.UUID) (*matchmaking_vo.PrizeDistribution, error) {
	if p.Status != PrizePoolStatusLocked && p.Status != PrizePoolStatusInEscrow {
		return nil, fmt.Errorf("cannot calculate distribution in status: %s", p.Status)
	}

	if len(rankedPlayerIDs) < 1 {
		return nil, fmt.Errorf("must have at least one player (winner)")
	}

	return p.DistributionRule.Calculate(p.TotalAmount, rankedPlayerIDs, mvpPlayerID)
}

// Distribute marks the pool as distributed and records winners
func (p *PrizePool) Distribute(distribution *matchmaking_vo.PrizeDistribution) error {
	if p.Status != PrizePoolStatusInEscrow {
		return fmt.Errorf("can only distribute from escrow status, current: %s", p.Status)
	}

	if p.EscrowEndTime != nil && time.Now().Before(*p.EscrowEndTime) {
		return fmt.Errorf("escrow period not yet ended: %s remaining", time.Until(*p.EscrowEndTime))
	}

	// Validate distribution total equals pool total
	if !distribution.Total.Equals(p.TotalAmount) {
		return fmt.Errorf("distribution total %s does not match pool total %s",
			distribution.Total.String(), p.TotalAmount.String())
	}

	// Record winners
	p.Winners = []PrizeWinner{}
	now := time.Now()

	if distribution.WinnerAmount.IsPositive() && distribution.WinnerPlayerID != uuid.Nil {
		p.Winners = append(p.Winners, PrizeWinner{
			PlayerID: distribution.WinnerPlayerID,
			Rank:     1,
			Amount:   distribution.WinnerAmount,
			PaidAt:   &now,
		})
	}

	if distribution.RunnerUpAmount.IsPositive() && distribution.RunnerUpPlayerID != uuid.Nil {
		p.Winners = append(p.Winners, PrizeWinner{
			PlayerID: distribution.RunnerUpPlayerID,
			Rank:     2,
			Amount:   distribution.RunnerUpAmount,
			PaidAt:   &now,
		})
	}

	if distribution.ThirdPlaceAmount.IsPositive() && distribution.ThirdPlacePlayerID != uuid.Nil {
		p.Winners = append(p.Winners, PrizeWinner{
			PlayerID: distribution.ThirdPlacePlayerID,
			Rank:     3,
			Amount:   distribution.ThirdPlaceAmount,
			PaidAt:   &now,
		})
	}

	if distribution.MVPBonus.IsPositive() && distribution.MVPPlayerID != uuid.Nil {
		p.MVPPlayerID = &distribution.MVPPlayerID
	}

	p.Status = PrizePoolStatusDistributed
	p.DistributedAt = &now
	p.UpdatedAt = now

	return nil
}

// Cancel cancels the prize pool and prepares for refunds
func (p *PrizePool) Cancel(reason string) error {
	if p.Status == PrizePoolStatusDistributed {
		return fmt.Errorf("cannot cancel already distributed prize pool")
	}

	now := time.Now()
	p.Status = PrizePoolStatusCancelled
	p.UpdatedAt = now

	return nil
}

// validateTotalAmount ensures the invariant: total = platform + sum(player contributions)
func (p *PrizePool) validateTotalAmount() error {
	expectedTotal := p.PlatformContribution

	for _, contribution := range p.PlayerContributions {
		expectedTotal = expectedTotal.Add(contribution)
	}

	if !expectedTotal.Equals(p.TotalAmount) {
		return fmt.Errorf("total amount %s does not equal expected sum %s",
			p.TotalAmount.String(), expectedTotal.String())
	}

	return nil
}

// GetPlayerCount returns the number of contributing players
func (p *PrizePool) GetPlayerCount() int {
	return len(p.PlayerContributions)
}

// GetPlayerContribution returns a player's total contribution
func (p *PrizePool) GetPlayerContribution(playerID uuid.UUID) wallet_vo.Amount {
	if amount, exists := p.PlayerContributions[playerID]; exists {
		return amount
	}
	return wallet_vo.NewAmount(0)
}

// Validate ensures prize pool invariants
func (p *PrizePool) Validate() error {
	if p.MatchID == uuid.Nil {
		return fmt.Errorf("match_id cannot be nil")
	}

	if !p.Currency.IsValid() {
		return fmt.Errorf("invalid currency: %s", p.Currency)
	}

	if p.TotalAmount.IsNegative() {
		return fmt.Errorf("total amount cannot be negative")
	}

	if err := p.validateTotalAmount(); err != nil {
		return err
	}

	return nil
}
