package wallet_usecases

import (
	"context"
	"log/slog"

	shared "github.com/resource-ownership/go-common/pkg/common"
	wallet_in "github.com/replay-api/replay-api/pkg/domain/wallet/ports/in"
	wallet_out "github.com/replay-api/replay-api/pkg/domain/wallet/ports/out"
	wallet_services "github.com/replay-api/replay-api/pkg/domain/wallet/services"
)

// GetTransactionsUseCase handles fetching wallet transactions
type GetTransactionsUseCase struct {
	walletRepo     wallet_out.WalletRepository
	walletQuerySvc *wallet_services.WalletQueryService
	ledgerRepo     wallet_out.LedgerRepository
}

// NewGetTransactionsUseCase creates a new get transactions use case
func NewGetTransactionsUseCase(
	walletRepo wallet_out.WalletRepository,
	walletQuerySvc *wallet_services.WalletQueryService,
	ledgerRepo wallet_out.LedgerRepository,
) *GetTransactionsUseCase {
	return &GetTransactionsUseCase{
		walletRepo:     walletRepo,
		walletQuerySvc: walletQuerySvc,
		ledgerRepo:     ledgerRepo,
	}
}

// GetTransactions retrieves the transaction history for a user's wallet
func (uc *GetTransactionsUseCase) GetTransactions(ctx context.Context, query wallet_in.GetTransactionsQuery) (*wallet_in.TransactionsResult, error) {
	// Validate query
	if err := query.Validate(); err != nil {
		slog.WarnContext(ctx, "GetTransactions: invalid query", "error", err)
		return nil, shared.NewErrInvalidInput(err.Error())
	}

	// Auth check - user must be authenticated
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, shared.NewErrUnauthorized()
	}

	slog.InfoContext(ctx, "GetTransactions: fetching transactions",
		"user_id", query.UserID,
		"limit", query.Filters.Limit,
		"offset", query.Filters.Offset,
	)

	// Try to find user's wallet first
	wallet, err := uc.walletQuerySvc.FindByUserID(ctx, query.UserID)
	if err != nil || wallet == nil {
		slog.InfoContext(ctx, "GetTransactions: wallet not found, returning empty",
			"user_id", query.UserID,
			"error", err,
		)
		// User has no wallet yet - return empty transactions
		return &wallet_in.TransactionsResult{
			Transactions: []wallet_in.TransactionDTO{},
			TotalCount:   0,
			Limit:        query.Filters.Limit,
			Offset:       query.Filters.Offset,
		}, nil
	}

	// Convert query filters to repository filters
	repoFilters := wallet_out.HistoryFilters{
		Currency:      query.Filters.Currency,
		AssetType:     query.Filters.AssetType,
		EntryType:     query.Filters.EntryType,
		OperationType: query.Filters.OperationType,
		FromDate:      query.Filters.FromDate,
		ToDate:        query.Filters.ToDate,
		Limit:         query.Filters.Limit,
		Offset:        query.Filters.Offset,
		SortBy:        query.Filters.SortBy,
		SortOrder:     query.Filters.SortOrder,
	}

	// Set default sorting
	if repoFilters.SortBy == "" {
		repoFilters.SortBy = "created_at"
	}
	if repoFilters.SortOrder == "" {
		repoFilters.SortOrder = "desc"
	}

	// Fetch transaction history from ledger
	entries, totalCount, err := uc.ledgerRepo.GetAccountHistory(ctx, wallet.ID, repoFilters)
	if err != nil {
		slog.ErrorContext(ctx, "GetTransactions: failed to fetch ledger entries",
			"wallet_id", wallet.ID,
			"error", err,
		)
		return nil, err
	}

	// Convert ledger entries to DTOs
	transactions := make([]wallet_in.TransactionDTO, 0, len(entries))
	for _, entry := range entries {
		tx := wallet_in.TransactionDTO{
			ID:            entry.ID,
			TransactionID: entry.TransactionID,
			Type:          entry.Metadata.OperationType,
			EntryType:     string(entry.EntryType),
			AssetType:     string(entry.AssetType),
			Currency:      entry.Currency,
			Amount:        entry.Amount.String(),
			BalanceAfter:  entry.BalanceAfter.String(),
			Description:   entry.Description,
			CreatedAt:     entry.CreatedAt,
			IsReversed:    entry.IsReversed,
		}
		transactions = append(transactions, tx)
	}

	return &wallet_in.TransactionsResult{
		Transactions: transactions,
		TotalCount:   totalCount,
		Limit:        query.Filters.Limit,
		Offset:       query.Filters.Offset,
	}, nil
}
