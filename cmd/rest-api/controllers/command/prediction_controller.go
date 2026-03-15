package cmd_controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	prediction_entities "github.com/replay-api/replay-api/pkg/domain/prediction/entities"
	prediction_in "github.com/replay-api/replay-api/pkg/domain/prediction/ports/in"
)

type PredictionCommandController struct {
	marketCommand prediction_in.MarketCommand
	betCommand    prediction_in.BetCommand
}

func NewPredictionCommandController(c container.Container) *PredictionCommandController {
	var marketCommand prediction_in.MarketCommand
	var betCommand prediction_in.BetCommand

	if err := c.Resolve(&marketCommand); err != nil {
		slog.Warn("MarketCommand not available", "error", err)
	}
	if err := c.Resolve(&betCommand); err != nil {
		slog.Warn("BetCommand not available", "error", err)
	}

	return &PredictionCommandController{
		marketCommand: marketCommand,
		betCommand:    betCommand,
	}
}

// ==================== Market Handlers ====================

// CreateMarketHandler handles POST /predictions/markets
func (ctrl *PredictionCommandController) CreateMarketHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ctrl.marketCommand == nil {
			http.Error(w, `{"success":false,"error":"prediction service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		var req struct {
			MatchID     string                            `json:"match_id"`
			GameID      string                            `json:"game_id"`
			BetType     string                            `json:"bet_type"`
			Title       string                            `json:"title"`
			Description string                            `json:"description"`
			Options     []prediction_entities.MarketOption `json:"options"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"success":false,"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}

		matchID, err := uuid.Parse(req.MatchID)
		if err != nil {
			http.Error(w, `{"success":false,"error":"invalid match_id"}`, http.StatusBadRequest)
			return
		}

		cmd := prediction_in.CreateMarketCommand{
			MatchID:     matchID,
			GameID:      req.GameID,
			BetType:     prediction_entities.BetType(req.BetType),
			Title:       req.Title,
			Description: req.Description,
			Options:     req.Options,
		}

		if err := cmd.Validate(); err != nil {
			http.Error(w, `{"success":false,"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		market, err := ctrl.marketCommand.CreateMarket(r.Context(), cmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to create prediction market", "error", err)
			http.Error(w, `{"success":false,"error":"failed to create market"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(market)
	}
}

// LockMarketHandler handles POST /predictions/markets/{market_id}/lock
func (ctrl *PredictionCommandController) LockMarketHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ctrl.marketCommand == nil {
			http.Error(w, `{"success":false,"error":"prediction service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		vars := mux.Vars(r)
		marketID, err := uuid.Parse(vars["market_id"])
		if err != nil {
			http.Error(w, `{"success":false,"error":"invalid market_id"}`, http.StatusBadRequest)
			return
		}

		if err := ctrl.marketCommand.LockMarket(r.Context(), marketID); err != nil {
			slog.ErrorContext(r.Context(), "failed to lock market", "error", err)
			http.Error(w, `{"success":false,"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}
}

// ResolveMarketHandler handles POST /predictions/markets/{market_id}/resolve
func (ctrl *PredictionCommandController) ResolveMarketHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ctrl.marketCommand == nil {
			http.Error(w, `{"success":false,"error":"prediction service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		vars := mux.Vars(r)
		marketID, err := uuid.Parse(vars["market_id"])
		if err != nil {
			http.Error(w, `{"success":false,"error":"invalid market_id"}`, http.StatusBadRequest)
			return
		}

		var req struct {
			OutcomeKey string `json:"outcome_key"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"success":false,"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}

		cmd := prediction_in.ResolveMarketCommand{
			MarketID:   marketID,
			OutcomeKey: req.OutcomeKey,
		}

		if err := cmd.Validate(); err != nil {
			http.Error(w, `{"success":false,"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		if err := ctrl.marketCommand.ResolveMarket(r.Context(), cmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve market", "error", err)
			http.Error(w, `{"success":false,"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}
}

// CancelMarketHandler handles POST /predictions/markets/{market_id}/cancel
func (ctrl *PredictionCommandController) CancelMarketHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ctrl.marketCommand == nil {
			http.Error(w, `{"success":false,"error":"prediction service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		vars := mux.Vars(r)
		marketID, err := uuid.Parse(vars["market_id"])
		if err != nil {
			http.Error(w, `{"success":false,"error":"invalid market_id"}`, http.StatusBadRequest)
			return
		}

		if err := ctrl.marketCommand.CancelMarket(r.Context(), marketID); err != nil {
			slog.ErrorContext(r.Context(), "failed to cancel market", "error", err)
			http.Error(w, `{"success":false,"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}
}

// ==================== Bet Handlers ====================

// PlaceBetHandler handles POST /predictions/markets/{market_id}/bets
func (ctrl *PredictionCommandController) PlaceBetHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ctrl.betCommand == nil {
			http.Error(w, `{"success":false,"error":"prediction service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		vars := mux.Vars(r)
		marketID, err := uuid.Parse(vars["market_id"])
		if err != nil {
			http.Error(w, `{"success":false,"error":"invalid market_id"}`, http.StatusBadRequest)
			return
		}

		var req struct {
			OptionKey string `json:"option_key"`
			Amount    int64  `json:"amount"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"success":false,"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}

		cmd := prediction_in.PlaceBetCommand{
			MarketID:  marketID,
			OptionKey: req.OptionKey,
			Amount:    req.Amount,
		}

		if err := cmd.Validate(); err != nil {
			http.Error(w, `{"success":false,"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		bet, err := ctrl.betCommand.PlaceBet(r.Context(), cmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to place bet", "error", err)
			http.Error(w, `{"success":false,"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(bet)
	}
}
