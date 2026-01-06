package payment_usecases

import (
	"context"
	"log/slog"

	shared "github.com/resource-ownership/go-common/pkg/common"
	payment_in "github.com/replay-api/replay-api/pkg/domain/payment/ports/in"
	payment_out "github.com/replay-api/replay-api/pkg/domain/payment/ports/out"
)

// GetUserPaymentsUseCase handles fetching payments for a user
type GetUserPaymentsUseCase struct {
	paymentRepo payment_out.PaymentRepository
}

// NewGetUserPaymentsUseCase creates a new get user payments use case
func NewGetUserPaymentsUseCase(paymentRepo payment_out.PaymentRepository) *GetUserPaymentsUseCase {
	return &GetUserPaymentsUseCase{
		paymentRepo: paymentRepo,
	}
}

// GetUserPayments retrieves all payments for a user
func (uc *GetUserPaymentsUseCase) GetUserPayments(ctx context.Context, query payment_in.GetUserPaymentsQuery) (*payment_in.PaymentsResult, error) {
	// Validate query
	if err := query.Validate(); err != nil {
		slog.WarnContext(ctx, "GetUserPayments: invalid query", "error", err)
		return nil, shared.NewErrInvalidInput(err.Error())
	}

	// Auth check - user must be authenticated
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, shared.NewErrUnauthorized()
	}

	slog.InfoContext(ctx, "GetUserPayments: fetching payments",
		"user_id", query.UserID,
		"limit", query.Filters.Limit,
		"offset", query.Filters.Offset,
	)

	// Convert query filters to repository filters
	repoFilters := payment_out.PaymentFilters{
		UserID:   &query.UserID,
		Provider: query.Filters.Provider,
		Status:   query.Filters.Status,
		Type:     query.Filters.Type,
		Limit:    query.Filters.Limit,
		Offset:   query.Filters.Offset,
	}

	// Fetch payments
	payments, err := uc.paymentRepo.FindByUserID(ctx, query.UserID, repoFilters)
	if err != nil {
		slog.ErrorContext(ctx, "GetUserPayments: failed to fetch payments",
			"user_id", query.UserID,
			"error", err,
		)
		return nil, shared.NewErrBadRequest("failed to fetch payments")
	}

	// Convert to DTOs
	dtos := make([]payment_in.PaymentDTO, len(payments))
	for i, p := range payments {
		dtos[i] = payment_in.PaymentToDTO(p)
	}

	// Note: For proper pagination, repository should return total count
	// For now, we estimate based on returned results
	totalCount := int64(len(payments))
	if len(payments) == query.Filters.Limit {
		// There might be more, estimate higher
		totalCount = int64(query.Filters.Offset + len(payments) + 1)
	}

	return &payment_in.PaymentsResult{
		Payments:   dtos,
		TotalCount: totalCount,
		Limit:      query.Filters.Limit,
		Offset:     query.Filters.Offset,
	}, nil
}
