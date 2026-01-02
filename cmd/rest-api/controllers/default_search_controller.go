package controllers

import (
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

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
//
// Generic query parameters supported:
// - q: Text search value (used with search_fields)
// - search_fields: Comma-separated list of fields to search with 'q' (OR logic)
// - Any other query param is treated as an exact filter (field name = param key)
// - limit, offset, page: Pagination
// - sort, order: Sorting
//
// The SDK is responsible for sending correct PascalCase field names that match
// the Go struct fields. The validation of which fields are queryable is handled
// by the Compile method which checks against the queryableFields defined in each
// entity's query service.
func (c *DefaultSearchController[T]) buildSearchFromQueryParams(r *http.Request) common.Search {
	query := r.URL.Query()
	var exactFilters []common.SearchableValue
	var textSearchParams []common.SearchableValue

	// Reserved params that are not filters
	reservedParams := map[string]bool{
		"q": true, "search_fields": true,
		"limit": true, "offset": true, "page": true,
		"sort": true, "order": true,
	}

	// Generic text search: use 'q' value across fields specified in 'search_fields'
	// SDK specifies which fields to search, validation happens in Compile
	if q := query.Get("q"); q != "" {
		searchFields := query.Get("search_fields")
		if searchFields != "" {
			// SDK specified which fields to search (PascalCase expected)
			for _, field := range strings.Split(searchFields, ",") {
				field = strings.TrimSpace(field)
				if field != "" {
					textSearchParams = append(textSearchParams, common.SearchableValue{
						Field:    field,
						Values:   []interface{}{q},
						Operator: common.ContainsOperator,
					})
				}
			}
		}
	}

	// Any non-reserved query param is treated as an exact filter
	// SDK must send field names in PascalCase matching Go struct fields
	for key, values := range query {
		if reservedParams[key] || len(values) == 0 || values[0] == "" {
			continue
		}

		value := values[0]
		operator := common.EqualsOperator

		// Support wildcard search: value containing * uses ContainsOperator
		if strings.Contains(value, "*") {
			operator = common.ContainsOperator
			value = strings.ReplaceAll(value, "*", "")
		}

		exactFilters = append(exactFilters, common.SearchableValue{
			Field:    key, // SDK sends PascalCase field names directly
			Values:   []interface{}{value},
			Operator: operator,
		})
	}

	// Build SearchAggregation - combine exact filters (AND) with text search (OR)
	var searchParams []common.SearchAggregation

	// Add exact filter params (AND logic)
	if len(exactFilters) > 0 {
		searchParams = append(searchParams, common.SearchAggregation{
			Params: []common.SearchParameter{
				{
					ValueParams:       exactFilters,
					AggregationClause: common.AndAggregationClause,
				},
			},
			AggregationClause: common.AndAggregationClause,
		})
	}

	// Add text search params (OR logic) - match ANY of the specified search fields
	if len(textSearchParams) > 0 {
		searchParams = append(searchParams, common.SearchAggregation{
			Params: []common.SearchParameter{
				{
					ValueParams:       textSearchParams,
					AggregationClause: common.OrAggregationClause, // OR between text fields
				},
			},
			AggregationClause: common.AndAggregationClause, // AND with other filters
		})
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
