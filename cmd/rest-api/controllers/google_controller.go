package controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/golobby/container/v3"
	google_entity "github.com/replay-api/replay-api/pkg/domain/google/entities"
	google_in "github.com/replay-api/replay-api/pkg/domain/google/ports/in"
)

type GoogleController struct {
	OnboardGoogleUserCommand google_in.OnboardGoogleUserCommand
}

func NewGoogleController(container *container.Container) *GoogleController {
	var onboardGoogleUserCommand google_in.OnboardGoogleUserCommand
	err := container.Resolve(&onboardGoogleUserCommand)

	if err != nil {
		slog.Warn("Cannot resolve google_in.OnboardGoogleUserCommand for new GoogleController - Google auth will be disabled", "err", err)
	}

	return &GoogleController{
		OnboardGoogleUserCommand: onboardGoogleUserCommand,
	}
}

func (c *GoogleController) OnboardGoogleUser(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		corsOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
		if corsOrigin == "" {
			corsOrigin = "http://localhost:3030"
		}
		w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "X-Resource-Owner-ID, X-Intended-Audience")

		// Check if OnboardGoogleUserCommand is available
		if c.OnboardGoogleUserCommand == nil {
			slog.WarnContext(r.Context(), "Google user onboarding not available - OnboardGoogleUserCommand not registered")
			http.Error(w, "Service Temporarily Unavailable", http.StatusServiceUnavailable)
			return
		}

		if r.Body == nil {
			slog.ErrorContext(r.Context(), "no request body", "request.Body", r.Body)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		decoder := json.NewDecoder(r.Body)
		var googleUserParams google_entity.GoogleUser
		err := decoder.Decode(&googleUserParams)

		if err != nil {
			slog.ErrorContext(r.Context(), "error decoding google user from request", "err", err, "request.body", r.Body)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		err = c.OnboardGoogleUserCommand.Validate(r.Context(), &googleUserParams)

		if err != nil {
			slog.ErrorContext(r.Context(), "error validating google user", "err", err, "googleUserParams", googleUserParams)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		googleUser, ridToken, err := c.OnboardGoogleUserCommand.Exec(r.Context(), &googleUserParams)

		if err != nil {
			slog.ErrorContext(r.Context(), "error onboarding google user", "err", err, "googleUserParams.Email", googleUserParams.Email, "googleUserParams.VHash", googleUserParams.VHash)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if ridToken == nil {
			slog.ErrorContext(r.Context(), "error onboarding google user", "err", "controller: ridToken is nil", "googleUserParams.Email", googleUserParams.Email, "googleUserParams.VHash", googleUserParams.VHash)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set(ResourceOwnerIDHeaderKey, ridToken.GetID().String())
		w.Header().Set(ResourceOwnerAudTypeHeaderKey, string(ridToken.IntendedAudience))
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(googleUser)
	}
}
