package query_controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	wallet_in "github.com/replay-api/replay-api/pkg/domain/wallet/ports/in"
)

type WalletQueryController struct {
	walletQuery wallet_in.WalletQuery
}

func NewWalletQueryController(c container.Container) *WalletQueryController {
	var walletQuery wallet_in.WalletQuery

	if err := c.Resolve(&walletQuery); err != nil {
		slog.Error("Failed to resolve WalletQuery", "error", err)
		panic(err)
	}

	return &WalletQueryController{
		walletQuery: walletQuery,
	}
}

// requireAuthentication checks if the user is authenticated and returns their resource owner
// Returns nil and writes an error response if not authenticated
func requireAuthentication(w http.ResponseWriter, r *http.Request) *common.ResourceOwner {
	ctx := r.Context()
	
	// Check if authenticated
	authenticated, ok := ctx.Value(common.AuthenticatedKey).(bool)
	if !ok || !authenticated {
		slog.WarnContext(ctx, "Wallet access attempted without authentication")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"success":false,"error":"Authentication required to access wallet"}`))
		return nil
	}
	
	// Get resource owner
	resourceOwner := common.GetResourceOwner(ctx)
	if resourceOwner.UserID == uuid.Nil {
		slog.WarnContext(ctx, "Wallet access attempted without valid user ID")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"success":false,"error":"Valid user authentication required to access wallet"}`))
		return nil
	}
	
	return &resourceOwner
}

// GetWalletBalanceHandler handles GET /wallet/balance
// @Summary Get wallet balance
// @Description Returns the authenticated user's wallet balance for all assets
// @Tags Wallet
// @Produce json
// @Security BearerAuth
// @Success 200 {object} wallet_in.WalletBalanceResult
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /wallet/balance [get]
// SECURITY: This endpoint requires authentication and only returns the authenticated user's balance
func (c *WalletQueryController) GetWalletBalanceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// SECURITY: Verify authentication (critical for financial data)
	resourceOwner := requireAuthentication(w, r)
	if resourceOwner == nil {
		return // Response already written
	}

	// Create query - only allow access to the authenticated user's wallet
	query := wallet_in.GetWalletBalanceQuery{
		UserID: resourceOwner.UserID,
	}

	// Execute use case
	result, err := c.walletQuery.GetBalance(ctx, query)
	if err != nil {
		slog.ErrorContext(ctx, "GetWalletBalanceHandler: error from use case", "error", err)
		common.WriteErrorFromDomainError(w, err)
		return
	}

	common.WriteSuccess(w, result)
}

// GetWalletTransactionsHandler handles GET /wallet/transactions
// @Summary Get wallet transactions
// @Description Returns the authenticated user's wallet transactions with pagination and filtering
// @Tags Wallet
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit results" default(50) maximum(200)
// @Param offset query int false "Offset for pagination" default(0)
// @Param sort_by query string false "Sort field" Enums(created_at, amount, type, status) default(created_at)
// @Param sort_order query string false "Sort order" Enums(asc, desc) default(desc)
// @Success 200 {object} wallet_in.TransactionsResult
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /wallet/transactions [get]
// SECURITY: This endpoint requires authentication and only returns the authenticated user's transactions
func (c *WalletQueryController) GetWalletTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// SECURITY: Verify authentication (critical for financial data)
	resourceOwner := requireAuthentication(w, r)
	if resourceOwner == nil {
		return // Response already written
	}

	// Parse query params with safe defaults
	filters := wallet_in.TransactionFilters{
		Limit:  50,
		Offset: 0,
	}

	// Parse limit with maximum cap for security/performance
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			if limit > 200 {
				limit = 200 // Cap maximum to prevent abuse
			}
			filters.Limit = limit
		}
	}

	// Parse offset
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	// Parse sort parameters with validation
	sortBy := r.URL.Query().Get("sort_by")
	// Only allow specific sort fields for security
	allowedSortFields := map[string]bool{
		"created_at": true,
		"amount":     true,
		"type":       true,
		"status":     true,
	}
	if allowedSortFields[sortBy] {
		filters.SortBy = sortBy
	}
	
	sortOrder := r.URL.Query().Get("sort_order")
	if sortOrder == "asc" || sortOrder == "desc" {
		filters.SortOrder = sortOrder
	}

	// Create query - only allow access to the authenticated user's transactions
	query := wallet_in.GetTransactionsQuery{
		UserID:  resourceOwner.UserID,
		Filters: filters,
	}

	// Execute use case
	result, err := c.walletQuery.GetTransactions(ctx, query)
	if err != nil {
		slog.ErrorContext(ctx, "GetWalletTransactionsHandler: error from use case", "error", err)
		common.WriteErrorFromDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		slog.ErrorContext(ctx, "GetWalletTransactionsHandler: error encoding response", "error", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
}
