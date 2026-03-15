package exchange_services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	exchange_entities "github.com/replay-api/replay-api/pkg/domain/exchange/entities"
	exchange_in "github.com/replay-api/replay-api/pkg/domain/exchange/ports/in"
	exchange_out "github.com/replay-api/replay-api/pkg/domain/exchange/ports/out"
	exchange_vo "github.com/replay-api/replay-api/pkg/domain/exchange/value-objects"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// WalletOperations defines the wallet operations needed by the order service
type WalletOperations interface {
	DepositBTC(ctx context.Context, userID uuid.UUID, amount wallet_vo.BtcAmount, reference string) error
	WithdrawBTC(ctx context.Context, userID uuid.UUID, amount wallet_vo.BtcAmount, reference string) error
	CreditUSD(ctx context.Context, userID uuid.UUID, amount wallet_vo.Amount, reference string) error
}

// StripeOperations defines the Stripe operations needed by the order service
type StripeOperations interface {
	CreatePaymentIntent(ctx context.Context, amountCents int64, currency string, metadata map[string]string) (paymentIntentID string, clientSecret string, err error)
	RefundPayment(ctx context.Context, paymentIntentID string) error
}

// OrderService orchestrates the buy/sell Bitcoin flow
type OrderService struct {
	orderRepo      exchange_out.OrderRepository
	quoteRepo      exchange_out.QuoteRepository
	smartRouter    *SmartRouter
	pricingService *PricingService
	feeService     *FeeService
	walletOps      WalletOperations
	stripeOps      StripeOperations
	eventPublisher exchange_out.ExchangeEventPublisher
	resourceOwner  shared.ResourceOwner
}

// NewOrderService creates a new order service
func NewOrderService(
	orderRepo exchange_out.OrderRepository,
	quoteRepo exchange_out.QuoteRepository,
	smartRouter *SmartRouter,
	pricingService *PricingService,
	feeService *FeeService,
	walletOps WalletOperations,
	stripeOps StripeOperations,
	eventPublisher exchange_out.ExchangeEventPublisher,
	resourceOwner shared.ResourceOwner,
) *OrderService {
	return &OrderService{
		orderRepo:      orderRepo,
		quoteRepo:      quoteRepo,
		smartRouter:    smartRouter,
		pricingService: pricingService,
		feeService:     feeService,
		walletOps:      walletOps,
		stripeOps:      stripeOps,
		eventPublisher: eventPublisher,
		resourceOwner:  resourceOwner,
	}
}

// BuyBitcoin initiates a Bitcoin buy order
// Flow: Validate -> Get Quote/Price -> Calculate Fee -> Create Stripe PaymentIntent -> Create Order -> Publish Event
func (s *OrderService) BuyBitcoin(ctx context.Context, cmd exchange_in.BuyBitcoinCommand) (*exchange_in.BuyBitcoinResult, error) {
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid buy command: %w", err)
	}

	// Check idempotency
	existing, err := s.orderRepo.FindByIdempotencyKey(ctx, cmd.IdempotencyKey)
	if err == nil && existing != nil {
		return &exchange_in.BuyBitcoinResult{
			OrderID:               existing.ID,
			Status:                string(existing.Status),
			AmountUSD:             existing.RequestedAmountUSD.Dollars(),
			EstimatedBTC:          existing.ExecutedAmountBTC.ToBTC(),
			FeeUSD:                existing.FeeAmountUSD.Dollars(),
			FeePercent:            existing.FeePercent,
			StripePaymentIntentID: existing.StripePaymentIntentID,
		}, nil
	}

	// Rate limiting: max 10 orders per hour per user
	hourAgo := time.Now().Add(-1 * time.Hour)
	orderCount, _ := s.orderRepo.CountByUserIDSince(ctx, cmd.UserID, hourAgo)
	if orderCount >= 10 {
		return nil, fmt.Errorf("rate limit exceeded: maximum 10 orders per hour")
	}

	// Get price (from quote or live)
	var btcPrice float64
	var quoteID *uuid.UUID

	if cmd.QuoteID != nil {
		quote, err := s.quoteRepo.FindByID(ctx, *cmd.QuoteID)
		if err != nil {
			return nil, fmt.Errorf("quote not found: %w", err)
		}
		if !quote.IsUsable() {
			return nil, fmt.Errorf("quote %s is expired or already consumed", cmd.QuoteID)
		}
		if quote.UserID != cmd.UserID {
			return nil, fmt.Errorf("quote does not belong to this user")
		}
		btcPrice = quote.BTCPriceUSD
		quoteID = cmd.QuoteID
		if err := quote.MarkConsumed(); err != nil {
			return nil, err
		}
		if err := s.quoteRepo.MarkConsumed(ctx, quote.ID); err != nil {
			log.Printf("[OrderService] Failed to mark quote consumed: %v", err)
		}
	} else {
		pricing, err := s.pricingService.GetCurrentPrice(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get BTC price: %w", err)
		}
		btcPrice = pricing.MedianPrice
	}

	// Calculate fee
	feeResult, err := s.feeService.CalculateFee(ctx, cmd.UserID, exchange_vo.OrderSideBuy, cmd.AmountUSD)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate fee: %w", err)
	}

	amountUSD := wallet_vo.NewAmount(cmd.AmountUSD)
	feeAmountUSD := wallet_vo.NewAmount(feeResult.FeeAmount)
	netForExchange := cmd.AmountUSD - feeResult.FeeAmount
	estimatedBTC := netForExchange / btcPrice

	// Create order entity
	order, err := exchange_entities.NewBuyOrder(
		s.resourceOwner,
		cmd.UserID,
		cmd.WalletID,
		amountUSD,
		feeResult.FeePercent,
		feeAmountUSD,
		cmd.IdempotencyKey,
		quoteID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Create Stripe PaymentIntent
	metadata := map[string]string{
		"order_id":   order.ID.String(),
		"user_id":    cmd.UserID.String(),
		"order_type": "btc_buy",
		"btc_price":  fmt.Sprintf("%.2f", btcPrice),
	}
	// SECURITY: Use wallet_vo.Amount for safe float→cents conversion (avoids truncation bugs)
	amountCents := amountUSD.Cents()
	paymentIntentID, clientSecret, err := s.stripeOps.CreatePaymentIntent(ctx, amountCents, "usd", metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe payment: %w", err)
	}

	order.StripePaymentIntentID = paymentIntentID

	// Save order
	if err := s.orderRepo.Save(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	// Publish event
	if s.eventPublisher != nil {
		go func() {
			if err := s.eventPublisher.PublishOrderCreated(context.Background(), order); err != nil {
				log.Printf("[OrderService] Failed to publish order created event: %v", err)
			}
		}()
	}

	return &exchange_in.BuyBitcoinResult{
		OrderID:               order.ID,
		Status:                string(order.Status),
		AmountUSD:             cmd.AmountUSD,
		EstimatedBTC:          estimatedBTC,
		FeeUSD:                feeResult.FeeAmount,
		FeePercent:            feeResult.FeePercent,
		StripeClientSecret:    clientSecret,
		StripePaymentIntentID: paymentIntentID,
	}, nil
}

// SellBitcoin initiates a Bitcoin sell order
// Flow: Validate -> Get Price -> Calculate Fee -> Debit BTC from wallet -> Route to Exchange -> Credit USD
func (s *OrderService) SellBitcoin(ctx context.Context, cmd exchange_in.SellBitcoinCommand) (*exchange_in.SellBitcoinResult, error) {
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid sell command: %w", err)
	}

	// Check idempotency
	existing, err := s.orderRepo.FindByIdempotencyKey(ctx, cmd.IdempotencyKey)
	if err == nil && existing != nil {
		return &exchange_in.SellBitcoinResult{
			OrderID:      existing.ID,
			Status:       string(existing.Status),
			AmountBTC:    existing.RequestedAmountBTC.ToBTC(),
			EstimatedUSD: existing.NetAmountUSD.Dollars(),
			FeeUSD:       existing.FeeAmountUSD.Dollars(),
			FeePercent:   existing.FeePercent,
		}, nil
	}

	// Get price
	var btcPrice float64
	var quoteID *uuid.UUID

	if cmd.QuoteID != nil {
		quote, err := s.quoteRepo.FindByID(ctx, *cmd.QuoteID)
		if err != nil {
			return nil, fmt.Errorf("quote not found: %w", err)
		}
		if !quote.IsUsable() {
			return nil, fmt.Errorf("quote is expired or consumed")
		}
		btcPrice = quote.BTCPriceUSD
		quoteID = cmd.QuoteID
		_ = quote.MarkConsumed()
		_ = s.quoteRepo.MarkConsumed(ctx, quote.ID)
	} else {
		pricing, err := s.pricingService.GetCurrentPrice(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get BTC price: %w", err)
		}
		btcPrice = pricing.MedianPrice
	}

	amountBTC := wallet_vo.NewBtcAmount(cmd.AmountBTC)
	grossUSD := amountBTC.ToUSD(btcPrice)

	// Calculate fee
	feeResult, err := s.feeService.CalculateFee(ctx, cmd.UserID, exchange_vo.OrderSideSell, grossUSD.Dollars())
	if err != nil {
		return nil, fmt.Errorf("failed to calculate fee: %w", err)
	}

	feeAmountUSD := wallet_vo.NewAmount(feeResult.FeeAmount)

	// Create order
	order, err := exchange_entities.NewSellOrder(
		s.resourceOwner,
		cmd.UserID,
		cmd.WalletID,
		amountBTC,
		grossUSD,
		feeResult.FeePercent,
		feeAmountUSD,
		cmd.IdempotencyKey,
		quoteID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create sell order: %w", err)
	}

	// Debit BTC from user wallet (lock funds)
	if err := s.walletOps.WithdrawBTC(ctx, cmd.UserID, amountBTC, order.ID.String()); err != nil {
		return nil, fmt.Errorf("insufficient BTC balance: %w", err)
	}

	// Save order
	if err := s.orderRepo.Save(ctx, order); err != nil {
		// Rollback: re-credit BTC
		_ = s.walletOps.DepositBTC(ctx, cmd.UserID, amountBTC, fmt.Sprintf("rollback-%s", order.ID.String()))
		return nil, fmt.Errorf("failed to save sell order: %w", err)
	}

	// Publish event for async exchange execution
	if s.eventPublisher != nil {
		go func() {
			_ = s.eventPublisher.PublishOrderCreated(context.Background(), order)
		}()
	}

	return &exchange_in.SellBitcoinResult{
		OrderID:        order.ID,
		Status:         string(order.Status),
		AmountBTC:      cmd.AmountBTC,
		EstimatedUSD:   grossUSD.Dollars(),
		FeeUSD:         feeResult.FeeAmount,
		FeePercent:     feeResult.FeePercent,
		NetProceedsUSD: feeResult.NetAmount,
	}, nil
}

// CancelOrder cancels a pending order
func (s *OrderService) CancelOrder(ctx context.Context, cmd exchange_in.CancelOrderCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("invalid cancel command: %w", err)
	}

	order, err := s.orderRepo.FindByID(ctx, cmd.OrderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	if order.UserID != cmd.UserID {
		return fmt.Errorf("order does not belong to this user")
	}

	if err := order.MarkCancelled(); err != nil {
		return err
	}

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	// Refund if needed
	if order.NeedsRefund() && s.stripeOps != nil {
		go func() {
			if err := s.stripeOps.RefundPayment(context.Background(), order.StripePaymentIntentID); err != nil {
				log.Printf("[OrderService] Failed to refund Stripe payment %s: %v", order.StripePaymentIntentID, err)
			}
		}()
	}

	// Re-credit BTC for sell orders
	if order.Side == exchange_vo.OrderSideSell {
		go func() {
			_ = s.walletOps.DepositBTC(context.Background(), order.UserID, order.RequestedAmountBTC, fmt.Sprintf("cancel-refund-%s", order.ID))
		}()
	}

	if s.eventPublisher != nil {
		go func() {
			_ = s.eventPublisher.PublishOrderCancelled(context.Background(), order)
		}()
	}

	return nil
}

// ExecuteOrderOnExchange executes an order on the best exchange (called by consumer)
func (s *OrderService) ExecuteOrderOnExchange(ctx context.Context, orderID uuid.UUID) error {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	if order.Status.IsTerminal() {
		return fmt.Errorf("order %s is already in terminal state: %s", orderID, order.Status)
	}

	// Route to best exchange
	adapter, err := s.smartRouter.RouteOrder(ctx, order.Side, order.Pair)
	if err != nil {
		_ = order.MarkFailed(fmt.Sprintf("routing failed: %v", err))
		_ = s.orderRepo.Update(ctx, order)
		return err
	}

	// Mark executing
	_ = order.MarkExecuting(adapter.GetProvider(), "")
	_ = s.orderRepo.Update(ctx, order)

	if s.eventPublisher != nil {
		_ = s.eventPublisher.PublishOrderExecuting(ctx, order)
	}

	// Execute on exchange
	var result *exchange_out.ExchangeOrderResult
	if order.Side == exchange_vo.OrderSideBuy {
		result, err = adapter.PlaceMarketBuyOrder(ctx, order.Pair, order.NetAmountUSD.Dollars())
	} else {
		result, err = adapter.PlaceMarketSellOrder(ctx, order.Pair, order.RequestedAmountBTC.ToBTC())
	}

	if err != nil {
		_ = order.MarkFailed(fmt.Sprintf("exchange error: %v", err))
		_ = s.orderRepo.Update(ctx, order)
		if s.eventPublisher != nil {
			_ = s.eventPublisher.PublishOrderFailed(ctx, order)
		}
		return err
	}

	order.ExchangeOrderID = result.OrderID

	// Mark filled
	executedBTC := wallet_vo.NewBtcAmount(result.FilledQtyBTC)
	if err := order.MarkFilled(executedBTC, result.AvgPriceUSD); err != nil {
		return err
	}
	_ = s.orderRepo.Update(ctx, order)

	// Settle: credit user wallet
	if order.Side == exchange_vo.OrderSideBuy {
		if err := s.walletOps.DepositBTC(ctx, order.UserID, executedBTC, order.ID.String()); err != nil {
			log.Printf("[OrderService] CRITICAL: Failed to credit BTC to user %s for order %s: %v", order.UserID, order.ID, err)
			_ = order.MarkFailed(fmt.Sprintf("settlement failed: %v", err))
			_ = s.orderRepo.Update(ctx, order)
			return err
		}
	} else {
		// Sell: credit USD
		netUSD := wallet_vo.NewAmount(result.FilledQtyBTC * result.AvgPriceUSD)
		feeUSD := wallet_vo.NewAmount(netUSD.Dollars() * (order.FeePercent / 100.0))
		creditUSD := netUSD.Subtract(feeUSD)
		order.FeeAmountUSD = feeUSD
		order.NetAmountUSD = creditUSD

		if err := s.walletOps.CreditUSD(ctx, order.UserID, creditUSD, order.ID.String()); err != nil {
			log.Printf("[OrderService] CRITICAL: Failed to credit USD to user %s for order %s: %v", order.UserID, order.ID, err)
			_ = order.MarkFailed(fmt.Sprintf("settlement failed: %v", err))
			_ = s.orderRepo.Update(ctx, order)
			return err
		}
	}

	// Mark completed
	_ = order.MarkCompleted()
	_ = s.orderRepo.Update(ctx, order)

	if s.eventPublisher != nil {
		_ = s.eventPublisher.PublishOrderFilled(ctx, order)
	}

	log.Printf("[OrderService] Order %s completed: %s %.8f BTC @ $%.2f", order.ID, order.Side, executedBTC.ToBTC(), result.AvgPriceUSD)
	return nil
}
