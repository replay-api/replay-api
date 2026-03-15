package cmd_controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	oracle_in "github.com/replay-api/replay-api/pkg/domain/oracle/ports/in"
	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

// OracleCommandController handles write operations for oracle results
type OracleCommandController struct {
	commandHandler oracle_in.OracleCommandHandler
}

// NewOracleCommandController creates a new controller resolving deps from DI container
func NewOracleCommandController(c container.Container) *OracleCommandController {
	var commandHandler oracle_in.OracleCommandHandler

	if err := c.Resolve(&commandHandler); err != nil {
		slog.Warn("OracleCommandHandler not available", "error", err)
	}

	return &OracleCommandController{
		commandHandler: commandHandler,
	}
}

// --- Request DTOs ---

type IngestExternalScoreRequest struct {
	ExternalMatchID *string `json:"external_match_id,omitempty"`
	MatchID         *string `json:"match_id,omitempty"`
	GameID          string  `json:"game_id"`
	SourceType      string  `json:"source_type"`
	ProviderMatchID string  `json:"provider_match_id"`
	WinnerTeamID    *string `json:"winner_team_id,omitempty"`
	IsDraw          bool    `json:"is_draw"`
	TeamAID         string  `json:"team_a_id"`
	TeamBID         string  `json:"team_b_id"`
	TeamAScore      int     `json:"team_a_score"`
	TeamBScore      int     `json:"team_b_score"`
	RoundsPlayed    int     `json:"rounds_played"`
	MVPPlayerID     *string `json:"mvp_player_id,omitempty"`
}

type CreateExternalMatchOracleRequest struct {
	ExternalMatchID string `json:"external_match_id"`
	GameID          string `json:"game_id"`
}

type PublishToChainRequest struct {
	ChainIDs []int64 `json:"chain_ids,omitempty"`
}

type DisputeEscalationRequest struct {
	Reason     string `json:"reason"`
	DisputedBy string `json:"disputed_by"`
}

type TriggerIngestionRequest struct {
	MatchID         *string `json:"match_id,omitempty"`
	ExternalMatchID *string `json:"external_match_id,omitempty"`
	GameID          string  `json:"game_id"`
}

// --- Handlers ---

// IngestExternalScoreHandler handles POST /oracle/results/{id}/ingest
func (c *OracleCommandController) IngestExternalScoreHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if c.commandHandler == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		ctx := r.Context()
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		vars := mux.Vars(r)
		resultID, err := uuid.Parse(vars["id"])
		if err != nil {
			http.Error(w, "invalid oracle result id", http.StatusBadRequest)
			return
		}

		var req IngestExternalScoreRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		var matchID *uuid.UUID
		if req.MatchID != nil {
			parsed, err := uuid.Parse(*req.MatchID)
			if err != nil {
				http.Error(w, "invalid match_id", http.StatusBadRequest)
				return
			}
			matchID = &parsed
		}

		var winnerTeamID *uuid.UUID
		if req.WinnerTeamID != nil {
			parsed, err := uuid.Parse(*req.WinnerTeamID)
			if err != nil {
				http.Error(w, "invalid winner_team_id", http.StatusBadRequest)
				return
			}
			winnerTeamID = &parsed
		}

		teamAID, err := uuid.Parse(req.TeamAID)
		if err != nil {
			http.Error(w, "invalid team_a_id", http.StatusBadRequest)
			return
		}
		teamBID, err := uuid.Parse(req.TeamBID)
		if err != nil {
			http.Error(w, "invalid team_b_id", http.StatusBadRequest)
			return
		}

		var mvpPlayerID *uuid.UUID
		if req.MVPPlayerID != nil {
			parsed, err := uuid.Parse(*req.MVPPlayerID)
			if err != nil {
				http.Error(w, "invalid mvp_player_id", http.StatusBadRequest)
				return
			}
			mvpPlayerID = &parsed
		}

		cmd := oracle_in.IngestExternalScoreCommand{
			OracleResultID:  &resultID,
			ExternalMatchID: req.ExternalMatchID,
			MatchID:         matchID,
			GameID:          replay_common.GameIDKey(req.GameID),
			SourceType:      oracle_vo.OracleSourceType(req.SourceType),
			ProviderMatchID: req.ProviderMatchID,
			WinnerTeamID:    winnerTeamID,
			IsDraw:          req.IsDraw,
			TeamAID:         teamAID,
			TeamBID:         teamBID,
			TeamAScore:      req.TeamAScore,
			TeamBScore:      req.TeamBScore,
			RoundsPlayed:    req.RoundsPlayed,
			MVPPlayerID:     mvpPlayerID,
		}

		if err := c.commandHandler.IngestExternalScore(ctx, cmd); err != nil {
			slog.ErrorContext(ctx, "failed to ingest external score", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"status": "accepted"})
	}
}

// CreateExternalMatchOracleHandler handles POST /oracle/results
func (c *OracleCommandController) CreateExternalMatchOracleHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if c.commandHandler == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		var req CreateExternalMatchOracleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		cmd := oracle_in.CreateExternalMatchOracleCommand{
			ExternalMatchID: req.ExternalMatchID,
			GameID:          replay_common.GameIDKey(req.GameID),
		}

		if err := cmd.Validate(); err != nil {
			http.Error(w, "invalid request parameters", http.StatusBadRequest)
			return
		}

		result, err := c.commandHandler.CreateExternalMatchOracle(r.Context(), cmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to create external match oracle", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(oracle_in.MapOracleResultToDTO(result))
	}
}

// PublishToChainHandler handles POST /oracle/results/{id}/publish
func (c *OracleCommandController) PublishToChainHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if c.commandHandler == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		vars := mux.Vars(r)
		resultID, err := uuid.Parse(vars["id"])
		if err != nil {
			http.Error(w, "invalid oracle result id", http.StatusBadRequest)
			return
		}

		var req PublishToChainRequest
		json.NewDecoder(r.Body).Decode(&req)

		cmd := oracle_in.PublishToChainCommand{
			OracleResultID: resultID,
		}

		for _, id := range req.ChainIDs {
			cmd.ChainIDs = append(cmd.ChainIDs, oracle_vo.ChainID(id))
		}

		if err := c.commandHandler.PublishToChain(r.Context(), cmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to publish to chain", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"status": "publishing"})
	}
}

// DisputeEscalationHandler handles POST /oracle/results/{id}/dispute
func (c *OracleCommandController) DisputeEscalationHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if c.commandHandler == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		vars := mux.Vars(r)
		resultID, err := uuid.Parse(vars["id"])
		if err != nil {
			http.Error(w, "invalid oracle result id", http.StatusBadRequest)
			return
		}

		var req DisputeEscalationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		disputedBy, err := uuid.Parse(req.DisputedBy)
		if err != nil {
			http.Error(w, "invalid disputed_by", http.StatusBadRequest)
			return
		}

		cmd := oracle_in.HandleDisputeCommand{
			OracleResultID: resultID,
			Reason:         req.Reason,
			DisputedBy:     disputedBy,
		}

		if err := cmd.Validate(); err != nil {
			http.Error(w, "invalid request parameters", http.StatusBadRequest)
			return
		}

		if err := c.commandHandler.HandleDisputeEscalation(r.Context(), cmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to handle dispute escalation", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "disputed"})
	}
}

// TriggerIngestionHandler handles POST /oracle/results/trigger-ingestion
func (c *OracleCommandController) TriggerIngestionHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if c.commandHandler == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		var req TriggerIngestionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		var matchID *uuid.UUID
		if req.MatchID != nil {
			parsed, err := uuid.Parse(*req.MatchID)
			if err != nil {
				http.Error(w, "invalid match_id", http.StatusBadRequest)
				return
			}
			matchID = &parsed
		}

		gameID := replay_common.GameIDKey(req.GameID)

		cmd := oracle_in.TriggerIngestionCommand{
			MatchID:         matchID,
			ExternalMatchID: req.ExternalMatchID,
			GameID:          gameID,
		}

		if err := cmd.Validate(); err != nil {
			http.Error(w, "invalid request parameters", http.StatusBadRequest)
			return
		}

		if err := c.commandHandler.TriggerIngestionFromAllProviders(r.Context(), cmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to trigger ingestion", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"status": "ingestion_triggered"})
	}
}
