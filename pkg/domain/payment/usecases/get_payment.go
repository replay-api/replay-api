package payment_usecases

import (
	"context"
	"log/slog"

	common "github.com/replay-api/replay-api/pkg/domain"
	payment_in "github.com/replay-api/replay-api/pkg/domain/payment/ports/in"
	payment_out "github.com/replay-api/replay-api/pkg/domain/payment/ports/out"
)

// GetPaymentUseCase handles fetching a single payment
type GetPaymentUseCase struct {
	paymentRepo payment_out.PaymentRepository
}

// NewGetPaymentUseCase creates a new get payment use case
func NewGetPaymentUseCase(paymentRepo payment_out.PaymentRepository) *GetPaymentUseCase {
	return &GetPaymentUseCase{
		paymentRepo: paymentRepo,
	}
}

// GetPayment retrieves a payment by ID
func (uc *GetPaymentUseCase) GetPayment(ctx context.Context, query payment_in.GetPaymentQuery) (*payment_in.PaymentDTO, error) {
	// Validate query
	if err := query.Validate(); err != nil {
		slog.WarnContext(ctx, "GetPayment: invalid query", "error", err)
		return nil, common.NewErrInvalidInput(err.Error())
	}

	// Auth check - user must be authenticated
	isAuthenticated := ctx.Value(common.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, common.NewErrUnauthorized()
	}

	slog.InfoContext(ctx, "GetPayment: fetching payment",
		"payment_id", query.PaymentID,
		"user_id", query.UserID,
	)

	// Fetch payment
	payment, err := uc.paymentRepo.FindByID(ctx, query.PaymentID)
	if err != nil {
		slog.WarnContext(ctx, "GetPayment: payment not found",
			"payment_id", query.PaymentID,
			"error", err,
		)
		return nil, common.NewErrNotFound(common.ResourceTypePayment, "id", query.PaymentID)
	}

	// Authorization check - user can only view their own payments
	if payment.UserID != query.UserID {
		slog.WarnContext(ctx, "GetPayment: unauthorized access attempt",
			"payment_id", query.PaymentID,
			"payment_user_id", payment.UserID,
			"requesting_user_id", query.UserID,
		)
		return nil, common.NewErrUnauthorized()
	}

	dto := payment_in.PaymentToDTO(payment)
	return &dto, nil
}
