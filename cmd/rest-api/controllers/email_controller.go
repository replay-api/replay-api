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
		slog.Error("Cannot resolve email_in.OnboardEmailUserCommand for new EmailController", "err", err)
		panic(err)
	}

	var loginEmailUserCommand email_in.LoginEmailUserCommand
	err = container.Resolve(&loginEmailUserCommand)

	if err != nil {
		slog.Error("Cannot resolve email_in.LoginEmailUserCommand for new EmailController", "err", err)
		panic(err)
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

		// Create EmailUser entity from request
		emailUser := &email_entities.EmailUser{
			Email:       req.Email,
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
			// Check for specific errors
			errMsg := err.Error()
			if strings.Contains(errMsg, "already exists") {
				http.Error(w, "Conflict: "+errMsg, http.StatusConflict)
				return
			}
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if ridToken == nil {
			slog.ErrorContext(r.Context(), "error onboarding email user", "err", "controller: ridToken is nil", "email", req.Email)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set(ResourceOwnerIDHeaderKey, ridToken.GetID().String())
		w.Header().Set(ResourceOwnerAudTypeHeaderKey, string(ridToken.IntendedAudience))
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resultUser)
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

		err = c.LoginEmailUserCommand.Validate(r.Context(), req.Email, req.Password, req.VHash)

		if err != nil {
			slog.ErrorContext(r.Context(), "error validating login request", "err", err, "email", req.Email)
			http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
			return
		}

		emailUser, ridToken, err := c.LoginEmailUserCommand.Exec(r.Context(), req.Email, req.Password, req.VHash)

		if err != nil {
			slog.ErrorContext(r.Context(), "error logging in email user", "err", err, "email", req.Email)
			errMsg := err.Error()
			if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "invalid password") {
				http.Error(w, "Unauthorized: invalid credentials", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if ridToken == nil {
			slog.ErrorContext(r.Context(), "error logging in email user", "err", "controller: ridToken is nil", "email", req.Email)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set(ResourceOwnerIDHeaderKey, ridToken.GetID().String())
		w.Header().Set(ResourceOwnerAudTypeHeaderKey, string(ridToken.IntendedAudience))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(emailUser)
	}
}
