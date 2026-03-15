package exchange_vo

import (
	"fmt"
	"os"
	"strconv"
)

// FeeConfig holds configurable fee percentages for Bitcoin trading
// All values are percentages (e.g., 1.5 means 1.5%)
type FeeConfig struct {
	BuyFeePercent  float64 `json:"buy_fee_percent" bson:"buy_fee_percent"`
	SellFeePercent float64 `json:"sell_fee_percent" bson:"sell_fee_percent"`
	MinFeeUSD      float64 `json:"min_fee_usd" bson:"min_fee_usd"`
	MaxFeeUSD      float64 `json:"max_fee_usd" bson:"max_fee_usd"`
}

// PlanTier represents subscription plan tiers for fee calculation
type PlanTier string

const (
	PlanTierFree       PlanTier = "free"
	PlanTierStarter    PlanTier = "starter"
	PlanTierPro        PlanTier = "pro"
	PlanTierTeam       PlanTier = "team"
	PlanTierBusiness   PlanTier = "business"
	PlanTierCustom     PlanTier = "custom"
)

// DefaultFeeConfigs returns default fee configurations per plan tier
func DefaultFeeConfigs() map[PlanTier]FeeConfig {
	return map[PlanTier]FeeConfig{
		PlanTierFree: {
			BuyFeePercent:  2.0,
			SellFeePercent: 1.5,
			MinFeeUSD:      0.50,
			MaxFeeUSD:      500.00,
		},
		PlanTierStarter: {
			BuyFeePercent:  1.5,
			SellFeePercent: 1.0,
			MinFeeUSD:      0.50,
			MaxFeeUSD:      500.00,
		},
		PlanTierPro: {
			BuyFeePercent:  1.0,
			SellFeePercent: 0.75,
			MinFeeUSD:      0.25,
			MaxFeeUSD:      250.00,
		},
		PlanTierTeam: {
			BuyFeePercent:  0.75,
			SellFeePercent: 0.50,
			MinFeeUSD:      0.10,
			MaxFeeUSD:      200.00,
		},
		PlanTierBusiness: {
			BuyFeePercent:  0.50,
			SellFeePercent: 0.35,
			MinFeeUSD:      0.10,
			MaxFeeUSD:      100.00,
		},
		PlanTierCustom: {
			BuyFeePercent:  0.25,
			SellFeePercent: 0.15,
			MinFeeUSD:      0.00,
			MaxFeeUSD:      50.00,
		},
	}
}

// GetFeeConfig returns the fee config for a given plan tier, with env var overrides
func GetFeeConfig(tier PlanTier) FeeConfig {
	defaults := DefaultFeeConfigs()
	config, ok := defaults[tier]
	if !ok {
		config = defaults[PlanTierFree] // fallback to free tier
	}

	// Allow env var overrides for global fee adjustments
	if v := os.Getenv("BTC_BUY_FEE_PERCENT"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			config.BuyFeePercent = f
		}
	}
	if v := os.Getenv("BTC_SELL_FEE_PERCENT"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			config.SellFeePercent = f
		}
	}
	if v := os.Getenv("BTC_MIN_FEE_USD"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			config.MinFeeUSD = f
		}
	}
	if v := os.Getenv("BTC_MAX_FEE_USD"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			config.MaxFeeUSD = f
		}
	}

	return config
}

// CalculateFee calculates the fee for a given side and USD amount
func (fc FeeConfig) CalculateFee(side OrderSide, amountUSD float64) float64 {
	var feePercent float64
	if side == OrderSideBuy {
		feePercent = fc.BuyFeePercent
	} else {
		feePercent = fc.SellFeePercent
	}

	fee := amountUSD * (feePercent / 100.0)

	// Apply min/max bounds
	if fee < fc.MinFeeUSD {
		fee = fc.MinFeeUSD
	}
	if fee > fc.MaxFeeUSD {
		fee = fc.MaxFeeUSD
	}

	return fee
}

// Validate ensures fee config has reasonable values
func (fc FeeConfig) Validate() error {
	if fc.BuyFeePercent < 0 || fc.BuyFeePercent > 10 {
		return fmt.Errorf("buy fee percent must be between 0%% and 10%%, got %.2f%%", fc.BuyFeePercent)
	}
	if fc.SellFeePercent < 0 || fc.SellFeePercent > 10 {
		return fmt.Errorf("sell fee percent must be between 0%% and 10%%, got %.2f%%", fc.SellFeePercent)
	}
	if fc.MinFeeUSD < 0 {
		return fmt.Errorf("min fee must be non-negative, got $%.2f", fc.MinFeeUSD)
	}
	if fc.MaxFeeUSD < fc.MinFeeUSD {
		return fmt.Errorf("max fee ($%.2f) must be >= min fee ($%.2f)", fc.MaxFeeUSD, fc.MinFeeUSD)
	}
	return nil
}

// FeeResult holds the calculated fee details for a trade
type FeeResult struct {
	FeePercent float64  `json:"fee_percent"`
	FeeAmount  float64  `json:"fee_amount_usd"`
	NetAmount  float64  `json:"net_amount_usd"`
	PlanTier   PlanTier `json:"plan_tier"`
}

// BTCWithdrawalFeeConfig holds configurable fees for Bitcoin withdrawals
type BTCWithdrawalFeeConfig struct {
	OnChainFeePercent  float64 `json:"onchain_fee_percent"`   // Platform fee % for on-chain (default 0.5%)
	LightningFeePercent float64 `json:"lightning_fee_percent"` // Platform fee % for Lightning (default 0.1%)
	MinLightningSats   int64   `json:"min_lightning_sats"`    // Minimum Lightning fee in sats (default 100)
	PassthroughNetworkFee bool  `json:"passthrough_network_fee"` // Whether miner fees are passed to user
}

// DefaultBTCWithdrawalFeeConfig returns default Bitcoin withdrawal fee config
func DefaultBTCWithdrawalFeeConfig() BTCWithdrawalFeeConfig {
	return BTCWithdrawalFeeConfig{
		OnChainFeePercent:    0.5,
		LightningFeePercent:  0.1,
		MinLightningSats:     100,
		PassthroughNetworkFee: true,
	}
}
