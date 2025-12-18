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
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
)

// WithdrawalController handles withdrawal HTTP requests
type WithdrawalController struct {
	container         container.Container
	withdrawalCommand billing_in.WithdrawalCommand
}

// NewWithdrawalController creates a new WithdrawalController
func NewWithdrawalController(c container.Container) *WithdrawalController {
	ctrl := &WithdrawalController{container: c}

	if err := c.Resolve(&ctrl.withdrawalCommand); err != nil {
		slog.Warn("WithdrawalCommand not available", "error", err)
	}

	return ctrl
}

// CreateWithdrawalRequest represents the request body for creating a withdrawal
type CreateWithdrawalRequest struct {
	WalletID    string                           `json:"wallet_id"`
	Amount      float64                          `json:"amount"`
	Currency    string                           `json:"currency"`
	Method      billing_entities.WithdrawalMethod `json:"method"`
	BankDetails billing_entities.BankDetails     `json:"bank_details"`
}

// CreateWithdrawalHandler handles POST /withdrawals
func (ctrl *WithdrawalController) CreateWithdrawalHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if ctrl.withdrawalCommand == nil {
			http.Error(w, `{"error":"withdrawal service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		// Check authentication
		authenticated, ok := ctx.Value(common.AuthenticatedKey).(bool)
		if !ok || !authenticated {
			http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
			return
		}

		resourceOwner := common.GetResourceOwner(ctx)
		if resourceOwner.UserID == uuid.Nil {
			http.Error(w, `{"error":"valid user required"}`, http.StatusUnauthorized)
			return
		}

		var req CreateWithdrawalRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}

		walletID, err := uuid.Parse(req.WalletID)
		if err != nil {
			http.Error(w, `{"error":"invalid wallet_id"}`, http.StatusBadRequest)
			return
		}

		cmd := billing_in.CreateWithdrawalCommand{
			UserID:      resourceOwner.UserID,
			WalletID:    walletID,
			Amount:      req.Amount,
			Currency:    req.Currency,
			Method:      req.Method,
			BankDetails: req.BankDetails,
		}

		withdrawal, err := ctrl.withdrawalCommand.Create(ctx, cmd)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to create withdrawal", "error", err)
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(withdrawal)
	}
}

// CancelWithdrawalHandler handles POST /withdrawals/{id}/cancel
func (ctrl *WithdrawalController) CancelWithdrawalHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if ctrl.withdrawalCommand == nil {
			http.Error(w, `{"error":"withdrawal service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		// Check authentication
		authenticated, ok := ctx.Value(common.AuthenticatedKey).(bool)
		if !ok || !authenticated {
			http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		withdrawalIDStr := vars["id"]

		withdrawalID, err := uuid.Parse(withdrawalIDStr)
		if err != nil {
			http.Error(w, `{"error":"invalid withdrawal ID"}`, http.StatusBadRequest)
			return
		}

		withdrawal, err := ctrl.withdrawalCommand.Cancel(ctx, withdrawalID)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to cancel withdrawal", "error", err)
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(withdrawal)
	}
}

// GetWithdrawalHandler handles GET /withdrawals/{id}
func (ctrl *WithdrawalController) GetWithdrawalHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if ctrl.withdrawalCommand == nil {
			http.Error(w, `{"error":"withdrawal service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		// Check authentication
		authenticated, ok := ctx.Value(common.AuthenticatedKey).(bool)
		if !ok || !authenticated {
			http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		withdrawalIDStr := vars["id"]

		withdrawalID, err := uuid.Parse(withdrawalIDStr)
		if err != nil {
			http.Error(w, `{"error":"invalid withdrawal ID"}`, http.StatusBadRequest)
			return
		}

		withdrawal, err := ctrl.withdrawalCommand.GetByID(ctx, withdrawalID)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get withdrawal", "error", err)
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(withdrawal)
	}
}

// ListWithdrawalsHandler handles GET /withdrawals
func (ctrl *WithdrawalController) ListWithdrawalsHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if ctrl.withdrawalCommand == nil {
			http.Error(w, `{"error":"withdrawal service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		// Check authentication
		authenticated, ok := ctx.Value(common.AuthenticatedKey).(bool)
		if !ok || !authenticated {
			http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
			return
		}

		resourceOwner := common.GetResourceOwner(ctx)
		if resourceOwner.UserID == uuid.Nil {
			http.Error(w, `{"error":"valid user required"}`, http.StatusUnauthorized)
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

		withdrawals, err := ctrl.withdrawalCommand.GetByUserID(ctx, resourceOwner.UserID, limit, offset)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to list withdrawals", "error", err)
			http.Error(w, `{"error":"failed to list withdrawals"}`, http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"withdrawals": withdrawals,
			"count":       len(withdrawals),
			"limit":       limit,
			"offset":      offset,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}
}

