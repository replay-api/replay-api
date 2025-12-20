package iam_use_cases

import (
	"context"
	"log/slog"

	common "github.com/replay-api/replay-api/pkg/domain"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	iam_out "github.com/replay-api/replay-api/pkg/domain/iam/ports/out"
)

// safeUintToInt converts uint to int with overflow protection.
// Returns 0 if the value exceeds int max value (should not happen in practice for pagination).
func safeUintToInt(val uint) int {
	const maxInt = int(^uint(0) >> 1) // Maximum value for int
	if val > uint(maxInt) {
		// In practice, pagination values should never exceed int max
		// Return max int as a safe fallback
		return maxInt
	}
	return int(val)
}

// SearchProfileUseCase handles profile search operations.
//
// This use case provides:
//   - Search profiles by various criteria (name, source key, etc.)
//   - Respects visibility and audience settings
//   - Pagination support for large result sets
//
// Use this for:
//   - Finding user profiles by search term
//   - Looking up profiles by external identifiers (Steam ID, Email)
//   - Admin user lookups
type SearchProfileUseCase struct {
	ProfileReader iam_out.ProfileReader
}

// NewSearchProfileUseCase creates a new SearchProfileUseCase instance.
func NewSearchProfileUseCase(profileReader iam_out.ProfileReader) *SearchProfileUseCase {
	return &SearchProfileUseCase{
		ProfileReader: profileReader,
	}
}

// SearchProfileResult contains paginated profile search results.
type SearchProfileResult struct {
	Profiles   []iam_entities.Profile
	TotalCount int
	Page       int
	PageSize   int
}

// Exec searches for profiles matching the provided criteria.
//
// Parameters:
//   - ctx: Context containing authentication and audience information
//   - search: Search parameters including filters, pagination, and visibility
//
// Returns:
//   - *SearchProfileResult: Paginated list of matching profiles
//   - error: Database or validation errors
func (uc *SearchProfileUseCase) Exec(ctx context.Context, search common.Search) (*SearchProfileResult, error) {
	slog.InfoContext(ctx, "searching profiles",
		"skip", search.ResultOptions.Skip,
		"limit", search.ResultOptions.Limit,
	)

	profiles, err := uc.ProfileReader.Search(ctx, search)
	if err != nil {
		slog.ErrorContext(ctx, "failed to search profiles", "error", err)
		return nil, err
	}

	slog.InfoContext(ctx, "profile search completed", "results_count", len(profiles))

	limit := search.ResultOptions.Limit
	if limit == 0 {
		limit = 1
	}
	page := safeUintToInt(search.ResultOptions.Skip / limit)

	return &SearchProfileResult{
		Profiles:   profiles,
		TotalCount: len(profiles),
		Page:       page,
		PageSize:   safeUintToInt(search.ResultOptions.Limit),
	}, nil
}
