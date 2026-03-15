package query_controllers

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	common "github.com/replay-api/replay-api/pkg/domain"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_in "github.com/replay-api/replay-api/pkg/domain/wallet/ports/in"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// VaultQueryController handles team vault query operations
type VaultQueryController struct {
	vaultQuery wallet_in.TeamVaultQuery
}

// NewVaultQueryController creates a new vault query controller
func NewVaultQueryController(c container.Container) *VaultQueryController {
	var vaultQuery wallet_in.TeamVaultQuery

	if err := c.Resolve(&vaultQuery); err != nil {
		slog.Error("Failed to resolve TeamVaultQuery", "error", err)
		panic(err)
	}

	return &VaultQueryController{
		vaultQuery: vaultQuery,
	}
}

// requireVaultQueryAuth checks authentication for vault query operations
func requireVaultQueryAuth(w http.ResponseWriter, r *http.Request) *shared.ResourceOwner {
	ctx := r.Context()
	authenticated, ok := ctx.Value(shared.AuthenticatedKey).(bool)
	if !ok || !authenticated {
		slog.WarnContext(ctx, "Vault query attempted without authentication")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"success":false,"error":"Authentication required to access vault"}`))
		return nil
	}
	resourceOwner := shared.GetResourceOwner(ctx)
	if resourceOwner.UserID == uuid.Nil {
		slog.WarnContext(ctx, "Vault query attempted without valid user ID")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"success":false,"error":"Valid user authentication required"}`))
		return nil
	}
	return &resourceOwner
}

func getSquadIDFromQueryPath(r *http.Request) (uuid.UUID, error) {
	vars := mux.Vars(r)
	squadIDStr := vars["squad_id"]
	if squadIDStr == "" {
		return uuid.Nil, &wallet_in.ValidationError{Field: "squad_id", Message: "squad_id is required in path"}
	}
	squadID, err := uuid.Parse(squadIDStr)
	if err != nil {
		return uuid.Nil, &wallet_in.ValidationError{Field: "squad_id", Message: "invalid squad_id format"}
	}
	return squadID, nil
}

func parseLimitOffset(r *http.Request) (int, int) {
	limit := 20
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	return limit, offset
}

// GetVaultHandler handles GET /squads/{squad_id}/vault
func (c *VaultQueryController) GetVaultHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	resourceOwner := requireVaultQueryAuth(w, r)
	if resourceOwner == nil {
		return
	}

	squadID, err := getSquadIDFromQueryPath(r)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
		return
	}

	vault, err := c.vaultQuery.GetVaultBySquadID(ctx, squadID)
	if err != nil {
		slog.ErrorContext(ctx, "GetVaultHandler: error", "error", err)
		common.WriteErrorFromDomainError(w, err)
		return
	}

	common.WriteSuccess(w, vault)
}

// GetVaultBalanceHandler handles GET /squads/{squad_id}/vault/balance
func (c *VaultQueryController) GetVaultBalanceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	resourceOwner := requireVaultQueryAuth(w, r)
	if resourceOwner == nil {
		return
	}

	squadID, err := getSquadIDFromQueryPath(r)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
		return
	}

	result, err := c.vaultQuery.GetVaultBalance(ctx, squadID)
	if err != nil {
		slog.ErrorContext(ctx, "GetVaultBalanceHandler: error", "error", err)
		common.WriteErrorFromDomainError(w, err)
		return
	}

	common.WriteSuccess(w, result)
}

// GetVaultProposalsHandler handles GET /squads/{squad_id}/vault/proposals
func (c *VaultQueryController) GetVaultProposalsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	resourceOwner := requireVaultQueryAuth(w, r)
	if resourceOwner == nil {
		return
	}

	squadID, err := getSquadIDFromQueryPath(r)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
		return
	}

	limit, offset := parseLimitOffset(r)

	filters := wallet_in.ProposalFilters{
		Limit:  limit,
		Offset: offset,
	}

	if v := r.URL.Query().Get("status"); v != "" {
		status := wallet_entities.ProposalStatus(v)
		filters.Status = &status
	}

	if v := r.URL.Query().Get("type"); v != "" {
		proposalType := wallet_entities.ProposalType(v)
		filters.Type = &proposalType
	}

	result, err := c.vaultQuery.GetProposals(ctx, squadID, filters)
	if err != nil {
		slog.ErrorContext(ctx, "GetVaultProposalsHandler: error", "error", err)
		common.WriteErrorFromDomainError(w, err)
		return
	}

	common.WriteSuccess(w, result)
}

// GetVaultProposalByIDHandler handles GET /squads/{squad_id}/vault/proposals/{proposal_id}
func (c *VaultQueryController) GetVaultProposalByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	resourceOwner := requireVaultQueryAuth(w, r)
	if resourceOwner == nil {
		return
	}

	vars := mux.Vars(r)
	proposalIDStr := vars["proposal_id"]
	if proposalIDStr == "" {
		common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "proposal_id is required", "")
		return
	}

	proposalID, err := uuid.Parse(proposalIDStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid proposal_id format", "")
		return
	}

	proposal, err := c.vaultQuery.GetProposalByID(ctx, proposalID)
	if err != nil {
		slog.ErrorContext(ctx, "GetVaultProposalByIDHandler: error", "error", err)
		common.WriteErrorFromDomainError(w, err)
		return
	}

	common.WriteSuccess(w, proposal)
}

// GetVaultActivityHandler handles GET /squads/{squad_id}/vault/activity
func (c *VaultQueryController) GetVaultActivityHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	resourceOwner := requireVaultQueryAuth(w, r)
	if resourceOwner == nil {
		return
	}

	squadID, err := getSquadIDFromQueryPath(r)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
		return
	}

	limit, offset := parseLimitOffset(r)

	filters := wallet_in.ActivityFilters{
		Limit:  limit,
		Offset: offset,
	}

	if v := r.URL.Query().Get("activity_type"); v != "" {
		activityType := wallet_entities.VaultActivityType(v)
		filters.ActivityType = &activityType
	}

	if v := r.URL.Query().Get("actor_id"); v != "" {
		actorID, err := uuid.Parse(v)
		if err == nil {
			filters.ActorID = &actorID
		}
	}

	if v := r.URL.Query().Get("from_date"); v != "" {
		fromDate, err := time.Parse(time.RFC3339, v)
		if err == nil {
			filters.FromDate = &fromDate
		}
	}

	if v := r.URL.Query().Get("to_date"); v != "" {
		toDate, err := time.Parse(time.RFC3339, v)
		if err == nil {
			filters.ToDate = &toDate
		}
	}

	result, err := c.vaultQuery.GetVaultActivity(ctx, squadID, filters)
	if err != nil {
		slog.ErrorContext(ctx, "GetVaultActivityHandler: error", "error", err)
		common.WriteErrorFromDomainError(w, err)
		return
	}

	common.WriteSuccess(w, result)
}

// GetVaultAnalyticsHandler handles GET /squads/{squad_id}/vault/analytics
func (c *VaultQueryController) GetVaultAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	resourceOwner := requireVaultQueryAuth(w, r)
	if resourceOwner == nil {
		return
	}

	squadID, err := getSquadIDFromQueryPath(r)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
		return
	}

	// Default: last 30 days
	now := time.Now().UTC()
	timeRange := wallet_in.VaultAnalyticsTimeRange{
		From: now.AddDate(0, -1, 0),
		To:   now,
	}

	if v := r.URL.Query().Get("from"); v != "" {
		from, err := time.Parse(time.RFC3339, v)
		if err == nil {
			timeRange.From = from
		}
	}

	if v := r.URL.Query().Get("to"); v != "" {
		to, err := time.Parse(time.RFC3339, v)
		if err == nil {
			timeRange.To = to
		}
	}

	result, err := c.vaultQuery.GetVaultAnalytics(ctx, squadID, timeRange)
	if err != nil {
		slog.ErrorContext(ctx, "GetVaultAnalyticsHandler: error", "error", err)
		common.WriteErrorFromDomainError(w, err)
		return
	}

	common.WriteSuccess(w, result)
}

// GetVaultInventoryHandler handles GET /squads/{squad_id}/vault/inventory
func (c *VaultQueryController) GetVaultInventoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	resourceOwner := requireVaultQueryAuth(w, r)
	if resourceOwner == nil {
		return
	}

	squadID, err := getSquadIDFromQueryPath(r)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
		return
	}

	limit, offset := parseLimitOffset(r)

	filters := wallet_in.InventoryFilters{
		Limit:  limit,
		Offset: offset,
	}

	if v := r.URL.Query().Get("item_type"); v != "" {
		itemType := wallet_entities.InventoryItemType(v)
		filters.ItemType = &itemType
	}

	if v := r.URL.Query().Get("rarity"); v != "" {
		rarity := wallet_entities.ItemRarity(v)
		filters.Rarity = &rarity
	}

	if v := r.URL.Query().Get("game_id"); v != "" {
		filters.GameID = &v
	}

	if v := r.URL.Query().Get("status"); v != "" {
		status := wallet_entities.InventoryItemStatus(v)
		filters.Status = &status
	}

	result, err := c.vaultQuery.GetVaultInventory(ctx, squadID, filters)
	if err != nil {
		slog.ErrorContext(ctx, "GetVaultInventoryHandler: error", "error", err)
		common.WriteErrorFromDomainError(w, err)
		return
	}

	common.WriteSuccess(w, result)
}
