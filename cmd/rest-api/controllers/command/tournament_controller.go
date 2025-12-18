package cmd_controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	common "github.com/replay-api/replay-api/pkg/domain"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
	tournament_in "github.com/replay-api/replay-api/pkg/domain/tournament/ports/in"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

type TournamentCommandController struct {
	tournamentCommand tournament_in.TournamentCommand
}

func NewTournamentCommandController(c container.Container) *TournamentCommandController {
	var tournamentCommand tournament_in.TournamentCommand

	if err := c.Resolve(&tournamentCommand); err != nil {
		slog.Error("Failed to resolve TournamentCommand", "error", err)
		panic(err)
	}

	return &TournamentCommandController{
		tournamentCommand: tournamentCommand,
	}
}

// CreateTournamentRequest represents the request to create a tournament
type CreateTournamentRequest struct {
	Name                string                              `json:"name"`
	Description         string                              `json:"description"`
	GameID              string                              `json:"game_id"`
	GameMode            string                              `json:"game_mode"`
	Region              string                              `json:"region"`
	Format              string                              `json:"format"`
	MaxParticipants     int                                 `json:"max_participants"`
	MinParticipants     int                                 `json:"min_participants"`
	EntryFee            float64                             `json:"entry_fee"`
	Currency            string                              `json:"currency"`
	StartTime           string                              `json:"start_time"` // RFC3339
	RegistrationOpen    string                              `json:"registration_open"`
	RegistrationClose   string                              `json:"registration_close"`
	Rules               tournament_entities.TournamentRules `json:"rules"`
	OrganizerID         string                              `json:"organizer_id"`
}

// CreateTournamentHandler handles POST /tournaments
func (c *TournamentCommandController) CreateTournamentHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateTournamentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Parse UUIDs and timestamps
		organizerID, err := uuid.Parse(req.OrganizerID)
		if err != nil {
			http.Error(w, "invalid organizer_id", http.StatusBadRequest)
			return
		}

		startTime, err := parseRFC3339(req.StartTime)
		if err != nil {
			http.Error(w, "invalid start_time format", http.StatusBadRequest)
			return
		}

		registrationOpen, err := parseRFC3339(req.RegistrationOpen)
		if err != nil {
			http.Error(w, "invalid registration_open format", http.StatusBadRequest)
			return
		}

		registrationClose, err := parseRFC3339(req.RegistrationClose)
		if err != nil {
			http.Error(w, "invalid registration_close format", http.StatusBadRequest)
			return
		}

		// Create command - use resource owner from context
		resourceOwner := common.GetResourceOwner(r.Context())
		cmd := tournament_in.CreateTournamentCommand{
			ResourceOwner:     resourceOwner,
			Name:              req.Name,
			Description:       req.Description,
			GameID:            common.GameIDKey(req.GameID),
			GameMode:          req.GameMode,
			Region:            req.Region,
			Format:            tournament_entities.TournamentFormat(req.Format),
			MaxParticipants:   req.MaxParticipants,
			MinParticipants:   req.MinParticipants,
			EntryFee:          wallet_vo.NewAmount(req.EntryFee),
			Currency:          wallet_vo.Currency(req.Currency),
			StartTime:         startTime,
			RegistrationOpen:  registrationOpen,
			RegistrationClose: registrationClose,
			Rules:             req.Rules,
			OrganizerID:       organizerID,
		}

		tournament, err := c.tournamentCommand.CreateTournament(r.Context(), cmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to create tournament", "error", err)
			http.Error(w, "failed to create tournament: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(tournament); err != nil {
			slog.ErrorContext(r.Context(), "failed to encode response", "error", err)
			http.Error(w, "error encoding response", http.StatusInternalServerError)
			return
		}
	}
}

// UpdateTournamentHandler handles PUT /tournaments/:id
func (c *TournamentCommandController) UpdateTournamentHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tournamentIDStr := vars["id"]

		tournamentID, err := uuid.Parse(tournamentIDStr)
		if err != nil {
			http.Error(w, "invalid tournament id", http.StatusBadRequest)
			return
		}

		var req struct {
			Name              *string                              `json:"name,omitempty"`
			Description       *string                              `json:"description,omitempty"`
			MaxParticipants   *int                                 `json:"max_participants,omitempty"`
			StartTime         *string                              `json:"start_time,omitempty"`
			RegistrationClose *string                              `json:"registration_close,omitempty"`
			Rules             *tournament_entities.TournamentRules `json:"rules,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		cmd := tournament_in.UpdateTournamentCommand{
			TournamentID:      tournamentID,
			Name:              req.Name,
			Description:       req.Description,
			MaxParticipants:   req.MaxParticipants,
			Rules:             req.Rules,
		}

		if req.StartTime != nil {
			t, err := parseRFC3339(*req.StartTime)
			if err != nil {
				http.Error(w, "invalid start_time format", http.StatusBadRequest)
				return
			}
			cmd.StartTime = &t
		}

		if req.RegistrationClose != nil {
			t, err := parseRFC3339(*req.RegistrationClose)
			if err != nil {
				http.Error(w, "invalid registration_close format", http.StatusBadRequest)
				return
			}
			cmd.RegistrationClose = &t
		}

		tournament, err := c.tournamentCommand.UpdateTournament(r.Context(), cmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to update tournament", "error", err)
			http.Error(w, "failed to update tournament: "+err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(tournament); err != nil {
			slog.ErrorContext(r.Context(), "failed to encode response", "error", err)
			http.Error(w, "error encoding response", http.StatusInternalServerError)
			return
		}
	}
}

// DeleteTournamentHandler handles DELETE /tournaments/:id
func (c *TournamentCommandController) DeleteTournamentHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tournamentIDStr := vars["id"]

		tournamentID, err := uuid.Parse(tournamentIDStr)
		if err != nil {
			http.Error(w, "invalid tournament id", http.StatusBadRequest)
			return
		}

		if err := c.tournamentCommand.DeleteTournament(r.Context(), tournamentID); err != nil {
			slog.ErrorContext(r.Context(), "failed to delete tournament", "error", err)
			http.Error(w, "failed to delete tournament: "+err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// RegisterPlayerHandler handles POST /tournaments/:id/register
func (c *TournamentCommandController) RegisterPlayerHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tournamentIDStr := vars["id"]

		tournamentID, err := uuid.Parse(tournamentIDStr)
		if err != nil {
			http.Error(w, "invalid tournament id", http.StatusBadRequest)
			return
		}

		var req struct {
			PlayerID    string `json:"player_id"`
			DisplayName string `json:"display_name"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		playerID, err := uuid.Parse(req.PlayerID)
		if err != nil {
			http.Error(w, "invalid player_id", http.StatusBadRequest)
			return
		}

		cmd := tournament_in.RegisterPlayerCommand{
			TournamentID: tournamentID,
			PlayerID:     playerID,
			DisplayName:  req.DisplayName,
		}

		if err := c.tournamentCommand.RegisterPlayer(r.Context(), cmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to register player", "error", err)
			http.Error(w, "failed to register: "+err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "registered"})
	}
}

// UnregisterPlayerHandler handles DELETE /tournaments/:id/register
func (c *TournamentCommandController) UnregisterPlayerHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tournamentIDStr := vars["id"]

		tournamentID, err := uuid.Parse(tournamentIDStr)
		if err != nil {
			http.Error(w, "invalid tournament id", http.StatusBadRequest)
			return
		}

		var req struct {
			PlayerID string `json:"player_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		playerID, err := uuid.Parse(req.PlayerID)
		if err != nil {
			http.Error(w, "invalid player_id", http.StatusBadRequest)
			return
		}

		cmd := tournament_in.UnregisterPlayerCommand{
			TournamentID: tournamentID,
			PlayerID:     playerID,
		}

		if err := c.tournamentCommand.UnregisterPlayer(r.Context(), cmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to unregister player", "error", err)
			http.Error(w, "failed to unregister: "+err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "unregistered"})
	}
}

// StartTournamentHandler handles POST /tournaments/:id/start
func (c *TournamentCommandController) StartTournamentHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tournamentIDStr := vars["id"]

		tournamentID, err := uuid.Parse(tournamentIDStr)
		if err != nil {
			http.Error(w, "invalid tournament id", http.StatusBadRequest)
			return
		}

		if err := c.tournamentCommand.StartTournament(r.Context(), tournamentID); err != nil {
			slog.ErrorContext(r.Context(), "failed to start tournament", "error", err)
			http.Error(w, "failed to start tournament: "+err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "started"})
	}
}

// GenerateBracketsHandler handles POST /tournaments/:id/brackets
// Generates tournament brackets based on format (single/double elimination, round-robin, swiss)
func (c *TournamentCommandController) GenerateBracketsHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tournamentIDStr := vars["id"]

		tournamentID, err := uuid.Parse(tournamentIDStr)
		if err != nil {
			http.Error(w, "invalid tournament id", http.StatusBadRequest)
			return
		}

		if err := c.tournamentCommand.GenerateBrackets(r.Context(), tournamentID); err != nil {
			slog.ErrorContext(r.Context(), "failed to generate brackets", "error", err)
			http.Error(w, "failed to generate brackets: "+err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "brackets_generated"})
	}
}

// ScheduleMatchesHandler handles POST /tournaments/:id/schedule
// Automatically schedules all tournament matches
func (c *TournamentCommandController) ScheduleMatchesHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tournamentIDStr := vars["id"]

		tournamentID, err := uuid.Parse(tournamentIDStr)
		if err != nil {
			http.Error(w, "invalid tournament id", http.StatusBadRequest)
			return
		}

		var req struct {
			StartTime         *string `json:"start_time,omitempty"`
			MatchDurationMins int     `json:"match_duration_mins,omitempty"`
			BreakBetweenMins  int     `json:"break_between_mins,omitempty"`
			ConcurrentMatches int     `json:"concurrent_matches,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		cmd := tournament_in.ScheduleMatchesCommand{
			TournamentID:      tournamentID,
			MatchDurationMins: req.MatchDurationMins,
			BreakBetweenMins:  req.BreakBetweenMins,
			ConcurrentMatches: req.ConcurrentMatches,
		}

		if req.StartTime != nil {
			t, err := parseRFC3339(*req.StartTime)
			if err != nil {
				http.Error(w, "invalid start_time format", http.StatusBadRequest)
				return
			}
			cmd.StartTime = &t
		}

		if err := c.tournamentCommand.ScheduleMatches(r.Context(), cmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to schedule matches", "error", err)
			http.Error(w, "failed to schedule matches: "+err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "matches_scheduled"})
	}
}

// RescheduleMatchHandler handles PUT /tournaments/:id/matches/:match_id/schedule
// Reschedules a specific match
func (c *TournamentCommandController) RescheduleMatchHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tournamentIDStr := vars["id"]
		matchIDStr := vars["match_id"]

		tournamentID, err := uuid.Parse(tournamentIDStr)
		if err != nil {
			http.Error(w, "invalid tournament id", http.StatusBadRequest)
			return
		}

		matchID, err := uuid.Parse(matchIDStr)
		if err != nil {
			http.Error(w, "invalid match id", http.StatusBadRequest)
			return
		}

		var req struct {
			NewTime string `json:"new_time"`
			Reason  string `json:"reason,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		newTime, err := parseRFC3339(req.NewTime)
		if err != nil {
			http.Error(w, "invalid new_time format", http.StatusBadRequest)
			return
		}

		cmd := tournament_in.RescheduleMatchCommand{
			TournamentID: tournamentID,
			MatchID:      matchID,
			NewTime:      newTime,
			Reason:       req.Reason,
		}

		if err := c.tournamentCommand.RescheduleMatch(r.Context(), cmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to reschedule match", "error", err)
			http.Error(w, "failed to reschedule match: "+err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "match_rescheduled"})
	}
}

// ReportMatchResultHandler handles POST /tournaments/:id/matches/:match_id/result
// Reports the result of a completed match
func (c *TournamentCommandController) ReportMatchResultHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tournamentIDStr := vars["id"]
		matchIDStr := vars["match_id"]

		tournamentID, err := uuid.Parse(tournamentIDStr)
		if err != nil {
			http.Error(w, "invalid tournament id", http.StatusBadRequest)
			return
		}

		matchID, err := uuid.Parse(matchIDStr)
		if err != nil {
			http.Error(w, "invalid match id", http.StatusBadRequest)
			return
		}

		var req struct {
			WinnerID string `json:"winner_id"`
			Score    string `json:"score"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		winnerID, err := uuid.Parse(req.WinnerID)
		if err != nil {
			http.Error(w, "invalid winner_id", http.StatusBadRequest)
			return
		}

		cmd := tournament_in.ReportMatchResultCommand{
			TournamentID: tournamentID,
			MatchID:      matchID,
			WinnerID:     winnerID,
			Score:        req.Score,
			ReportedBy:   common.GetResourceOwner(r.Context()).UserID,
		}

		if err := c.tournamentCommand.ReportMatchResult(r.Context(), cmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to report match result", "error", err)
			http.Error(w, "failed to report match result: "+err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "result_reported"})
	}
}

// Helper function to parse RFC3339 timestamps
func parseRFC3339(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
