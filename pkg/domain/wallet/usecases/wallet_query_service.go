package wallet_usecases

import (
	"context"

	wallet_in "github.com/replay-api/replay-api/pkg/domain/wallet/ports/in"
)

// WalletQueryService implements the WalletQuery interface
// by delegating to the domain query service
type WalletQueryService struct {
	getBalanceUseCase     *GetWalletBalanceUseCase
	getTransactionsUseCase *GetTransactionsUseCase
}

// NewWalletQueryService creates a new wallet query service
func NewWalletQueryService(
	getBalanceUseCase *GetWalletBalanceUseCase,
	getTransactionsUseCase *GetTransactionsUseCase,
) *WalletQueryService {
	return &WalletQueryService{
		getBalanceUseCase:     getBalanceUseCase,
		getTransactionsUseCase: getTransactionsUseCase,
	}
}

// GetBalance retrieves the wallet balance for a user
func (s *WalletQueryService) GetBalance(ctx context.Context, query wallet_in.GetWalletBalanceQuery) (*wallet_in.WalletBalanceResult, error) {
	return s.getBalanceUseCase.GetBalance(ctx, query)
}

// GetTransactions retrieves the transaction history for a user's wallet
func (s *WalletQueryService) GetTransactions(ctx context.Context, query wallet_in.GetTransactionsQuery) (*wallet_in.TransactionsResult, error) {
	return s.getTransactionsUseCase.GetTransactions(ctx, query)
}

// Ensure WalletQueryService implements wallet_in.WalletQuery
var _ wallet_in.WalletQuery = (*WalletQueryService)(nil)
