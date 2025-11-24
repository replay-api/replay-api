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
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	squad_in "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/ports/in"
	squad_out "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/ports/out"
	squad_value_objects "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/value-objects"
	"go.mongodb.org/mongo-driver/mongo"
)

type SquadController struct {
	container container.Container
}

func NewSquadController(container container.Container) *SquadController {
	return &SquadController{container: container}
}

func (ctrl *SquadController) CreateSquadHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		var createSquadCommand squad_in.CreateSquadCommand
		err := json.NewDecoder(r.Body).Decode(&createSquadCommand)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to decode request", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var createSquadCommandHandler squad_in.CreateSquadCommandHandler
		err = ctrl.container.Resolve(&createSquadCommandHandler)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to resolve CreateSquadCommandHandler", "err", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		squad, err := createSquadCommandHandler.Exec(r.Context(), createSquadCommand)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to create squad", "err", err)
			if err.Error() == "Unauthorized" {
				w.WriteHeader(http.StatusUnauthorized)
			}
			return
		}

		err = json.NewEncoder(w).Encode(squad)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to encode response", "err", err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
	}
}

type AddMemberRequest struct {
	PlayerID string                              `json:"player_id"`
	Type     squad_value_objects.SquadMembershipType `json:"type"`
	Roles    []string                            `json:"roles"`
}

// AddMemberHandler handles POST /squads/{id}/members
func (ctrl *SquadController) AddMemberHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		squadID := vars["id"]

		if squadID == "" {
			http.Error(w, "squad_id is required", http.StatusBadRequest)
			return
		}

		var req AddMemberRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.ErrorContext(r.Context(), "Failed to decode request", "err", err)
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.PlayerID == "" {
			http.Error(w, "player_id is required", http.StatusBadRequest)
			return
		}

		var squadReader squad_out.SquadReader
		if err := ctrl.container.Resolve(&squadReader); err != nil {
			slog.ErrorContext(r.Context(), "Failed to resolve SquadReader", "err", err)
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		idUUID, err := uuid.Parse(squadID)
		if err != nil {
			http.Error(w, "invalid squad_id format", http.StatusBadRequest)
			return
		}
		valueParams := []common.SearchableValue{{Field: "ID", Values: []interface{}{idUUID}, Operator: common.EqualsOperator}}
		search := common.NewSearchByValues(r.Context(), valueParams, common.SearchResultOptions{Limit: 1}, common.UserAudienceIDKey)
		results, err := squadReader.Search(r.Context(), search)
		if err != nil || len(results) == 0 {
			http.Error(w, "squad not found", http.StatusNotFound)
			return
		}

		squad := results[0]

		// Add member to membership map
		if squad.Membership == nil {
			squad.Membership = make(map[string]squad_value_objects.SquadMembership)
		}

		now := time.Now()
		memberType := req.Type
		if memberType == "" {
			memberType = squad_value_objects.SquadMembershipTypeMember
		}

		squad.Membership[req.PlayerID] = squad_value_objects.SquadMembership{
			Type:  memberType,
			Roles: req.Roles,
			Status: map[time.Time]squad_value_objects.SquadMembershipStatus{
				now: squad_value_objects.SquadMembershipStatusActive,
			},
			History: map[time.Time]squad_value_objects.SquadMembershipType{
				now: memberType,
			},
		}

		// Update squad in database
		var squadWriter squad_out.SquadWriter
		if err := ctrl.container.Resolve(&squadWriter); err != nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		updatedSquad, err := squadWriter.Update(r.Context(), &squad)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to update squad", "err", err)
			http.Error(w, "failed to add member", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updatedSquad)
	}
}

// RemoveMemberHandler handles DELETE /squads/{id}/members/{player_id}
func (ctrl *SquadController) RemoveMemberHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		squadID := vars["id"]
		playerID := vars["player_id"]

		if squadID == "" || playerID == "" {
			http.Error(w, "squad_id and player_id are required", http.StatusBadRequest)
			return
		}

		var squadReader squad_out.SquadReader
		if err := ctrl.container.Resolve(&squadReader); err != nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		idUUID, err := uuid.Parse(squadID)
		if err != nil {
			http.Error(w, "invalid squad_id format", http.StatusBadRequest)
			return
		}
		valueParams := []common.SearchableValue{{Field: "ID", Values: []interface{}{idUUID}, Operator: common.EqualsOperator}}
		search := common.NewSearchByValues(r.Context(), valueParams, common.SearchResultOptions{Limit: 1}, common.UserAudienceIDKey)
		results, err := squadReader.Search(r.Context(), search)
		if err != nil || len(results) == 0 {
			http.Error(w, "squad not found", http.StatusNotFound)
			return
		}

		squad := results[0]

		// Check if member exists
		if squad.Membership == nil || squad.Membership[playerID].Type == "" {
			http.Error(w, "member not found in squad", http.StatusNotFound)
			return
		}

		// Remove member from membership map
		delete(squad.Membership, playerID)

		// Update squad in database
		var squadWriter squad_out.SquadWriter
		if err := ctrl.container.Resolve(&squadWriter); err != nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		_, err = squadWriter.Update(r.Context(), &squad)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to update squad", "err", err)
			http.Error(w, "failed to remove member", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

type UpdateMemberRoleRequest struct {
	Type  squad_value_objects.SquadMembershipType `json:"type"`
	Roles []string                            `json:"roles"`
}

// UpdateMemberRoleHandler handles PUT /squads/{id}/members/{player_id}/role
func (ctrl *SquadController) UpdateMemberRoleHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		squadID := vars["id"]
		playerID := vars["player_id"]

		if squadID == "" || playerID == "" {
			http.Error(w, "squad_id and player_id are required", http.StatusBadRequest)
			return
		}

		var req UpdateMemberRoleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		var squadReader squad_out.SquadReader
		if err := ctrl.container.Resolve(&squadReader); err != nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		idUUID, err := uuid.Parse(squadID)
		if err != nil {
			http.Error(w, "invalid squad_id format", http.StatusBadRequest)
			return
		}
		valueParams := []common.SearchableValue{{Field: "ID", Values: []interface{}{idUUID}, Operator: common.EqualsOperator}}
		search := common.NewSearchByValues(r.Context(), valueParams, common.SearchResultOptions{Limit: 1}, common.UserAudienceIDKey)
		results, err := squadReader.Search(r.Context(), search)
		if err != nil || len(results) == 0 {
			http.Error(w, "squad not found", http.StatusNotFound)
			return
		}

		squad := results[0]

		if squad.Membership == nil || squad.Membership[playerID].Type == "" {
			http.Error(w, "member not found in squad", http.StatusNotFound)
			return
		}

		// Update member role
		membership := squad.Membership[playerID]
		if req.Type != "" {
			membership.Type = req.Type
			if membership.History == nil {
				membership.History = make(map[time.Time]squad_value_objects.SquadMembershipType)
			}
			membership.History[time.Now()] = req.Type
		}
		if req.Roles != nil {
			membership.Roles = req.Roles
		}
		squad.Membership[playerID] = membership

		var squadWriter squad_out.SquadWriter
		if err := ctrl.container.Resolve(&squadWriter); err != nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		updatedSquad, err := squadWriter.Update(r.Context(), &squad)
		if err != nil {
			http.Error(w, "failed to update member role", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updatedSquad)
	}
}

// GetSquadHandler handles GET /squads/{id}
func (ctrl *SquadController) GetSquadHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		squadID := vars["id"]

		if squadID == "" {
			http.Error(w, "squad_id is required", http.StatusBadRequest)
			return
		}

		var squadReader squad_out.SquadReader
		if err := ctrl.container.Resolve(&squadReader); err != nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		idUUID, err := uuid.Parse(squadID)
		if err != nil {
			http.Error(w, "invalid squad_id format", http.StatusBadRequest)
			return
		}
		valueParams := []common.SearchableValue{{Field: "ID", Values: []interface{}{idUUID}, Operator: common.EqualsOperator}}
		search := common.NewSearchByValues(r.Context(), valueParams, common.SearchResultOptions{Limit: 1}, common.UserAudienceIDKey)
		results, err := squadReader.Search(r.Context(), search)
		if err != nil {
			http.Error(w, "error fetching squad", http.StatusInternalServerError)
			return
		}

		if len(results) == 0 {
			http.Error(w, "squad not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(results[0])
	}
}

type UpdateSquadRequest struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Description string `json:"description"`
	LogoURI     string `json:"logo_uri"`
	BannerURI   string `json:"banner_uri"`
}

// UpdateSquadHandler handles PUT /squads/{id}
func (ctrl *SquadController) UpdateSquadHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		squadID := vars["id"]

		if squadID == "" {
			http.Error(w, "squad_id is required", http.StatusBadRequest)
			return
		}

		var req UpdateSquadRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		var squadReader squad_out.SquadReader
		if err := ctrl.container.Resolve(&squadReader); err != nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		idUUID, err := uuid.Parse(squadID)
		if err != nil {
			http.Error(w, "invalid squad_id format", http.StatusBadRequest)
			return
		}
		valueParams := []common.SearchableValue{{Field: "ID", Values: []interface{}{idUUID}, Operator: common.EqualsOperator}}
		search := common.NewSearchByValues(r.Context(), valueParams, common.SearchResultOptions{Limit: 1}, common.UserAudienceIDKey)
		results, err := squadReader.Search(r.Context(), search)
		if err != nil || len(results) == 0 {
			http.Error(w, "squad not found", http.StatusNotFound)
			return
		}

		squad := results[0]

		// Update fields
		if req.Name != "" {
			squad.Name = req.Name
		}
		if req.Symbol != "" {
			squad.Symbol = req.Symbol
		}
		if req.Description != "" {
			squad.Description = req.Description
		}
		if req.LogoURI != "" {
			squad.LogoURI = req.LogoURI
		}
		if req.BannerURI != "" {
			squad.BannerURI = req.BannerURI
		}
		squad.UpdatedAt = time.Now()

		var squadWriter squad_out.SquadWriter
		if err := ctrl.container.Resolve(&squadWriter); err != nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		updatedSquad, err := squadWriter.Update(r.Context(), &squad)
		if err != nil {
			http.Error(w, "failed to update squad", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updatedSquad)
	}
}

// DeleteSquadHandler handles DELETE /squads/{id}
func (ctrl *SquadController) DeleteSquadHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		squadID := vars["id"]

		if squadID == "" {
			http.Error(w, "squad_id is required", http.StatusBadRequest)
			return
		}

		var squadWriter squad_out.SquadWriter
		if err := ctrl.container.Resolve(&squadWriter); err != nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		squadUUID, err := uuid.Parse(squadID)
		if err != nil {
			http.Error(w, "invalid squad_id format", http.StatusBadRequest)
			return
		}

		if err := squadWriter.Delete(r.Context(), squadUUID); err != nil {
			if err == mongo.ErrNoDocuments {
				http.Error(w, "squad not found", http.StatusNotFound)
				return
			}
			http.Error(w, "failed to delete squad", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
