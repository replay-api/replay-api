package controllers

import (
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"

	common "github.com/replay-api/replay-api/pkg/domain"
)

type DefaultSearchController[T any] struct {
	common.Searchable[T]
	helper *ControllerHelper
}

func NewDefaultSearchController[T any](service common.Searchable[T]) *DefaultSearchController[T] {
	return &DefaultSearchController[T]{
		Searchable: service,
		helper:     NewControllerHelper(),
	}
}

func (c *DefaultSearchController[T]) DefaultSearchHandler(w http.ResponseWriter, r *http.Request) {
	base64SearchHeader := r.Header.Get("x-search")
	if base64SearchHeader == "" {
		slog.Error("(DefaultSearchHandler) No search source provided: missing 'x-search' in request header")
		c.helper.HandleError(w, r, common.NewAPIError(http.StatusBadRequest, "BAD_REQUEST", "missing 'x-search' header"), "missing search header")
		return
	}

	searchJSON, err := base64.StdEncoding.DecodeString(base64SearchHeader)
	if err != nil {
		slog.Error("(DefaultSearchHandler) Error decoding search source base64 header", "error", err)
		c.helper.HandleError(w, r, common.NewAPIError(http.StatusBadRequest, "BAD_REQUEST", "invalid base64 search header"), "error decoding search header")
		return
	}

	var s common.Search
	err = json.Unmarshal([]byte(searchJSON), &s)
	if err != nil {
		slog.Error("(DefaultSearchHandler) Error unmarshalling search source", "error", err, "searchJSON", searchJSON, "base64SearchHeader", base64SearchHeader)
		c.helper.HandleError(w, r, common.NewAPIError(http.StatusBadRequest, "BAD_REQUEST", "invalid search JSON"), "error unmarshalling search")
		return
	}

	if s.SearchParams == nil {
		slog.Error("(DefaultSearchHandler) Empty search source provided")
		c.helper.HandleError(w, r, common.NewAPIError(http.StatusBadRequest, "BAD_REQUEST", "empty search parameters"), "empty search parameters")
		return
	}

	compiledSearch, err := c.Compile(r.Context(), s.SearchParams, s.ResultOptions)
	if c.helper.HandleError(w, r, err, "error validating search request") {
		return // Error handled by helper
	}

	if compiledSearch == nil {
		slog.Error("(DefaultSearchHandler) Error validating search request", "error", "search is nil")
		c.helper.HandleError(w, r, common.NewAPIError(http.StatusUnprocessableEntity, "UNPROCESSABLE_ENTITY", "compiled search is nil"), "compiled search validation failed")
		return
	}

	results, err := c.Search(r.Context(), *compiledSearch)
	if c.helper.HandleError(w, r, err, "error filtering search request") {
		return // Error handled by helper
	}

	// Write successful response
	c.helper.WriteOK(w, r, results)
}
