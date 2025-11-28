package payment_usecases

import (
	"context"

	payment_in "github.com/replay-api/replay-api/pkg/domain/payment/ports/in"
	payment_out "github.com/replay-api/replay-api/pkg/domain/payment/ports/out"
)

// PaymentQueryService aggregates all payment query use cases
type PaymentQueryService struct {
	getPayment      *GetPaymentUseCase
	getUserPayments *GetUserPaymentsUseCase
}

// NewPaymentQueryService creates a new payment query service
func NewPaymentQueryService(paymentRepo payment_out.PaymentRepository) *PaymentQueryService {
	return &PaymentQueryService{
		getPayment:      NewGetPaymentUseCase(paymentRepo),
		getUserPayments: NewGetUserPaymentsUseCase(paymentRepo),
	}
}

// GetPayment retrieves a payment by ID
func (s *PaymentQueryService) GetPayment(ctx context.Context, query payment_in.GetPaymentQuery) (*payment_in.PaymentDTO, error) {
	return s.getPayment.GetPayment(ctx, query)
}

// GetUserPayments retrieves all payments for a user
func (s *PaymentQueryService) GetUserPayments(ctx context.Context, query payment_in.GetUserPaymentsQuery) (*payment_in.PaymentsResult, error) {
	return s.getUserPayments.GetUserPayments(ctx, query)
}

// Ensure PaymentQueryService implements PaymentQuery interface
var _ payment_in.PaymentQuery = (*PaymentQueryService)(nil)
