package cmd_controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
	tournament_in "github.com/replay-api/replay-api/pkg/domain/tournament/ports/in"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type TournamentCommandController struct {
	tournamentCommand tournament_in.TournamentCommand
}

func NewTournamentCommandController(c container.Container) *TournamentCommandController {
	var tournamentCommand tournament_in.TournamentCommand

	if err := c.Resolve(&tournamentCommand); err != nil {
		slog.Warn("TournamentCommand not available", "error", err)
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
		resourceOwner := shared.GetResourceOwner(r.Context())
		cmd := tournament_in.CreateTournamentCommand{
			ResourceOwner:     resourceOwner,
			Name:              req.Name,
			Description:       req.Description,
			GameID:            replay_common.GameIDKey(req.GameID),
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
			http.Error(w, `{"success":false,"error":"Failed to create tournament"}`, http.StatusInternalServerError)
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
			http.Error(w, `{"success":false,"error":"Failed to update tournament"}`, http.StatusBadRequest)
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
			http.Error(w, `{"success":false,"error":"Failed to delete tournament"}`, http.StatusBadRequest)
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
			http.Error(w, `{"success":false,"error":"Failed to register for tournament"}`, http.StatusBadRequest)
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
			http.Error(w, `{"success":false,"error":"Failed to unregister from tournament"}`, http.StatusBadRequest)
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
			http.Error(w, `{"success":false,"error":"Failed to start tournament"}`, http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "started"})
	}
}

// OpenRegistrationHandler handles POST /tournaments/:id/registration/open
func (c *TournamentCommandController) OpenRegistrationHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tournamentIDStr := vars["id"]

		tournamentID, err := uuid.Parse(tournamentIDStr)
		if err != nil {
			http.Error(w, "invalid tournament id", http.StatusBadRequest)
			return
		}

		if err := c.tournamentCommand.OpenRegistration(r.Context(), tournamentID); err != nil {
			slog.ErrorContext(r.Context(), "failed to open registration", "error", err)
			http.Error(w, sanitizeTournamentError(err, "Failed to open registration"), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "registration_open"})
	}
}

// CloseRegistrationHandler handles POST /tournaments/:id/registration/close
func (c *TournamentCommandController) CloseRegistrationHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tournamentIDStr := vars["id"]

		tournamentID, err := uuid.Parse(tournamentIDStr)
		if err != nil {
			http.Error(w, "invalid tournament id", http.StatusBadRequest)
			return
		}

		if err := c.tournamentCommand.CloseRegistration(r.Context(), tournamentID); err != nil {
			slog.ErrorContext(r.Context(), "failed to close registration", "error", err)
			http.Error(w, sanitizeTournamentError(err, "Failed to close registration"), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "registration_closed"})
	}
}

// CompleteTournamentHandler handles POST /tournaments/:id/complete
func (c *TournamentCommandController) CompleteTournamentHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tournamentIDStr := vars["id"]

		tournamentID, err := uuid.Parse(tournamentIDStr)
		if err != nil {
			http.Error(w, "invalid tournament id", http.StatusBadRequest)
			return
		}

		var req struct {
			Winners []struct {
				PlayerID  string  `json:"player_id"`
				Placement int     `json:"placement"`
				Prize     float64 `json:"prize"`
			} `json:"winners"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		winners := make([]tournament_entities.TournamentWinner, 0, len(req.Winners))
		for _, wr := range req.Winners {
			playerID, err := uuid.Parse(wr.PlayerID)
			if err != nil {
				http.Error(w, "invalid player_id in winners", http.StatusBadRequest)
				return
			}
			winners = append(winners, tournament_entities.TournamentWinner{
				PlayerID:  playerID,
				Placement: wr.Placement,
				Prize:     wallet_vo.NewAmount(wr.Prize),
			})
		}

		cmd := tournament_in.CompleteTournamentCommand{
			TournamentID: tournamentID,
			Winners:      winners,
		}

		if err := c.tournamentCommand.CompleteTournament(r.Context(), cmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to complete tournament", "error", err)
			http.Error(w, sanitizeTournamentError(err, "Failed to complete tournament"), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "completed"})
	}
}

// CancelTournamentHandler handles POST /tournaments/:id/cancel
func (c *TournamentCommandController) CancelTournamentHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tournamentIDStr := vars["id"]

		tournamentID, err := uuid.Parse(tournamentIDStr)
		if err != nil {
			http.Error(w, "invalid tournament id", http.StatusBadRequest)
			return
		}

		var req struct {
			Reason string `json:"reason"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		cmd := tournament_in.CancelTournamentCommand{
			TournamentID: tournamentID,
			Reason:       req.Reason,
		}

		if err := c.tournamentCommand.CancelTournament(r.Context(), cmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to cancel tournament", "error", err)
			http.Error(w, sanitizeTournamentError(err, "Failed to cancel tournament"), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "cancelled"})
	}
}

// CheckInHandler handles POST /tournaments/:id/check-in
func (c *TournamentCommandController) CheckInHandler(apiContext context.Context) http.HandlerFunc {
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

		if err := c.tournamentCommand.CheckIn(r.Context(), tournamentID, playerID); err != nil {
			slog.ErrorContext(r.Context(), "failed to check in player", "error", err)
			http.Error(w, sanitizeTournamentError(err, "Failed to check in"), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "checked_in"})
	}
}

// RecordMatchResultHandler handles POST /tournaments/:id/matches/:match_id/result
func (c *TournamentCommandController) RecordMatchResultHandler(apiContext context.Context) http.HandlerFunc {
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

		if err := c.tournamentCommand.RecordMatchResult(r.Context(), tournamentID, matchID, winnerID); err != nil {
			slog.ErrorContext(r.Context(), "failed to record match result", "error", err)
			http.Error(w, sanitizeTournamentError(err, "Failed to record match result"), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "result_recorded"})
	}
}

// AdvanceBracketHandler handles POST /tournaments/:id/advance-bracket
func (c *TournamentCommandController) AdvanceBracketHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tournamentIDStr := vars["id"]

		tournamentID, err := uuid.Parse(tournamentIDStr)
		if err != nil {
			http.Error(w, "invalid tournament id", http.StatusBadRequest)
			return
		}

		if err := c.tournamentCommand.AdvanceBracket(r.Context(), tournamentID); err != nil {
			slog.ErrorContext(r.Context(), "failed to advance bracket", "error", err)
			http.Error(w, sanitizeTournamentError(err, "Failed to advance bracket"), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "bracket_advanced"})
	}
}

// sanitizeTournamentError sanitizes error messages for external consumption (financial-grade security)
func sanitizeTournamentError(err error, fallback string) string {
	msg := err.Error()
	// Allow domain-level validation errors through (safe for users)
	safePatterns := []string{
		"not open", "not enough", "already registered", "tournament is full",
		"registration", "status", "cannot", "must be", "not found",
		"player not found", "check-in", "match not found", "winner",
	}
	for _, pattern := range safePatterns {
		if containsInsensitive(msg, pattern) {
			return fmt.Sprintf(`{"success":false,"error":"%s"}`, msg)
		}
	}
	// Default: generic error to avoid leaking internal details
	return fmt.Sprintf(`{"success":false,"error":"%s"}`, fallback)
}

func containsInsensitive(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsLower(toLower(s), toLower(substr)))
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func containsLower(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper function to parse RFC3339 timestamps
func parseRFC3339(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
