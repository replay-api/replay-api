package exchange_usecases

import (
	"context"
	"fmt"
	"time"

	exchange_entities "github.com/replay-api/replay-api/pkg/domain/exchange/entities"
	exchange_in "github.com/replay-api/replay-api/pkg/domain/exchange/ports/in"
	exchange_out "github.com/replay-api/replay-api/pkg/domain/exchange/ports/out"
	exchange_vo "github.com/replay-api/replay-api/pkg/domain/exchange/value-objects"
	exchange_services "github.com/replay-api/replay-api/pkg/domain/exchange/services"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// GetQuoteUseCase generates locked price quotes for Bitcoin trades
type GetQuoteUseCase struct {
	pricingService *exchange_services.PricingService
	feeService     *exchange_services.FeeService
	quoteRepo      exchange_out.QuoteRepository
	eventPublisher exchange_out.ExchangeEventPublisher
	resourceOwner  shared.ResourceOwner
	quoteTTL       time.Duration
}

// NewGetQuoteUseCase creates a new GetQuoteUseCase
func NewGetQuoteUseCase(
	pricingService *exchange_services.PricingService,
	feeService *exchange_services.FeeService,
	quoteRepo exchange_out.QuoteRepository,
	eventPublisher exchange_out.ExchangeEventPublisher,
	resourceOwner shared.ResourceOwner,
) *GetQuoteUseCase {
	return &GetQuoteUseCase{
		pricingService: pricingService,
		feeService:     feeService,
		quoteRepo:      quoteRepo,
		eventPublisher: eventPublisher,
		resourceOwner:  resourceOwner,
		quoteTTL:       30 * time.Second,
	}
}

// Execute generates a locked price quote
func (uc *GetQuoteUseCase) Execute(ctx context.Context, query exchange_in.GetQuoteQuery) (*exchange_in.QuoteResult, error) {
	if err := query.Validate(); err != nil {
		return nil, fmt.Errorf("invalid quote query: %w", err)
	}

	// Get current price
	pricing, err := uc.pricingService.GetCurrentPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get price: %w", err)
	}

	// Calculate fee
	var amountUSD float64
	if query.Side == exchange_vo.OrderSideBuy {
		amountUSD = query.AmountUSD
	} else {
		amountUSD = query.AmountBTC * pricing.MedianPrice
	}

	feeResult, err := uc.feeService.CalculateFee(ctx, query.UserID, query.Side, amountUSD)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate fee: %w", err)
	}

	// Create quote entity
	var quote *exchange_entities.Quote
	if query.Side == exchange_vo.OrderSideBuy {
		amount := wallet_vo.NewAmount(query.AmountUSD)
		quote, err = exchange_entities.NewBuyQuote(
			uc.resourceOwner,
			query.UserID,
			amount,
			pricing.MedianPrice,
			feeResult.FeePercent,
			uc.quoteTTL,
			"median",
			pricing.Confidence,
		)
	} else {
		btcAmount := wallet_vo.NewBtcAmount(query.AmountBTC)
		quote, err = exchange_entities.NewSellQuote(
			uc.resourceOwner,
			query.UserID,
			btcAmount,
			pricing.MedianPrice,
			feeResult.FeePercent,
			uc.quoteTTL,
			"median",
			pricing.Confidence,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create quote: %w", err)
	}

	// Persist quote
	if err := uc.quoteRepo.Save(ctx, quote); err != nil {
		return nil, fmt.Errorf("failed to save quote: %w", err)
	}

	// Publish event
	if uc.eventPublisher != nil {
		go func() {
			_ = uc.eventPublisher.PublishQuoteCreated(context.Background(), quote)
		}()
	}

	return &exchange_in.QuoteResult{
		QuoteID:          quote.ID,
		Side:             string(query.Side),
		BTCPriceUSD:      quote.BTCPriceUSD,
		AmountUSD:        quote.AmountUSD.Dollars(),
		BTCAmount:        quote.BTCAmount.ToBTC(),
		FeePercent:       quote.FeePercent,
		FeeAmountUSD:     quote.FeeAmount.Dollars(),
		TotalCostUSD:     quote.TotalCostUSD.Dollars(),
		NetProceedsUSD:   quote.NetProceedsUSD.Dollars(),
		ExpiresAt:        quote.ExpiresAt.Format(time.RFC3339),
		RemainingSeconds: quote.RemainingSeconds(),
		PriceSource:      quote.PriceSource,
		Confidence:       quote.PriceConfidence,
	}, nil
}
