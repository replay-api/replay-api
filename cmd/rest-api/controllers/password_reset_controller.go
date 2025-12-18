// Package controllers provides HTTP handlers for the REST API.
package controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	auth_in "github.com/replay-api/replay-api/pkg/domain/auth/ports/in"
)

// PasswordResetController handles password reset endpoints
type PasswordResetController struct {
	container    *container.Container
	resetService auth_in.PasswordResetCommand
}

// NewPasswordResetController creates a new password reset controller
func NewPasswordResetController(c *container.Container) *PasswordResetController {
	ctrl := &PasswordResetController{container: c}

	if err := c.Resolve(&ctrl.resetService); err != nil {
		slog.Error("Failed to resolve PasswordResetCommand", "err", err)
	}

	return ctrl
}

// RequestPasswordResetRequest is the request payload for requesting password reset
type RequestPasswordResetRequest struct {
	Email string `json:"email"`
}

// ConfirmPasswordResetRequest is the request payload for confirming password reset
type ConfirmPasswordResetRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// RequestPasswordReset handles POST /auth/password-reset
func (ctrl *PasswordResetController) RequestPasswordReset(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RequestPasswordResetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Email == "" {
			writeJSONError(w, http.StatusBadRequest, "email is required")
			return
		}

		cmd := auth_in.RequestPasswordResetCommand{
			Email:     req.Email,
			IPAddress: getClientIP(r),
			UserAgent: r.UserAgent(),
		}

		reset, err := ctrl.resetService.RequestPasswordReset(r.Context(), cmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to request password reset", "error", err)
			// Don't reveal specific errors to prevent email enumeration
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"success": true,
				"message": "If an account exists with this email, a password reset link will be sent",
			})
			return
		}

		// Even if reset is nil (email not found), return success to prevent enumeration
		response := map[string]interface{}{
			"success": true,
			"message": "If an account exists with this email, a password reset link will be sent",
		}

		if reset != nil {
			response["expires_at"] = reset.ExpiresAt
		}

		writeJSON(w, http.StatusOK, response)
	}
}

// ConfirmPasswordReset handles POST /auth/password-reset/confirm
func (ctrl *PasswordResetController) ConfirmPasswordReset(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ConfirmPasswordResetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Token == "" {
			writeJSONError(w, http.StatusBadRequest, "token is required")
			return
		}

		if req.NewPassword == "" {
			writeJSONError(w, http.StatusBadRequest, "new_password is required")
			return
		}

		if len(req.NewPassword) < 8 {
			writeJSONError(w, http.StatusBadRequest, "password must be at least 8 characters")
			return
		}

		cmd := auth_in.ConfirmPasswordResetCommand{
			Token:       req.Token,
			NewPassword: req.NewPassword,
			IPAddress:   getClientIP(r),
			UserAgent:   r.UserAgent(),
		}

		result, err := ctrl.resetService.ConfirmPasswordReset(r.Context(), cmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to confirm password reset", "error", err)
			writeJSONError(w, http.StatusInternalServerError, "failed to reset password")
			return
		}

		if !result.Success {
			writeJSON(w, http.StatusBadRequest, result)
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

// ValidateResetToken handles GET /auth/password-reset/validate?token=xxx
func (ctrl *PasswordResetController) ValidateResetToken(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			writeJSONError(w, http.StatusBadRequest, "token is required")
			return
		}

		reset, err := ctrl.resetService.ValidateResetToken(r.Context(), token)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"valid":   false,
				"message": err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"valid":      true,
			"email":      reset.Email,
			"expires_at": reset.ExpiresAt,
		})
	}
}

