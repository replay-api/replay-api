package wallet_usecases_test

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_in "github.com/replay-api/replay-api/pkg/domain/wallet/ports/in"
	wallet_out "github.com/replay-api/replay-api/pkg/domain/wallet/ports/out"
	wallet_services "github.com/replay-api/replay-api/pkg/domain/wallet/services"
	wallet_usecases "github.com/replay-api/replay-api/pkg/domain/wallet/usecases"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLedgerRepository is a mock implementation of wallet_out.LedgerRepository
type MockLedgerRepository struct {
	mock.Mock
}

func (m *MockLedgerRepository) CreateTransaction(ctx context.Context, entries []*wallet_entities.LedgerEntry) error {
	args := m.Called(ctx, entries)
	return args.Error(0)
}

func (m *MockLedgerRepository) CreateEntry(ctx context.Context, entry *wallet_entities.LedgerEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *MockLedgerRepository) FindByID(ctx context.Context, id uuid.UUID) (*wallet_entities.LedgerEntry, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wallet_entities.LedgerEntry), args.Error(1)
}

func (m *MockLedgerRepository) FindByTransactionID(ctx context.Context, txID uuid.UUID) ([]*wallet_entities.LedgerEntry, error) {
	args := m.Called(ctx, txID)
	return args.Get(0).([]*wallet_entities.LedgerEntry), args.Error(1)
}

func (m *MockLedgerRepository) FindByAccountID(ctx context.Context, accountID uuid.UUID, limit int, offset int) ([]*wallet_entities.LedgerEntry, error) {
	args := m.Called(ctx, accountID, limit, offset)
	return args.Get(0).([]*wallet_entities.LedgerEntry), args.Error(1)
}

func (m *MockLedgerRepository) FindByAccountAndCurrency(ctx context.Context, accountID uuid.UUID, currency wallet_vo.Currency) ([]*wallet_entities.LedgerEntry, error) {
	args := m.Called(ctx, accountID, currency)
	return args.Get(0).([]*wallet_entities.LedgerEntry), args.Error(1)
}

func (m *MockLedgerRepository) FindByIdempotencyKey(ctx context.Context, key string) (*wallet_entities.LedgerEntry, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wallet_entities.LedgerEntry), args.Error(1)
}

func (m *MockLedgerRepository) ExistsByIdempotencyKey(ctx context.Context, key string) bool {
	args := m.Called(ctx, key)
	return args.Bool(0)
}

func (m *MockLedgerRepository) FindByDateRange(ctx context.Context, accountID uuid.UUID, from time.Time, to time.Time) ([]*wallet_entities.LedgerEntry, error) {
	args := m.Called(ctx, accountID, from, to)
	return args.Get(0).([]*wallet_entities.LedgerEntry), args.Error(1)
}

func (m *MockLedgerRepository) CalculateBalance(ctx context.Context, accountID uuid.UUID, currency wallet_vo.Currency) (wallet_vo.Amount, error) {
	args := m.Called(ctx, accountID, currency)
	return args.Get(0).(wallet_vo.Amount), args.Error(1)
}

func (m *MockLedgerRepository) GetAccountHistory(ctx context.Context, accountID uuid.UUID, filters wallet_out.HistoryFilters) ([]*wallet_entities.LedgerEntry, int64, error) {
	args := m.Called(ctx, accountID, filters)
	return args.Get(0).([]*wallet_entities.LedgerEntry), args.Get(1).(int64), args.Error(2)
}

func (m *MockLedgerRepository) FindPendingApprovals(ctx context.Context, limit int) ([]*wallet_entities.LedgerEntry, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*wallet_entities.LedgerEntry), args.Error(1)
}

func (m *MockLedgerRepository) UpdateApprovalStatus(ctx context.Context, entryID uuid.UUID, status wallet_entities.ApprovalStatus, approverID uuid.UUID) error {
	args := m.Called(ctx, entryID, status, approverID)
	return args.Error(0)
}

func (m *MockLedgerRepository) MarkAsReversed(ctx context.Context, entryID uuid.UUID, reversalEntryID uuid.UUID) error {
	args := m.Called(ctx, entryID, reversalEntryID)
	return args.Error(0)
}

func (m *MockLedgerRepository) GetDailyTransactionCount(ctx context.Context, accountID uuid.UUID) (int64, error) {
	args := m.Called(ctx, accountID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockLedgerRepository) GetDailyTransactionVolume(ctx context.Context, accountID uuid.UUID, currency wallet_vo.Currency) (wallet_vo.Amount, error) {
	args := m.Called(ctx, accountID, currency)
	return args.Get(0).(wallet_vo.Amount), args.Error(1)
}

func (m *MockLedgerRepository) FindByUserAndDateRange(ctx context.Context, userID uuid.UUID, from time.Time, to time.Time) ([]*wallet_entities.LedgerEntry, error) {
	args := m.Called(ctx, userID, from, to)
	return args.Get(0).([]*wallet_entities.LedgerEntry), args.Error(1)
}

// createTestEVMAddress is defined in get_wallet_balance_test.go

func TestGetTransactions_Success(t *testing.T) {
	mockWalletRepo := new(MockWalletRepository)
	mockLedgerRepo := new(MockLedgerRepository)
	walletQuerySvc := wallet_services.NewWalletQueryService(mockWalletRepo)
	usecase := wallet_usecases.NewGetTransactionsUseCase(mockWalletRepo, walletQuerySvc, mockLedgerRepo)

	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)

	walletID := uuid.New()
	now := time.Now()
	testWallet := &wallet_entities.UserWallet{
		BaseEntity: shared.BaseEntity{
			ID:        walletID,
			CreatedAt: now,
			UpdatedAt: now,
		},
		EVMAddress: createTestEVMAddress(),
		Balances: map[wallet_vo.Currency]wallet_vo.Amount{
			wallet_vo.CurrencyUSD: wallet_vo.NewAmount(100),
		},
	}

	mockWalletRepo.On("Search", mock.Anything, mock.AnythingOfType("shared.Search")).Return([]wallet_entities.UserWallet{*testWallet}, nil)

	// Create test ledger entries
	entries := []*wallet_entities.LedgerEntry{
		{
			ID:            uuid.New(),
			TransactionID: uuid.New(),
			AccountID:     walletID,
			EntryType:     wallet_entities.EntryTypeDebit,
			AssetType:     wallet_entities.AssetTypeFiat,
			Currency:      string(wallet_vo.CurrencyUSD),
			Amount:        big.NewFloat(50),
			BalanceAfter:  big.NewFloat(100),
			Description:   "Deposit",
			CreatedAt:     now,
			IsReversed:    false,
			Metadata: wallet_entities.LedgerMetadata{
				OperationType: "Deposit",
			},
		},
		{
			ID:            uuid.New(),
			TransactionID: uuid.New(),
			AccountID:     walletID,
			EntryType:     wallet_entities.EntryTypeCredit,
			AssetType:     wallet_entities.AssetTypeFiat,
			Currency:      string(wallet_vo.CurrencyUSD),
			Amount:        big.NewFloat(25),
			BalanceAfter:  big.NewFloat(75),
			Description:   "Entry Fee",
			CreatedAt:     now.Add(-time.Hour),
			IsReversed:    false,
			Metadata: wallet_entities.LedgerMetadata{
				OperationType: "EntryFee",
			},
		},
	}

	mockLedgerRepo.On("GetAccountHistory", mock.Anything, walletID, mock.Anything).Return(entries, int64(2), nil)

	query := wallet_in.GetTransactionsQuery{
		UserID: userID,
		Filters: wallet_in.TransactionFilters{
			Limit:  50,
			Offset: 0,
		},
	}

	result, err := usecase.GetTransactions(ctx, query)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(2), result.TotalCount)
	assert.Len(t, result.Transactions, 2)
	assert.Equal(t, "Deposit", result.Transactions[0].Type)
	assert.Contains(t, result.Transactions[0].Amount, "50")
	mockWalletRepo.AssertExpectations(t)
	mockLedgerRepo.AssertExpectations(t)
}

func TestGetTransactions_Unauthenticated(t *testing.T) {
	mockWalletRepo := new(MockWalletRepository)
	mockLedgerRepo := new(MockLedgerRepository)
	walletQuerySvc := wallet_services.NewWalletQueryService(mockWalletRepo)
	usecase := wallet_usecases.NewGetTransactionsUseCase(mockWalletRepo, walletQuerySvc, mockLedgerRepo)

	ctx := context.Background()
	// No authentication context

	query := wallet_in.GetTransactionsQuery{
		UserID: uuid.New(),
		Filters: wallet_in.TransactionFilters{
			Limit:  50,
			Offset: 0,
		},
	}

	result, err := usecase.GetTransactions(ctx, query)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Unauthorized")
}

func TestGetTransactions_InvalidQuery(t *testing.T) {
	mockWalletRepo := new(MockWalletRepository)
	mockLedgerRepo := new(MockLedgerRepository)
	walletQuerySvc := wallet_services.NewWalletQueryService(mockWalletRepo)
	usecase := wallet_usecases.NewGetTransactionsUseCase(mockWalletRepo, walletQuerySvc, mockLedgerRepo)

	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)

	// Empty UserID
	query := wallet_in.GetTransactionsQuery{
		UserID: uuid.Nil,
		Filters: wallet_in.TransactionFilters{
			Limit:  50,
			Offset: 0,
		},
	}

	result, err := usecase.GetTransactions(ctx, query)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetTransactions_WalletNotFound_ReturnsEmpty(t *testing.T) {
	mockWalletRepo := new(MockWalletRepository)
	mockLedgerRepo := new(MockLedgerRepository)
	walletQuerySvc := wallet_services.NewWalletQueryService(mockWalletRepo)
	usecase := wallet_usecases.NewGetTransactionsUseCase(mockWalletRepo, walletQuerySvc, mockLedgerRepo)

	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)

	// Wallet not found
	mockWalletRepo.On("Search", mock.Anything, mock.AnythingOfType("shared.Search")).Return([]wallet_entities.UserWallet{}, nil)

	query := wallet_in.GetTransactionsQuery{
		UserID: userID,
		Filters: wallet_in.TransactionFilters{
			Limit:  50,
			Offset: 0,
		},
	}

	result, err := usecase.GetTransactions(ctx, query)

	// Should return empty transactions, not an error
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Transactions)
	assert.Equal(t, int64(0), result.TotalCount)
	mockWalletRepo.AssertExpectations(t)
	mockLedgerRepo.AssertExpectations(t)
}

func TestGetTransactions_LedgerError(t *testing.T) {
	mockWalletRepo := new(MockWalletRepository)
	mockLedgerRepo := new(MockLedgerRepository)
	walletQuerySvc := wallet_services.NewWalletQueryService(mockWalletRepo)
	usecase := wallet_usecases.NewGetTransactionsUseCase(mockWalletRepo, walletQuerySvc, mockLedgerRepo)

	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)

	walletID := uuid.New()
	now := time.Now()
	testWallet := &wallet_entities.UserWallet{
		BaseEntity: shared.BaseEntity{
			ID:        walletID,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	mockWalletRepo.On("Search", mock.Anything, mock.AnythingOfType("shared.Search")).Return([]wallet_entities.UserWallet{*testWallet}, nil)
	mockLedgerRepo.On("GetAccountHistory", mock.Anything, walletID, mock.Anything).Return([]*wallet_entities.LedgerEntry{}, int64(0), errors.New("database error"))

	query := wallet_in.GetTransactionsQuery{
		UserID: userID,
		Filters: wallet_in.TransactionFilters{
			Limit:  50,
			Offset: 0,
		},
	}

	result, err := usecase.GetTransactions(ctx, query)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockWalletRepo.AssertExpectations(t)
	mockLedgerRepo.AssertExpectations(t)
}

func TestGetTransactions_EmptyTransactions(t *testing.T) {
	mockWalletRepo := new(MockWalletRepository)
	mockLedgerRepo := new(MockLedgerRepository)
	walletQuerySvc := wallet_services.NewWalletQueryService(mockWalletRepo)
	usecase := wallet_usecases.NewGetTransactionsUseCase(mockWalletRepo, walletQuerySvc, mockLedgerRepo)

	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)

	walletID := uuid.New()
	now := time.Now()
	testWallet := &wallet_entities.UserWallet{
		BaseEntity: shared.BaseEntity{
			ID:        walletID,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	mockWalletRepo.On("Search", mock.Anything, mock.AnythingOfType("shared.Search")).Return([]wallet_entities.UserWallet{*testWallet}, nil)
	mockLedgerRepo.On("GetAccountHistory", mock.Anything, walletID, mock.Anything).Return([]*wallet_entities.LedgerEntry{}, int64(0), nil)

	query := wallet_in.GetTransactionsQuery{
		UserID: userID,
		Filters: wallet_in.TransactionFilters{
			Limit:  50,
			Offset: 0,
		},
	}

	result, err := usecase.GetTransactions(ctx, query)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Transactions)
	assert.Equal(t, int64(0), result.TotalCount)
	assert.Equal(t, 50, result.Limit)
	assert.Equal(t, 0, result.Offset)
	mockWalletRepo.AssertExpectations(t)
	mockLedgerRepo.AssertExpectations(t)
}

func TestGetTransactions_WithPagination(t *testing.T) {
	mockWalletRepo := new(MockWalletRepository)
	mockLedgerRepo := new(MockLedgerRepository)
	walletQuerySvc := wallet_services.NewWalletQueryService(mockWalletRepo)
	usecase := wallet_usecases.NewGetTransactionsUseCase(mockWalletRepo, walletQuerySvc, mockLedgerRepo)

	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)

	walletID := uuid.New()
	now := time.Now()
	testWallet := &wallet_entities.UserWallet{
		BaseEntity: shared.BaseEntity{
			ID:        walletID,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	mockWalletRepo.On("Search", mock.Anything, mock.AnythingOfType("shared.Search")).Return([]wallet_entities.UserWallet{*testWallet}, nil)

	// Return one entry for page 2
	entries := []*wallet_entities.LedgerEntry{
		{
			ID:            uuid.New(),
			TransactionID: uuid.New(),
			AccountID:     walletID,
			EntryType:     wallet_entities.EntryTypeDebit,
			AssetType:     wallet_entities.AssetTypeFiat,
			Currency:      string(wallet_vo.CurrencyUSD),
			Amount:        big.NewFloat(100),
			BalanceAfter:  big.NewFloat(100),
			Description:   "Deposit",
			CreatedAt:     now,
			Metadata: wallet_entities.LedgerMetadata{
				OperationType: "Deposit",
			},
		},
	}

	mockLedgerRepo.On("GetAccountHistory", mock.Anything, walletID, mock.Anything).Return(entries, int64(25), nil)

	query := wallet_in.GetTransactionsQuery{
		UserID: userID,
		Filters: wallet_in.TransactionFilters{
			Limit:  10,
			Offset: 20,
		},
	}

	result, err := usecase.GetTransactions(ctx, query)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(25), result.TotalCount)
	assert.Equal(t, 10, result.Limit)
	assert.Equal(t, 20, result.Offset)
	mockWalletRepo.AssertExpectations(t)
	mockLedgerRepo.AssertExpectations(t)
}
