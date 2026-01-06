package iam_use_cases

import (
	"context"
	"log/slog"

	shared "github.com/resource-ownership/go-common/pkg/common"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	iam_out "github.com/replay-api/replay-api/pkg/domain/iam/ports/out"
)

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
func (uc *SearchProfileUseCase) Exec(ctx context.Context, search shared.Search) (*SearchProfileResult, error) {
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
	skip := search.ResultOptions.Skip
	if skip > 9223372036854775807 { // int max (assuming 64-bit)
		skip = 9223372036854775807
	}
	if limit > 9223372036854775807 { // int max (assuming 64-bit)
		limit = 9223372036854775807
	}
	page := int(skip / limit)

	return &SearchProfileResult{
		Profiles:   profiles,
		TotalCount: len(profiles),
		Page:       page,
		PageSize:   int(limit),
	}, nil
}
