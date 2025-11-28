package query_controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/golobby/container/v3"
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

// GetWalletBalanceHandler handles GET /wallet/balance
func (c *WalletQueryController) GetWalletBalanceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get resource owner from context (set by middleware)
	resourceOwner := common.GetResourceOwner(ctx)

	// Create query
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
func (c *WalletQueryController) GetWalletTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get resource owner from context (set by middleware)
	resourceOwner := common.GetResourceOwner(ctx)

	// Parse query params
	filters := wallet_in.TransactionFilters{
		Limit:  50,
		Offset: 0,
	}

	// Parse limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	}

	// Parse offset
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	// Parse sort parameters
	filters.SortBy = r.URL.Query().Get("sort_by")
	filters.SortOrder = r.URL.Query().Get("sort_order")

	// Create query
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
