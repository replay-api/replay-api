package cmd_controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	common "github.com/replay-api/replay-api/pkg/domain"
	exchange_in "github.com/replay-api/replay-api/pkg/domain/exchange/ports/in"
	exchange_out "github.com/replay-api/replay-api/pkg/domain/exchange/ports/out"
	exchange_services "github.com/replay-api/replay-api/pkg/domain/exchange/services"
	exchange_usecases "github.com/replay-api/replay-api/pkg/domain/exchange/usecases"
	exchange_vo "github.com/replay-api/replay-api/pkg/domain/exchange/value-objects"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// ExchangeController handles Bitcoin exchange HTTP requests (buy, sell, quotes, rates, orders, fees)
type ExchangeController struct {
	orderService     *exchange_services.OrderService
	getQuoteUC       *exchange_usecases.GetQuoteUseCase
	getExchangeRates *exchange_usecases.GetExchangeRatesUseCase
	feeService       *exchange_services.FeeService
	orderRepo        exchange_out.OrderRepository
}

// NewExchangeController creates a new exchange controller with dependencies from the DI container
func NewExchangeController(c container.Container) *ExchangeController {
	ctrl := &ExchangeController{}

	if err := c.Resolve(&ctrl.orderService); err != nil {
		slog.Warn("OrderService not available", "error", err)
	}

	if err := c.Resolve(&ctrl.getQuoteUC); err != nil {
		slog.Warn("GetQuoteUseCase not available", "error", err)
	}

	if err := c.Resolve(&ctrl.getExchangeRates); err != nil {
		slog.Warn("GetExchangeRatesUseCase not available", "error", err)
	}

	if err := c.Resolve(&ctrl.feeService); err != nil {
		slog.Warn("FeeService not available", "error", err)
	}

	if err := c.Resolve(&ctrl.orderRepo); err != nil {
		slog.Warn("OrderRepository not available", "error", err)
	}

	return ctrl
}

// requireExchangeAuth checks authentication and returns the resource owner's user ID.
// Returns uuid.Nil and writes an error response if not authenticated.
func requireExchangeAuth(w http.ResponseWriter, r *http.Request) uuid.UUID {
	ctx := r.Context()

	authenticated, ok := ctx.Value(shared.AuthenticatedKey).(bool)
	if !ok || !authenticated {
		slog.WarnContext(ctx, "Exchange endpoint accessed without authentication")
		common.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required for exchange operations", "")
		return uuid.Nil
	}

	resourceOwner := shared.GetResourceOwner(ctx)
	if resourceOwner.UserID == uuid.Nil {
		slog.WarnContext(ctx, "Exchange endpoint accessed without valid user ID")
		common.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Valid user authentication required for exchange operations", "")
		return uuid.Nil
	}

	return resourceOwner.UserID
}

// --- Request DTOs ---

// BuyBitcoinRequest represents the POST body for buying BTC
type BuyBitcoinRequest struct {
	WalletID            string  `json:"wallet_id"`
	AmountUSD           float64 `json:"amount_usd"`
	QuoteID             string  `json:"quote_id,omitempty"`
	StripePaymentMethod string  `json:"stripe_payment_method"`
	IdempotencyKey      string  `json:"idempotency_key"`
}

// SellBitcoinRequest represents the POST body for selling BTC
type SellBitcoinRequest struct {
	WalletID       string  `json:"wallet_id"`
	AmountBTC      float64 `json:"amount_btc"`
	QuoteID        string  `json:"quote_id,omitempty"`
	IdempotencyKey string  `json:"idempotency_key"`
}

// GetQuoteRequest represents the POST body for requesting a quote
type GetQuoteRequest struct {
	Side      string  `json:"side"`
	AmountUSD float64 `json:"amount_usd,omitempty"`
	AmountBTC float64 `json:"amount_btc,omitempty"`
}

// --- Handlers ---

// PostBuyBitcoin handles POST /v1/exchange/buy
func (ctrl *ExchangeController) PostBuyBitcoin(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if ctrl.orderService == nil {
			slog.ErrorContext(ctx, "OrderService not available")
			common.WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Exchange service not available", "")
			return
		}

		userID := requireExchangeAuth(w, r)
		if userID == uuid.Nil {
			return
		}

		var req BuyBitcoinRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", "")
			return
		}

		walletID, err := uuid.Parse(req.WalletID)
		if err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid wallet_id", "")
			return
		}

		var quoteID *uuid.UUID
		if req.QuoteID != "" {
			parsed, err := uuid.Parse(req.QuoteID)
			if err != nil {
				common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid quote_id", "")
				return
			}
			quoteID = &parsed
		}

		cmd := exchange_in.BuyBitcoinCommand{
			UserID:              userID,
			WalletID:            walletID,
			AmountUSD:           req.AmountUSD,
			QuoteID:             quoteID,
			StripePaymentMethod: req.StripePaymentMethod,
			IdempotencyKey:      req.IdempotencyKey,
		}

		result, err := ctrl.orderService.BuyBitcoin(ctx, cmd)
		if err != nil {
			slog.ErrorContext(ctx, "BuyBitcoin failed", "error", err, "user_id", userID)
			common.WriteErrorFromDomainError(w, err)
			return
		}

		common.WriteCreated(w, result)
	}
}

// PostSellBitcoin handles POST /v1/exchange/sell
func (ctrl *ExchangeController) PostSellBitcoin(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if ctrl.orderService == nil {
			slog.ErrorContext(ctx, "OrderService not available")
			common.WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Exchange service not available", "")
			return
		}

		userID := requireExchangeAuth(w, r)
		if userID == uuid.Nil {
			return
		}

		var req SellBitcoinRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", "")
			return
		}

		walletID, err := uuid.Parse(req.WalletID)
		if err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid wallet_id", "")
			return
		}

		var quoteID *uuid.UUID
		if req.QuoteID != "" {
			parsed, err := uuid.Parse(req.QuoteID)
			if err != nil {
				common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid quote_id", "")
				return
			}
			quoteID = &parsed
		}

		cmd := exchange_in.SellBitcoinCommand{
			UserID:         userID,
			WalletID:       walletID,
			AmountBTC:      req.AmountBTC,
			QuoteID:        quoteID,
			IdempotencyKey: req.IdempotencyKey,
		}

		result, err := ctrl.orderService.SellBitcoin(ctx, cmd)
		if err != nil {
			slog.ErrorContext(ctx, "SellBitcoin failed", "error", err, "user_id", userID)
			common.WriteErrorFromDomainError(w, err)
			return
		}

		common.WriteCreated(w, result)
	}
}

// GetQuote handles POST /v1/exchange/quote
func (ctrl *ExchangeController) GetQuote(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if ctrl.getQuoteUC == nil {
			slog.ErrorContext(ctx, "GetQuoteUseCase not available")
			common.WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Quote service not available", "")
			return
		}

		userID := requireExchangeAuth(w, r)
		if userID == uuid.Nil {
			return
		}

		var req GetQuoteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", "")
			return
		}

		side := exchange_vo.OrderSide(req.Side)
		if !side.IsValid() {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Side must be BUY or SELL", "")
			return
		}

		query := exchange_in.GetQuoteQuery{
			UserID:    userID,
			Side:      side,
			AmountUSD: req.AmountUSD,
			AmountBTC: req.AmountBTC,
		}

		result, err := ctrl.getQuoteUC.Execute(ctx, query)
		if err != nil {
			slog.ErrorContext(ctx, "GetQuote failed", "error", err, "user_id", userID)
			common.WriteErrorFromDomainError(w, err)
			return
		}

		common.WriteSuccess(w, result)
	}
}

// GetExchangeRates handles GET /v1/exchange/rates
func (ctrl *ExchangeController) GetExchangeRates(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if ctrl.getExchangeRates == nil {
			slog.ErrorContext(ctx, "GetExchangeRatesUseCase not available")
			common.WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Exchange rates service not available", "")
			return
		}

		result, err := ctrl.getExchangeRates.Execute(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "GetExchangeRates failed", "error", err)
			common.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get exchange rates", "")
			return
		}

		common.WriteSuccess(w, result)
	}
}

// GetOrderHistory handles GET /v1/exchange/orders
func (ctrl *ExchangeController) GetOrderHistory(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if ctrl.orderRepo == nil {
			slog.ErrorContext(ctx, "OrderRepository not available")
			common.WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Order service not available", "")
			return
		}

		userID := requireExchangeAuth(w, r)
		if userID == uuid.Nil {
			return
		}

		// Parse pagination
		limit := 20
		offset := 0

		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
				limit = parsed
			}
		}
		if o := r.URL.Query().Get("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
				offset = parsed
			}
		}

		orders, totalCount, err := ctrl.orderRepo.FindByUserID(ctx, userID, limit, offset)
		if err != nil {
			slog.ErrorContext(ctx, "GetOrderHistory failed", "error", err, "user_id", userID)
			common.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve order history", "")
			return
		}

		// Map to summary DTOs
		summaries := make([]exchange_in.OrderSummary, 0, len(orders))
		for _, order := range orders {
			summary := exchange_in.OrderSummary{
				OrderID:     order.ID,
				Side:        string(order.Side),
				Status:      string(order.Status),
				AmountUSD:   order.RequestedAmountUSD.Dollars(),
				AmountBTC:   order.RequestedAmountBTC.ToBTC(),
				FeeUSD:      order.FeeAmountUSD.Dollars(),
				BTCPriceUSD: order.ExecutedPriceUSD,
				CreatedAt:   order.CreatedAt.Format("2006-01-02T15:04:05Z"),
			}
			if order.SettledAt != nil {
				summary.CompletedAt = order.SettledAt.Format("2006-01-02T15:04:05Z")
			}
			summaries = append(summaries, summary)
		}

		result := exchange_in.OrderHistoryResult{
			Orders:     summaries,
			TotalCount: totalCount,
			Limit:      limit,
			Offset:     offset,
		}

		common.WriteSuccess(w, result)
	}
}

// GetOrderByID handles GET /v1/exchange/orders/{id}
func (ctrl *ExchangeController) GetOrderByID(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if ctrl.orderRepo == nil {
			slog.ErrorContext(ctx, "OrderRepository not available")
			common.WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Order service not available", "")
			return
		}

		userID := requireExchangeAuth(w, r)
		if userID == uuid.Nil {
			return
		}

		vars := mux.Vars(r)
		orderIDStr := vars["id"]

		orderID, err := uuid.Parse(orderIDStr)
		if err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid order ID", "")
			return
		}

		order, err := ctrl.orderRepo.FindByID(ctx, orderID)
		if err != nil {
			slog.ErrorContext(ctx, "GetOrderByID failed", "error", err, "order_id", orderIDStr)
			common.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Order not found", "")
			return
		}

		// Security: ensure order belongs to the authenticated user
		if order.UserID != userID {
			common.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Order not found", "")
			return
		}

		completedAt := ""
		if order.SettledAt != nil {
			completedAt = order.SettledAt.Format("2006-01-02T15:04:05Z")
		}

		result := exchange_in.OrderDetailResult{
			OrderSummary: exchange_in.OrderSummary{
				OrderID:     order.ID,
				Side:        string(order.Side),
				Status:      string(order.Status),
				AmountUSD:   order.RequestedAmountUSD.Dollars(),
				AmountBTC:   order.RequestedAmountBTC.ToBTC(),
				FeeUSD:      order.FeeAmountUSD.Dollars(),
				BTCPriceUSD: order.ExecutedPriceUSD,
				CreatedAt:   order.CreatedAt.Format("2006-01-02T15:04:05Z"),
				CompletedAt: completedAt,
			},
			ExchangeProvider:      string(order.ExchangeProvider),
			ExchangeOrderID:       order.ExchangeOrderID,
			StripePaymentIntentID: order.StripePaymentIntentID,
			FeePercent:            order.FeePercent,
			NetAmountUSD:          order.NetAmountUSD.Dollars(),
			FailureReason:         order.FailureReason,
			RetryCount:            order.RetryCount,
		}

		common.WriteSuccess(w, result)
	}
}

// PostCancelOrder handles POST /v1/exchange/orders/{id}/cancel
func (ctrl *ExchangeController) PostCancelOrder(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if ctrl.orderService == nil {
			slog.ErrorContext(ctx, "OrderService not available")
			common.WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Exchange service not available", "")
			return
		}

		userID := requireExchangeAuth(w, r)
		if userID == uuid.Nil {
			return
		}

		vars := mux.Vars(r)
		orderIDStr := vars["id"]

		orderID, err := uuid.Parse(orderIDStr)
		if err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid order ID", "")
			return
		}

		cmd := exchange_in.CancelOrderCommand{
			UserID:  userID,
			OrderID: orderID,
		}

		if err := ctrl.orderService.CancelOrder(ctx, cmd); err != nil {
			slog.ErrorContext(ctx, "CancelOrder failed", "error", err, "order_id", orderIDStr, "user_id", userID)
			common.WriteErrorFromDomainError(w, err)
			return
		}

		common.WriteSuccess(w, map[string]interface{}{
			"order_id": orderID,
			"status":   "cancelled",
			"message":  "Order cancelled successfully",
		})
	}
}

// GetFeeSchedule handles GET /v1/exchange/fees
func (ctrl *ExchangeController) GetFeeSchedule(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if ctrl.feeService == nil {
			slog.ErrorContext(ctx, "FeeService not available")
			common.WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Fee service not available", "")
			return
		}

		userID := requireExchangeAuth(w, r)
		if userID == uuid.Nil {
			return
		}

		// Get user's current tier
		tier := ctrl.feeService.GetUserTier(ctx, userID)
		config := exchange_vo.GetFeeConfig(tier)

		// Get all tiers for display
		allTiers := ctrl.feeService.GetFeeSchedule()
		tierMap := make(map[string]exchange_vo.FeeConfig, len(allTiers))
		for t, c := range allTiers {
			tierMap[string(t)] = c
		}

		result := exchange_in.FeeScheduleResult{
			PlanTier:       string(tier),
			BuyFeePercent:  config.BuyFeePercent,
			SellFeePercent: config.SellFeePercent,
			MinFeeUSD:      config.MinFeeUSD,
			MaxFeeUSD:      config.MaxFeeUSD,
			AllTiers:       tierMap,
		}

		common.WriteSuccess(w, result)
	}
}
