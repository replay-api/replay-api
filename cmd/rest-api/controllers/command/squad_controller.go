package cmd_controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	common "github.com/replay-api/replay-api/pkg/domain"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	squad_out "github.com/replay-api/replay-api/pkg/domain/squad/ports/out"
	squad_value_objects "github.com/replay-api/replay-api/pkg/domain/squad/value-objects"
	"go.mongodb.org/mongo-driver/mongo"
)

type SquadController struct {
	container                 container.Container
	createSquadCommandHandler squad_in.CreateSquadCommandHandler
}

func NewSquadController(container container.Container) *SquadController {
	var createSquadCommandHandler squad_in.CreateSquadCommandHandler
	var err = container.Resolve(&createSquadCommandHandler)
	if err != nil {
		slog.Error("Failed to resolve CreateSquadCommandHandler", "err", err)
		return nil
	}

	return &SquadController{
		container:                 container,
		createSquadCommandHandler: createSquadCommandHandler,
	}
}

func (ctrl *SquadController) CreateSquadHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		var createSquadCommand squad_in.CreateOrUpdatedSquadCommand
		err := json.NewDecoder(r.Body).Decode(&createSquadCommand)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to decode request", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		squad, err := ctrl.createSquadCommandHandler.Exec(r.Context(), createSquadCommand)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to create squad",
				"error", err,
				"squad_name", createSquadCommand.Name,
				"slug_uri", createSquadCommand.SlugURI)
			if err.Error() == "Unauthorized" {
				w.WriteHeader(http.StatusUnauthorized)
			} else if strings.Contains(err.Error(), "already exists") {
				w.WriteHeader(http.StatusConflict)
				errorJSON := map[string]string{
					"code":  "CONFLICT",
					"error": err.Error(),
				}

				err = json.NewEncoder(w).Encode(errorJSON)
				if err != nil {
					slog.ErrorContext(r.Context(), "Failed to encode response", "error", err)
				}
			} else if strings.Contains(err.Error(), "not found") {
				w.WriteHeader(http.StatusNotFound)
				errorJSON := map[string]string{
					"code":  "NOT_FOUND",
					"error": err.Error(),
				}

				err = json.NewEncoder(w).Encode(errorJSON)
				if err != nil {
					slog.ErrorContext(r.Context(), "Failed to encode response", "error", err)
				}
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		slog.InfoContext(r.Context(), "Squad created successfully",
			"squad_id", squad.ID,
			"squad_name", squad.Name,
			"slug_uri", squad.SlugURI,
			"group_id", squad.ResourceOwner.GroupID,
			"user_id", r.Context().Value(common.UserIDKey))

		err = json.NewEncoder(w).Encode(squad)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to encode response",
				"error", err,
				"squad_id", squad.ID)
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

		// Parse player ID
		playerUUID, err := uuid.Parse(req.PlayerID)
		if err != nil {
			http.Error(w, "invalid player_id format", http.StatusBadRequest)
			return
		}

		// Check if member already exists in slice
		for _, m := range squad.Membership {
			if m.PlayerProfileID == playerUUID {
				http.Error(w, "player already a member of this squad", http.StatusConflict)
				return
			}
		}

		// Add member to membership slice
		now := time.Now()
		memberType := req.Type
		if memberType == "" {
			memberType = squad_value_objects.SquadMembershipTypeMember
		}

		newMembership := squad_value_objects.SquadMembership{
			PlayerProfileID: playerUUID,
			Type:            memberType,
			Roles:           req.Roles,
			Status: map[time.Time]squad_value_objects.SquadMembershipStatus{
				now: squad_value_objects.SquadMembershipStatusActive,
			},
			History: map[time.Time]squad_value_objects.SquadMembershipType{
				now: memberType,
			},
		}
		squad.Membership = append(squad.Membership, newMembership)

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

		// Parse player ID
		playerUUID, err := uuid.Parse(playerID)
		if err != nil {
			http.Error(w, "invalid player_id format", http.StatusBadRequest)
			return
		}

		// Find and remove member from membership slice
		memberFound := false
		newMembership := make([]squad_value_objects.SquadMembership, 0, len(squad.Membership))
		for _, m := range squad.Membership {
			if m.PlayerProfileID == playerUUID {
				memberFound = true
				continue // Skip this member (effectively removing them)
			}
			newMembership = append(newMembership, m)
		}

		if !memberFound {
			http.Error(w, "member not found in squad", http.StatusNotFound)
			return
		}

		squad.Membership = newMembership

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

		// Parse player ID
		playerUUID, err := uuid.Parse(playerID)
		if err != nil {
			http.Error(w, "invalid player_id format", http.StatusBadRequest)
			return
		}

		// Find member in slice
		memberIndex := -1
		for i, m := range squad.Membership {
			if m.PlayerProfileID == playerUUID {
				memberIndex = i
				break
			}
		}

		if memberIndex == -1 {
			http.Error(w, "member not found in squad", http.StatusNotFound)
			return
		}

		// Update member role
		membership := squad.Membership[memberIndex]
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
		squad.Membership[memberIndex] = membership

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
