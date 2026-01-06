package wallet_services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_in "github.com/replay-api/replay-api/pkg/domain/wallet/ports/in"
	wallet_out "github.com/replay-api/replay-api/pkg/domain/wallet/ports/out"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

// WalletService implements wallet business logic with ledger integration
// All wallet operations are recorded in an immutable ledger for audit compliance
// Uses transaction coordinator for atomic operations with automatic rollback
type WalletService struct {
	walletRepo     wallet_out.WalletRepository
	walletQuerySvc *WalletQueryService
	coordinator    *TransactionCoordinator
}

func NewWalletService(
	walletRepo wallet_out.WalletRepository,
	walletQuerySvc *WalletQueryService,
	coordinator *TransactionCoordinator,
) wallet_in.WalletCommand {
	return &WalletService{
		walletRepo:     walletRepo,
		walletQuerySvc: walletQuerySvc,
		coordinator:    coordinator,
	}
}

func (s *WalletService) CreateWallet(ctx context.Context, cmd wallet_in.CreateWalletCommand) (*wallet_entities.UserWallet, error) {
	evmAddress, err := wallet_vo.NewEVMAddress(cmd.EVMAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid EVM address: %w", err)
	}

	resourceOwner := shared.GetResourceOwner(ctx)
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
	wallet, err := s.walletQuerySvc.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("wallet not found: %w", err)
	}

	currency, err := wallet_vo.ParseCurrency(cmd.Currency)
	if err != nil {
		return fmt.Errorf("invalid currency: %w", err)
	}

	amount := wallet_vo.NewAmount(cmd.Amount)

	paymentID, err := uuid.Parse(cmd.TxHash)
	if err != nil {
		return fmt.Errorf("invalid transaction hash: %w", err)
	}

	// Use transaction coordinator for atomic operation with automatic rollback
	ledgerTxID, err := s.coordinator.ExecuteDeposit(
		ctx,
		wallet,
		currency,
		amount,
		paymentID,
		wallet_entities.LedgerMetadata{}, // TODO: Add request metadata
	)
	if err != nil {
		slog.ErrorContext(ctx, "deposit transaction failed",
			"wallet_id", wallet.ID,
			"amount", amount.String(),
			"error", err)
		return fmt.Errorf("deposit failed: %w", err)
	}

	wallet.AddPendingTransaction(paymentID)

	slog.InfoContext(ctx, "deposit completed successfully",
		"wallet_id", wallet.ID,
		"amount", amount.String(),
		"currency", currency,
		"ledger_tx_id", ledgerTxID)

	return nil
}

func (s *WalletService) Withdraw(ctx context.Context, cmd wallet_in.WithdrawCommand) error {
	wallet, err := s.walletQuerySvc.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("wallet not found: %w", err)
	}

	currency, err := wallet_vo.ParseCurrency(cmd.Currency)
	if err != nil {
		return fmt.Errorf("invalid currency: %w", err)
	}

	amount := wallet_vo.NewAmount(cmd.Amount)

	// Use transaction coordinator for atomic operation
	ledgerTxID, err := s.coordinator.ExecuteWithdrawal(
		ctx,
		wallet,
		currency,
		amount,
		cmd.ToAddress,
		wallet_entities.LedgerMetadata{},
	)
	if err != nil {
		slog.ErrorContext(ctx, "withdrawal transaction failed",
			"wallet_id", wallet.ID,
			"amount", amount.String(),
			"error", err)
		return fmt.Errorf("withdraw failed: %w", err)
	}

	slog.InfoContext(ctx, "withdrawal completed successfully",
		"wallet_id", wallet.ID,
		"amount", amount.String(),
		"currency", currency,
		"to_address", cmd.ToAddress,
		"ledger_tx_id", ledgerTxID)

	return nil
}

func (s *WalletService) DeductEntryFee(ctx context.Context, cmd wallet_in.DeductEntryFeeCommand) error {
	wallet, err := s.walletQuerySvc.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("wallet not found: %w", err)
	}

	currency, err := wallet_vo.ParseCurrency(cmd.Currency)
	if err != nil {
		return fmt.Errorf("invalid currency: %w", err)
	}

	amount := wallet_vo.NewAmount(cmd.Amount)

	// Use transaction coordinator for atomic operation
	ledgerTxID, err := s.coordinator.ExecuteEntryFee(
		ctx,
		wallet,
		currency,
		amount,
		nil, // TODO: Add matchID to command
		nil, // TODO: Add tournamentID to command
		wallet_entities.LedgerMetadata{},
	)
	if err != nil {
		slog.ErrorContext(ctx, "entry fee transaction failed",
			"wallet_id", wallet.ID,
			"amount", amount.String(),
			"error", err)
		return fmt.Errorf("insufficient balance: %w", err)
	}

	slog.InfoContext(ctx, "entry fee deducted successfully",
		"wallet_id", wallet.ID,
		"amount", amount.String(),
		"currency", currency,
		"ledger_tx_id", ledgerTxID)

	return nil
}

func (s *WalletService) AddPrize(ctx context.Context, cmd wallet_in.AddPrizeCommand) error {
	wallet, err := s.walletQuerySvc.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("wallet not found: %w", err)
	}

	currency, err := wallet_vo.ParseCurrency(cmd.Currency)
	if err != nil {
		return fmt.Errorf("invalid currency: %w", err)
	}

	amount := wallet_vo.NewAmount(cmd.Amount)
	maxDailyWinnings := wallet_vo.NewAmount(50.00) // $50/day limit

	// Use transaction coordinator for atomic operation
	ledgerTxID, err := s.coordinator.ExecutePrizeWinning(
		ctx,
		wallet,
		currency,
		amount,
		nil, // TODO: Add matchID to command
		nil, // TODO: Add tournamentID to command
		maxDailyWinnings,
		wallet_entities.LedgerMetadata{},
	)
	if err != nil {
		slog.ErrorContext(ctx, "prize transaction failed",
			"wallet_id", wallet.ID,
			"amount", amount.String(),
			"error", err)
		return fmt.Errorf("failed to add prize: %w", err)
	}

	slog.InfoContext(ctx, "prize added successfully",
		"wallet_id", wallet.ID,
		"amount", amount.String(),
		"currency", currency,
		"ledger_tx_id", ledgerTxID)

	return nil
}

func (s *WalletService) Refund(ctx context.Context, cmd wallet_in.RefundCommand) error {
	wallet, err := s.walletQuerySvc.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("wallet not found: %w", err)
	}

	currency, err := wallet_vo.ParseCurrency(cmd.Currency)
	if err != nil {
		return fmt.Errorf("invalid currency: %w", err)
	}

	amount := wallet_vo.NewAmount(cmd.Amount)

	// TODO: RefundCommand should include original transaction ID
	// For now, record refund as a new deposit with refund metadata
	// Ideally: s.coordinator.ExecuteRefund(ctx, originalTxID, cmd.Reason)

	// Record refund as deposit using transaction coordinator
	refundPaymentID := uuid.New()
	ledgerTxID, err := s.coordinator.ExecuteDeposit(
		ctx,
		wallet,
		currency,
		amount,
		refundPaymentID,
		wallet_entities.LedgerMetadata{
			OperationType: "Refund",
			Notes:         cmd.Reason,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to record refund in ledger: %w", err)
	}

	// Update wallet balance
	if err := wallet.Deposit(currency, amount); err != nil {
		slog.ErrorContext(ctx, "wallet refund failed after ledger write",
			"wallet_id", wallet.ID,
			"ledger_tx_id", ledgerTxID,
			"error", err)
		// TODO: Implement automatic reversal
		return fmt.Errorf("refund failed: %w", err)
	}

	if err := s.walletRepo.Update(ctx, wallet); err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	slog.InfoContext(ctx, "refund completed successfully",
		"wallet_id", wallet.ID,
		"amount", amount.String(),
		"currency", currency,
		"reason", cmd.Reason,
		"ledger_tx_id", ledgerTxID)

	return nil
}

// DebitWallet debits an amount from the user's wallet
func (s *WalletService) DebitWallet(ctx context.Context, cmd wallet_in.DebitWalletCommand) (*wallet_entities.WalletTransaction, error) {
	wallet, err := s.walletQuerySvc.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	currency, err := wallet_vo.ParseCurrency(cmd.Currency)
	if err != nil {
		return nil, fmt.Errorf("invalid currency: %w", err)
	}

	// Use transaction coordinator for atomic operation
	ledgerTxID, err := s.coordinator.ExecuteWithdrawal(
		ctx,
		wallet,
		currency,
		cmd.Amount,
		"internal_debit",
		wallet_entities.LedgerMetadata{
			OperationType: "Debit",
			Notes:         cmd.Description,
		},
	)
	if err != nil {
		slog.ErrorContext(ctx, "debit transaction failed",
			"wallet_id", wallet.ID,
			"amount", cmd.Amount.String(),
			"error", err)
		return nil, fmt.Errorf("debit failed: %w", err)
	}

	slog.InfoContext(ctx, "debit completed successfully",
		"wallet_id", wallet.ID,
		"amount", cmd.Amount.String(),
		"currency", currency,
		"ledger_tx_id", ledgerTxID)

	now := time.Now()
	return &wallet_entities.WalletTransaction{
		ID:          uuid.New(),
		WalletID:    wallet.ID,
		Type:        "Debit",
		Status:      wallet_entities.TransactionStatusCompleted,
		LedgerTxID:  &ledgerTxID,
		StartedAt:   now,
		CompletedAt: &now,
		Metadata:    cmd.Metadata,
	}, nil
}

// CreditWallet credits an amount to the user's wallet
func (s *WalletService) CreditWallet(ctx context.Context, cmd wallet_in.CreditWalletCommand) (*wallet_entities.WalletTransaction, error) {
	wallet, err := s.walletQuerySvc.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	currency, err := wallet_vo.ParseCurrency(cmd.Currency)
	if err != nil {
		return nil, fmt.Errorf("invalid currency: %w", err)
	}

	// Use transaction coordinator for atomic operation
	paymentID := uuid.New()
	ledgerTxID, err := s.coordinator.ExecuteDeposit(
		ctx,
		wallet,
		currency,
		cmd.Amount,
		paymentID,
		wallet_entities.LedgerMetadata{
			OperationType: "Credit",
			Notes:         cmd.Description,
		},
	)
	if err != nil {
		slog.ErrorContext(ctx, "credit transaction failed",
			"wallet_id", wallet.ID,
			"amount", cmd.Amount.String(),
			"error", err)
		return nil, fmt.Errorf("credit failed: %w", err)
	}

	slog.InfoContext(ctx, "credit completed successfully",
		"wallet_id", wallet.ID,
		"amount", cmd.Amount.String(),
		"currency", currency,
		"ledger_tx_id", ledgerTxID)

	now := time.Now()
	return &wallet_entities.WalletTransaction{
		ID:          uuid.New(),
		WalletID:    wallet.ID,
		Type:        "Credit",
		Status:      wallet_entities.TransactionStatusCompleted,
		LedgerTxID:  &ledgerTxID,
		StartedAt:   now,
		CompletedAt: &now,
		Metadata:    cmd.Metadata,
	}, nil
}
