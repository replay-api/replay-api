package controllers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	common "github.com/replay-api/replay-api/pkg/domain"
	google_entity "github.com/replay-api/replay-api/pkg/domain/google/entities"
	google_in "github.com/replay-api/replay-api/pkg/domain/google/ports/in"
)

type GoogleController struct {
	OnboardGoogleUserCommand google_in.OnboardGoogleUserCommand
	helper                   *ControllerHelper
}

func NewGoogleController(container *container.Container) *GoogleController {
	var onboardGoogleUserCommand google_in.OnboardGoogleUserCommand
	err := container.Resolve(&onboardGoogleUserCommand)

	if err != nil {
		slog.Error("Cannot resolve google_in.OnboardGoogleUserCommand for new GoogleController", "err", err)
		panic(err)
	}

	return &GoogleController{
		OnboardGoogleUserCommand: onboardGoogleUserCommand,
		helper:                   NewControllerHelper(),
	}
}

func (c *GoogleController) OnboardGoogleUser(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "localhost:3000") // TODO: >>> config
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		var googleUserParams google_entity.GoogleUser

		// Decode request using helper
		if err := c.helper.DecodeJSONRequest(w, r, &googleUserParams); err != nil {
			return // Error already handled by helper
		}

		// Validate the google user parameters
		err := c.OnboardGoogleUserCommand.Validate(r.Context(), &googleUserParams)
		if c.helper.HandleError(w, r, err, "error validating google user") {
			return // Error handled by helper
		}

		// Execute the onboard command
		googleUser, ridToken, err := c.OnboardGoogleUserCommand.Exec(r.Context(), &googleUserParams)
		if c.helper.HandleError(w, r, err, "error onboarding google user") {
			return // Error handled by helper
		}

		// Validate ridToken is not nil
		if ridToken == nil {
			slog.ErrorContext(r.Context(), "error onboarding google user", "err", "controller: ridToken is nil", "googleUserParams.Email", googleUserParams.Email, "googleUserParams.VHash", googleUserParams.VHash)
			c.helper.HandleError(w, r, common.NewAPIError(http.StatusInternalServerError, "INTERNAL_ERROR", "ridToken is nil"), "ridToken validation failed")
			return
		}

		// Set additional headers for resource owner information
		w.Header().Set(ResourceOwnerIDHeaderKey, ridToken.GetID().String())
		w.Header().Set(ResourceOwnerAudTypeHeaderKey, string(ridToken.IntendedAudience))

		// Write successful response
		c.helper.WriteCreated(w, r, googleUser)
	}
}
