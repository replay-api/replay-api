package wallet_usecases

import (
	"context"

	"github.com/google/uuid"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_in "github.com/replay-api/replay-api/pkg/domain/wallet/ports/in"
	wallet_services "github.com/replay-api/replay-api/pkg/domain/wallet/services"
)

// WalletQueryService implements the WalletQuery interface
// by delegating to the domain query service
type WalletQueryService struct {
	getBalanceUseCase      *GetWalletBalanceUseCase
	getTransactionsUseCase *GetTransactionsUseCase
	walletQuerySvc         *wallet_services.WalletQueryService
}

// NewWalletQueryService creates a new wallet query service
func NewWalletQueryService(
	getBalanceUseCase *GetWalletBalanceUseCase,
	getTransactionsUseCase *GetTransactionsUseCase,
) *WalletQueryService {
	return &WalletQueryService{
		getBalanceUseCase:      getBalanceUseCase,
		getTransactionsUseCase: getTransactionsUseCase,
	}
}

// SetWalletQuerySvc sets the domain wallet query service for direct wallet lookups
func (s *WalletQueryService) SetWalletQuerySvc(svc *wallet_services.WalletQueryService) {
	s.walletQuerySvc = svc
}

// GetBalance retrieves the wallet balance for a user
func (s *WalletQueryService) GetBalance(ctx context.Context, query wallet_in.GetWalletBalanceQuery) (*wallet_in.WalletBalanceResult, error) {
	return s.getBalanceUseCase.GetBalance(ctx, query)
}

// GetTransactions retrieves the transaction history for a user's wallet
func (s *WalletQueryService) GetTransactions(ctx context.Context, query wallet_in.GetTransactionsQuery) (*wallet_in.TransactionsResult, error) {
	return s.getTransactionsUseCase.GetTransactions(ctx, query)
}

// GetWalletByUserID returns the wallet entity for a given user ID, used for ownership verification
func (s *WalletQueryService) GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*wallet_entities.UserWallet, error) {
	if s.walletQuerySvc == nil {
		return nil, nil
	}
	return s.walletQuerySvc.FindByUserID(ctx, userID)
}

// Ensure WalletQueryService implements wallet_in.WalletQuery
var _ wallet_in.WalletQuery = (*WalletQueryService)(nil)
