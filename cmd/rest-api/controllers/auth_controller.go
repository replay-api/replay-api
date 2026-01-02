package controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	iam_in "github.com/replay-api/replay-api/pkg/domain/iam/ports/in"
)

// AuthController handles authentication operations like token refresh and logout
type AuthController struct {
	RefreshRIDTokenCommand iam_in.RefreshRIDTokenCommand
	RevokeRIDTokenCommand  iam_in.RevokeRIDTokenCommand
	CreateRIDTokenCommand  iam_in.CreateRIDTokenCommand
}

// NewAuthController creates a new AuthController with resolved dependencies
func NewAuthController(container *container.Container) *AuthController {
	var refreshRIDTokenCommand iam_in.RefreshRIDTokenCommand
	err := container.Resolve(&refreshRIDTokenCommand)
	if err != nil {
		slog.Warn("RefreshRIDTokenCommand not available", "error", err)
	}

	var revokeRIDTokenCommand iam_in.RevokeRIDTokenCommand
	err = container.Resolve(&revokeRIDTokenCommand)
	if err != nil {
		slog.Warn("RevokeRIDTokenCommand not available", "error", err)
	}

	var createRIDTokenCommand iam_in.CreateRIDTokenCommand
	err = container.Resolve(&createRIDTokenCommand)
	if err != nil {
		slog.Warn("CreateRIDTokenCommand not available", "error", err)
	}

	return &AuthController{
		RefreshRIDTokenCommand: refreshRIDTokenCommand,
		RevokeRIDTokenCommand:  revokeRIDTokenCommand,
		CreateRIDTokenCommand:  createRIDTokenCommand,
	}
}

// RefreshTokenRequest represents the request body for token refresh
type RefreshTokenRequest struct {
	TokenID string `json:"token_id"`
}

// RefreshTokenResponse represents the response for token refresh
type RefreshTokenResponse struct {
	Success bool   `json:"success"`
	TokenID string `json:"token_id,omitempty"`
	Message string `json:"message,omitempty"`
}

// LogoutRequest represents the request body for logout
type LogoutRequest struct {
	TokenID string `json:"token_id"`
}

// LogoutResponse represents the response for logout
type LogoutResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func setCORSHeaders(w http.ResponseWriter) {
	corsOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "http://localhost:3030"
	}
	w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Resource-Owner-ID")
	w.Header().Set("Access-Control-Expose-Headers", "X-Resource-Owner-ID, X-Intended-Audience")
}

// RefreshToken handles POST /auth/refresh
// Refreshes an existing token with a new expiration time
func (c *AuthController) RefreshToken(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setCORSHeaders(w)

		// Check authentication
		authenticated, ok := r.Context().Value(common.AuthenticatedKey).(bool)
		if !ok || !authenticated {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(RefreshTokenResponse{
				Success: false,
				Message: "Authentication required",
			})
			return
		}

		if r.Body == nil {
			slog.ErrorContext(r.Context(), "no request body for token refresh")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(RefreshTokenResponse{
				Success: false,
				Message: "Request body is required",
			})
			return
		}

		decoder := json.NewDecoder(r.Body)
		var req RefreshTokenRequest
		err := decoder.Decode(&req)

		if err != nil {
			slog.ErrorContext(r.Context(), "error decoding refresh token request", "err", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(RefreshTokenResponse{
				Success: false,
				Message: "Invalid request body",
			})
			return
		}

		// Also check for token ID from header if not in body
		if req.TokenID == "" {
			req.TokenID = r.Header.Get(ResourceOwnerIDHeaderKey)
		}

		if req.TokenID == "" {
			slog.ErrorContext(r.Context(), "token_id is required for refresh")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(RefreshTokenResponse{
				Success: false,
				Message: "token_id is required",
			})
			return
		}

		tokenID, err := uuid.Parse(req.TokenID)
		if err != nil {
			slog.ErrorContext(r.Context(), "invalid token_id format", "tokenID", req.TokenID, "err", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(RefreshTokenResponse{
				Success: false,
				Message: "Invalid token_id format",
			})
			return
		}

		newToken, err := c.RefreshRIDTokenCommand.Exec(r.Context(), tokenID)
		if err != nil {
			slog.ErrorContext(r.Context(), "error refreshing token", "err", err, "tokenID", req.TokenID)
			
			statusCode := http.StatusInternalServerError
			message := "Failed to refresh token"
			
			errMsg := err.Error()
			if strings.Contains(errMsg, "expired") {
				statusCode = http.StatusUnauthorized
				message = "Token has expired, please log in again"
			} else if strings.Contains(errMsg, "not found") {
				statusCode = http.StatusNotFound
				message = "Token not found"
			} else if strings.Contains(errMsg, "unauthorized") {
				statusCode = http.StatusForbidden
				message = "You are not authorized to refresh this token"
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			_ = json.NewEncoder(w).Encode(RefreshTokenResponse{
				Success: false,
				Message: message,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set(ResourceOwnerIDHeaderKey, newToken.ID.String())
		w.Header().Set(ResourceOwnerAudTypeHeaderKey, string(newToken.IntendedAudience))
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(RefreshTokenResponse{
			Success: true,
			TokenID: newToken.ID.String(),
			Message: "Token refreshed successfully",
		})
	}
}

// Logout handles POST /auth/logout
// Revokes the current token, effectively logging the user out
func (c *AuthController) Logout(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setCORSHeaders(w)

		// Check authentication - logout should work even for expired tokens
		// but we need some form of token identification

		if r.Body == nil {
			slog.ErrorContext(r.Context(), "no request body for logout")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(LogoutResponse{
				Success: false,
				Message: "Request body is required",
			})
			return
		}

		decoder := json.NewDecoder(r.Body)
		var req LogoutRequest
		err := decoder.Decode(&req)

		if err != nil {
			slog.ErrorContext(r.Context(), "error decoding logout request", "err", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(LogoutResponse{
				Success: false,
				Message: "Invalid request body",
			})
			return
		}

		// Also check for token ID from header if not in body
		if req.TokenID == "" {
			req.TokenID = r.Header.Get(ResourceOwnerIDHeaderKey)
		}

		if req.TokenID == "" {
			slog.ErrorContext(r.Context(), "token_id is required for logout")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(LogoutResponse{
				Success: false,
				Message: "token_id is required",
			})
			return
		}

		tokenID, err := uuid.Parse(req.TokenID)
		if err != nil {
			slog.ErrorContext(r.Context(), "invalid token_id format for logout", "tokenID", req.TokenID, "err", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(LogoutResponse{
				Success: false,
				Message: "Invalid token_id format",
			})
			return
		}

		err = c.RevokeRIDTokenCommand.Exec(r.Context(), tokenID)
		if err != nil {
			slog.ErrorContext(r.Context(), "error revoking token", "err", err, "tokenID", req.TokenID)
			
			// Most logout errors should still return success to the client
			// We don't want to expose internal errors or prevent users from "logging out" in the UI
			errMsg := err.Error()
			if strings.Contains(errMsg, "unauthorized") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(LogoutResponse{
					Success: false,
					Message: "You are not authorized to revoke this token",
				})
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(LogoutResponse{
			Success: true,
			Message: "Logged out successfully",
		})
	}
}

// GuestTokenResponse represents the response for guest token creation
type GuestTokenResponse struct {
	Success       bool   `json:"success"`
	TokenID       string `json:"token_id,omitempty"`
	UserID        string `json:"user_id,omitempty"`
	ExpiresAt     string `json:"expires_at,omitempty"`
	Message       string `json:"message,omitempty"`
}

// CreateGuestToken handles POST /auth/guest
// Creates a new guest token for unauthenticated users
// This allows guests to have a session and potentially convert to full accounts later
func (c *AuthController) CreateGuestToken(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setCORSHeaders(w)

		// Check if CreateRIDTokenCommand is available
		if c.CreateRIDTokenCommand == nil {
			slog.ErrorContext(r.Context(), "CreateRIDTokenCommand not available for guest token creation")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(GuestTokenResponse{
				Success: false,
				Message: "Guest token service unavailable",
			})
			return
		}

		// Generate a unique user ID for the guest
		guestUserID := uuid.New()
		guestGroupID := uuid.New()

		// Create resource owner for the guest
		resourceOwner := common.ResourceOwner{
			TenantID: common.TeamPROTenantID,
			ClientID: common.TeamPROAppClientID,
			GroupID:  guestGroupID,
			UserID:   guestUserID,
		}

		// Create the guest token
		ridToken, err := c.CreateRIDTokenCommand.Exec(
			r.Context(),
			resourceOwner,
			iam_entities.RIDSource_Guest,
			common.UserAudienceIDKey,
		)

		if err != nil {
			slog.ErrorContext(r.Context(), "error creating guest token", "err", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(GuestTokenResponse{
				Success: false,
				Message: "Failed to create guest token",
			})
			return
		}

		if ridToken == nil {
			slog.ErrorContext(r.Context(), "guest token creation returned nil")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(GuestTokenResponse{
				Success: false,
				Message: "Failed to create guest token",
			})
			return
		}

		// Set headers for the new token
		w.Header().Set(ResourceOwnerIDHeaderKey, ridToken.GetID().String())
		w.Header().Set(ResourceOwnerAudTypeHeaderKey, string(ridToken.IntendedAudience))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(GuestTokenResponse{
			Success:   true,
			TokenID:   ridToken.GetID().String(),
			UserID:    guestUserID.String(),
			ExpiresAt: ridToken.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
			Message:   "Guest token created successfully",
		})
	}
}