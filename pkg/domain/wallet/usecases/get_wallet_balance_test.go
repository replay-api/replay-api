package wallet_usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_in "github.com/replay-api/replay-api/pkg/domain/wallet/ports/in"
	wallet_services "github.com/replay-api/replay-api/pkg/domain/wallet/services"
	wallet_usecases "github.com/replay-api/replay-api/pkg/domain/wallet/usecases"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWalletRepository is a mock implementation of wallet_out.WalletRepository
type MockWalletRepository struct {
	mock.Mock
}

func (m *MockWalletRepository) Save(ctx context.Context, wallet *wallet_entities.UserWallet) error {
	args := m.Called(ctx, wallet)
	return args.Error(0)
}

func (m *MockWalletRepository) FindByID(ctx context.Context, id uuid.UUID) (*wallet_entities.UserWallet, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wallet_entities.UserWallet), args.Error(1)
}

func (m *MockWalletRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*wallet_entities.UserWallet, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wallet_entities.UserWallet), args.Error(1)
}

func (m *MockWalletRepository) FindByEVMAddress(ctx context.Context, address string) (*wallet_entities.UserWallet, error) {
	args := m.Called(ctx, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wallet_entities.UserWallet), args.Error(1)
}

func (m *MockWalletRepository) Update(ctx context.Context, wallet *wallet_entities.UserWallet) error {
	args := m.Called(ctx, wallet)
	return args.Error(0)
}

func (m *MockWalletRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWalletRepository) GetByID(ctx context.Context, id uuid.UUID) (*wallet_entities.UserWallet, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wallet_entities.UserWallet), args.Error(1)
}

func (m *MockWalletRepository) Search(ctx context.Context, s shared.Search) ([]wallet_entities.UserWallet, error) {
	args := m.Called(ctx, s)
	return args.Get(0).([]wallet_entities.UserWallet), args.Error(1)
}

func (m *MockWalletRepository) Compile(ctx context.Context, searchParams []shared.SearchAggregation, resultOptions shared.SearchResultOptions) (*shared.Search, error) {
	args := m.Called(ctx, searchParams, resultOptions)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shared.Search), args.Error(1)
}

func createTestEVMAddress() wallet_vo.EVMAddress {
	addr, _ := wallet_vo.NewEVMAddress("0x1234567890123456789012345678901234567890")
	return addr
}

func TestGetWalletBalance_Success(t *testing.T) {
	mockWalletRepo := new(MockWalletRepository)
	walletQuerySvc := wallet_services.NewWalletQueryService(mockWalletRepo)
	usecase := wallet_usecases.NewGetWalletBalanceUseCase(walletQuerySvc)

	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)

	// Create a test wallet
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
			wallet_vo.CurrencyUSD:  wallet_vo.NewAmount(100.50),
			wallet_vo.CurrencyUSDC: wallet_vo.NewAmount(50.25),
		},
		TotalDeposited: wallet_vo.NewAmount(200),
		TotalWithdrawn: wallet_vo.NewAmount(50),
		TotalPrizesWon: wallet_vo.NewAmount(10),
		IsLocked:       false,
	}

	mockWalletRepo.On("Search", mock.Anything, mock.AnythingOfType("shared.Search")).Return([]wallet_entities.UserWallet{*testWallet}, nil)

	query := wallet_in.GetWalletBalanceQuery{
		UserID: userID,
	}

	result, err := usecase.GetBalance(ctx, query)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, walletID, result.WalletID)
	assert.Equal(t, userID, result.UserID)
	assert.Contains(t, result.Balances["USD"], "100.50")
	assert.Contains(t, result.Balances["USDC"], "50.25")
	assert.Contains(t, result.TotalDeposited, "200")
	assert.Contains(t, result.TotalWithdrawn, "50")
	assert.Contains(t, result.TotalPrizesWon, "10")
	assert.False(t, result.IsLocked)
	mockWalletRepo.AssertExpectations(t)
}

func TestGetWalletBalance_Unauthenticated(t *testing.T) {
	mockWalletRepo := new(MockWalletRepository)
	walletQuerySvc := wallet_services.NewWalletQueryService(mockWalletRepo)
	usecase := wallet_usecases.NewGetWalletBalanceUseCase(walletQuerySvc)

	ctx := context.Background()
	// No authentication context

	query := wallet_in.GetWalletBalanceQuery{
		UserID: uuid.New(),
	}

	result, err := usecase.GetBalance(ctx, query)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Unauthorized")
}

func TestGetWalletBalance_InvalidQuery(t *testing.T) {
	mockWalletRepo := new(MockWalletRepository)
	walletQuerySvc := wallet_services.NewWalletQueryService(mockWalletRepo)
	usecase := wallet_usecases.NewGetWalletBalanceUseCase(walletQuerySvc)

	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)

	// Empty UserID
	query := wallet_in.GetWalletBalanceQuery{
		UserID: uuid.Nil,
	}

	result, err := usecase.GetBalance(ctx, query)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetWalletBalance_WalletNotFound_ReturnsDefault(t *testing.T) {
	mockWalletRepo := new(MockWalletRepository)
	walletQuerySvc := wallet_services.NewWalletQueryService(mockWalletRepo)
	usecase := wallet_usecases.NewGetWalletBalanceUseCase(walletQuerySvc)

	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)

	// Wallet not found
	mockWalletRepo.On("Search", mock.Anything, mock.AnythingOfType("shared.Search")).Return([]wallet_entities.UserWallet{}, nil)

	query := wallet_in.GetWalletBalanceQuery{
		UserID: userID,
	}

	result, err := usecase.GetBalance(ctx, query)

	// Should return default wallet for new users, not an error
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, "0.00", result.Balances["USD"])
	assert.Equal(t, "0.00", result.TotalDeposited)
	assert.False(t, result.IsLocked)
	mockWalletRepo.AssertExpectations(t)
}

func TestGetWalletBalance_LockedWallet(t *testing.T) {
	mockWalletRepo := new(MockWalletRepository)
	walletQuerySvc := wallet_services.NewWalletQueryService(mockWalletRepo)
	usecase := wallet_usecases.NewGetWalletBalanceUseCase(walletQuerySvc)

	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)

	// Create a locked wallet
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
		TotalDeposited: wallet_vo.NewAmount(100),
		TotalWithdrawn: wallet_vo.NewAmount(0),
		TotalPrizesWon: wallet_vo.NewAmount(0),
		IsLocked:       true,
		LockReason:     "Fraud investigation",
	}

	mockWalletRepo.On("Search", mock.Anything, mock.AnythingOfType("shared.Search")).Return([]wallet_entities.UserWallet{*testWallet}, nil)

	query := wallet_in.GetWalletBalanceQuery{
		UserID: userID,
	}

	result, err := usecase.GetBalance(ctx, query)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsLocked)
	assert.Equal(t, "Fraud investigation", result.LockReason)
	mockWalletRepo.AssertExpectations(t)
}
