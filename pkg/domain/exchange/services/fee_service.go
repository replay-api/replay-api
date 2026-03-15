package exchange_services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	exchange_vo "github.com/replay-api/replay-api/pkg/domain/exchange/value-objects"
)

// SubscriptionPlanResolver resolves a user's current plan tier
type SubscriptionPlanResolver interface {
	GetUserPlanTier(ctx context.Context, userID uuid.UUID) (exchange_vo.PlanTier, error)
}

// FeeService calculates trading fees based on user's subscription plan tier
type FeeService struct {
	planResolver SubscriptionPlanResolver
	feeConfigs   map[exchange_vo.PlanTier]exchange_vo.FeeConfig
}

// NewFeeService creates a new fee service
func NewFeeService(planResolver SubscriptionPlanResolver) *FeeService {
	return &FeeService{
		planResolver: planResolver,
		feeConfigs:   exchange_vo.DefaultFeeConfigs(),
	}
}

// CalculateFee calculates the fee for a given user, side, and amount
func (s *FeeService) CalculateFee(ctx context.Context, userID uuid.UUID, side exchange_vo.OrderSide, amountUSD float64) (*exchange_vo.FeeResult, error) {
	tier := exchange_vo.PlanTierFree
	if s.planResolver != nil {
		resolved, err := s.planResolver.GetUserPlanTier(ctx, userID)
		if err != nil {
			// Log but don't fail - use free tier as fallback
			fmt.Printf("[FeeService] Failed to resolve plan tier for user %s: %v, using free tier\n", userID, err)
		} else {
			tier = resolved
		}
	}

	config := exchange_vo.GetFeeConfig(tier)
	feeAmount := config.CalculateFee(side, amountUSD)
	var feePercent float64
	if side == exchange_vo.OrderSideBuy {
		feePercent = config.BuyFeePercent
	} else {
		feePercent = config.SellFeePercent
	}

	return &exchange_vo.FeeResult{
		FeePercent: feePercent,
		FeeAmount:  feeAmount,
		NetAmount:  amountUSD - feeAmount,
		PlanTier:   tier,
	}, nil
}

// GetFeeSchedule returns all fee tiers for display
func (s *FeeService) GetFeeSchedule() map[exchange_vo.PlanTier]exchange_vo.FeeConfig {
	return s.feeConfigs
}

// GetUserTier resolves the user's current plan tier
func (s *FeeService) GetUserTier(ctx context.Context, userID uuid.UUID) exchange_vo.PlanTier {
	if s.planResolver == nil {
		return exchange_vo.PlanTierFree
	}
	tier, err := s.planResolver.GetUserPlanTier(ctx, userID)
	if err != nil {
		return exchange_vo.PlanTierFree
	}
	return tier
}