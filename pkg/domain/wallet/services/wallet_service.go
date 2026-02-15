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

// WalletEventPublisher defines the interface for publishing wallet events to Kafka
// This is a domain-layer port to avoid importing infrastructure (kafka) package directly
type WalletEventPublisher interface {
	PublishWalletCreated(ctx context.Context, walletID, userID uuid.UUID) error
	PublishWalletDeposit(ctx context.Context, walletID, userID uuid.UUID, amount float64, currency string, ledgerTxID uuid.UUID) error
	PublishWalletWithdrawal(ctx context.Context, walletID, userID uuid.UUID, amount float64, currency string, toAddress string, ledgerTxID uuid.UUID) error
	PublishWalletEntryFee(ctx context.Context, walletID, userID uuid.UUID, amount float64, currency string, matchID, tournamentID *uuid.UUID, ledgerTxID uuid.UUID) error
	PublishWalletPrize(ctx context.Context, walletID, userID uuid.UUID, amount float64, currency string, matchID, tournamentID *uuid.UUID, ledgerTxID uuid.UUID) error
	PublishWalletRefund(ctx context.Context, walletID, userID uuid.UUID, amount float64, currency string, reason string, ledgerTxID uuid.UUID) error
}

// WalletService implements wallet business logic with ledger integration
// All wallet operations are recorded in an immutable ledger for audit compliance
// Uses transaction coordinator for atomic operations with automatic rollback
// Publishes Kafka events after successful mutations for downstream consumers
type WalletService struct {
	walletRepo     wallet_out.WalletRepository
	walletQuerySvc *WalletQueryService
	coordinator    *TransactionCoordinator
	eventPublisher WalletEventPublisher // Optional — nil-safe for environments without Kafka
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

// NewWalletServiceWithEvents creates a WalletService with Kafka event publishing
func NewWalletServiceWithEvents(
	walletRepo wallet_out.WalletRepository,
	walletQuerySvc *WalletQueryService,
	coordinator *TransactionCoordinator,
	eventPublisher WalletEventPublisher,
) wallet_in.WalletCommand {
	return &WalletService{
		walletRepo:     walletRepo,
		walletQuerySvc: walletQuerySvc,
		coordinator:    coordinator,
		eventPublisher: eventPublisher,
	}
}

// publishEvent is a nil-safe helper that publishes events without blocking the main flow
func (s *WalletService) publishEvent(ctx context.Context, eventName string, fn func() error) {
	if s.eventPublisher == nil {
		return
	}
	if err := fn(); err != nil {
		slog.WarnContext(ctx, "failed to publish wallet event (non-blocking)",
			"event", eventName,
			"error", err)
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

	// Publish wallet created event (non-blocking)
	s.publishEvent(ctx, "wallet.created", func() error {
		return s.eventPublisher.PublishWalletCreated(ctx, wallet.ID, resourceOwner.UserID)
	})

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

	// Resolve contract address for on-chain tracking
	var contractAddr string
	if cmd.ChainID != 0 {
		chainID, _ := wallet_vo.ParseChainID(cmd.ChainID)
		contractAddr, _ = wallet_vo.ChainContractAddress(currency, chainID)
	}

	// Use transaction coordinator for atomic operation with automatic rollback
	ledgerTxID, err := s.coordinator.ExecuteDeposit(
		ctx,
		wallet,
		currency,
		amount,
		paymentID,
		wallet_entities.LedgerMetadata{
			OperationType:   "Deposit",
			ChainID:         cmd.ChainID,
			PaymentMethod:   cmd.PaymentMethod,
			ContractAddress: contractAddr,
			SourceIP:        cmd.SourceIP,
			UserAgent:       cmd.UserAgent,
		},
	)
	if err != nil {
		slog.ErrorContext(ctx, "deposit transaction failed",
			"wallet_id", wallet.ID,
			"amount", amount.String(),
			"error", err)
		return fmt.Errorf("deposit failed: %w", err)
	}

	wallet.AddPendingTransaction(paymentID) //nolint:errcheck // non-critical tracking

	slog.InfoContext(ctx, "deposit completed successfully",
		"wallet_id", wallet.ID,
		"amount", amount.String(),
		"currency", currency,
		"ledger_tx_id", ledgerTxID)

	// Publish deposit event (non-blocking)
	s.publishEvent(ctx, "wallet.deposit", func() error {
		return s.eventPublisher.PublishWalletDeposit(ctx, wallet.ID, cmd.UserID, cmd.Amount, cmd.Currency, ledgerTxID)
	})

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

	// Resolve contract address for on-chain tracking
	var contractAddr string
	if cmd.ChainID != 0 {
		chainID, _ := wallet_vo.ParseChainID(cmd.ChainID)
		contractAddr, _ = wallet_vo.ChainContractAddress(currency, chainID)
	}

	// Use transaction coordinator for atomic operation
	ledgerTxID, err := s.coordinator.ExecuteWithdrawal(
		ctx,
		wallet,
		currency,
		amount,
		cmd.ToAddress,
		wallet_entities.LedgerMetadata{
			OperationType:   "Withdrawal",
			ChainID:         cmd.ChainID,
			PaymentMethod:   cmd.PaymentMethod,
			ContractAddress: contractAddr,
			SourceIP:        cmd.SourceIP,
			UserAgent:       cmd.UserAgent,
		},
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

	// Publish withdrawal event (non-blocking)
	s.publishEvent(ctx, "wallet.withdrawal", func() error {
		return s.eventPublisher.PublishWalletWithdrawal(ctx, wallet.ID, cmd.UserID, cmd.Amount, cmd.Currency, cmd.ToAddress, ledgerTxID)
	})

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
		cmd.MatchID,
		cmd.TournamentID,
		wallet_entities.LedgerMetadata{
			OperationType: "EntryFee",
			MatchID:       cmd.MatchID,
			TournamentID:  cmd.TournamentID,
		},
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

	// Publish entry fee event (non-blocking)
	s.publishEvent(ctx, "wallet.entry_fee", func() error {
		return s.eventPublisher.PublishWalletEntryFee(ctx, wallet.ID, cmd.UserID, cmd.Amount, cmd.Currency, cmd.MatchID, cmd.TournamentID, ledgerTxID)
	})

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
		cmd.MatchID,
		cmd.TournamentID,
		maxDailyWinnings,
		wallet_entities.LedgerMetadata{
			OperationType: "PrizeWinning",
			MatchID:       cmd.MatchID,
			TournamentID:  cmd.TournamentID,
		},
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

	// Publish prize event (non-blocking)
	s.publishEvent(ctx, "wallet.prize", func() error {
		return s.eventPublisher.PublishWalletPrize(ctx, wallet.ID, cmd.UserID, cmd.Amount, cmd.Currency, cmd.MatchID, cmd.TournamentID, ledgerTxID)
	})

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

	// Record refund as deposit using transaction coordinator
	// NOTE: ExecuteDeposit already updates the wallet balance and persists it
	// via the saga pattern (ledger → wallet.Deposit → walletRepo.Update)
	// Do NOT call wallet.Deposit or walletRepo.Update again — that would double-credit.
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
		return fmt.Errorf("failed to process refund: %w", err)
	}

	slog.InfoContext(ctx, "refund completed successfully",
		"wallet_id", wallet.ID,
		"amount", amount.String(),
		"currency", currency,
		"reason", cmd.Reason,
		"ledger_tx_id", ledgerTxID)

	// Publish refund event (non-blocking)
	s.publishEvent(ctx, "wallet.refund", func() error {
		return s.eventPublisher.PublishWalletRefund(ctx, wallet.ID, cmd.UserID, cmd.Amount, cmd.Currency, cmd.Reason, ledgerTxID)
	})

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
