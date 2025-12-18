package cmd_controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
)

// SubscriptionController handles subscription management endpoints
type SubscriptionController struct {
	upgradeHandler   billing_in.UpgradeSubscriptionCommandHandler
	downgradeHandler billing_in.DowngradeSubscriptionCommandHandler
}

// NewSubscriptionController creates a new SubscriptionController
func NewSubscriptionController(c container.Container) *SubscriptionController {
	ctrl := &SubscriptionController{}

	if err := c.Resolve(&ctrl.upgradeHandler); err != nil {
		slog.Warn("UpgradeSubscriptionCommandHandler not registered", "error", err)
	}

	if err := c.Resolve(&ctrl.downgradeHandler); err != nil {
		slog.Warn("DowngradeSubscriptionCommandHandler not registered", "error", err)
	}

	return ctrl
}

// UpgradeSubscriptionRequest represents the upgrade request body
type UpgradeSubscriptionRequest struct {
	PlanID         string                 `json:"plan_id"`
	BillingPeriod  string                 `json:"billing_period,omitempty"`
	PaymentMethod  string                 `json:"payment_method,omitempty"`
	Args           map[string]interface{} `json:"args,omitempty"`
}

// DowngradeSubscriptionRequest represents the downgrade request body
type DowngradeSubscriptionRequest struct {
	PlanID string                 `json:"plan_id"`
	Args   map[string]interface{} `json:"args,omitempty"`
}

// UpgradeSubscriptionHandler handles POST /subscriptions/upgrade
func (ctrl *SubscriptionController) UpgradeSubscriptionHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ctrl.upgradeHandler == nil {
			http.Error(w, "subscription service not available", http.StatusServiceUnavailable)
			return
		}

		var req UpgradeSubscriptionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		planID, err := uuid.Parse(req.PlanID)
		if err != nil {
			http.Error(w, "invalid plan_id", http.StatusBadRequest)
			return
		}

		cmd := billing_in.UpgradeSubscriptionCommand{
			PlanID: planID,
			Args:   req.Args,
		}

		if err := ctrl.upgradeHandler.Exec(r.Context(), cmd); err != nil {
			slog.ErrorContext(r.Context(), "Failed to upgrade subscription", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Subscription upgraded successfully",
		})
	}
}

// DowngradeSubscriptionHandler handles POST /subscriptions/downgrade
func (ctrl *SubscriptionController) DowngradeSubscriptionHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ctrl.downgradeHandler == nil {
			http.Error(w, "subscription service not available", http.StatusServiceUnavailable)
			return
		}

		var req DowngradeSubscriptionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		planID, err := uuid.Parse(req.PlanID)
		if err != nil {
			http.Error(w, "invalid plan_id", http.StatusBadRequest)
			return
		}

		cmd := billing_in.DowngradeSubscriptionCommand{
			PlanID: planID,
			Args:   req.Args,
		}

		if err := ctrl.downgradeHandler.Exec(r.Context(), cmd); err != nil {
			slog.ErrorContext(r.Context(), "Failed to downgrade subscription", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Subscription downgrade processed",
		})
	}
}

