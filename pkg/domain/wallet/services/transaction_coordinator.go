package wallet_services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_out "github.com/replay-api/replay-api/pkg/domain/wallet/ports/out"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

// TransactionCoordinator implements the Saga pattern for distributed wallet transactions
// Ensures atomicity across ledger and wallet updates with automatic rollback on failure
// This is CRITICAL for financial integrity - prevents money from being lost or duplicated
type TransactionCoordinator struct {
	walletRepo    wallet_out.WalletRepository
	ledgerService *LedgerService
	// TODO: Add distributed lock service for multi-instance deployments
}

// NewTransactionCoordinator creates a new transaction coordinator
func NewTransactionCoordinator(
	walletRepo wallet_out.WalletRepository,
	ledgerService *LedgerService,
) *TransactionCoordinator {
	return &TransactionCoordinator{
		walletRepo:    walletRepo,
		ledgerService: ledgerService,
	}
}

// TransactionStep represents a single step in the saga
type TransactionStep struct {
	Name     string
	Execute  func(ctx context.Context) error
	Rollback func(ctx context.Context) error
}

// SagaExecutor executes a series of transaction steps with automatic rollback
type SagaExecutor struct {
	steps          []TransactionStep
	executedSteps  []TransactionStep
	coordinator    *TransactionCoordinator
}

// NewSagaExecutor creates a new saga executor
func (c *TransactionCoordinator) NewSaga() *SagaExecutor {
	return &SagaExecutor{
		steps:          []TransactionStep{},
		executedSteps:  []TransactionStep{},
		coordinator:    c,
	}
}

// AddStep adds a step to the saga
func (s *SagaExecutor) AddStep(step TransactionStep) *SagaExecutor {
	s.steps = append(s.steps, step)
	return s
}

// Execute runs all steps in order, rolling back on any failure
func (s *SagaExecutor) Execute(ctx context.Context) error {
	for i, step := range s.steps {
		slog.DebugContext(ctx, "executing saga step",
			"step_name", step.Name,
			"step_index", i,
			"total_steps", len(s.steps))

		if err := step.Execute(ctx); err != nil {
			slog.ErrorContext(ctx, "saga step failed, initiating rollback",
				"step_name", step.Name,
				"step_index", i,
				"error", err)

			// Rollback all executed steps in reverse order
			if rollbackErr := s.rollback(ctx); rollbackErr != nil {
				// CRITICAL: Rollback failed - requires manual intervention
				slog.ErrorContext(ctx, "CRITICAL: saga rollback failed, manual intervention required",
					"original_error", err,
					"rollback_error", rollbackErr,
					"executed_steps", len(s.executedSteps))

				return fmt.Errorf("transaction failed and rollback failed: %w (rollback error: %v)", err, rollbackErr)
			}

			return fmt.Errorf("transaction rolled back due to: %w", err)
		}

		s.executedSteps = append(s.executedSteps, step)
	}

	slog.DebugContext(ctx, "saga completed successfully",
		"total_steps", len(s.steps))

	return nil
}

// rollback executes rollback functions in reverse order
func (s *SagaExecutor) rollback(ctx context.Context) error {
	// Rollback in reverse order
	for i := len(s.executedSteps) - 1; i >= 0; i-- {
		step := s.executedSteps[i]

		if step.Rollback == nil {
			slog.WarnContext(ctx, "step has no rollback function, skipping",
				"step_name", step.Name,
				"step_index", i)
			continue
		}

		slog.DebugContext(ctx, "rolling back step",
			"step_name", step.Name,
			"step_index", i)

		if err := step.Rollback(ctx); err != nil {
			return fmt.Errorf("failed to rollback step %s: %w", step.Name, err)
		}
	}

	return nil
}

// ExecuteDeposit performs atomic deposit: ledger + wallet update with rollback
func (c *TransactionCoordinator) ExecuteDeposit(
	ctx context.Context,
	wallet *wallet_entities.UserWallet,
	currency wallet_vo.Currency,
	amount wallet_vo.Amount,
	paymentID uuid.UUID,
	metadata wallet_entities.LedgerMetadata,
) (ledgerTxID uuid.UUID, err error) {
	var tempLedgerTxID uuid.UUID

	saga := c.NewSaga()

	// Step 1: Record in ledger
	saga.AddStep(TransactionStep{
		Name: "RecordDepositLedger",
		Execute: func(ctx context.Context) error {
			txID, err := c.ledgerService.RecordDeposit(
				ctx,
				wallet.ID,
				currency,
				amount,
				paymentID,
				metadata,
			)
			if err != nil {
				return fmt.Errorf("failed to record deposit in ledger: %w", err)
			}
			tempLedgerTxID = txID
			return nil
		},
		Rollback: func(ctx context.Context) error {
			if tempLedgerTxID == uuid.Nil {
				return nil // Nothing to rollback
			}

			// Reverse the ledger entry
			_, err := c.ledgerService.RecordRefund(
				ctx,
				tempLedgerTxID,
				"Automatic rollback: wallet update failed",
			)
			if err != nil {
				return fmt.Errorf("failed to reverse ledger entry: %w", err)
			}

			slog.WarnContext(ctx, "ledger entry reversed due to wallet update failure",
				"ledger_tx_id", tempLedgerTxID,
				"wallet_id", wallet.ID)

			return nil
		},
	})

	// Step 2: Update wallet
	saga.AddStep(TransactionStep{
		Name: "UpdateWalletBalance",
		Execute: func(ctx context.Context) error {
			if err := wallet.Deposit(currency, amount); err != nil {
				return fmt.Errorf("failed to update wallet balance: %w", err)
			}
			return nil
		},
		Rollback: func(ctx context.Context) error {
			// Reverse wallet balance change
			if err := wallet.Withdraw(currency, amount); err != nil {
				return fmt.Errorf("failed to reverse wallet balance: %w", err)
			}
			return nil
		},
	})

	// Step 3: Persist wallet
	saga.AddStep(TransactionStep{
		Name: "PersistWallet",
		Execute: func(ctx context.Context) error {
			if err := c.walletRepo.Update(ctx, wallet); err != nil {
				return fmt.Errorf("failed to persist wallet: %w", err)
			}
			return nil
		},
		Rollback: func(ctx context.Context) error {
			// Reload wallet from DB to revert in-memory changes
			// Note: This assumes the DB update succeeded but we're rolling back due to a later step
			// In practice, if DB update fails, we don't need to rollback DB
			return nil
		},
	})

	// Execute saga
	if err := saga.Execute(ctx); err != nil {
		return uuid.Nil, err
	}

	return tempLedgerTxID, nil
}

// ExecuteWithdrawal performs atomic withdrawal: ledger + wallet update with rollback
func (c *TransactionCoordinator) ExecuteWithdrawal(
	ctx context.Context,
	wallet *wallet_entities.UserWallet,
	currency wallet_vo.Currency,
	amount wallet_vo.Amount,
	toAddress string,
	metadata wallet_entities.LedgerMetadata,
) (ledgerTxID uuid.UUID, err error) {
	var tempLedgerTxID uuid.UUID

	saga := c.NewSaga()

	// Step 1: Record in ledger
	saga.AddStep(TransactionStep{
		Name: "RecordWithdrawalLedger",
		Execute: func(ctx context.Context) error {
			txID, err := c.ledgerService.RecordWithdrawal(
				ctx,
				wallet.ID,
				currency,
				amount,
				toAddress,
				metadata,
			)
			if err != nil {
				return fmt.Errorf("failed to record withdrawal in ledger: %w", err)
			}
			tempLedgerTxID = txID
			return nil
		},
		Rollback: func(ctx context.Context) error {
			if tempLedgerTxID == uuid.Nil {
				return nil
			}

			_, err := c.ledgerService.RecordRefund(
				ctx,
				tempLedgerTxID,
				"Automatic rollback: wallet update failed",
			)
			return err
		},
	})

	// Step 2: Update wallet
	saga.AddStep(TransactionStep{
		Name: "UpdateWalletBalance",
		Execute: func(ctx context.Context) error {
			if err := wallet.Withdraw(currency, amount); err != nil {
				return fmt.Errorf("failed to update wallet balance: %w", err)
			}
			return nil
		},
		Rollback: func(ctx context.Context) error {
			if err := wallet.Deposit(currency, amount); err != nil {
				return fmt.Errorf("failed to reverse wallet balance: %w", err)
			}
			return nil
		},
	})

	// Step 3: Persist wallet
	saga.AddStep(TransactionStep{
		Name: "PersistWallet",
		Execute: func(ctx context.Context) error {
			return c.walletRepo.Update(ctx, wallet)
		},
		Rollback: nil,
	})

	if err := saga.Execute(ctx); err != nil {
		return uuid.Nil, err
	}

	return tempLedgerTxID, nil
}

// ExecuteEntryFee performs atomic entry fee deduction with rollback
func (c *TransactionCoordinator) ExecuteEntryFee(
	ctx context.Context,
	wallet *wallet_entities.UserWallet,
	currency wallet_vo.Currency,
	amount wallet_vo.Amount,
	matchID *uuid.UUID,
	tournamentID *uuid.UUID,
	metadata wallet_entities.LedgerMetadata,
) (ledgerTxID uuid.UUID, err error) {
	var tempLedgerTxID uuid.UUID

	saga := c.NewSaga()

	saga.AddStep(TransactionStep{
		Name: "RecordEntryFeeLedger",
		Execute: func(ctx context.Context) error {
			txID, err := c.ledgerService.RecordEntryFee(
				ctx,
				wallet.ID,
				currency,
				amount,
				matchID,
				tournamentID,
				metadata,
			)
			if err != nil {
				return err
			}
			tempLedgerTxID = txID
			return nil
		},
		Rollback: func(ctx context.Context) error {
			if tempLedgerTxID == uuid.Nil {
				return nil
			}
			_, err := c.ledgerService.RecordRefund(
				ctx,
				tempLedgerTxID,
				"Automatic rollback: wallet update failed",
			)
			return err
		},
	})

	saga.AddStep(TransactionStep{
		Name: "DeductEntryFee",
		Execute: func(ctx context.Context) error {
			return wallet.DeductEntryFee(currency, amount)
		},
		Rollback: func(ctx context.Context) error {
			return wallet.Deposit(currency, amount)
		},
	})

	saga.AddStep(TransactionStep{
		Name: "PersistWallet",
		Execute: func(ctx context.Context) error {
			return c.walletRepo.Update(ctx, wallet)
		},
		Rollback: nil,
	})

	if err := saga.Execute(ctx); err != nil {
		return uuid.Nil, err
	}

	return tempLedgerTxID, nil
}

// ExecutePrizeWinning performs atomic prize addition with rollback
func (c *TransactionCoordinator) ExecutePrizeWinning(
	ctx context.Context,
	wallet *wallet_entities.UserWallet,
	currency wallet_vo.Currency,
	amount wallet_vo.Amount,
	matchID *uuid.UUID,
	tournamentID *uuid.UUID,
	maxDailyWinnings wallet_vo.Amount,
	metadata wallet_entities.LedgerMetadata,
) (ledgerTxID uuid.UUID, err error) {
	var tempLedgerTxID uuid.UUID

	saga := c.NewSaga()

	saga.AddStep(TransactionStep{
		Name: "RecordPrizeLedger",
		Execute: func(ctx context.Context) error {
			txID, err := c.ledgerService.RecordPrizeWinning(
				ctx,
				wallet.ID,
				currency,
				amount,
				matchID,
				tournamentID,
				metadata,
			)
			if err != nil {
				return err
			}
			tempLedgerTxID = txID
			return nil
		},
		Rollback: func(ctx context.Context) error {
			if tempLedgerTxID == uuid.Nil {
				return nil
			}
			_, err := c.ledgerService.RecordRefund(
				ctx,
				tempLedgerTxID,
				"Automatic rollback: wallet update failed",
			)
			return err
		},
	})

	saga.AddStep(TransactionStep{
		Name: "AddPrizeToWallet",
		Execute: func(ctx context.Context) error {
			return wallet.AddPrize(currency, amount, maxDailyWinnings)
		},
		Rollback: func(ctx context.Context) error {
			return wallet.Withdraw(currency, amount)
		},
	})

	saga.AddStep(TransactionStep{
		Name: "PersistWallet",
		Execute: func(ctx context.Context) error {
			return c.walletRepo.Update(ctx, wallet)
		},
		Rollback: nil,
	})

	if err := saga.Execute(ctx); err != nil {
		return uuid.Nil, err
	}

	return tempLedgerTxID, nil
}

// WalletTransaction represents a coordinated wallet transaction
type WalletTransaction struct {
	ID           uuid.UUID                    `json:"id"`
	WalletID     uuid.UUID                    `json:"wallet_id"`
	Type         string                       `json:"type"` // Deposit, Withdrawal, EntryFee, Prize
	Status       TransactionStatus            `json:"status"`
	LedgerTxID   *uuid.UUID                   `json:"ledger_tx_id,omitempty"`
	StartedAt    time.Time                    `json:"started_at"`
	CompletedAt  *time.Time                   `json:"completed_at,omitempty"`
	ErrorMessage string                       `json:"error_message,omitempty"`
	Metadata     map[string]interface{}       `json:"metadata"`
}

// TransactionStatus represents the status of a coordinated transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "Pending"
	TransactionStatusCompleted TransactionStatus = "Completed"
	TransactionStatusFailed    TransactionStatus = "Failed"
	TransactionStatusRolledBack TransactionStatus = "RolledBack"
)
