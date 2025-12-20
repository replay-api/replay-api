package query_controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
	tournament_in "github.com/replay-api/replay-api/pkg/domain/tournament/ports/in"
)

type TournamentQueryController struct {
	tournamentReader tournament_in.TournamentReader
}

func NewTournamentQueryController(c container.Container) *TournamentQueryController {
	var tournamentReader tournament_in.TournamentReader

	if err := c.Resolve(&tournamentReader); err != nil {
		slog.Error("Failed to resolve TournamentReader", "error", err)
		panic(err)
	}

	return &TournamentQueryController{
		tournamentReader: tournamentReader,
	}
}

// GetTournamentHandler handles GET /tournaments/{id}
// @Summary Get tournament by ID
// @Description Returns detailed information about a specific tournament
// @Tags Tournaments
// @Produce json
// @Param id path string true "Tournament ID"
// @Success 200 {object} tournament_entities.Tournament
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Router /tournaments/{id} [get]
func (c *TournamentQueryController) GetTournamentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentIDStr := vars["id"]

	tournamentID, err := uuid.Parse(tournamentIDStr)
	if err != nil {
		http.Error(w, "invalid tournament id", http.StatusBadRequest)
		return
	}

	tournament, err := c.tournamentReader.GetTournament(r.Context(), tournamentID)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to fetch tournament", "tournament_id", tournamentID, "error", err)
		http.Error(w, "tournament not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tournament); err != nil {
		slog.ErrorContext(r.Context(), "failed to encode response", "error", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
}

// ListTournamentsHandler handles GET /tournaments
// @Summary List tournaments
// @Description Returns a list of tournaments with optional filtering
// @Tags Tournaments
// @Produce json
// @Param game_id query string false "Game ID filter"
// @Param region query string false "Region filter"
// @Param status query string false "Status filter"
// @Param limit query int false "Limit results" default(50)
// @Success 200 {object} object "ListResponse"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tournaments [get]
func (c *TournamentQueryController) ListTournamentsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query params
	gameID := r.URL.Query().Get("game_id")
	region := r.URL.Query().Get("region")
	statusStr := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")

	limit := 50 // Default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	var statusFilter []tournament_entities.TournamentStatus
	if statusStr != "" {
		statusFilter = []tournament_entities.TournamentStatus{
			tournament_entities.TournamentStatus(statusStr),
		}
	}

	tournaments, err := c.tournamentReader.ListTournaments(r.Context(), gameID, region, statusFilter, limit)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to list tournaments", "error", err)
		http.Error(w, "error fetching tournaments", http.StatusInternalServerError)
		return
	}

	type ListResponse struct {
		Total       int                                    `json:"total"`
		Tournaments []*tournament_entities.Tournament      `json:"tournaments"`
	}

	response := ListResponse{
		Total:       len(tournaments),
		Tournaments: tournaments,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.ErrorContext(r.Context(), "failed to encode response", "error", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
}

// GetUpcomingTournamentsHandler handles GET /tournaments/upcoming
// @Summary Get upcoming tournaments
// @Description Returns a list of upcoming tournaments for a specific game
// @Tags Tournaments
// @Produce json
// @Param game_id query string true "Game ID"
// @Param limit query int false "Limit results" default(20)
// @Success 200 {object} object "UpcomingResponse"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tournaments/upcoming [get]
func (c *TournamentQueryController) GetUpcomingTournamentsHandler(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("game_id")
	if gameID == "" {
		http.Error(w, "game_id is required", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 20 // Default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	tournaments, err := c.tournamentReader.GetUpcomingTournaments(r.Context(), gameID, limit)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to fetch upcoming tournaments", "error", err)
		http.Error(w, "error fetching tournaments", http.StatusInternalServerError)
		return
	}

	type UpcomingResponse struct {
		Total       int                                    `json:"total"`
		Tournaments []*tournament_entities.Tournament      `json:"tournaments"`
	}

	response := UpcomingResponse{
		Total:       len(tournaments),
		Tournaments: tournaments,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.ErrorContext(r.Context(), "failed to encode response", "error", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
}

// GetPlayerTournamentsHandler handles GET /players/{player_id}/tournaments
// @Summary Get player tournaments
// @Description Returns all tournaments a player has participated in
// @Tags Tournaments
// @Produce json
// @Param player_id path string true "Player ID"
// @Success 200 {object} object "PlayerTournamentsResponse"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /players/{player_id}/tournaments [get]
func (c *TournamentQueryController) GetPlayerTournamentsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	playerIDStr := vars["player_id"]

	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "invalid player_id", http.StatusBadRequest)
		return
	}

	tournaments, err := c.tournamentReader.GetPlayerTournaments(r.Context(), playerID)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to fetch player tournaments", "player_id", playerID, "error", err)
		http.Error(w, "error fetching tournaments", http.StatusInternalServerError)
		return
	}

	type PlayerTournamentsResponse struct {
		PlayerID    string                                 `json:"player_id"`
		Total       int                                    `json:"total"`
		Tournaments []*tournament_entities.Tournament      `json:"tournaments"`
	}

	response := PlayerTournamentsResponse{
		PlayerID:    playerID.String(),
		Total:       len(tournaments),
		Tournaments: tournaments,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.ErrorContext(r.Context(), "failed to encode response", "error", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
}

// GetOrganizerTournamentsHandler handles GET /organizers/{organizer_id}/tournaments
// @Summary Get organizer tournaments
// @Description Returns all tournaments organized by a specific organizer
// @Tags Tournaments
// @Produce json
// @Param organizer_id path string true "Organizer ID"
// @Success 200 {object} object "OrganizerTournamentsResponse"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /organizers/{organizer_id}/tournaments [get]
func (c *TournamentQueryController) GetOrganizerTournamentsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizerIDStr := vars["organizer_id"]

	organizerID, err := uuid.Parse(organizerIDStr)
	if err != nil {
		http.Error(w, "invalid organizer_id", http.StatusBadRequest)
		return
	}

	tournaments, err := c.tournamentReader.GetOrganizerTournaments(r.Context(), organizerID)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to fetch organizer tournaments", "organizer_id", organizerID, "error", err)
		http.Error(w, "error fetching tournaments", http.StatusInternalServerError)
		return
	}

	type OrganizerTournamentsResponse struct {
		OrganizerID string                                 `json:"organizer_id"`
		Total       int                                    `json:"total"`
		Tournaments []*tournament_entities.Tournament      `json:"tournaments"`
	}

	response := OrganizerTournamentsResponse{
		OrganizerID: organizerID.String(),
		Total:       len(tournaments),
		Tournaments: tournaments,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.ErrorContext(r.Context(), "failed to encode response", "error", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
}
