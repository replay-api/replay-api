package cmd_controllers

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	shared "github.com/resource-ownership/go-common/pkg/common"
	payment_entities "github.com/replay-api/replay-api/pkg/domain/payment/entities"
	payment_in "github.com/replay-api/replay-api/pkg/domain/payment/ports/in"
	payment_out "github.com/replay-api/replay-api/pkg/domain/payment/ports/out"
)

// PaymentController handles payment-related HTTP requests
type PaymentController struct {
	container       container.Container
	paymentService  payment_in.PaymentCommand
	paymentQuery    payment_in.PaymentQuery
	paymentRepo     payment_out.PaymentRepository // TODO: Remove once all handlers use use cases
}

// NewPaymentController creates a new payment controller
func NewPaymentController(container container.Container) *PaymentController {
	ctrl := &PaymentController{container: container}

	// Resolve dependencies
	if err := container.Resolve(&ctrl.paymentService); err != nil {
		slog.Error("Failed to resolve PaymentCommand", "err", err)
	}
	if err := container.Resolve(&ctrl.paymentQuery); err != nil {
		slog.Error("Failed to resolve PaymentQuery", "err", err)
	}
	if err := container.Resolve(&ctrl.paymentRepo); err != nil {
		slog.Error("Failed to resolve PaymentRepository", "err", err)
	}

	return ctrl
}

// CreatePaymentIntentRequest represents a request to create a payment intent
type CreatePaymentIntentRequest struct {
	WalletID    string                            `json:"wallet_id"`
	Amount      int64                             `json:"amount"`
	Currency    string                            `json:"currency"`
	PaymentType payment_entities.PaymentType      `json:"payment_type"`
	Provider    payment_entities.PaymentProvider  `json:"provider"`
	Metadata    map[string]any                    `json:"metadata,omitempty"`
}

// CreatePaymentIntentResponse represents the response from creating a payment intent
type CreatePaymentIntentResponse struct {
	PaymentID    string `json:"payment_id"`
	ClientSecret string `json:"client_secret,omitempty"`
	RedirectURL  string `json:"redirect_url,omitempty"`
	CryptoAddress string `json:"crypto_address,omitempty"`
	Status       string `json:"status"`
}

// PaymentResponse represents a payment in API responses
type PaymentResponse struct {
	ID                string                            `json:"id"`
	UserID            string                            `json:"user_id"`
	WalletID          string                            `json:"wallet_id"`
	Type              payment_entities.PaymentType      `json:"type"`
	Provider          payment_entities.PaymentProvider  `json:"provider"`
	Status            payment_entities.PaymentStatus    `json:"status"`
	Amount            int64                             `json:"amount"`
	Currency          string                            `json:"currency"`
	Fee               int64                             `json:"fee"`
	NetAmount         int64                             `json:"net_amount"`
	ProviderPaymentID string                            `json:"provider_payment_id,omitempty"`
	CreatedAt         string                            `json:"created_at"`
	UpdatedAt         string                            `json:"updated_at"`
	CompletedAt       *string                           `json:"completed_at,omitempty"`
	FailureReason     string                            `json:"failure_reason,omitempty"`
}

// CreatePaymentIntentHandler handles requests to create a payment intent
func (ctrl *PaymentController) CreatePaymentIntentHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Get user ID from context (set by auth middleware)
		userID, ok := r.Context().Value(shared.UserIDKey).(uuid.UUID)
		if !ok || userID == uuid.Nil {
			http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
			return
		}

		var req CreatePaymentIntentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Amount <= 0 {
			http.Error(w, `{"error": "amount must be positive"}`, http.StatusBadRequest)
			return
		}
		if req.Currency == "" {
			req.Currency = "usd"
		}
		if req.Provider == "" {
			req.Provider = payment_entities.PaymentProviderStripe
		}
		if req.PaymentType == "" {
			req.PaymentType = payment_entities.PaymentTypeDeposit
		}

		walletID, err := uuid.Parse(req.WalletID)
		if err != nil {
			http.Error(w, `{"error": "invalid wallet_id"}`, http.StatusBadRequest)
			return
		}

		// Create payment intent
		result, err := ctrl.paymentService.CreatePaymentIntent(r.Context(), payment_in.CreatePaymentIntentCommand{
			UserID:      userID,
			WalletID:    walletID,
			Amount:      req.Amount,
			Currency:    req.Currency,
			PaymentType: req.PaymentType,
			Provider:    req.Provider,
			Metadata:    req.Metadata,
		})

		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to create payment intent", "err", err)
			http.Error(w, `{"error": "failed to create payment intent"}`, http.StatusInternalServerError)
			return
		}

		response := CreatePaymentIntentResponse{
			PaymentID:    result.Payment.ID.String(),
			ClientSecret: result.ClientSecret,
			RedirectURL:  result.RedirectURL,
			CryptoAddress: result.CryptoAddress,
			Status:       string(result.Payment.Status),
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
	}
}

// GetPaymentHandler handles requests to get a specific payment
func (ctrl *PaymentController) GetPaymentHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		vars := mux.Vars(r)
		paymentIDStr := vars["payment_id"]

		paymentID, err := uuid.Parse(paymentIDStr)
		if err != nil {
			http.Error(w, `{"error": "invalid payment_id"}`, http.StatusBadRequest)
			return
		}

		// Get user ID from context (set by auth middleware)
		userID, ok := r.Context().Value(shared.UserIDKey).(uuid.UUID)
		if !ok || userID == uuid.Nil {
			http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// Use query handler with proper authentication context
		ctx := r.Context()
		ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)

		dto, err := ctrl.paymentQuery.GetPayment(ctx, payment_in.GetPaymentQuery{
			PaymentID: paymentID,
			UserID:    userID,
		})
		if err != nil {
			// Check if it's a not found or unauthorized error
			if shared.IsNotFoundError(err) {
				http.Error(w, `{"error": "payment not found"}`, http.StatusNotFound)
				return
			}
			if shared.IsUnauthorizedError(err) {
				http.Error(w, `{"error": "forbidden"}`, http.StatusForbidden)
				return
			}
			slog.ErrorContext(r.Context(), "Failed to get payment", "err", err)
			http.Error(w, `{"error": "failed to get payment"}`, http.StatusInternalServerError)
			return
		}

		response := dtoToResponse(dto)
		_ = json.NewEncoder(w).Encode(response)
	}
}

// GetUserPaymentsHandler handles requests to get all payments for the current user
func (ctrl *PaymentController) GetUserPaymentsHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		userID, ok := r.Context().Value(shared.UserIDKey).(uuid.UUID)
		if !ok || userID == uuid.Nil {
			http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// Use query handler with proper authentication context
		ctx := r.Context()
		ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)

		// Parse query params for filters
		filters := payment_in.PaymentQueryFilters{
			Limit:  50,
			Offset: 0,
		}

		result, err := ctrl.paymentQuery.GetUserPayments(ctx, payment_in.GetUserPaymentsQuery{
			UserID:  userID,
			Filters: filters,
		})
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to get user payments", "err", err)
			http.Error(w, `{"error": "failed to retrieve payments"}`, http.StatusInternalServerError)
			return
		}

		// Convert DTOs to response format
		response := make([]PaymentResponse, len(result.Payments))
		for i, dto := range result.Payments {
			response[i] = dtoToResponse(&dto)
		}

		_ = json.NewEncoder(w).Encode(response)
	}
}

// ConfirmPaymentRequest represents a request to confirm a payment
type ConfirmPaymentRequest struct {
	PaymentMethodID string `json:"payment_method_id"`
}

// ConfirmPaymentHandler handles requests to confirm a payment
func (ctrl *PaymentController) ConfirmPaymentHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		vars := mux.Vars(r)
		paymentIDStr := vars["payment_id"]

		paymentID, err := uuid.Parse(paymentIDStr)
		if err != nil {
			http.Error(w, `{"error": "invalid payment_id"}`, http.StatusBadRequest)
			return
		}

		// Verify user owns this payment before confirming
		existingPayment, err := ctrl.paymentRepo.FindByID(r.Context(), paymentID)
		if err != nil {
			http.Error(w, `{"error": "payment not found"}`, http.StatusNotFound)
			return
		}

		userID, ok := r.Context().Value(shared.UserIDKey).(uuid.UUID)
		if !ok || (userID != existingPayment.UserID && !shared.IsAdmin(r.Context())) {
			slog.WarnContext(r.Context(), "unauthorized payment confirm attempt",
				"payment_id", paymentID,
				"payment_owner", existingPayment.UserID,
				"requesting_user", userID,
			)
			http.Error(w, `{"error": "forbidden"}`, http.StatusForbidden)
			return
		}

		var req ConfirmPaymentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
			return
		}

		payment, err := ctrl.paymentService.ConfirmPayment(r.Context(), payment_in.ConfirmPaymentCommand{
			PaymentID:       paymentID,
			PaymentMethodID: req.PaymentMethodID,
		})

		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to confirm payment", "err", err)
			http.Error(w, `{"error": "failed to confirm payment"}`, http.StatusInternalServerError)
			return
		}

		response := paymentToResponse(payment)
		_ = json.NewEncoder(w).Encode(response)
	}
}

// CancelPaymentHandler handles requests to cancel a payment
func (ctrl *PaymentController) CancelPaymentHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		vars := mux.Vars(r)
		paymentIDStr := vars["payment_id"]

		paymentID, err := uuid.Parse(paymentIDStr)
		if err != nil {
			http.Error(w, `{"error": "invalid payment_id"}`, http.StatusBadRequest)
			return
		}

		// Verify user owns this payment before canceling
		existingPayment, err := ctrl.paymentRepo.FindByID(r.Context(), paymentID)
		if err != nil {
			http.Error(w, `{"error": "payment not found"}`, http.StatusNotFound)
			return
		}

		userID, ok := r.Context().Value(shared.UserIDKey).(uuid.UUID)
		if !ok || (userID != existingPayment.UserID && !shared.IsAdmin(r.Context())) {
			slog.WarnContext(r.Context(), "unauthorized payment cancel attempt",
				"payment_id", paymentID,
				"payment_owner", existingPayment.UserID,
				"requesting_user", userID,
			)
			http.Error(w, `{"error": "forbidden"}`, http.StatusForbidden)
			return
		}

		payment, err := ctrl.paymentService.CancelPayment(r.Context(), payment_in.CancelPaymentCommand{
			PaymentID: paymentID,
		})

		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to cancel payment", "err", err)
			http.Error(w, `{"error": "failed to cancel payment"}`, http.StatusInternalServerError)
			return
		}

		response := paymentToResponse(payment)
		_ = json.NewEncoder(w).Encode(response)
	}
}

// RefundPaymentRequest represents a request to refund a payment
type RefundPaymentRequest struct {
	Amount int64  `json:"amount,omitempty"`
	Reason string `json:"reason"`
}

// RefundPaymentHandler handles requests to refund a payment (admin only)
func (ctrl *PaymentController) RefundPaymentHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Check admin access
		if !shared.IsAdmin(r.Context()) {
			http.Error(w, `{"error": "admin access required"}`, http.StatusForbidden)
			return
		}

		vars := mux.Vars(r)
		paymentIDStr := vars["payment_id"]

		paymentID, err := uuid.Parse(paymentIDStr)
		if err != nil {
			http.Error(w, `{"error": "invalid payment_id"}`, http.StatusBadRequest)
			return
		}

		var req RefundPaymentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
			return
		}

		payment, err := ctrl.paymentService.RefundPayment(r.Context(), payment_in.RefundPaymentCommand{
			PaymentID: paymentID,
			Amount:    req.Amount,
			Reason:    req.Reason,
		})

		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to refund payment", "err", err)
			http.Error(w, `{"error": "failed to refund payment"}`, http.StatusInternalServerError)
			return
		}

		response := paymentToResponse(payment)
		_ = json.NewEncoder(w).Encode(response)
	}
}

// StripeWebhookHandler handles Stripe webhook events
func (ctrl *PaymentController) StripeWebhookHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const maxBodyBytes = int64(65536)
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

		payload, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("Failed to read webhook body", "err", err)
			http.Error(w, "Error reading body", http.StatusBadRequest)
			return
		}

		// Get Stripe signature header
		signature := r.Header.Get("Stripe-Signature")
		if signature == "" {
			http.Error(w, "Missing Stripe-Signature header", http.StatusBadRequest)
			return
		}

		// Process webhook
		err = ctrl.paymentService.ProcessWebhook(r.Context(), payment_in.ProcessWebhookCommand{
			Provider:  payment_entities.PaymentProviderStripe,
			Payload:   payload,
			Signature: signature,
		})

		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to process Stripe webhook", "err", err)
			// Return 200 to acknowledge receipt even on processing errors
			// Stripe will retry if we return an error
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// paymentToResponse converts a payment entity to an API response
func paymentToResponse(p *payment_entities.Payment) PaymentResponse {
	response := PaymentResponse{
		ID:                p.ID.String(),
		UserID:            p.UserID.String(),
		WalletID:          p.WalletID.String(),
		Type:              p.Type,
		Provider:          p.Provider,
		Status:            p.Status,
		Amount:            p.Amount,
		Currency:          p.Currency,
		Fee:               p.Fee,
		NetAmount:         p.NetAmount,
		ProviderPaymentID: p.ProviderPaymentID,
		CreatedAt:         p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:         p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		FailureReason:     p.FailureReason,
	}

	if p.CompletedAt != nil {
		completedAt := p.CompletedAt.Format("2006-01-02T15:04:05Z")
		response.CompletedAt = &completedAt
	}

	return response
}

// dtoToResponse converts a PaymentDTO to an API response
func dtoToResponse(dto *payment_in.PaymentDTO) PaymentResponse {
	response := PaymentResponse{
		ID:                dto.ID.String(),
		UserID:            dto.UserID.String(),
		WalletID:          dto.WalletID.String(),
		Type:              dto.Type,
		Provider:          dto.Provider,
		Status:            dto.Status,
		Amount:            dto.Amount,
		Currency:          dto.Currency,
		Fee:               dto.Fee,
		NetAmount:         dto.NetAmount,
		ProviderPaymentID: dto.ProviderPaymentID,
		CreatedAt:         dto.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:         dto.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		FailureReason:     dto.FailureReason,
	}

	if dto.CompletedAt != nil {
		completedAt := dto.CompletedAt.Format("2006-01-02T15:04:05Z")
		response.CompletedAt = &completedAt
	}

	return response
}
