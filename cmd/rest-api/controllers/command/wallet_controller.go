package cmd_controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	wallet_in "github.com/replay-api/replay-api/pkg/domain/wallet/ports/in"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// WalletCommandController handles wallet command operations (deposit, withdraw, etc.)
type WalletCommandController struct {
	walletCommand wallet_in.WalletCommand
}

// NewWalletCommandController creates a new wallet command controller
func NewWalletCommandController(c container.Container) *WalletCommandController {
	var walletCommand wallet_in.WalletCommand

	if err := c.Resolve(&walletCommand); err != nil {
		slog.Error("Failed to resolve WalletCommand", "error", err)
		panic(err)
	}

	return &WalletCommandController{
		walletCommand: walletCommand,
	}
}

// CreateWalletRequest represents a request to create a new wallet
type CreateWalletRequest struct {
	EVMAddress string `json:"evm_address"`
}

// DepositRequest represents a request to deposit funds
type DepositRequest struct {
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	TxHash         string  `json:"tx_hash"`
	ChainID        int     `json:"chain_id,omitempty"`
	PaymentMethod  string  `json:"payment_method,omitempty"`
	IdempotencyKey string  `json:"idempotency_key,omitempty"` // Optional client-side duplicate prevention
}

// WithdrawRequest represents a request to withdraw funds
type WithdrawRequest struct {
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	ToAddress      string  `json:"to_address"`
	ChainID        int     `json:"chain_id,omitempty"`
	PaymentMethod  string  `json:"payment_method,omitempty"`
	IdempotencyKey string  `json:"idempotency_key,omitempty"` // Optional client-side duplicate prevention
}

// DeductEntryFeeRequest represents a request to deduct entry fee
type DeductEntryFeeRequest struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// AddPrizeRequest represents a request to add prize winnings
type AddPrizeRequest struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// CreditWalletRequest represents a generic credit request
type CreditWalletRequest struct {
	Amount      float64                `json:"amount"`
	Currency    string                 `json:"currency"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// DebitWalletRequest represents a generic debit request
type DebitWalletRequest struct {
	Amount      float64                `json:"amount"`
	Currency    string                 `json:"currency"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// WalletResponse represents a wallet in API responses
type WalletResponse struct {
	WalletID   string `json:"wallet_id"`
	EVMAddress string `json:"evm_address"`
	CreatedAt  string `json:"created_at"`
}

// requireWalletAuth checks authentication and returns user ID
// Returns uuid.Nil and writes error if not authenticated
func requireWalletAuth(w http.ResponseWriter, r *http.Request) uuid.UUID {
	ctx := r.Context()

	// Check authentication
	authenticated, ok := ctx.Value(shared.AuthenticatedKey).(bool)
	if !ok || !authenticated {
		slog.WarnContext(ctx, "Wallet command attempted without authentication")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"success":false,"error":"Authentication required for wallet operations"}`))
		return uuid.Nil
	}

	// Get resource owner
	resourceOwner := shared.GetResourceOwner(ctx)
	if resourceOwner.UserID == uuid.Nil {
		slog.WarnContext(ctx, "Wallet command attempted without valid user ID")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"success":false,"error":"Valid user authentication required for wallet operations"}`))
		return uuid.Nil
	}

	return resourceOwner.UserID
}

// CreateWalletHandler handles POST /wallet
// Creates a new wallet for the authenticated user
func (c *WalletCommandController) CreateWalletHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := requireWalletAuth(w, r)
		if userID == uuid.Nil {
			return
		}

		var req CreateWalletRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", "")
			return
		}

		if req.EVMAddress == "" {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "EVM address is required", "")
			return
		}

		cmd := wallet_in.CreateWalletCommand{
			UserID:     userID,
			EVMAddress: req.EVMAddress,
		}

		if err := cmd.Validate(); err != nil {
			slog.ErrorContext(ctx, "CreateWalletHandler: validation error", "error", err)
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid wallet parameters", "")
			return
		}

		wallet, err := c.walletCommand.CreateWallet(ctx, cmd)
		if err != nil {
			slog.ErrorContext(ctx, "CreateWalletHandler: error creating wallet", "error", err)
			common.WriteErrorFromDomainError(w, err)
			return
		}

		response := WalletResponse{
			WalletID:   wallet.ID.String(),
			EVMAddress: wallet.EVMAddress.String(),
			CreatedAt:  wallet.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    response,
		})
	}
}

// DepositHandler handles POST /wallet/deposit
// Records a deposit to the user's wallet
func (c *WalletCommandController) DepositHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := requireWalletAuth(w, r)
		if userID == uuid.Nil {
			return
		}

		var req DepositRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", "")
			return
		}

		cmd := wallet_in.DepositCommand{
			UserID:        userID,
			Amount:        req.Amount,
			Currency:      req.Currency,
			TxHash:        req.TxHash,
			ChainID:       req.ChainID,
			PaymentMethod: req.PaymentMethod,
			SourceIP:      extractClientIP(r),
			UserAgent:     r.UserAgent(),
		}

		if err := cmd.Validate(); err != nil {
			slog.ErrorContext(ctx, "DepositHandler: validation error", "error", err)
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid deposit parameters", "")
			return
		}

		if err := c.walletCommand.Deposit(ctx, cmd); err != nil {
			slog.ErrorContext(ctx, "DepositHandler: error processing deposit", "error", err)
			common.WriteErrorFromDomainError(w, err)
			return
		}

		common.WriteSuccess(w, map[string]interface{}{
			"message": "Deposit processed successfully",
		})
	}
}

// WithdrawHandler handles POST /wallet/withdraw
// Initiates a withdrawal from the user's wallet
func (c *WalletCommandController) WithdrawHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := requireWalletAuth(w, r)
		if userID == uuid.Nil {
			return
		}

		var req WithdrawRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", "")
			return
		}

		cmd := wallet_in.WithdrawCommand{
			UserID:        userID,
			Amount:        req.Amount,
			Currency:      req.Currency,
			ToAddress:     req.ToAddress,
			ChainID:       req.ChainID,
			PaymentMethod: req.PaymentMethod,
			SourceIP:      extractClientIP(r),
			UserAgent:     r.UserAgent(),
		}

		if err := cmd.Validate(); err != nil {
			slog.ErrorContext(ctx, "WithdrawHandler: validation error", "error", err)
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid withdrawal parameters", "")
			return
		}

		if err := c.walletCommand.Withdraw(ctx, cmd); err != nil {
			slog.ErrorContext(ctx, "WithdrawHandler: error processing withdrawal", "error", err)
			common.WriteErrorFromDomainError(w, err)
			return
		}

		common.WriteSuccess(w, map[string]interface{}{
			"message": "Withdrawal initiated successfully",
		})
	}
}

// DeductEntryFeeHandler handles POST /wallet/entry-fee
// Deducts an entry fee from the user's wallet
func (c *WalletCommandController) DeductEntryFeeHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := requireWalletAuth(w, r)
		if userID == uuid.Nil {
			return
		}

		var req DeductEntryFeeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", "")
			return
		}

		cmd := wallet_in.DeductEntryFeeCommand{
			UserID:   userID,
			Amount:   req.Amount,
			Currency: req.Currency,
		}

		if err := cmd.Validate(); err != nil {
			slog.ErrorContext(ctx, "DeductEntryFeeHandler: validation error", "error", err)
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid entry fee parameters", "")
			return
		}

		if err := c.walletCommand.DeductEntryFee(ctx, cmd); err != nil {
			slog.ErrorContext(ctx, "DeductEntryFeeHandler: error deducting entry fee", "error", err)
			common.WriteErrorFromDomainError(w, err)
			return
		}

		common.WriteSuccess(w, map[string]interface{}{
			"message": "Entry fee deducted successfully",
		})
	}
}

// AddPrizeHandler handles POST /wallet/prize
// Adds prize winnings to the user's wallet
func (c *WalletCommandController) AddPrizeHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := requireWalletAuth(w, r)
		if userID == uuid.Nil {
			return
		}

		var req AddPrizeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", "")
			return
		}

		cmd := wallet_in.AddPrizeCommand{
			UserID:   userID,
			Amount:   req.Amount,
			Currency: req.Currency,
		}

		if err := cmd.Validate(); err != nil {
			slog.ErrorContext(ctx, "AddPrizeHandler: validation error", "error", err)
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid prize parameters", "")
			return
		}

		if err := c.walletCommand.AddPrize(ctx, cmd); err != nil {
			slog.ErrorContext(ctx, "AddPrizeHandler: error adding prize", "error", err)
			common.WriteErrorFromDomainError(w, err)
			return
		}

		common.WriteSuccess(w, map[string]interface{}{
			"message": "Prize added successfully",
		})
	}
}

// CreditWalletHandler handles POST /wallet/credit
// Generic credit operation for the wallet
func (c *WalletCommandController) CreditWalletHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := requireWalletAuth(w, r)
		if userID == uuid.Nil {
			return
		}

		var req CreditWalletRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", "")
			return
		}

		amount := wallet_vo.NewAmount(req.Amount)
		cmd := wallet_in.CreditWalletCommand{
			UserID:      userID,
			Amount:      amount,
			Currency:    req.Currency,
			Description: req.Description,
			Metadata:    req.Metadata,
		}

		if err := cmd.Validate(); err != nil {
			slog.ErrorContext(ctx, "CreditWalletHandler: validation error", "error", err)
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid credit parameters", "")
			return
		}

		tx, err := c.walletCommand.CreditWallet(ctx, cmd)
		if err != nil {
			slog.ErrorContext(ctx, "CreditWalletHandler: error crediting wallet", "error", err)
			common.WriteErrorFromDomainError(w, err)
			return
		}

		common.WriteSuccess(w, map[string]interface{}{
			"message":        "Credit completed successfully",
			"transaction_id": tx.ID.String(),
		})
	}
}

// DebitWalletHandler handles POST /wallet/debit
// Generic debit operation for the wallet
func (c *WalletCommandController) DebitWalletHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := requireWalletAuth(w, r)
		if userID == uuid.Nil {
			return
		}

		var req DebitWalletRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", "")
			return
		}

		amount := wallet_vo.NewAmount(req.Amount)
		cmd := wallet_in.DebitWalletCommand{
			UserID:      userID,
			Amount:      amount,
			Currency:    req.Currency,
			Description: req.Description,
			Metadata:    req.Metadata,
		}

		if err := cmd.Validate(); err != nil {
			slog.ErrorContext(ctx, "DebitWalletHandler: validation error", "error", err)
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid debit parameters", "")
			return
		}

		tx, err := c.walletCommand.DebitWallet(ctx, cmd)
		if err != nil {
			slog.ErrorContext(ctx, "DebitWalletHandler: error debiting wallet", "error", err)
			common.WriteErrorFromDomainError(w, err)
			return
		}

		common.WriteSuccess(w, map[string]interface{}{
			"message":        "Debit completed successfully",
			"transaction_id": tx.ID.String(),
		})
	}
}

// extractClientIP extracts the client IP from an HTTP request,
// respecting X-Forwarded-For and X-Real-IP headers for proxied requests.
// Falls back to RemoteAddr if no proxy headers are present.
func extractClientIP(r *http.Request) string {
	// Check X-Forwarded-For first (most common for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs — take the first (client IP)
		if idx := len(xff); idx > 0 {
			for i, c := range xff {
				if c == ',' {
					return xff[:i]
				}
			}
			return xff
		}
	}

	// Check X-Real-IP (nginx proxy)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
