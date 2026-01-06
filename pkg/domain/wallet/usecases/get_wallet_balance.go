package wallet_usecases

import (
	"context"
	"log/slog"

	shared "github.com/resource-ownership/go-common/pkg/common"
	wallet_in "github.com/replay-api/replay-api/pkg/domain/wallet/ports/in"
	wallet_out "github.com/replay-api/replay-api/pkg/domain/wallet/ports/out"
)

// GetWalletBalanceUseCase retrieves current wallet balances for authenticated users.
//
// This use case provides:
//   - Multi-currency balance retrieval (USD, USDC, USDT)
//   - Default wallet creation for new users (zero balances)
//   - Account lock status and reason display
//   - Total deposited/withdrawn/prizes tracking
//
// Security:
//   - Requires authenticated context
//   - Users can only access their own wallet balance
//
// Returns default zeroed wallet if wallet doesn't exist yet, ensuring
// seamless UX for new users before their first deposit.
//
// Dependencies:
//   - WalletRepository: For wallet lookup by user ID
type GetWalletBalanceUseCase struct {
	walletRepo wallet_out.WalletRepository
}

// NewGetWalletBalanceUseCase creates a new get wallet balance use case
func NewGetWalletBalanceUseCase(walletRepo wallet_out.WalletRepository) *GetWalletBalanceUseCase {
	return &GetWalletBalanceUseCase{
		walletRepo: walletRepo,
	}
}

// GetBalance retrieves the wallet balance for a user
func (uc *GetWalletBalanceUseCase) GetBalance(ctx context.Context, query wallet_in.GetWalletBalanceQuery) (*wallet_in.WalletBalanceResult, error) {
	// Validate query
	if err := query.Validate(); err != nil {
		slog.WarnContext(ctx, "GetWalletBalance: invalid query", "error", err)
		return nil, shared.NewErrInvalidInput(err.Error())
	}

	// Auth check - user must be authenticated
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, shared.NewErrUnauthorized()
	}

	slog.InfoContext(ctx, "GetWalletBalance: fetching balance", "user_id", query.UserID)

	// Try to fetch wallet
	wallet, err := uc.walletRepo.FindByUserID(ctx, query.UserID)
	if err != nil {
		slog.InfoContext(ctx, "GetWalletBalance: wallet not found, returning default",
			"user_id", query.UserID,
			"error", err,
		)
		// Return default wallet for new users
		return &wallet_in.WalletBalanceResult{
			UserID:         query.UserID,
			Balances:       map[string]string{"USD": "0.00", "USDC": "0.00", "USDT": "0.00"},
			TotalDeposited: "0.00",
			TotalWithdrawn: "0.00",
			TotalPrizesWon: "0.00",
			IsLocked:       false,
		}, nil
	}

	// Convert balances to string map
	balancesStr := make(map[string]string)
	for currency, amount := range wallet.Balances {
		balancesStr[string(currency)] = amount.String()
	}

	createdAt := wallet.CreatedAt
	updatedAt := wallet.UpdatedAt

	return &wallet_in.WalletBalanceResult{
		WalletID:       wallet.ID,
		UserID:         query.UserID,
		EVMAddress:     wallet.EVMAddress.String(),
		Balances:       balancesStr,
		TotalDeposited: wallet.TotalDeposited.String(),
		TotalWithdrawn: wallet.TotalWithdrawn.String(),
		TotalPrizesWon: wallet.TotalPrizesWon.String(),
		IsLocked:       wallet.IsLocked,
		LockReason:     wallet.LockReason,
		CreatedAt:      &createdAt,
		UpdatedAt:      &updatedAt,
	}, nil
}
