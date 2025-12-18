// Package controllers provides HTTP handlers for the REST API.
package controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	auth_entities "github.com/replay-api/replay-api/pkg/domain/auth/entities"
	auth_in "github.com/replay-api/replay-api/pkg/domain/auth/ports/in"
	email_out "github.com/replay-api/replay-api/pkg/domain/email/ports/out"
)

// EmailVerificationController handles email verification endpoints
type EmailVerificationController struct {
	container          *container.Container
	verificationService auth_in.EmailVerificationCommand
	emailUserReader    email_out.EmailUserReader
	emailUserWriter    email_out.EmailUserWriter
}

// NewEmailVerificationController creates a new email verification controller
func NewEmailVerificationController(c *container.Container) *EmailVerificationController {
	ctrl := &EmailVerificationController{container: c}

	if err := c.Resolve(&ctrl.verificationService); err != nil {
		slog.Error("Failed to resolve EmailVerificationCommand", "err", err)
	}
	if err := c.Resolve(&ctrl.emailUserReader); err != nil {
		slog.Error("Failed to resolve EmailUserReader", "err", err)
	}
	if err := c.Resolve(&ctrl.emailUserWriter); err != nil {
		slog.Error("Failed to resolve EmailUserWriter", "err", err)
	}

	return ctrl
}

// SendVerificationRequest is the request payload for sending verification
type SendVerificationRequest struct {
	Email string `json:"email"`
}

// VerifyEmailRequest is the request payload for email verification
type VerifyEmailRequest struct {
	Token string `json:"token,omitempty"`
	Code  string `json:"code,omitempty"`
}

// ResendVerificationRequest is the request payload for resending verification
type ResendVerificationRequest struct {
	Email string `json:"email"`
}

// SendVerificationEmail handles POST /auth/verification/send
func (ctrl *EmailVerificationController) SendVerificationEmail(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resourceOwner := common.GetResourceOwner(r.Context())
		if resourceOwner.UserID == uuid.Nil {
			writeJSONError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		var req SendVerificationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Email == "" {
			writeJSONError(w, http.StatusBadRequest, "email is required")
			return
		}

		// Get client IP and user agent for audit
		ipAddress := getClientIP(r)
		userAgent := r.UserAgent()

		cmd := auth_in.SendVerificationEmailCommand{
			UserID:    resourceOwner.UserID,
			Email:     req.Email,
			Type:      auth_entities.VerificationTypeEmail,
			IPAddress: ipAddress,
			UserAgent: userAgent,
		}

		verification, err := ctrl.verificationService.SendVerificationEmail(r.Context(), cmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to send verification email", "error", err)
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Return limited info (don't expose token/code)
		response := map[string]interface{}{
			"message":    "Verification email sent successfully",
			"email":      req.Email,
			"expires_at": verification.ExpiresAt,
		}

		writeJSON(w, http.StatusOK, response)
	}
}

// VerifyEmail handles POST /auth/verify-email
func (ctrl *EmailVerificationController) VerifyEmail(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req VerifyEmailRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Token == "" && req.Code == "" {
			writeJSONError(w, http.StatusBadRequest, "token or code is required")
			return
		}

		// Get client IP and user agent for audit
		ipAddress := getClientIP(r)
		userAgent := r.UserAgent()

		cmd := auth_in.VerifyEmailCommand{
			Token:     req.Token,
			Code:      req.Code,
			IPAddress: ipAddress,
			UserAgent: userAgent,
		}

		result, err := ctrl.verificationService.VerifyEmail(r.Context(), cmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to verify email", "error", err)
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if !result.Success {
			status := http.StatusBadRequest
			if result.RemainingAttempts == 0 {
				status = http.StatusTooManyRequests
			}
			writeJSON(w, status, result)
			return
		}

		// TODO: Update email user's EmailVerified field to true
		// This would require looking up the verification to get the user ID

		writeJSON(w, http.StatusOK, result)
	}
}

// VerifyEmailByToken handles GET /auth/verify-email?token=xxx
// This is for email link verification
func (ctrl *EmailVerificationController) VerifyEmailByToken(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			writeJSONError(w, http.StatusBadRequest, "token is required")
			return
		}

		cmd := auth_in.VerifyEmailCommand{
			Token:     token,
			IPAddress: getClientIP(r),
			UserAgent: r.UserAgent(),
		}

		result, err := ctrl.verificationService.VerifyEmail(r.Context(), cmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to verify email", "error", err)
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

// ResendVerification handles POST /auth/verification/resend
func (ctrl *EmailVerificationController) ResendVerification(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resourceOwner := common.GetResourceOwner(r.Context())
		if resourceOwner.UserID == uuid.Nil {
			writeJSONError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		var req ResendVerificationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Email == "" {
			writeJSONError(w, http.StatusBadRequest, "email is required")
			return
		}

		cmd := auth_in.ResendVerificationCommand{
			UserID:    resourceOwner.UserID,
			Email:     req.Email,
			IPAddress: getClientIP(r),
			UserAgent: r.UserAgent(),
		}

		verification, err := ctrl.verificationService.ResendVerification(r.Context(), cmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to resend verification", "error", err)
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}

		response := map[string]interface{}{
			"message":    "Verification email resent successfully",
			"email":      req.Email,
			"expires_at": verification.ExpiresAt,
		}

		writeJSON(w, http.StatusOK, response)
	}
}

// GetVerificationStatus handles GET /auth/verification/status
func (ctrl *EmailVerificationController) GetVerificationStatus(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resourceOwner := common.GetResourceOwner(r.Context())
		if resourceOwner.UserID == uuid.Nil {
			writeJSONError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		verification, err := ctrl.verificationService.GetVerificationStatus(r.Context(), resourceOwner.UserID)
		if err != nil {
			// No verification found is not an error
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"status":  "none",
				"message": "No pending verification",
			})
			return
		}

		response := map[string]interface{}{
			"status":             verification.Status,
			"type":               verification.Type,
			"email":              verification.Email,
			"expires_at":         verification.ExpiresAt,
			"remaining_attempts": verification.RemainingAttempts(),
			"is_expired":         verification.IsExpired(),
			"is_verified":        verification.IsVerified(),
		}

		writeJSON(w, http.StatusOK, response)
	}
}

// Helper functions

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxied requests)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

