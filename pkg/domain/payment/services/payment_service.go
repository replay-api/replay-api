// Package payment_services implements payment business logic
package payment_services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	payment_entities "github.com/replay-api/replay-api/pkg/domain/payment/entities"
	payment_in "github.com/replay-api/replay-api/pkg/domain/payment/ports/in"
	payment_out "github.com/replay-api/replay-api/pkg/domain/payment/ports/out"
	wallet_in "github.com/replay-api/replay-api/pkg/domain/wallet/ports/in"
)

// PaymentService implements payment business logic
type PaymentService struct {
	paymentRepo      payment_out.PaymentRepository
	providerAdapters map[payment_entities.PaymentProvider]payment_out.PaymentProviderAdapter
	walletService    wallet_in.WalletCommand
}

// NewPaymentService creates a new payment service
func NewPaymentService(
	paymentRepo payment_out.PaymentRepository,
	walletService wallet_in.WalletCommand,
	adapters ...payment_out.PaymentProviderAdapter,
) payment_in.PaymentCommand {
	adapterMap := make(map[payment_entities.PaymentProvider]payment_out.PaymentProviderAdapter)
	for _, adapter := range adapters {
		adapterMap[adapter.GetProvider()] = adapter
	}

	return &PaymentService{
		paymentRepo:      paymentRepo,
		providerAdapters: adapterMap,
		walletService:    walletService,
	}
}

// CreatePaymentIntent creates a new payment intent with the specified provider
func (s *PaymentService) CreatePaymentIntent(ctx context.Context, cmd payment_in.CreatePaymentIntentCommand) (*payment_in.PaymentIntentResult, error) {
	// Get the provider adapter
	adapter, ok := s.providerAdapters[cmd.Provider]
	if !ok {
		return nil, fmt.Errorf("unsupported payment provider: %s", cmd.Provider)
	}

	// Create payment entity
	payment := payment_entities.NewPayment(
		cmd.UserID,
		cmd.WalletID,
		cmd.PaymentType,
		cmd.Provider,
		cmd.Amount,
		cmd.Currency,
	)

	if cmd.Metadata != nil {
		payment.Metadata = cmd.Metadata
	}

	// Check for idempotency - if payment with same key exists, return it
	existingPayment, err := s.paymentRepo.FindByIdempotencyKey(ctx, payment.IdempotencyKey)
	if err == nil && existingPayment != nil {
		slog.InfoContext(ctx, "returning existing payment for idempotency key",
			"idempotency_key", payment.IdempotencyKey,
			"payment_id", existingPayment.ID)
		return &payment_in.PaymentIntentResult{
			Payment: existingPayment,
		}, nil
	}

	// Create payment intent with provider
	intentReq := payment_out.CreateIntentRequest{
		Amount:         cmd.Amount,
		Currency:       cmd.Currency,
		Description:    fmt.Sprintf("LeetGaming %s - %s", cmd.PaymentType, payment.ID.String()),
		Metadata:       cmd.Metadata,
		IdempotencyKey: payment.IdempotencyKey,
	}

	intentResp, err := adapter.CreatePaymentIntent(ctx, intentReq)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create payment intent with provider",
			"provider", cmd.Provider,
			"error", err)
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	// Update payment with provider details
	payment.ProviderPaymentID = intentResp.ProviderPaymentID
	payment.MarkProcessing()

	// Save payment to database
	if err := s.paymentRepo.Save(ctx, payment); err != nil {
		slog.ErrorContext(ctx, "failed to save payment",
			"payment_id", payment.ID,
			"error", err)
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	slog.InfoContext(ctx, "payment intent created",
		"payment_id", payment.ID,
		"provider_payment_id", intentResp.ProviderPaymentID,
		"provider", cmd.Provider)

	return &payment_in.PaymentIntentResult{
		Payment:       payment,
		ClientSecret:  intentResp.ClientSecret,
		RedirectURL:   intentResp.RedirectURL,
		CryptoAddress: intentResp.CryptoAddress,
	}, nil
}

// ConfirmPayment confirms a payment with the payment method
func (s *PaymentService) ConfirmPayment(ctx context.Context, cmd payment_in.ConfirmPaymentCommand) (*payment_entities.Payment, error) {
	// Get payment from database
	payment, err := s.paymentRepo.FindByID(ctx, cmd.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	// Get the provider adapter
	adapter, ok := s.providerAdapters[payment.Provider]
	if !ok {
		return nil, fmt.Errorf("unsupported payment provider: %s", payment.Provider)
	}

	// Confirm with provider
	confirmResp, err := adapter.ConfirmPayment(ctx, payment_out.ConfirmPaymentRequest{
		ProviderPaymentID: payment.ProviderPaymentID,
		PaymentMethodID:   cmd.PaymentMethodID,
	})
	if err != nil {
		payment.MarkFailed(err.Error())
		s.paymentRepo.Update(ctx, payment)
		return nil, fmt.Errorf("failed to confirm payment: %w", err)
	}

	// Update payment status based on response
	if confirmResp.Status == "succeeded" {
		payment.MarkSucceeded(confirmResp.ProviderFee)

		// Credit wallet for deposits
		if payment.Type == payment_entities.PaymentTypeDeposit {
			if err := s.creditWallet(ctx, payment); err != nil {
				slog.ErrorContext(ctx, "failed to credit wallet after successful payment",
					"payment_id", payment.ID,
					"error", err)
				// Payment succeeded but wallet credit failed - needs manual intervention
				payment.Metadata["wallet_credit_error"] = err.Error()
			}
		}
	}

	// Save updated payment
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	slog.InfoContext(ctx, "payment confirmed",
		"payment_id", payment.ID,
		"status", payment.Status)

	return payment, nil
}

// RefundPayment refunds a payment (full or partial)
func (s *PaymentService) RefundPayment(ctx context.Context, cmd payment_in.RefundPaymentCommand) (*payment_entities.Payment, error) {
	// Get payment from database
	payment, err := s.paymentRepo.FindByID(ctx, cmd.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	// Validate payment can be refunded
	if payment.Status != payment_entities.PaymentStatusSucceeded {
		return nil, fmt.Errorf("payment cannot be refunded in status: %s", payment.Status)
	}

	// Get the provider adapter
	adapter, ok := s.providerAdapters[payment.Provider]
	if !ok {
		return nil, fmt.Errorf("unsupported payment provider: %s", payment.Provider)
	}

	// Determine refund amount
	refundAmount := cmd.Amount
	if refundAmount == 0 {
		refundAmount = payment.Amount
	}

	// Process refund with provider
	refundResp, err := adapter.RefundPayment(ctx, payment_out.RefundRequest{
		ProviderPaymentID: payment.ProviderPaymentID,
		Amount:            refundAmount,
		Reason:            cmd.Reason,
		IdempotencyKey:    uuid.New().String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to process refund: %w", err)
	}

	// Update payment status
	payment.MarkRefunded()
	payment.Metadata["refund_id"] = refundResp.RefundID
	payment.Metadata["refund_amount"] = refundResp.Amount
	payment.Metadata["refund_reason"] = cmd.Reason

	// Save updated payment
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	// TODO: Debit wallet for refunded deposits

	slog.InfoContext(ctx, "payment refunded",
		"payment_id", payment.ID,
		"refund_amount", refundAmount)

	return payment, nil
}

// CancelPayment cancels a pending payment
func (s *PaymentService) CancelPayment(ctx context.Context, cmd payment_in.CancelPaymentCommand) (*payment_entities.Payment, error) {
	// Get payment from database
	payment, err := s.paymentRepo.FindByID(ctx, cmd.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	// Validate payment can be cancelled
	if !payment.IsPending() && payment.Status != payment_entities.PaymentStatusProcessing {
		return nil, fmt.Errorf("payment cannot be cancelled in status: %s", payment.Status)
	}

	// Get the provider adapter
	adapter, ok := s.providerAdapters[payment.Provider]
	if !ok {
		return nil, fmt.Errorf("unsupported payment provider: %s", payment.Provider)
	}

	// Cancel with provider
	_, err = adapter.CancelPayment(ctx, payment_out.CancelRequest{
		ProviderPaymentID: payment.ProviderPaymentID,
		Reason:            cmd.Reason,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to cancel payment: %w", err)
	}

	// Update payment status
	payment.MarkCanceled()

	// Save updated payment
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	slog.InfoContext(ctx, "payment cancelled",
		"payment_id", payment.ID)

	return payment, nil
}

// ProcessWebhook processes a webhook event from a payment provider
func (s *PaymentService) ProcessWebhook(ctx context.Context, cmd payment_in.ProcessWebhookCommand) error {
	// Get the provider adapter
	adapter, ok := s.providerAdapters[cmd.Provider]
	if !ok {
		return fmt.Errorf("unsupported payment provider: %s", cmd.Provider)
	}

	// Parse and validate webhook
	event, err := adapter.ParseWebhook(cmd.Payload, cmd.Signature)
	if err != nil {
		return fmt.Errorf("failed to parse webhook: %w", err)
	}

	// Find payment by provider payment ID
	payment, err := s.paymentRepo.FindByProviderPaymentID(ctx, event.ProviderPaymentID)
	if err != nil {
		slog.WarnContext(ctx, "payment not found for webhook event",
			"provider_payment_id", event.ProviderPaymentID,
			"event_type", event.EventType)
		return nil // Don't error - payment might not exist yet
	}

	// Update payment based on event
	switch event.Status {
	case payment_entities.PaymentStatusSucceeded:
		payment.MarkSucceeded(event.ProviderFee)

		// Credit wallet for deposits
		if payment.Type == payment_entities.PaymentTypeDeposit {
			if err := s.creditWallet(ctx, payment); err != nil {
				slog.ErrorContext(ctx, "failed to credit wallet from webhook",
					"payment_id", payment.ID,
					"error", err)
				payment.Metadata["wallet_credit_error"] = err.Error()
			}
		}

	case payment_entities.PaymentStatusFailed:
		payment.MarkFailed(event.FailureReason)

	case payment_entities.PaymentStatusCanceled:
		payment.MarkCanceled()

	case payment_entities.PaymentStatusRefunded:
		payment.MarkRefunded()

	case payment_entities.PaymentStatusProcessing:
		payment.MarkProcessing()
	}

	// Save updated payment
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return fmt.Errorf("failed to update payment from webhook: %w", err)
	}

	slog.InfoContext(ctx, "webhook processed",
		"payment_id", payment.ID,
		"event_type", event.EventType,
		"new_status", payment.Status)

	return nil
}

// creditWallet credits the user's wallet after successful deposit
func (s *PaymentService) creditWallet(ctx context.Context, payment *payment_entities.Payment) error {
	if s.walletService == nil {
		return fmt.Errorf("wallet service not configured")
	}

	// Get resource owner from context or payment
	resourceOwner := common.GetResourceOwner(ctx)
	if resourceOwner.UserID == uuid.Nil {
		resourceOwner.UserID = payment.UserID
	}

	// Credit wallet with net amount (after fees)
	depositCmd := wallet_in.DepositCommand{
		UserID:   payment.UserID,
		Amount:   float64(payment.NetAmount) / 100, // Convert from cents
		Currency: payment.Currency,
		TxHash:   payment.ProviderPaymentID,
	}

	return s.walletService.Deposit(ctx, depositCmd)
}

// Ensure PaymentService implements PaymentCommand
var _ payment_in.PaymentCommand = (*PaymentService)(nil)
