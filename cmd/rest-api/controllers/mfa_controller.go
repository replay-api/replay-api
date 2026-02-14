package controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	auth_in "github.com/replay-api/replay-api/pkg/domain/auth/ports/in"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type MFAController struct {
	mfaCommand auth_in.MFACommand
}

func NewMFAController(c container.Container) *MFAController {
	var mfaCommand auth_in.MFACommand
	if err := c.Resolve(&mfaCommand); err != nil {
		slog.Warn("MFACommand not available", "error", err)
		return &MFAController{}
	}

	return &MFAController{
		mfaCommand: mfaCommand,
	}
}

// SetupMFAHandler handles POST /auth/mfa/setup
func (c *MFAController) SetupMFAHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if c.mfaCommand == nil {
			http.Error(w, `{"error":"MFA service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		// Check authentication
		authenticated, ok := r.Context().Value(shared.AuthenticatedKey).(bool)
		if !ok || !authenticated {
			http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
			return
		}

		resourceOwner := shared.GetResourceOwner(r.Context())
		if resourceOwner.UserID == uuid.Nil {
			http.Error(w, `{"error":"valid user authentication required"}`, http.StatusUnauthorized)
			return
		}

		var req struct {
			Email string `json:"email"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}

		setupResponse, err := c.mfaCommand.SetupTOTP(r.Context(), resourceOwner.UserID, req.Email)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to setup MFA", "error", err, "user_id", resourceOwner.UserID)
			http.Error(w, `{"error":"Failed to setup MFA"}`, http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(setupResponse)
	}
}

// VerifyMFAHandler handles POST /auth/mfa/verify
func (c *MFAController) VerifyMFAHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if c.mfaCommand == nil {
			http.Error(w, `{"error":"MFA service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		// Check authentication
		authenticated, ok := r.Context().Value(shared.AuthenticatedKey).(bool)
		if !ok || !authenticated {
			http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
			return
		}

		resourceOwner := shared.GetResourceOwner(r.Context())
		if resourceOwner.UserID == uuid.Nil {
			http.Error(w, `{"error":"valid user authentication required"}`, http.StatusUnauthorized)
			return
		}

		var req struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}

		err := c.mfaCommand.VerifyAndActivate(r.Context(), resourceOwner.UserID, req.Code)
		if err != nil {
			slog.WarnContext(r.Context(), "MFA verification failed", "error", err, "user_id", resourceOwner.UserID)
			http.Error(w, `{"error":"MFA verification failed"}`, http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]bool{"verified": true})
	}
}

// ValidateMFAHandler handles POST /auth/mfa/validate (for login)
// SECURITY: Derives user ID from auth context, NOT from request body
func (c *MFAController) ValidateMFAHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if c.mfaCommand == nil {
			http.Error(w, `{"error":"MFA service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		// SECURITY: Derive user ID from auth context, NOT from request body.
		// During MFA validation, the user must already have partial authentication
		// (e.g., password verified), so auth context should be set.
		authenticated, ok := r.Context().Value(shared.AuthenticatedKey).(bool)
		if !ok || !authenticated {
			http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
			return
		}

		resourceOwner := shared.GetResourceOwner(r.Context())
		if resourceOwner.UserID == uuid.Nil {
			http.Error(w, `{"error":"valid user authentication required"}`, http.StatusUnauthorized)
			return
		}
		userID := resourceOwner.UserID

		var req struct {
			Code      string `json:"code"`
			UseBackup bool   `json:"use_backup"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}

		var err error
		if req.UseBackup {
			err = c.mfaCommand.VerifyBackupCode(r.Context(), userID, req.Code)
		} else {
			err = c.mfaCommand.Verify(r.Context(), userID, req.Code)
		}

		if err != nil {
			slog.WarnContext(r.Context(), "MFA validation failed", "error", err, "user_id", userID, "use_backup", req.UseBackup)
			http.Error(w, `{"error":"invalid code"}`, http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]bool{"valid": true})
	}
}

// DisableMFAHandler handles POST /auth/mfa/disable
func (c *MFAController) DisableMFAHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if c.mfaCommand == nil {
			http.Error(w, `{"error":"MFA service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		// Check authentication
		authenticated, ok := r.Context().Value(shared.AuthenticatedKey).(bool)
		if !ok || !authenticated {
			http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
			return
		}

		resourceOwner := shared.GetResourceOwner(r.Context())
		if resourceOwner.UserID == uuid.Nil {
			http.Error(w, `{"error":"valid user authentication required"}`, http.StatusUnauthorized)
			return
		}

		var req struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}

		// Verify code before disabling
		err := c.mfaCommand.Verify(r.Context(), resourceOwner.UserID, req.Code)
		if err != nil {
			http.Error(w, `{"error":"invalid code"}`, http.StatusUnauthorized)
			return
		}

		err = c.mfaCommand.Disable(r.Context(), resourceOwner.UserID)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to disable MFA", "error", err, "user_id", resourceOwner.UserID)
			http.Error(w, `{"error":"Failed to disable MFA"}`, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]bool{"disabled": true})
	}
}

// GetMFAStatusHandler handles GET /auth/mfa/status
func (c *MFAController) GetMFAStatusHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if c.mfaCommand == nil {
			http.Error(w, `{"error":"MFA service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		// Check authentication
		authenticated, ok := r.Context().Value(shared.AuthenticatedKey).(bool)
		if !ok || !authenticated {
			http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
			return
		}

		resourceOwner := shared.GetResourceOwner(r.Context())
		if resourceOwner.UserID == uuid.Nil {
			http.Error(w, `{"error":"valid user authentication required"}`, http.StatusUnauthorized)
			return
		}

		status, err := c.mfaCommand.GetStatus(r.Context(), resourceOwner.UserID)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to get MFA status", "error", err, "user_id", resourceOwner.UserID)
			http.Error(w, `{"error":"Failed to get MFA status"}`, http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"enabled": false,
			"method":  nil,
		}

		if status != nil {
			response["enabled"] = status.IsActive()
			response["method"] = status.Method
			response["backup_codes_left"] = status.BackupCodesLeft
			if status.VerifiedAt != nil {
				response["verified_at"] = status.VerifiedAt
			}
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}

// RegenerateBackupCodesHandler handles POST /auth/mfa/backup-codes/regenerate
func (c *MFAController) RegenerateBackupCodesHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if c.mfaCommand == nil {
			http.Error(w, `{"error":"MFA service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		// Check authentication
		authenticated, ok := r.Context().Value(shared.AuthenticatedKey).(bool)
		if !ok || !authenticated {
			http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
			return
		}

		resourceOwner := shared.GetResourceOwner(r.Context())
		if resourceOwner.UserID == uuid.Nil {
			http.Error(w, `{"error":"valid user authentication required"}`, http.StatusUnauthorized)
			return
		}

		var req struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}

		newCodes, err := c.mfaCommand.RegenerateBackupCodes(r.Context(), resourceOwner.UserID, req.Code)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to regenerate backup codes", "error", err, "user_id", resourceOwner.UserID)
			http.Error(w, `{"error":"Failed to regenerate backup codes"}`, http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"backup_codes": newCodes,
		})
	}
}

