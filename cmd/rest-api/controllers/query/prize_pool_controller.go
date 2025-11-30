package query_controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_out "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/out"
)

type PrizePoolQueryController struct {
	prizePoolRepo matchmaking_out.PrizePoolRepository
}

func NewPrizePoolQueryController(c container.Container) *PrizePoolQueryController {
	var prizePoolRepo matchmaking_out.PrizePoolRepository

	if err := c.Resolve(&prizePoolRepo); err != nil {
		slog.Error("Failed to resolve PrizePoolRepository", "error", err)
		panic(err)
	}

	return &PrizePoolQueryController{
		prizePoolRepo: prizePoolRepo,
	}
}

// GetPrizePoolHandler handles GET /prize-pools/:id
func (c *PrizePoolQueryController) GetPrizePoolHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	poolIDStr := vars["id"]

	if poolIDStr == "" {
		slog.Error("GetPrizePoolHandler: missing pool id")
		http.Error(w, "pool id is required", http.StatusBadRequest)
		return
	}

	poolID, err := uuid.Parse(poolIDStr)
	if err != nil {
		slog.Error("GetPrizePoolHandler: invalid pool id", "id", poolIDStr, "error", err)
		http.Error(w, "invalid pool id format", http.StatusBadRequest)
		return
	}

	slog.Info("Fetching prize pool", "pool_id", poolID)

	pool, err := c.prizePoolRepo.FindByID(r.Context(), poolID)
	if err != nil {
		slog.Error("GetPrizePoolHandler: error fetching pool", "pool_id", poolID, "error", err)
		http.Error(w, "prize pool not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(pool); err != nil {
		slog.Error("GetPrizePoolHandler: error encoding response", "error", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
}

// GetPrizePoolByMatchHandler handles GET /matches/:match_id/prize-pool
func (c *PrizePoolQueryController) GetPrizePoolByMatchHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchIDStr := vars["match_id"]

	if matchIDStr == "" {
		slog.Error("GetPrizePoolByMatchHandler: missing match id")
		http.Error(w, "match id is required", http.StatusBadRequest)
		return
	}

	matchID, err := uuid.Parse(matchIDStr)
	if err != nil {
		slog.Error("GetPrizePoolByMatchHandler: invalid match id", "id", matchIDStr, "error", err)
		http.Error(w, "invalid match id format", http.StatusBadRequest)
		return
	}

	slog.Info("Fetching prize pool by match", "match_id", matchID)

	pool, err := c.prizePoolRepo.FindByMatchID(r.Context(), matchID)
	if err != nil {
		slog.Error("GetPrizePoolByMatchHandler: error fetching pool", "match_id", matchID, "error", err)
		http.Error(w, "prize pool not found for match", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(pool); err != nil {
		slog.Error("GetPrizePoolByMatchHandler: error encoding response", "error", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
}

// GetPrizePoolHistoryHandler handles GET /prize-pools/:id/history
func (c *PrizePoolQueryController) GetPrizePoolHistoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	poolIDStr := vars["id"]

	if poolIDStr == "" {
		slog.Error("GetPrizePoolHistoryHandler: missing pool id")
		http.Error(w, "pool id is required", http.StatusBadRequest)
		return
	}

	poolID, err := uuid.Parse(poolIDStr)
	if err != nil {
		slog.Error("GetPrizePoolHistoryHandler: invalid pool id", "id", poolIDStr, "error", err)
		http.Error(w, "invalid pool id format", http.StatusBadRequest)
		return
	}

	slog.Info("Fetching prize pool history", "pool_id", poolID)

	pool, err := c.prizePoolRepo.FindByID(r.Context(), poolID)
	if err != nil {
		slog.Error("GetPrizePoolHistoryHandler: error fetching pool", "pool_id", poolID, "error", err)
		http.Error(w, "prize pool not found", http.StatusNotFound)
		return
	}

	// Return prize pool with winners/distribution history
	type HistoryResponse struct {
		PoolID            uuid.UUID                                `json:"pool_id"`
		MatchID           uuid.UUID                                `json:"match_id"`
		Status            matchmaking_entities.PrizePoolStatus     `json:"status"`
		TotalAmount       string                                   `json:"total_amount"`
		Currency          string                                   `json:"currency"`
		Winners           []matchmaking_entities.PrizeWinner       `json:"winners"`
		MVPPlayerID       *uuid.UUID                               `json:"mvp_player_id,omitempty"`
		DistributedAt     *string                                  `json:"distributed_at,omitempty"`
		PlayerCount       int                                      `json:"player_count"`
		PlatformContribution string                                `json:"platform_contribution"`
	}

	var distributedAt *string
	if pool.DistributedAt != nil {
		t := pool.DistributedAt.Format("2006-01-02T15:04:05Z07:00")
		distributedAt = &t
	}

	response := HistoryResponse{
		PoolID:               pool.ID,
		MatchID:              pool.MatchID,
		Status:               pool.Status,
		TotalAmount:          pool.TotalAmount.String(),
		Currency:             string(pool.Currency),
		Winners:              pool.Winners,
		MVPPlayerID:          pool.MVPPlayerID,
		DistributedAt:        distributedAt,
		PlayerCount:          pool.GetPlayerCount(),
		PlatformContribution: pool.PlatformContribution.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("GetPrizePoolHistoryHandler: error encoding response", "error", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
}

// GetPendingDistributionsHandler handles GET /prize-pools/pending-distributions (admin only)
func (c *PrizePoolQueryController) GetPendingDistributionsHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("Fetching pending prize pool distributions")

	// Default limit 100
	limit := 100

	pools, err := c.prizePoolRepo.FindPendingDistributions(r.Context(), limit)
	if err != nil {
		slog.Error("GetPendingDistributionsHandler: error fetching pools", "error", err)
		http.Error(w, "error fetching pending distributions", http.StatusInternalServerError)
		return
	}

	type PendingDistributionResponse struct {
		Total int                                  `json:"total"`
		Pools []*matchmaking_entities.PrizePool    `json:"pools"`
	}

	response := PendingDistributionResponse{
		Total: len(pools),
		Pools: pools,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("GetPendingDistributionsHandler: error encoding response", "error", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
}
