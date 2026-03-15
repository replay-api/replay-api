package query_controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	prediction_entities "github.com/replay-api/replay-api/pkg/domain/prediction/entities"
	prediction_in "github.com/replay-api/replay-api/pkg/domain/prediction/ports/in"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type PredictionQueryController struct {
	marketQuery prediction_in.MarketQuery
	betQuery    prediction_in.BetQuery
}

func NewPredictionQueryController(c container.Container) *PredictionQueryController {
	var marketQuery prediction_in.MarketQuery
	var betQuery prediction_in.BetQuery

	if err := c.Resolve(&marketQuery); err != nil {
		slog.Warn("MarketQuery not available", "error", err)
	}
	if err := c.Resolve(&betQuery); err != nil {
		slog.Warn("BetQuery not available", "error", err)
	}

	return &PredictionQueryController{
		marketQuery: marketQuery,
		betQuery:    betQuery,
	}
}

// ==================== Market Query Handlers ====================

// GetMarketHandler handles GET /predictions/markets/{market_id}
func (ctrl *PredictionQueryController) GetMarketHandler(w http.ResponseWriter, r *http.Request) {
	if ctrl.marketQuery == nil {
		http.Error(w, `{"success":false,"error":"prediction service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	marketID, err := uuid.Parse(vars["market_id"])
	if err != nil {
		http.Error(w, `{"success":false,"error":"invalid market_id"}`, http.StatusBadRequest)
		return
	}

	market, err := ctrl.marketQuery.GetMarket(r.Context(), marketID)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to get market", "error", err)
		http.Error(w, `{"success":false,"error":"market not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(market)
}

// ListMatchMarketsHandler handles GET /predictions/matches/{match_id}/markets
func (ctrl *PredictionQueryController) ListMatchMarketsHandler(w http.ResponseWriter, r *http.Request) {
	if ctrl.marketQuery == nil {
		http.Error(w, `{"success":false,"error":"prediction service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	matchID, err := uuid.Parse(vars["match_id"])
	if err != nil {
		http.Error(w, `{"success":false,"error":"invalid match_id"}`, http.StatusBadRequest)
		return
	}

	limit, offset := parsePredictionPagination(r)
	status := r.URL.Query().Get("status")

	query := prediction_in.ListMatchMarketsQuery{
		MatchID: matchID,
		Status:  prediction_entities.PredictionStatus(status),
		Limit:   limit,
		Offset:  offset,
	}

	result, err := ctrl.marketQuery.ListMatchMarkets(r.Context(), query)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to list match markets", "error", err)
		http.Error(w, `{"success":false,"error":"failed to list markets"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ==================== Bet Query Handlers ====================

// GetMarketBetsHandler handles GET /predictions/markets/{market_id}/bets
func (ctrl *PredictionQueryController) GetMarketBetsHandler(w http.ResponseWriter, r *http.Request) {
	if ctrl.betQuery == nil {
		http.Error(w, `{"success":false,"error":"prediction service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	marketID, err := uuid.Parse(vars["market_id"])
	if err != nil {
		http.Error(w, `{"success":false,"error":"invalid market_id"}`, http.StatusBadRequest)
		return
	}

	limit, offset := parsePredictionPagination(r)

	result, err := ctrl.betQuery.GetMarketBets(r.Context(), marketID, limit, offset)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to get market bets", "error", err)
		http.Error(w, `{"success":false,"error":"failed to get market bets"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetUserBetsHandler handles GET /predictions/bets/me
func (ctrl *PredictionQueryController) GetUserBetsHandler(w http.ResponseWriter, r *http.Request) {
	if ctrl.betQuery == nil {
		http.Error(w, `{"success":false,"error":"prediction service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	userID, ok := r.Context().Value(shared.UserIDKey).(uuid.UUID)
	if !ok || userID == uuid.Nil {
		http.Error(w, `{"success":false,"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	limit, offset := parsePredictionPagination(r)
	status := r.URL.Query().Get("status")

	query := prediction_in.GetUserBetsQuery{
		UserID: userID,
		Status: prediction_entities.BetStatus(status),
		Limit:  limit,
		Offset: offset,
	}

	result, err := ctrl.betQuery.GetUserBets(r.Context(), query)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to get user bets", "error", err)
		http.Error(w, `{"success":false,"error":"failed to get user bets"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetUserBetSummaryHandler handles GET /predictions/markets/{market_id}/summary
func (ctrl *PredictionQueryController) GetUserBetSummaryHandler(w http.ResponseWriter, r *http.Request) {
	if ctrl.betQuery == nil {
		http.Error(w, `{"success":false,"error":"prediction service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	marketID, err := uuid.Parse(vars["market_id"])
	if err != nil {
		http.Error(w, `{"success":false,"error":"invalid market_id"}`, http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(shared.UserIDKey).(uuid.UUID)
	if !ok || userID == uuid.Nil {
		http.Error(w, `{"success":false,"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	summary, err := ctrl.betQuery.GetUserBetSummary(r.Context(), marketID, userID)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to get bet summary", "error", err)
		http.Error(w, `{"success":false,"error":"failed to get bet summary"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// GetLeaderboardHandler handles GET /predictions/leaderboard
func (ctrl *PredictionQueryController) GetLeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	if ctrl.betQuery == nil {
		http.Error(w, `{"success":false,"error":"prediction service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	entries, err := ctrl.betQuery.GetLeaderboard(r.Context(), limit)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to get leaderboard", "error", err)
		http.Error(w, `{"success":false,"error":"failed to get leaderboard"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"entries": entries,
		"limit":   limit,
	})
}

// parsePredictionPagination extracts limit and offset from query params
func parsePredictionPagination(r *http.Request) (int, int) {
	limit := 20
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	if limit > 100 {
		limit = 100
	}

	return limit, offset
}
