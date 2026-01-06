package billing_usecases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	billing_out "github.com/replay-api/replay-api/pkg/domain/billing/ports/out"
)

// WithdrawalFeeConfig holds fee configuration for withdrawals
type WithdrawalFeeConfig struct {
	FixedFee    float64 // Fixed fee per withdrawal
	PercentFee  float64 // Percentage fee (0.02 = 2%)
	MinFee      float64 // Minimum fee
	MaxFee      float64 // Maximum fee (0 = no max)
	MinAmount   float64 // Minimum withdrawal amount
	MaxAmount   float64 // Maximum withdrawal amount (0 = no max)
}

// DefaultWithdrawalFeeConfig provides default fee settings
var DefaultWithdrawalFeeConfig = WithdrawalFeeConfig{
	FixedFee:   0.50,      // $0.50 fixed fee
	PercentFee: 0.02,      // 2% percentage fee
	MinFee:     0.50,      // Minimum $0.50 fee
	MaxFee:     50.00,     // Maximum $50 fee
	MinAmount:  10.00,     // Minimum $10 withdrawal
	MaxAmount:  10000.00,  // Maximum $10,000 withdrawal
}

// WithdrawalUseCase implements withdrawal operations
type WithdrawalUseCase struct {
	withdrawalRepo billing_out.WithdrawalRepository
	walletReader   billing_out.WalletReader
	walletDebiter  billing_out.WalletDebiter
	feeConfig      WithdrawalFeeConfig
}

// NewWithdrawalUseCase creates a new WithdrawalUseCase
func NewWithdrawalUseCase(
	withdrawalRepo billing_out.WithdrawalRepository,
	walletReader billing_out.WalletReader,
	walletDebiter billing_out.WalletDebiter,
) *WithdrawalUseCase {
	return &WithdrawalUseCase{
		withdrawalRepo: withdrawalRepo,
		walletReader:   walletReader,
		walletDebiter:  walletDebiter,
		feeConfig:      DefaultWithdrawalFeeConfig,
	}
}

// Create creates a new withdrawal request
func (uc *WithdrawalUseCase) Create(ctx context.Context, cmd billing_in.CreateWithdrawalCommand) (*billing_entities.Withdrawal, error) {
	slog.InfoContext(ctx, "Creating withdrawal request",
		"user_id", cmd.UserID,
		"amount", cmd.Amount,
		"currency", cmd.Currency,
		"method", cmd.Method,
	)

	// Validate user is authenticated
	resourceOwner := shared.GetResourceOwner(ctx)
	if resourceOwner.UserID == uuid.Nil {
		return nil, shared.NewErrUnauthorized()
	}

	// Validate user can only create withdrawals for themselves
	if cmd.UserID != resourceOwner.UserID {
		slog.WarnContext(ctx, "User attempted to create withdrawal for another user",
			"requesting_user", resourceOwner.UserID,
			"target_user", cmd.UserID,
		)
		return nil, shared.NewErrForbidden("cannot create withdrawal for another user")
	}

	// Validate amount limits
	if cmd.Amount < uc.feeConfig.MinAmount {
		return nil, fmt.Errorf("minimum withdrawal amount is %.2f %s", uc.feeConfig.MinAmount, cmd.Currency)
	}
	if uc.feeConfig.MaxAmount > 0 && cmd.Amount > uc.feeConfig.MaxAmount {
		return nil, fmt.Errorf("maximum withdrawal amount is %.2f %s", uc.feeConfig.MaxAmount, cmd.Currency)
	}

	// Calculate fee
	fee := uc.calculateFee(cmd.Amount)

	// Check wallet balance
	balance, err := uc.walletReader.GetBalance(ctx, cmd.UserID, cmd.Currency)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get wallet balance", "error", err)
		return nil, fmt.Errorf("failed to verify wallet balance: %w", err)
	}

	if balance < cmd.Amount {
		return nil, fmt.Errorf("insufficient balance: have %.2f %s, need %.2f %s",
			balance, cmd.Currency, cmd.Amount, cmd.Currency)
	}

	// Create withdrawal entity
	withdrawal := billing_entities.NewWithdrawal(
		cmd.UserID,
		cmd.WalletID,
		cmd.Amount,
		cmd.Currency,
		cmd.Method,
		cmd.BankDetails,
		fee,
		resourceOwner,
	)

	// Validate withdrawal
	if err := withdrawal.Validate(); err != nil {
		return nil, fmt.Errorf("invalid withdrawal: %w", err)
	}

	// Reserve funds (debit wallet)
	debitRef := fmt.Sprintf("withdrawal:%s", withdrawal.ID.String())
	if err := uc.walletDebiter.Debit(ctx, cmd.WalletID, cmd.Amount, debitRef); err != nil {
		slog.ErrorContext(ctx, "Failed to debit wallet for withdrawal", "error", err)
		return nil, fmt.Errorf("failed to reserve funds: %w", err)
	}

	// Persist withdrawal
	created, err := uc.withdrawalRepo.Create(ctx, withdrawal)
	if err != nil {
		// Rollback wallet debit
		_ = uc.walletDebiter.Credit(ctx, cmd.WalletID, cmd.Amount, debitRef+":rollback")
		slog.ErrorContext(ctx, "Failed to create withdrawal, rolling back wallet debit", "error", err)
		return nil, fmt.Errorf("failed to create withdrawal: %w", err)
	}

	slog.InfoContext(ctx, "Withdrawal created successfully",
		"withdrawal_id", created.ID,
		"amount", created.Amount,
		"fee", created.Fee,
		"net_amount", created.NetAmount,
	)

	return created, nil
}

// Cancel cancels a pending withdrawal
func (uc *WithdrawalUseCase) Cancel(ctx context.Context, withdrawalID uuid.UUID) (*billing_entities.Withdrawal, error) {
	resourceOwner := shared.GetResourceOwner(ctx)
	if resourceOwner.UserID == uuid.Nil {
		return nil, shared.NewErrUnauthorized()
	}

	withdrawal, err := uc.withdrawalRepo.GetByID(ctx, withdrawalID)
	if err != nil {
		return nil, fmt.Errorf("withdrawal not found: %w", err)
	}

	// Validate ownership
	if withdrawal.UserID != resourceOwner.UserID {
		return nil, shared.NewErrForbidden("cannot cancel another user's withdrawal")
	}

	// Check if cancelable
	if !withdrawal.IsCancelable() {
		return nil, fmt.Errorf("withdrawal cannot be canceled in status: %s", withdrawal.Status)
	}

	// Cancel and refund
	withdrawal.Cancel()

	// Return funds to wallet
	creditRef := fmt.Sprintf("withdrawal-cancel:%s", withdrawal.ID.String())
	if err := uc.walletDebiter.Credit(ctx, withdrawal.WalletID, withdrawal.Amount, creditRef); err != nil {
		slog.ErrorContext(ctx, "Failed to refund wallet after cancellation", "error", err)
		// Still update the withdrawal status, but log the error
	}

	updated, err := uc.withdrawalRepo.Update(ctx, withdrawal)
	if err != nil {
		return nil, fmt.Errorf("failed to update withdrawal: %w", err)
	}

	slog.InfoContext(ctx, "Withdrawal canceled", "withdrawal_id", withdrawalID)
	return updated, nil
}

// GetByID retrieves a withdrawal by ID
func (uc *WithdrawalUseCase) GetByID(ctx context.Context, withdrawalID uuid.UUID) (*billing_entities.Withdrawal, error) {
	resourceOwner := shared.GetResourceOwner(ctx)

	withdrawal, err := uc.withdrawalRepo.GetByID(ctx, withdrawalID)
	if err != nil {
		return nil, err
	}

	// Only owner or admin can view
	if withdrawal.UserID != resourceOwner.UserID && !shared.IsAdmin(ctx) {
		return nil, shared.NewErrForbidden("cannot view another user's withdrawal")
	}

	return withdrawal, nil
}

// GetByUserID retrieves all withdrawals for a user
func (uc *WithdrawalUseCase) GetByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]billing_entities.Withdrawal, error) {
	resourceOwner := shared.GetResourceOwner(ctx)

	// Only owner or admin can view
	if userID != resourceOwner.UserID && !shared.IsAdmin(ctx) {
		return nil, shared.NewErrForbidden("cannot view another user's withdrawals")
	}

	return uc.withdrawalRepo.GetByUserID(ctx, userID, limit, offset)
}

// calculateFee calculates the withdrawal fee
func (uc *WithdrawalUseCase) calculateFee(amount float64) float64 {
	fee := uc.feeConfig.FixedFee + (amount * uc.feeConfig.PercentFee)

	if fee < uc.feeConfig.MinFee {
		fee = uc.feeConfig.MinFee
	}
	if uc.feeConfig.MaxFee > 0 && fee > uc.feeConfig.MaxFee {
		fee = uc.feeConfig.MaxFee
	}

	return fee
}

