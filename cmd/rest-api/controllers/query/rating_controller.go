package query_controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	common "github.com/replay-api/replay-api/pkg/domain"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
)

// RatingController handles player rating and leaderboard queries
type RatingController struct {
	ratingService matchmaking_in.RatingService
}

// NewRatingController creates a new RatingController
func NewRatingController(c container.Container) *RatingController {
	var ratingService matchmaking_in.RatingService

	if err := c.Resolve(&ratingService); err != nil {
		slog.Warn("RatingService not registered, rating endpoints will not work", "error", err)
	}

	return &RatingController{
		ratingService: ratingService,
	}
}

// GetPlayerRatingHandler handles GET /players/{id}/rating
// @Summary Get player rating
// @Description Returns the rating and MMR information for a player
// @Tags Rating
// @Produce json
// @Param id path string true "Player ID"
// @Param game_id query string false "Game ID" default(cs2)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {string} string "Bad Request"
// @Failure 503 {string} string "Service Unavailable"
// @Router /players/{id}/rating [get]
func (c *RatingController) GetPlayerRatingHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	playerIDStr := vars["id"]

	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "invalid player_id", http.StatusBadRequest)
		return
	}

	if c.ratingService == nil {
		http.Error(w, "rating service not available", http.StatusServiceUnavailable)
		return
	}

	// Parse optional game_id (default to CS2)
	gameID := common.CS2.ID
	if gameIDStr := r.URL.Query().Get("game_id"); gameIDStr != "" {
		gameID = common.GameIDKey(gameIDStr)
	}

	rating, err := c.ratingService.GetPlayerRating(r.Context(), playerID, gameID)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to get player rating", "player_id", playerID, "error", err)
		http.Error(w, "failed to get player rating", http.StatusInternalServerError)
		return
	}

	// Build response with additional computed fields
	response := map[string]interface{}{
		"player_id":        rating.PlayerID,
		"game_id":          rating.GameID,
		"rating":           rating.Rating,
		"mmr":              rating.GetMMR(),
		"rank":             rating.GetRank(),
		"rating_deviation": rating.RatingDeviation,
		"volatility":       rating.Volatility,
		"confidence":       rating.GetConfidence(),
		"is_provisional":   rating.IsProvisional(),
		"matches_played":   rating.MatchesPlayed,
		"wins":             rating.Wins,
		"losses":           rating.Losses,
		"draws":            rating.Draws,
		"win_rate":         rating.GetWinRate(),
		"win_streak":       rating.WinStreak,
		"peak_rating":      rating.PeakRating,
		"last_match_at":    rating.LastMatchAt,
		"recent_matches":   rating.RatingHistory,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// GetLeaderboardHandler handles GET /leaderboard
// @Summary Get leaderboard
// @Description Returns the leaderboard for a game with optional filtering
// @Tags Rating
// @Produce json
// @Param game_id query string false "Game ID" default(cs2)
// @Param region query string false "Region filter"
// @Param tier query string false "Tier filter"
// @Param limit query int false "Limit results" default(100) maximum(500)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 503 {string} string "Service Unavailable"
// @Router /leaderboard [get]
func (c *RatingController) GetLeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	if c.ratingService == nil {
		http.Error(w, "rating service not available", http.StatusServiceUnavailable)
		return
	}

	// Parse query parameters
	gameID := common.CS2.ID
	if gameIDStr := r.URL.Query().Get("game_id"); gameIDStr != "" {
		gameID = common.GameIDKey(gameIDStr)
	}

	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 500 {
			limit = parsedLimit
		}
	}

	leaderboard, err := c.ratingService.GetLeaderboard(r.Context(), gameID, limit)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to get leaderboard", "game_id", gameID, "error", err)
		http.Error(w, "failed to get leaderboard", http.StatusInternalServerError)
		return
	}

	// Build response with ranks
	entries := make([]map[string]interface{}, 0, len(leaderboard))
	for i, rating := range leaderboard {
		entries = append(entries, map[string]interface{}{
			"position":       i + 1,
			"player_id":      rating.PlayerID,
			"rating":         rating.Rating,
			"rank":           rating.GetRank(),
			"matches_played": rating.MatchesPlayed,
			"wins":           rating.Wins,
			"losses":         rating.Losses,
			"win_rate":       rating.GetWinRate(),
			"peak_rating":    rating.PeakRating,
		})
	}

	response := map[string]interface{}{
		"game_id":     gameID,
		"total":       len(entries),
		"leaderboard": entries,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// GetRankDistributionHandler handles GET /ranks/distribution
// @Summary Get rank distribution
// @Description Returns the distribution of players across ranks for a game
// @Tags Rating
// @Produce json
// @Param game_id query string false "Game ID" default(cs2)
// @Success 200 {object} map[string]interface{}
// @Failure 503 {string} string "Service Unavailable"
// @Router /ranks/distribution [get]
func (c *RatingController) GetRankDistributionHandler(w http.ResponseWriter, r *http.Request) {
	if c.ratingService == nil {
		http.Error(w, "rating service not available", http.StatusServiceUnavailable)
		return
	}

	gameID := common.CS2.ID
	if gameIDStr := r.URL.Query().Get("game_id"); gameIDStr != "" {
		gameID = common.GameIDKey(gameIDStr)
	}

	distribution, err := c.ratingService.GetRankDistribution(r.Context(), gameID)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to get rank distribution", "game_id", gameID, "error", err)
		http.Error(w, "failed to get rank distribution", http.StatusInternalServerError)
		return
	}

	// Calculate total and percentages
	var total int
	for _, count := range distribution {
		total += count
	}

	distributionWithPercentage := make([]map[string]interface{}, 0)
	ranks := []string{"bronze", "silver", "gold", "platinum", "diamond", "master", "grandmaster", "challenger"}
	for _, rank := range ranks {
		count := distribution[matchmaking_entities.Rank(rank)]
		var percentage float64
		if total > 0 {
			percentage = float64(count) / float64(total) * 100
		}
		distributionWithPercentage = append(distributionWithPercentage, map[string]interface{}{
			"rank":       rank,
			"count":      count,
			"percentage": percentage,
		})
	}

	response := map[string]interface{}{
		"game_id":      gameID,
		"total":        total,
		"distribution": distributionWithPercentage,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
