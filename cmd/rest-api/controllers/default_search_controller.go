package controllers

import (
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	common "github.com/replay-api/replay-api/pkg/domain"
)

type DefaultSearchController[T any] struct {
	common.Searchable[T]
}

func NewDefaultSearchController[T any](service common.Searchable[T]) *DefaultSearchController[T] {
	return &DefaultSearchController[T]{
		service,
	}
}

func (c *DefaultSearchController[T]) DefaultSearchHandler(w http.ResponseWriter, r *http.Request) {
	var s common.Search

	// Try x-search header first (legacy method)
	base64SearchHeader := r.Header.Get("x-search")
	if base64SearchHeader != "" {
		searchJSON, err := base64.StdEncoding.DecodeString(base64SearchHeader)
		if err != nil {
			slog.Error("(DefaultSearchHandler) Error decoding search source base64 header", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = json.Unmarshal([]byte(searchJSON), &s)
		if err != nil {
			slog.Error("(DefaultSearchHandler) Error unmarshalling search source", "error", err, "searchJSON", searchJSON, "base64SearchHeader", base64SearchHeader)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		// Fallback: Build search from query parameters
		s = c.buildSearchFromQueryParams(r)
	}

	compiledSearch, err := c.Compile(r.Context(), s.SearchParams, s.ResultOptions)
	if err != nil {
		slog.Error("(DefaultSearchHandler) Error validating search request", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if compiledSearch == nil {
		slog.Error("(DefaultSearchHandler) Error validating search request", "error", "search is nil")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	results, err := c.Search(r.Context(), *compiledSearch)
	if err != nil {
		slog.Error("(DefaultSearchHandler) Error filtering search request", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(results)
}

// buildSearchFromQueryParams builds a Search object from URL query parameters
// This allows the frontend SDK to use query params instead of the x-search header
func (c *DefaultSearchController[T]) buildSearchFromQueryParams(r *http.Request) common.Search {
	query := r.URL.Query()
	var valueParams []common.SearchableValue

	// Common search parameters - build SearchableValue for each filter
	if gameID := query.Get("game_id"); gameID != "" {
		valueParams = append(valueParams, common.SearchableValue{
			Field:    "GameID",
			Values:   []interface{}{gameID},
			Operator: common.EqualsOperator,
		})
	}

	// Full-text search on name field using Contains operator
	if q := query.Get("q"); q != "" {
		valueParams = append(valueParams, common.SearchableValue{
			Field:    "FullName",
			Values:   []interface{}{q},
			Operator: common.ContainsOperator,
		})
	}

	// Exact or contains match on name
	if name := query.Get("name"); name != "" {
		valueParams = append(valueParams, common.SearchableValue{
			Field:    "FullName",
			Values:   []interface{}{name},
			Operator: common.ContainsOperator,
		})
	}

	// Nickname/ShortName search
	if nickname := query.Get("nickname"); nickname != "" {
		valueParams = append(valueParams, common.SearchableValue{
			Field:    "ShortName",
			Values:   []interface{}{nickname},
			Operator: common.ContainsOperator,
		})
	}

	// Build SearchAggregation from value params
	var searchParams []common.SearchAggregation
	if len(valueParams) > 0 {
		searchParams = []common.SearchAggregation{
			{
				Params: []common.SearchParameter{
					{
						ValueParams:       valueParams,
						AggregationClause: common.AndAggregationClause,
					},
				},
				AggregationClause: common.AndAggregationClause,
			},
		}
	}

	// Result options
	var skip uint = 0
	var limit uint = 50

	if limitStr := query.Get("limit"); limitStr != "" {
		if l, err := strconv.ParseUint(limitStr, 10, 32); err == nil {
			limit = uint(l)
		}
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		if o, err := strconv.ParseUint(offsetStr, 10, 32); err == nil {
			skip = uint(o)
		}
	}

	if pageStr := query.Get("page"); pageStr != "" {
		if p, err := strconv.ParseUint(pageStr, 10, 32); err == nil && p > 0 {
			skip = uint((p - 1)) * limit
		}
	}

	resultOptions := common.SearchResultOptions{
		Skip:  skip,
		Limit: limit,
	}

	// Sort options
	var sortOptions []common.SortableField
	if sort := query.Get("sort"); sort != "" {
		direction := common.AscendingIDKey
		if order := query.Get("order"); order == "desc" {
			direction = common.DescendingIDKey
		}
		sortOptions = append(sortOptions, common.SortableField{
			Field:     sort,
			Direction: direction,
		})
	}

	return common.Search{
		SearchParams:  searchParams,
		ResultOptions: resultOptions,
		SortOptions:   sortOptions,
	}
}
