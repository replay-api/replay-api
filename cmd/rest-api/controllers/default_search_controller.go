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
}

func NewDefaultSearchController[T any](service common.Searchable[T]) *DefaultSearchController[T] {
	return &DefaultSearchController[T]{
		service,
	}
}

func (c *DefaultSearchController[T]) DefaultSearchHandler(w http.ResponseWriter, r *http.Request) {
	base64SearchHeader := r.Header.Get("x-search")
	if base64SearchHeader == "" {
		slog.Error("(DefaultSearchHandler) No search source provided: missing 'x-search' in request header")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	searchJSON, err := base64.StdEncoding.DecodeString(base64SearchHeader)
	if err != nil {
		slog.Error("(DefaultSearchHandler) Error decoding search source base64 header", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var s common.Search

	err = json.Unmarshal([]byte(searchJSON), &s)
	if err != nil {
		slog.Error("(DefaultSearchHandler) Error unmarshalling search source", "error", err, "searchJSON", searchJSON, "base64SearchHeader", base64SearchHeader)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if s.SearchParams == nil {
		slog.Error("(DefaultSearchHandler) Empty search source provided")
		w.WriteHeader(http.StatusBadRequest)
		return
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
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(results)
}
