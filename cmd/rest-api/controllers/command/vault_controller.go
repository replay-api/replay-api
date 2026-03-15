package cmd_controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	common "github.com/replay-api/replay-api/pkg/domain"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_in "github.com/replay-api/replay-api/pkg/domain/wallet/ports/in"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// VaultCommandController handles team vault command operations
type VaultCommandController struct {
	vaultCommand wallet_in.TeamVaultCommand
}

// NewVaultCommandController creates a new vault command controller
func NewVaultCommandController(c container.Container) *VaultCommandController {
	var vaultCommand wallet_in.TeamVaultCommand

	if err := c.Resolve(&vaultCommand); err != nil {
		slog.Error("Failed to resolve TeamVaultCommand", "error", err)
		panic(err)
	}

	return &VaultCommandController{
		vaultCommand: vaultCommand,
	}
}

// --- Request DTOs ---

type CreateVaultRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type VaultDepositRequest struct {
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	IdempotencyKey string  `json:"idempotency_key,omitempty"`
}

type ProposeTransactionRequest struct {
	Type        string  `json:"type"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Destination string  `json:"destination,omitempty"`
}

type ApproveProposalRequest struct {
	Reason        string  `json:"reason,omitempty"`
	SignatureHash *string `json:"signature_hash,omitempty"`
}

type RejectProposalRequest struct {
	Reason string `json:"reason"`
}

type UpdateVaultSettingsRequest struct {
	ApprovalPolicy     *wallet_vo.ApprovalPolicy `json:"approval_policy,omitempty"`
	OnChainThreshold   *float64                  `json:"on_chain_threshold,omitempty"`
	DailyWithdrawLimit *float64                  `json:"daily_withdraw_limit,omitempty"`
}

type DepositItemRequest struct {
	ItemID string `json:"item_id"`
}

type ProposeItemTransferRequest struct {
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	InventoryItemIDs  []string `json:"inventory_item_ids"`
	DestinationUserID string   `json:"destination_user_id"`
}

// --- Auth Helpers ---

func requireVaultAuth(w http.ResponseWriter, r *http.Request) uuid.UUID {
	ctx := r.Context()
	authenticated, ok := ctx.Value(shared.AuthenticatedKey).(bool)
	if !ok || !authenticated {
		slog.WarnContext(ctx, "Vault command attempted without authentication")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"success":false,"error":"Authentication required for vault operations"}`))
		return uuid.Nil
	}
	resourceOwner := shared.GetResourceOwner(ctx)
	if resourceOwner.UserID == uuid.Nil {
		slog.WarnContext(ctx, "Vault command attempted without valid user ID")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"success":false,"error":"Valid user authentication required for vault operations"}`))
		return uuid.Nil
	}
	return resourceOwner.UserID
}

func getSquadIDFromPath(r *http.Request) (uuid.UUID, error) {
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

func getProposalIDFromPath(r *http.Request) (uuid.UUID, error) {
	vars := mux.Vars(r)
	proposalIDStr := vars["proposal_id"]
	if proposalIDStr == "" {
		return uuid.Nil, &wallet_in.ValidationError{Field: "proposal_id", Message: "proposal_id is required in path"}
	}
	proposalID, err := uuid.Parse(proposalIDStr)
	if err != nil {
		return uuid.Nil, &wallet_in.ValidationError{Field: "proposal_id", Message: "invalid proposal_id format"}
	}
	return proposalID, nil
}

// --- Handlers ---

// CreateVaultHandler handles POST /squads/{squad_id}/vault
func (c *VaultCommandController) CreateVaultHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := requireVaultAuth(w, r)
		if userID == uuid.Nil {
			return
		}

		squadID, err := getSquadIDFromPath(r)
		if err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
			return
		}

		var req CreateVaultRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", "")
			return
		}

		cmd := wallet_in.CreateVaultCommand{
			SquadID:     squadID,
			Name:        req.Name,
			Description: req.Description,
			UserID:      userID,
		}

		if err := cmd.Validate(); err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
			return
		}

		vault, err := c.vaultCommand.CreateVault(ctx, cmd)
		if err != nil {
			slog.ErrorContext(ctx, "CreateVaultHandler: error", "error", err)
			common.WriteErrorFromDomainError(w, err)
			return
		}

		common.WriteCreated(w, vault)
	}
}

// DepositToVaultHandler handles POST /squads/{squad_id}/vault/deposit
func (c *VaultCommandController) DepositToVaultHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := requireVaultAuth(w, r)
		if userID == uuid.Nil {
			return
		}

		squadID, err := getSquadIDFromPath(r)
		if err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
			return
		}

		var req VaultDepositRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", "")
			return
		}

		cmd := wallet_in.VaultDepositCommand{
			SquadID:        squadID,
			UserID:         userID,
			Currency:       req.Currency,
			Amount:         req.Amount,
			IdempotencyKey: req.IdempotencyKey,
		}

		if err := cmd.Validate(); err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
			return
		}

		if err := c.vaultCommand.DepositToVault(ctx, cmd); err != nil {
			slog.ErrorContext(ctx, "DepositToVaultHandler: error", "error", err)
			common.WriteErrorFromDomainError(w, err)
			return
		}

		common.WriteSuccess(w, map[string]string{"status": "deposited"})
	}
}

// ProposeTransactionHandler handles POST /squads/{squad_id}/vault/proposals
func (c *VaultCommandController) ProposeTransactionHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := requireVaultAuth(w, r)
		if userID == uuid.Nil {
			return
		}

		squadID, err := getSquadIDFromPath(r)
		if err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
			return
		}

		var req ProposeTransactionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", "")
			return
		}

		cmd := wallet_in.ProposeTransactionCommand{
			SquadID:     squadID,
			UserID:      userID,
			Type:        wallet_entities.ProposalType(req.Type),
			Title:       req.Title,
			Description: req.Description,
			Amount:      req.Amount,
			Currency:    req.Currency,
			Destination: req.Destination,
		}

		if err := cmd.Validate(); err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
			return
		}

		proposal, err := c.vaultCommand.ProposeTransaction(ctx, cmd)
		if err != nil {
			slog.ErrorContext(ctx, "ProposeTransactionHandler: error", "error", err)
			common.WriteErrorFromDomainError(w, err)
			return
		}

		common.WriteCreated(w, proposal)
	}
}

// ApproveProposalHandler handles POST /squads/{squad_id}/vault/proposals/{proposal_id}/approve
func (c *VaultCommandController) ApproveProposalHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := requireVaultAuth(w, r)
		if userID == uuid.Nil {
			return
		}

		squadID, err := getSquadIDFromPath(r)
		if err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
			return
		}

		proposalID, err := getProposalIDFromPath(r)
		if err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
			return
		}

		var req ApproveProposalRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// Allow empty body for approve
			req = ApproveProposalRequest{}
		}

		cmd := wallet_in.ApproveProposalCommand{
			SquadID:       squadID,
			ProposalID:    proposalID,
			UserID:        userID,
			Reason:        req.Reason,
			SignatureHash: req.SignatureHash,
		}

		if err := cmd.Validate(); err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
			return
		}

		if err := c.vaultCommand.ApproveProposal(ctx, cmd); err != nil {
			slog.ErrorContext(ctx, "ApproveProposalHandler: error", "error", err)
			common.WriteErrorFromDomainError(w, err)
			return
		}

		common.WriteSuccess(w, map[string]string{"status": "approved"})
	}
}

// RejectProposalHandler handles POST /squads/{squad_id}/vault/proposals/{proposal_id}/reject
func (c *VaultCommandController) RejectProposalHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := requireVaultAuth(w, r)
		if userID == uuid.Nil {
			return
		}

		squadID, err := getSquadIDFromPath(r)
		if err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
			return
		}

		proposalID, err := getProposalIDFromPath(r)
		if err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
			return
		}

		var req RejectProposalRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", "")
			return
		}

		cmd := wallet_in.RejectProposalCommand{
			SquadID:    squadID,
			ProposalID: proposalID,
			UserID:     userID,
			Reason:     req.Reason,
		}

		if err := cmd.Validate(); err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
			return
		}

		if err := c.vaultCommand.RejectProposal(ctx, cmd); err != nil {
			slog.ErrorContext(ctx, "RejectProposalHandler: error", "error", err)
			common.WriteErrorFromDomainError(w, err)
			return
		}

		common.WriteSuccess(w, map[string]string{"status": "rejected"})
	}
}

// CancelProposalHandler handles POST /squads/{squad_id}/vault/proposals/{proposal_id}/cancel
func (c *VaultCommandController) CancelProposalHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := requireVaultAuth(w, r)
		if userID == uuid.Nil {
			return
		}

		squadID, err := getSquadIDFromPath(r)
		if err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
			return
		}

		proposalID, err := getProposalIDFromPath(r)
		if err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
			return
		}

		cmd := wallet_in.CancelProposalCommand{
			SquadID:    squadID,
			ProposalID: proposalID,
			UserID:     userID,
		}

		if err := cmd.Validate(); err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
			return
		}

		if err := c.vaultCommand.CancelProposal(ctx, cmd); err != nil {
			slog.ErrorContext(ctx, "CancelProposalHandler: error", "error", err)
			common.WriteErrorFromDomainError(w, err)
			return
		}

		common.WriteSuccess(w, map[string]string{"status": "cancelled"})
	}
}

// UpdateVaultSettingsHandler handles PUT /squads/{squad_id}/vault/settings
func (c *VaultCommandController) UpdateVaultSettingsHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := requireVaultAuth(w, r)
		if userID == uuid.Nil {
			return
		}

		squadID, err := getSquadIDFromPath(r)
		if err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
			return
		}

		var req UpdateVaultSettingsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", "")
			return
		}

		cmd := wallet_in.UpdateVaultSettingsCommand{
			SquadID:            squadID,
			UserID:             userID,
			ApprovalPolicy:     req.ApprovalPolicy,
			OnChainThreshold:   req.OnChainThreshold,
			DailyWithdrawLimit: req.DailyWithdrawLimit,
		}

		if err := cmd.Validate(); err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
			return
		}

		proposal, err := c.vaultCommand.UpdateVaultSettings(ctx, cmd)
		if err != nil {
			slog.ErrorContext(ctx, "UpdateVaultSettingsHandler: error", "error", err)
			common.WriteErrorFromDomainError(w, err)
			return
		}

		common.WriteSuccess(w, proposal)
	}
}

// DepositItemHandler handles POST /squads/{squad_id}/vault/inventory
func (c *VaultCommandController) DepositItemHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := requireVaultAuth(w, r)
		if userID == uuid.Nil {
			return
		}

		squadID, err := getSquadIDFromPath(r)
		if err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
			return
		}

		var req DepositItemRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", "")
			return
		}

		itemID, err := uuid.Parse(req.ItemID)
		if err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid item_id format", "")
			return
		}

		cmd := wallet_in.DepositItemCommand{
			SquadID: squadID,
			UserID:  userID,
			ItemID:  itemID,
		}

		if err := cmd.Validate(); err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
			return
		}

		if err := c.vaultCommand.DepositItem(ctx, cmd); err != nil {
			slog.ErrorContext(ctx, "DepositItemHandler: error", "error", err)
			common.WriteErrorFromDomainError(w, err)
			return
		}

		common.WriteSuccess(w, map[string]string{"status": "deposited"})
	}
}

// ProposeItemTransferHandler handles POST /squads/{squad_id}/vault/inventory/transfer
func (c *VaultCommandController) ProposeItemTransferHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := requireVaultAuth(w, r)
		if userID == uuid.Nil {
			return
		}

		squadID, err := getSquadIDFromPath(r)
		if err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
			return
		}

		var req ProposeItemTransferRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", "")
			return
		}

		var itemIDs []uuid.UUID
		for _, idStr := range req.InventoryItemIDs {
			id, err := uuid.Parse(idStr)
			if err != nil {
				common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid inventory_item_id format: "+idStr, "")
				return
			}
			itemIDs = append(itemIDs, id)
		}

		destUserID, err := uuid.Parse(req.DestinationUserID)
		if err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid destination_user_id format", "")
			return
		}

		cmd := wallet_in.ProposeItemTransferCommand{
			SquadID:           squadID,
			UserID:            userID,
			Title:             req.Title,
			Description:       req.Description,
			InventoryItemIDs:  itemIDs,
			DestinationUserID: destUserID,
		}

		if err := cmd.Validate(); err != nil {
			common.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "")
			return
		}

		proposal, err := c.vaultCommand.ProposeItemTransfer(ctx, cmd)
		if err != nil {
			slog.ErrorContext(ctx, "ProposeItemTransferHandler: error", "error", err)
			common.WriteErrorFromDomainError(w, err)
			return
		}

		common.WriteCreated(w, proposal)
	}
}
