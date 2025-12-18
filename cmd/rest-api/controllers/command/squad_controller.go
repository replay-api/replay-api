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
		_ = json.NewEncoder(w).Encode(updatedSquad)
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
		_ = json.NewEncoder(w).Encode(updatedSquad)
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
		_ = json.NewEncoder(w).Encode(results[0])
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
		_ = json.NewEncoder(w).Encode(updatedSquad)
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

// GetSquadStatsHandler handles GET /squads/{id}/stats
func (ctrl *SquadController) GetSquadStatsHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		squadID := vars["id"]

		if squadID == "" {
			http.Error(w, "squad_id is required", http.StatusBadRequest)
			return
		}

		squadUUID, err := uuid.Parse(squadID)
		if err != nil {
			http.Error(w, "invalid squad_id format", http.StatusBadRequest)
			return
		}

		gameID := r.URL.Query().Get("game_id")
		if gameID == "" {
			gameID = "cs2" // Default game
		}

		// Get squad reader
		var squadReader squad_in.SquadReader
		if err := ctrl.container.Resolve(&squadReader); err != nil {
			slog.ErrorContext(r.Context(), "Failed to resolve SquadReader", "error", err)
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		// Get squad to verify existence
		searchParams := []common.SearchAggregation{
			{
				Params: []common.SearchParameter{
					{
						ValueParams: []common.SearchableValue{
							{Field: "ID", Values: []interface{}{squadUUID.String()}, Operator: common.EqualsOperator},
						},
					},
				},
			},
		}
		resultOpts := common.SearchResultOptions{Limit: 1}

		compiledSearch, err := squadReader.Compile(r.Context(), searchParams, resultOpts)
		if err != nil {
			http.Error(w, "invalid search", http.StatusBadRequest)
			return
		}

		squads, err := squadReader.Search(r.Context(), *compiledSearch)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to fetch squad", "error", err)
			http.Error(w, "error fetching squad", http.StatusInternalServerError)
			return
		}

		if len(squads) == 0 {
			http.Error(w, "squad not found", http.StatusNotFound)
			return
		}

		squad := squads[0]

		// Build statistics from squad membership
		stats := map[string]interface{}{
			"squad_id":       squadUUID,
			"squad_name":     squad.Name,
			"game_id":        string(squad.GameID),
			"member_count":   len(squad.Membership),
			"total_matches":  0,
			"wins":           0,
			"losses":         0,
			"win_rate":       0.0,
			"current_streak": 0,
			"members":        []map[string]interface{}{},
			"last_updated":   time.Now(),
		}

		// Add member details
		members := []map[string]interface{}{}
		for _, member := range squad.Membership {
			members = append(members, map[string]interface{}{
				"player_id": member.PlayerProfileID,
				"type":      member.Type,
				"roles":     member.Roles,
			})
		}
		stats["members"] = members

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(stats)
	}
}

// InvitePlayerHandler handles POST /squads/{id}/invitations
func (ctrl *SquadController) InvitePlayerHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		squadID := vars["id"]

		if squadID == "" {
			http.Error(w, `{"error":"squad_id is required"}`, http.StatusBadRequest)
			return
		}

		squadUUID, err := uuid.Parse(squadID)
		if err != nil {
			http.Error(w, `{"error":"invalid squad_id format"}`, http.StatusBadRequest)
			return
		}

		var invitationCmd squad_in.InvitePlayerCommand
		if err := json.NewDecoder(r.Body).Decode(&invitationCmd); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		invitationCmd.SquadID = squadUUID

		var invitationUseCase squad_in.SquadInvitationCommand
		if err := ctrl.container.Resolve(&invitationUseCase); err != nil {
			http.Error(w, `{"error":"service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		invitation, err := invitationUseCase.InvitePlayer(r.Context(), invitationCmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to invite player", "error", err)
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(invitation)
	}
}

// RequestJoinHandler handles POST /squads/{id}/join-requests
func (ctrl *SquadController) RequestJoinHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		squadID := vars["id"]

		if squadID == "" {
			http.Error(w, `{"error":"squad_id is required"}`, http.StatusBadRequest)
			return
		}

		squadUUID, err := uuid.Parse(squadID)
		if err != nil {
			http.Error(w, `{"error":"invalid squad_id format"}`, http.StatusBadRequest)
			return
		}

		var joinCmd squad_in.RequestJoinCommand
		if err := json.NewDecoder(r.Body).Decode(&joinCmd); err != nil {
			// Allow empty body
			joinCmd = squad_in.RequestJoinCommand{}
		}
		joinCmd.SquadID = squadUUID

		var invitationUseCase squad_in.SquadInvitationCommand
		if err := ctrl.container.Resolve(&invitationUseCase); err != nil {
			http.Error(w, `{"error":"service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		invitation, err := invitationUseCase.RequestJoin(r.Context(), joinCmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to create join request", "error", err)
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(invitation)
	}
}

// RespondToInvitationHandler handles POST /invitations/{invitation_id}/respond
func (ctrl *SquadController) RespondToInvitationHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		invitationID := vars["invitation_id"]

		if invitationID == "" {
			http.Error(w, `{"error":"invitation_id is required"}`, http.StatusBadRequest)
			return
		}

		invitationUUID, err := uuid.Parse(invitationID)
		if err != nil {
			http.Error(w, `{"error":"invalid invitation_id format"}`, http.StatusBadRequest)
			return
		}

		var respondCmd squad_in.RespondToInvitationCommand
		if err := json.NewDecoder(r.Body).Decode(&respondCmd); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		respondCmd.InvitationID = invitationUUID

		var invitationUseCase squad_in.SquadInvitationCommand
		if err := ctrl.container.Resolve(&invitationUseCase); err != nil {
			http.Error(w, `{"error":"service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		invitation, err := invitationUseCase.RespondToInvitation(r.Context(), respondCmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to respond to invitation", "error", err)
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(invitation)
	}
}

// CancelInvitationHandler handles DELETE /invitations/{invitation_id}
func (ctrl *SquadController) CancelInvitationHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		invitationID := vars["invitation_id"]

		if invitationID == "" {
			http.Error(w, `{"error":"invitation_id is required"}`, http.StatusBadRequest)
			return
		}

		invitationUUID, err := uuid.Parse(invitationID)
		if err != nil {
			http.Error(w, `{"error":"invalid invitation_id format"}`, http.StatusBadRequest)
			return
		}

		var invitationUseCase squad_in.SquadInvitationCommand
		if err := ctrl.container.Resolve(&invitationUseCase); err != nil {
			http.Error(w, `{"error":"service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		err = invitationUseCase.CancelInvitation(r.Context(), invitationUUID)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to cancel invitation", "error", err)
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// GetSquadInvitationsHandler handles GET /squads/{id}/invitations
func (ctrl *SquadController) GetSquadInvitationsHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		squadID := vars["id"]

		if squadID == "" {
			http.Error(w, `{"error":"squad_id is required"}`, http.StatusBadRequest)
			return
		}

		squadUUID, err := uuid.Parse(squadID)
		if err != nil {
			http.Error(w, `{"error":"invalid squad_id format"}`, http.StatusBadRequest)
			return
		}

		var invitationUseCase squad_in.SquadInvitationCommand
		if err := ctrl.container.Resolve(&invitationUseCase); err != nil {
			http.Error(w, `{"error":"service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		invitations, err := invitationUseCase.GetSquadInvitations(r.Context(), squadUUID)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to get squad invitations", "error", err)
			http.Error(w, `{"error":"failed to get invitations"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"invitations": invitations,
			"count":       len(invitations),
		})
	}
}

// GetPlayerInvitationsHandler handles GET /players/{id}/invitations
func (ctrl *SquadController) GetPlayerInvitationsHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		playerID := vars["id"]

		if playerID == "" {
			http.Error(w, `{"error":"player_id is required"}`, http.StatusBadRequest)
			return
		}

		playerUUID, err := uuid.Parse(playerID)
		if err != nil {
			http.Error(w, `{"error":"invalid player_id format"}`, http.StatusBadRequest)
			return
		}

		var invitationUseCase squad_in.SquadInvitationCommand
		if err := ctrl.container.Resolve(&invitationUseCase); err != nil {
			http.Error(w, `{"error":"service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		invitations, err := invitationUseCase.GetPendingInvitations(r.Context(), playerUUID)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to get player invitations", "error", err)
			http.Error(w, `{"error":"failed to get invitations"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"invitations": invitations,
			"count":       len(invitations),
		})
	}
}
