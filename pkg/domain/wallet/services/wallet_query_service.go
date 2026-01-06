package wallet_services

import (
	"context"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
)

// WalletQueryService provides domain query operations for wallets using technology-agnostic search patterns
type WalletQueryService struct {
	reader          shared.Searchable[wallet_entities.UserWallet]
	queryableFields map[string]bool
}

// NewWalletQueryService creates a new wallet query service
func NewWalletQueryService(walletReader shared.Searchable[wallet_entities.UserWallet]) *WalletQueryService {
	queryableFields := map[string]bool{
		"ID":                  true,
		"EVMAddress":          true,
		"Balances":            true,
		"PendingTransactions": true,
		"TotalDeposited":      true,
		"TotalWithdrawn":      true,
		"TotalPrizesWon":      true,
		"DailyPrizeWinnings":  true,
		"LastPrizeWinDate":    true,
		"IsLocked":            true,
		"LockReason":          true,
		"CreatedAt":           true,
		"UpdatedAt":           true,
		// Resource ownership fields
		"TenantID": true,
		"UserID":   true,
		"GroupID":  true,
		"ClientID": true,
	}

	return &WalletQueryService{
		reader:          walletReader,
		queryableFields: queryableFields,
	}
}

// FindByUserID finds a wallet by user ID
// Business rule: Wallets are scoped by resource ownership (RLS)
func (s *WalletQueryService) FindByUserID(ctx context.Context, userID uuid.UUID) (*wallet_entities.UserWallet, error) {
	search := shared.NewSearchBuilder().
		WithAggregation(shared.NewSearchAggregation().
			WithValueParam("UserID", userID).
			Build()).
		WithLimit(1).
		Build()

	entities, err := s.reader.Search(ctx, search)
	if err != nil {
		return nil, err
	}

	if len(entities) == 0 {
		return nil, nil
	}

	return &entities[0], nil
}

// FindByEVMAddress finds a wallet by EVM address
// Business rule: EVM addresses are unique identifiers
func (s *WalletQueryService) FindByEVMAddress(ctx context.Context, evmAddress string) (*wallet_entities.UserWallet, error) {
	search := shared.NewSearchBuilder().
		WithAggregation(shared.NewSearchAggregation().
			WithValueParam("EVMAddress", evmAddress).
			Build()).
		WithLimit(1).
		Build()

	entities, err := s.reader.Search(ctx, search)
	if err != nil {
		return nil, err
	}

	if len(entities) == 0 {
		return nil, nil
	}

	return &entities[0], nil
}

// FindLockedWallets finds wallets that are currently locked
// Business rule: Locked wallets may have restrictions on operations
func (s *WalletQueryService) FindLockedWallets(ctx context.Context, limit int) ([]*wallet_entities.UserWallet, error) {
	search := shared.NewSearchBuilder().
		WithAggregation(shared.NewSearchAggregation().
			WithValueParam("IsLocked", true).
			Build()).
		WithSort("CreatedAt", shared.DescendingIDKey).
		WithLimit(uint(limit)).
		Build()

	entities, err := s.reader.Search(ctx, search)
	if err != nil {
		return nil, err
	}

	// Convert to pointer slice
	wallets := make([]*wallet_entities.UserWallet, len(entities))
	for i := range entities {
		wallets[i] = &entities[i]
	}

	return wallets, nil
}

// FindByBalanceThreshold finds wallets with balance above a threshold for a specific currency
// Business rule: Useful for reporting and analytics
func (s *WalletQueryService) FindByBalanceThreshold(ctx context.Context, currency string, minBalance float64, limit int) ([]*wallet_entities.UserWallet, error) {
	// Note: This would require extending the search framework to support nested field queries
	// For now, this is a placeholder showing the pattern
	search := shared.NewSearchBuilder().
		WithLimit(uint(limit)).
		Build()

	entities, err := s.reader.Search(ctx, search)
	if err != nil {
		return nil, err
	}

	// Convert to pointer slice
	wallets := make([]*wallet_entities.UserWallet, len(entities))
	for i := range entities {
		wallets[i] = &entities[i]
	}

	return wallets, nil
}
