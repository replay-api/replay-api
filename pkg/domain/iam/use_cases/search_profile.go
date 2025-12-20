package iam_use_cases

import (
	"context"
	"fmt"

	common "github.com/replay-api/replay-api/pkg/domain"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	iam_in "github.com/replay-api/replay-api/pkg/domain/iam/ports/in"
)

// safeUintToInt converts uint to int with overflow protection
// Returns an error if the value exceeds int max value
func safeUintToInt(val uint) (int, error) {
	const maxInt = int(^uint(0) >> 1) // Maximum value for int
	if val > uint(maxInt) {
		return 0, fmt.Errorf("value %d exceeds maximum int value %d", val, maxInt)
	}
	return int(val), nil
}

// ProfileSearchResult represents the result of a profile search
type ProfileSearchResult struct {
	Profiles []iam_entities.Profile `json:"profiles"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"page_size"`
}

// SearchProfile searches for profiles based on search criteria
func SearchProfile(ctx context.Context, search *common.Search, reader iam_in.ProfileReader) (*ProfileSearchResult, error) {
	profiles, err := reader.Search(ctx, *search)
	if err != nil {
		return nil, err
	}

	limit := search.ResultOptions.Limit
	if limit == 0 {
		limit = common.DefaultPageSize
	}

	// Safe conversion from uint to int
	pageSize, err := safeUintToInt(limit)
	if err != nil {
		return nil, fmt.Errorf("invalid page size: %w", err)
	}

	skip := search.ResultOptions.Skip
	// Safe conversion from uint to int for page calculation
	skipInt, err := safeUintToInt(skip)
	if err != nil {
		return nil, fmt.Errorf("invalid skip value: %w", err)
	}

	limitInt, err := safeUintToInt(limit)
	if err != nil {
		return nil, fmt.Errorf("invalid limit value: %w", err)
	}

	page := 0
	if limitInt > 0 {
		page = skipInt / limitInt
	}

	return &ProfileSearchResult{
		Profiles: profiles,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
