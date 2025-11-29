package wallet_services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_out "github.com/replay-api/replay-api/pkg/domain/wallet/ports/out"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

// System account IDs for double-entry accounting
var (
	SystemLiabilityAccountID = uuid.MustParse("00000000-0000-0000-0000-000000000001") // Platform owes users
	SystemRevenueAccountID   = uuid.MustParse("00000000-0000-0000-0000-000000000002") // Platform earnings
	SystemExpenseAccountID   = uuid.MustParse("00000000-0000-0000-0000-000000000003") // Platform costs
)

// LedgerService implements double-entry accounting for all wallet operations
// Every transaction creates TWO ledger entries: one debit and one credit
// This ensures the accounting equation always balances: Assets = Liabilities + Equity
type LedgerService struct {
	ledgerRepo       wallet_out.LedgerRepository
	idempotencyRepo  wallet_out.IdempotencyRepository
}

// NewLedgerService creates a new ledger service
func NewLedgerService(
	ledgerRepo wallet_out.LedgerRepository,
	idempotencyRepo wallet_out.IdempotencyRepository,
) *LedgerService {
	return &LedgerService{
		ledgerRepo:      ledgerRepo,
		idempotencyRepo: idempotencyRepo,
	}
}

// RecordDeposit records a fiat deposit using double-entry accounting
// Debit: User's asset account (increases balance)
// Credit: Platform's liability account (we owe user)
func (s *LedgerService) RecordDeposit(
	ctx context.Context,
	walletID uuid.UUID,
	currency wallet_vo.Currency,
	amount wallet_vo.Amount,
	paymentID uuid.UUID,
	metadata wallet_entities.LedgerMetadata,
) (uuid.UUID, error) {
	idempotencyKey := fmt.Sprintf("deposit_%s_%s", paymentID.String(), walletID.String())

	// Check idempotency
	if s.ledgerRepo.ExistsByIdempotencyKey(ctx, idempotencyKey) {
		entry, _ := s.ledgerRepo.FindByIdempotencyKey(ctx, idempotencyKey)
		slog.InfoContext(ctx, "deposit already recorded (idempotent)",
			"wallet_id", walletID,
			"payment_id", paymentID,
			"transaction_id", entry.TransactionID)
		return entry.TransactionID, nil
	}

	txID := uuid.New()
	userID := common.GetResourceOwner(ctx).UserID

	metadata.OperationType = "Deposit"
	metadata.PaymentID = &paymentID

	// Entry 1: DEBIT user's asset account (increases balance)
	debitEntry := wallet_entities.NewLedgerEntry(
		txID,
		walletID,
		wallet_entities.AccountTypeAsset,
		wallet_entities.EntryTypeDebit,
		wallet_entities.AssetTypeFiat,
		amount,
		fmt.Sprintf("Deposit via payment %s", paymentID.String()),
		idempotencyKey,
		userID,
	)
	debitEntry.SetCurrency(currency)
	debitEntry.SetMetadata(metadata)

	// Entry 2: CREDIT platform's liability account (we owe user)
	creditEntry := wallet_entities.NewLedgerEntry(
		txID,
		SystemLiabilityAccountID,
		wallet_entities.AccountTypeLiability,
		wallet_entities.EntryTypeCredit,
		wallet_entities.AssetTypeFiat,
		amount,
		fmt.Sprintf("User deposit liability from wallet %s", walletID.String()),
		idempotencyKey+"_liability",
		userID,
	)
	creditEntry.SetCurrency(currency)
	creditEntry.SetMetadata(metadata)

	// Validate entries
	if err := debitEntry.Validate(); err != nil {
		return uuid.Nil, fmt.Errorf("debit entry validation failed: %w", err)
	}
	if err := creditEntry.Validate(); err != nil {
		return uuid.Nil, fmt.Errorf("credit entry validation failed: %w", err)
	}

	// Atomic write (both entries or none)
	if err := s.ledgerRepo.CreateTransaction(ctx, []*wallet_entities.LedgerEntry{
		debitEntry,
		creditEntry,
	}); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create deposit transaction: %w", err)
	}

	slog.InfoContext(ctx, "deposit recorded in ledger",
		"transaction_id", txID,
		"wallet_id", walletID,
		"amount", amount.String(),
		"currency", currency)

	return txID, nil
}

// RecordWithdrawal records a fiat withdrawal using double-entry accounting
// Credit: User's asset account (decreases balance)
// Debit: Platform's liability account (user no longer owed)
func (s *LedgerService) RecordWithdrawal(
	ctx context.Context,
	walletID uuid.UUID,
	currency wallet_vo.Currency,
	amount wallet_vo.Amount,
	toAddress string,
	metadata wallet_entities.LedgerMetadata,
) (uuid.UUID, error) {
	idempotencyKey := fmt.Sprintf("withdrawal_%s_%d", walletID.String(), time.Now().Unix())

	txID := uuid.New()
	userID := common.GetResourceOwner(ctx).UserID

	metadata.OperationType = "Withdrawal"

	// Entry 1: CREDIT user's asset account (decreases balance)
	creditEntry := wallet_entities.NewLedgerEntry(
		txID,
		walletID,
		wallet_entities.AccountTypeAsset,
		wallet_entities.EntryTypeCredit,
		wallet_entities.AssetTypeFiat,
		amount,
		fmt.Sprintf("Withdrawal to %s", toAddress),
		idempotencyKey,
		userID,
	)
	creditEntry.SetCurrency(currency)
	creditEntry.SetMetadata(metadata)

	// Entry 2: DEBIT platform's liability account (user no longer owed)
	debitEntry := wallet_entities.NewLedgerEntry(
		txID,
		SystemLiabilityAccountID,
		wallet_entities.AccountTypeLiability,
		wallet_entities.EntryTypeDebit,
		wallet_entities.AssetTypeFiat,
		amount,
		fmt.Sprintf("User withdrawal processed for wallet %s", walletID.String()),
		idempotencyKey+"_liability",
		userID,
	)
	debitEntry.SetCurrency(currency)
	debitEntry.SetMetadata(metadata)

	// Atomic write
	if err := s.ledgerRepo.CreateTransaction(ctx, []*wallet_entities.LedgerEntry{
		creditEntry,
		debitEntry,
	}); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create withdrawal transaction: %w", err)
	}

	slog.InfoContext(ctx, "withdrawal recorded in ledger",
		"transaction_id", txID,
		"wallet_id", walletID,
		"amount", amount.String())

	return txID, nil
}

// RecordPrizeWinning records a prize award
// Debit: User's asset account (receives prize)
// Credit: Platform's expense account (prize cost to platform)
func (s *LedgerService) RecordPrizeWinning(
	ctx context.Context,
	walletID uuid.UUID,
	currency wallet_vo.Currency,
	amount wallet_vo.Amount,
	matchID *uuid.UUID,
	tournamentID *uuid.UUID,
	metadata wallet_entities.LedgerMetadata,
) (uuid.UUID, error) {
	idempotencyKey := fmt.Sprintf("prize_%s_%s", walletID.String(), uuid.New().String())

	txID := uuid.New()
	userID := common.GetResourceOwner(ctx).UserID

	metadata.OperationType = "PrizeWinning"
	metadata.MatchID = matchID
	metadata.TournamentID = tournamentID

	// Entry 1: DEBIT user's asset account (receives prize)
	debitEntry := wallet_entities.NewLedgerEntry(
		txID,
		walletID,
		wallet_entities.AccountTypeAsset,
		wallet_entities.EntryTypeDebit,
		wallet_entities.AssetTypeFiat,
		amount,
		"Prize winnings",
		idempotencyKey,
		userID,
	)
	debitEntry.SetCurrency(currency)
	debitEntry.SetMetadata(metadata)

	// Entry 2: CREDIT platform's expense account (prize cost)
	creditEntry := wallet_entities.NewLedgerEntry(
		txID,
		SystemExpenseAccountID,
		wallet_entities.AccountTypeExpense,
		wallet_entities.EntryTypeCredit,
		wallet_entities.AssetTypeFiat,
		amount,
		fmt.Sprintf("Prize paid to wallet %s", walletID.String()),
		idempotencyKey+"_expense",
		userID,
	)
	creditEntry.SetCurrency(currency)
	creditEntry.SetMetadata(metadata)

	// Atomic write
	if err := s.ledgerRepo.CreateTransaction(ctx, []*wallet_entities.LedgerEntry{
		debitEntry,
		creditEntry,
	}); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create prize transaction: %w", err)
	}

	slog.InfoContext(ctx, "prize recorded in ledger",
		"transaction_id", txID,
		"wallet_id", walletID,
		"amount", amount.String())

	return txID, nil
}

// RecordEntryFee records a matchmaking/tournament entry fee
// Credit: User's asset account (pays fee)
// Debit: Platform's revenue account (platform earns fee)
func (s *LedgerService) RecordEntryFee(
	ctx context.Context,
	walletID uuid.UUID,
	currency wallet_vo.Currency,
	amount wallet_vo.Amount,
	matchID *uuid.UUID,
	tournamentID *uuid.UUID,
	metadata wallet_entities.LedgerMetadata,
) (uuid.UUID, error) {
	idempotencyKey := fmt.Sprintf("entry_fee_%s_%s", walletID.String(), uuid.New().String())

	txID := uuid.New()
	userID := common.GetResourceOwner(ctx).UserID

	metadata.OperationType = "EntryFee"
	metadata.MatchID = matchID
	metadata.TournamentID = tournamentID

	// Entry 1: CREDIT user's asset account (pays fee)
	creditEntry := wallet_entities.NewLedgerEntry(
		txID,
		walletID,
		wallet_entities.AccountTypeAsset,
		wallet_entities.EntryTypeCredit,
		wallet_entities.AssetTypeFiat,
		amount,
		"Entry fee payment",
		idempotencyKey,
		userID,
	)
	creditEntry.SetCurrency(currency)
	creditEntry.SetMetadata(metadata)

	// Entry 2: DEBIT platform's revenue account (platform earns)
	debitEntry := wallet_entities.NewLedgerEntry(
		txID,
		SystemRevenueAccountID,
		wallet_entities.AccountTypeRevenue,
		wallet_entities.EntryTypeDebit,
		wallet_entities.AssetTypeFiat,
		amount,
		fmt.Sprintf("Entry fee from wallet %s", walletID.String()),
		idempotencyKey+"_revenue",
		userID,
	)
	debitEntry.SetCurrency(currency)
	debitEntry.SetMetadata(metadata)

	// Atomic write
	if err := s.ledgerRepo.CreateTransaction(ctx, []*wallet_entities.LedgerEntry{
		creditEntry,
		debitEntry,
	}); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create entry fee transaction: %w", err)
	}

	slog.InfoContext(ctx, "entry fee recorded in ledger",
		"transaction_id", txID,
		"wallet_id", walletID,
		"amount", amount.String())

	return txID, nil
}

// RecordRefund records a refund
// Reverses the original transaction
func (s *LedgerService) RecordRefund(
	ctx context.Context,
	originalTxID uuid.UUID,
	reason string,
) (uuid.UUID, error) {
	// Get original entries
	entries, err := s.ledgerRepo.FindByTransactionID(ctx, originalTxID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to find original transaction: %w", err)
	}

	if len(entries) == 0 {
		return uuid.Nil, fmt.Errorf("original transaction not found")
	}

	userID := entries[0].CreatedBy
	reversalTxID := uuid.New()
	reversalEntries := make([]*wallet_entities.LedgerEntry, 0, len(entries))

	// Create reversal entries (opposite entry types)
	for _, entry := range entries {
		reversalEntry := entry.Reverse(reason, userID)
		reversalEntry.TransactionID = reversalTxID // Group all reversals together
		reversalEntries = append(reversalEntries, reversalEntry)
	}

	// Atomic write of all reversals
	if err := s.ledgerRepo.CreateTransaction(ctx, reversalEntries); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create refund transaction: %w", err)
	}

	// Mark originals as reversed
	for i, entry := range entries {
		_ = s.ledgerRepo.MarkAsReversed(ctx, entry.ID, reversalEntries[i].ID)
	}

	slog.InfoContext(ctx, "refund recorded in ledger",
		"reversal_tx_id", reversalTxID,
		"original_tx_id", originalTxID,
		"reason", reason)

	return reversalTxID, nil
}

// CalculateBalance calculates current balance from ledger
// For asset accounts: Balance = SUM(debits) - SUM(credits)
func (s *LedgerService) CalculateBalance(
	ctx context.Context,
	walletID uuid.UUID,
	currency wallet_vo.Currency,
) (wallet_vo.Amount, error) {
	return s.ledgerRepo.CalculateBalance(ctx, walletID, currency)
}

// VerifyBalance verifies that wallet balance matches ledger
func (s *LedgerService) VerifyBalance(
	ctx context.Context,
	walletID uuid.UUID,
	currency wallet_vo.Currency,
	walletBalance wallet_vo.Amount,
) error {
	ledgerBalance, err := s.CalculateBalance(ctx, walletID, currency)
	if err != nil {
		return fmt.Errorf("failed to calculate ledger balance: %w", err)
	}

	if !ledgerBalance.Equals(walletBalance) {
		return fmt.Errorf("balance mismatch: ledger=%s, wallet=%s, difference=%s",
			ledgerBalance.String(),
			walletBalance.String(),
			ledgerBalance.Subtract(walletBalance).String())
	}

	return nil
}

// GetTransactionHistory retrieves paginated transaction history
func (s *LedgerService) GetTransactionHistory(
	ctx context.Context,
	walletID uuid.UUID,
	filters wallet_out.HistoryFilters,
) ([]*wallet_entities.LedgerEntry, int64, error) {
	return s.ledgerRepo.GetAccountHistory(ctx, walletID, filters)
}

// ExecuteWithIdempotency executes an operation with idempotency protection
func (s *LedgerService) ExecuteWithIdempotency(
	ctx context.Context,
	idempotencyKey string,
	operationType string,
	operation func() (uuid.UUID, error),
) (uuid.UUID, error) {
	// Check if operation already exists
	existing, err := s.idempotencyRepo.FindByKey(ctx, idempotencyKey)
	if err == nil {
		// Operation already processed
		switch existing.Status {
		case wallet_entities.OperationStatusCompleted:
			slog.InfoContext(ctx, "operation already completed (idempotent)",
				"idempotency_key", idempotencyKey,
				"result_id", existing.ResultID)
			return *existing.ResultID, nil
		case wallet_entities.OperationStatusProcessing:
			return uuid.Nil, fmt.Errorf("operation already in progress")
		case wallet_entities.OperationStatusFailed:
			// Allow retry of failed operations
			break
		}
	}

	// Create new idempotent operation
	op := wallet_entities.NewIdempotentOperation(
		idempotencyKey,
		operationType,
		nil, // TODO: Store request payload if needed
	)
	s.idempotencyRepo.Create(ctx, op)

	// Execute operation
	resultID, err := operation()
	if err != nil {
		op.MarkFailed(err.Error())
		s.idempotencyRepo.Update(ctx, op)
		return uuid.Nil, err
	}

	// Mark as completed
	op.MarkCompleted(resultID, nil)
	s.idempotencyRepo.Update(ctx, op)

	return resultID, nil
}
