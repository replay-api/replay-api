package wallet_services

import (
"context"
"fmt"

"github.com/google/uuid"
common "github.com/psavelis/team-pro/replay-api/pkg/domain"
wallet_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/wallet/entities"
wallet_in "github.com/psavelis/team-pro/replay-api/pkg/domain/wallet/ports/in"
wallet_out "github.com/psavelis/team-pro/replay-api/pkg/domain/wallet/ports/out"
wallet_vo "github.com/psavelis/team-pro/replay-api/pkg/domain/wallet/value-objects"
)

// WalletService implements wallet business logic
type WalletService struct {
	walletRepo wallet_out.WalletRepository
}

func NewWalletService(walletRepo wallet_out.WalletRepository) wallet_in.WalletCommand {
	return &WalletService{
		walletRepo: walletRepo,
	}
}

func (s *WalletService) CreateWallet(ctx context.Context, cmd wallet_in.CreateWalletCommand) (*wallet_entities.UserWallet, error) {
	evmAddress, err := wallet_vo.NewEVMAddress(cmd.EVMAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid EVM address: %w", err)
	}

	resourceOwner := common.GetResourceOwner(ctx)
	wallet, err := wallet_entities.NewUserWallet(resourceOwner, evmAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	if err := s.walletRepo.Save(ctx, wallet); err != nil {
		return nil, fmt.Errorf("failed to save wallet: %w", err)
	}

	return wallet, nil
}

func (s *WalletService) Deposit(ctx context.Context, cmd wallet_in.DepositCommand) error {
	wallet, err := s.walletRepo.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("wallet not found: %w", err)
	}

	currency, err := wallet_vo.ParseCurrency(cmd.Currency)
	if err != nil {
		return fmt.Errorf("invalid currency: %w", err)
	}

	amount := wallet_vo.NewAmount(cmd.Amount)

	if err := wallet.Deposit(currency, amount); err != nil {
		return fmt.Errorf("deposit failed: %w", err)
	}

	txHash, err := uuid.Parse(cmd.TxHash)
	if err != nil {
		return fmt.Errorf("invalid transaction hash: %w", err)
	}
	wallet.AddPendingTransaction(txHash)

	if err := s.walletRepo.Update(ctx, wallet); err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	return nil
}

func (s *WalletService) Withdraw(ctx context.Context, cmd wallet_in.WithdrawCommand) error {
	wallet, err := s.walletRepo.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("wallet not found: %w", err)
	}

	currency, err := wallet_vo.ParseCurrency(cmd.Currency)
	if err != nil {
		return fmt.Errorf("invalid currency: %w", err)
	}

	amount := wallet_vo.NewAmount(cmd.Amount)

	if err := wallet.Withdraw(currency, amount); err != nil {
		return fmt.Errorf("withdraw failed: %w", err)
	}

	if err := s.walletRepo.Update(ctx, wallet); err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	return nil
}

func (s *WalletService) DeductEntryFee(ctx context.Context, cmd wallet_in.DeductEntryFeeCommand) error {
	wallet, err := s.walletRepo.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("wallet not found: %w", err)
	}

	currency, err := wallet_vo.ParseCurrency(cmd.Currency)
	if err != nil {
		return fmt.Errorf("invalid currency: %w", err)
	}

	amount := wallet_vo.NewAmount(cmd.Amount)

	if err := wallet.DeductEntryFee(currency, amount); err != nil {
		return fmt.Errorf("insufficient balance: %w", err)
	}

	if err := s.walletRepo.Update(ctx, wallet); err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	return nil
}

func (s *WalletService) AddPrize(ctx context.Context, cmd wallet_in.AddPrizeCommand) error {
	wallet, err := s.walletRepo.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("wallet not found: %w", err)
	}

	currency, err := wallet_vo.ParseCurrency(cmd.Currency)
	if err != nil {
		return fmt.Errorf("invalid currency: %w", err)
	}

	amount := wallet_vo.NewAmount(cmd.Amount)

	maxDailyWinnings := wallet_vo.NewAmount(50.00) // $50/day limit

	if err := wallet.AddPrize(currency, amount, maxDailyWinnings); err != nil {
		return fmt.Errorf("failed to add prize: %w", err)
	}

	if err := s.walletRepo.Update(ctx, wallet); err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	return nil
}

func (s *WalletService) Refund(ctx context.Context, cmd wallet_in.RefundCommand) error {
	wallet, err := s.walletRepo.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("wallet not found: %w", err)
	}

	currency, err := wallet_vo.ParseCurrency(cmd.Currency)
	if err != nil {
		return fmt.Errorf("invalid currency: %w", err)
	}

	amount := wallet_vo.NewAmount(cmd.Amount)

	if err := wallet.Deposit(currency, amount); err != nil {
		return fmt.Errorf("refund failed: %w", err)
	}

	if err := s.walletRepo.Update(ctx, wallet); err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	return nil
}
