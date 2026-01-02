package controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/golobby/container/v3"
	email_entities "github.com/replay-api/replay-api/pkg/domain/email/entities"
	email_in "github.com/replay-api/replay-api/pkg/domain/email/ports/in"
)

type EmailController struct {
	OnboardEmailUserCommand email_in.OnboardEmailUserCommand
	LoginEmailUserCommand   email_in.LoginEmailUserCommand
}

func NewEmailController(container *container.Container) *EmailController {
	var onboardEmailUserCommand email_in.OnboardEmailUserCommand
	err := container.Resolve(&onboardEmailUserCommand)

	if err != nil {
		slog.Warn("Cannot resolve email_in.OnboardEmailUserCommand for new EmailController - Email registration will be disabled", "err", err)
	}

	var loginEmailUserCommand email_in.LoginEmailUserCommand
	err = container.Resolve(&loginEmailUserCommand)

	if err != nil {
		slog.Warn("Cannot resolve email_in.LoginEmailUserCommand for new EmailController - Email login will be disabled", "err", err)
	}

	return &EmailController{
		OnboardEmailUserCommand: onboardEmailUserCommand,
		LoginEmailUserCommand:   loginEmailUserCommand,
	}
}

type OnboardEmailRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	VHash       string `json:"v_hash"`
	DisplayName string `json:"display_name,omitempty"`
}

type LoginEmailRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	VHash    string `json:"v_hash"`
}

func (c *EmailController) OnboardEmailUser(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		corsOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
		if corsOrigin == "" {
			corsOrigin = "http://localhost:3030"
		}
		w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "X-Resource-Owner-ID, X-Intended-Audience")

		// Check if OnboardEmailUserCommand is available
		if c.OnboardEmailUserCommand == nil {
			slog.WarnContext(r.Context(), "Email user onboarding not available - OnboardEmailUserCommand not registered")
			http.Error(w, "Service Temporarily Unavailable", http.StatusServiceUnavailable)
			return
		}

		if r.Body == nil {
			slog.ErrorContext(r.Context(), "no request body", "request.Body", r.Body)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		decoder := json.NewDecoder(r.Body)
		var req OnboardEmailRequest
		err := decoder.Decode(&req)

		if err != nil {
			slog.ErrorContext(r.Context(), "error decoding email user from request", "err", err, "request.body", r.Body)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Email == "" || req.Password == "" {
			slog.ErrorContext(r.Context(), "missing required fields for email registration")
			http.Error(w, "Email and password are required", http.StatusBadRequest)
			return
		}

		// Create email user entity
		emailUser := &email_entities.EmailUser{
			Email:       strings.ToLower(strings.TrimSpace(req.Email)),
			VHash:       req.VHash,
			DisplayName: req.DisplayName,
		}

		err = c.OnboardEmailUserCommand.Validate(r.Context(), emailUser, req.Password)

		if err != nil {
			slog.ErrorContext(r.Context(), "error validating email user", "err", err, "email", req.Email)
			http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
			return
		}

		resultUser, ridToken, err := c.OnboardEmailUserCommand.Exec(r.Context(), emailUser, req.Password)

		if err != nil {
			slog.ErrorContext(r.Context(), "error onboarding email user", "err", err, "email", req.Email)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if ridToken == nil {
			slog.ErrorContext(r.Context(), "error onboarding email user - ridToken is nil", "email", req.Email)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set(ResourceOwnerIDHeaderKey, ridToken.GetID().String())
		w.Header().Set(ResourceOwnerAudTypeHeaderKey, string(ridToken.IntendedAudience))
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(resultUser)
	}
}

func (c *EmailController) LoginEmailUser(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		corsOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
		if corsOrigin == "" {
			corsOrigin = "http://localhost:3030"
		}
		w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "X-Resource-Owner-ID, X-Intended-Audience")

		// Check if LoginEmailUserCommand is available
		if c.LoginEmailUserCommand == nil {
			slog.WarnContext(r.Context(), "Email user login not available - LoginEmailUserCommand not registered")
			http.Error(w, "Service Temporarily Unavailable", http.StatusServiceUnavailable)
			return
		}

		if r.Body == nil {
			slog.ErrorContext(r.Context(), "no request body", "request.Body", r.Body)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		decoder := json.NewDecoder(r.Body)
		var req LoginEmailRequest
		err := decoder.Decode(&req)

		if err != nil {
			slog.ErrorContext(r.Context(), "error decoding login request", "err", err, "request.body", r.Body)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Email == "" || req.Password == "" {
			slog.ErrorContext(r.Context(), "missing required fields for email login")
			http.Error(w, "Email and password are required", http.StatusBadRequest)
			return
		}

		email := strings.ToLower(strings.TrimSpace(req.Email))

		emailUser, ridToken, err := c.LoginEmailUserCommand.Exec(r.Context(), email, req.Password, req.VHash)

		if err != nil {
			slog.ErrorContext(r.Context(), "error logging in email user", "err", err, "email", email)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if ridToken == nil {
			slog.ErrorContext(r.Context(), "error logging in email user - ridToken is nil", "email", email)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set(ResourceOwnerIDHeaderKey, ridToken.GetID().String())
		w.Header().Set(ResourceOwnerAudTypeHeaderKey, string(ridToken.IntendedAudience))
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(emailUser)
	}
}
