package cmd_controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// CheckoutController handles checkout flow HTTP requests (payment → subscription activation)
type CheckoutController struct {
	checkoutHandler billing_in.CheckoutSubscriptionCommandHandler
}

// NewCheckoutController creates a new checkout controller
func NewCheckoutController(c container.Container) *CheckoutController {
	ctrl := &CheckoutController{}

	if err := c.Resolve(&ctrl.checkoutHandler); err != nil {
		slog.Warn("CheckoutSubscriptionCommandHandler not registered", "error", err)
	}

	return ctrl
}

// CheckoutRequest represents the checkout request body
type CheckoutRequest struct {
	PlanID        string                 `json:"plan_id"`
	PaymentID     string                 `json:"payment_id,omitempty"`
	BillingPeriod string                 `json:"billing_period"` // "monthly", "quarterly", "yearly"
	Args          map[string]interface{} `json:"args,omitempty"`
}

// CheckoutResponse represents the checkout response
type CheckoutResponse struct {
	Success        bool   `json:"success"`
	SubscriptionID string `json:"subscription_id"`
	PlanID         string `json:"plan_id"`
	Status         string `json:"status"`
	Message        string `json:"message"`
}

// CheckoutHandler handles POST /checkout
func (ctrl *CheckoutController) CheckoutHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if ctrl.checkoutHandler == nil {
			slog.ErrorContext(r.Context(), "Checkout service not available")
			http.Error(w, `{"error": "checkout service not available"}`, http.StatusServiceUnavailable)
			return
		}

		// Get user ID from context (set by auth middleware)
		userID, ok := r.Context().Value(shared.UserIDKey).(uuid.UUID)
		if !ok || userID == uuid.Nil {
			http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
			return
		}

		var req CheckoutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
			return
		}

		planID, err := uuid.Parse(req.PlanID)
		if err != nil {
			http.Error(w, `{"error": "invalid plan_id"}`, http.StatusBadRequest)
			return
		}

		var paymentID uuid.UUID
		if req.PaymentID != "" {
			paymentID, err = uuid.Parse(req.PaymentID)
			if err != nil {
				http.Error(w, `{"error": "invalid payment_id"}`, http.StatusBadRequest)
				return
			}
		}

		// Validate billing period
		billingPeriod := req.BillingPeriod
		if billingPeriod == "" {
			billingPeriod = "monthly"
		}
		switch billingPeriod {
		case "monthly", "quarterly", "yearly":
			// valid
		default:
			http.Error(w, `{"error": "billing_period must be monthly, quarterly, or yearly"}`, http.StatusBadRequest)
			return
		}

		cmd := billing_in.CheckoutSubscriptionCommand{
			PlanID:        planID,
			PaymentID:     paymentID,
			BillingPeriod: billingPeriod,
			Args:          req.Args,
		}

		subscription, err := ctrl.checkoutHandler.Exec(r.Context(), cmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "Checkout failed", "error", err)

			// Determine appropriate status code
			if shared.IsUnauthorizedError(err) {
				http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
				return
			}

			http.Error(w, `{"error": "Failed to process checkout"}`, http.StatusBadRequest)
			return
		}

		response := CheckoutResponse{
			Success:        true,
			SubscriptionID: subscription.ID.String(),
			PlanID:         subscription.PlanID.String(),
			Status:         string(subscription.Status),
			Message:        "Checkout completed successfully",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}
