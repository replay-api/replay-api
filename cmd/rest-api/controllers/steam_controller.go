package controllers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	common "github.com/replay-api/replay-api/pkg/domain"
	steam_entity "github.com/replay-api/replay-api/pkg/domain/steam/entities"
	steam_in "github.com/replay-api/replay-api/pkg/domain/steam/ports/in"
)

type SteamController struct {
	OnboardSteamUserCommand steam_in.OnboardSteamUserCommand
	helper                  *ControllerHelper
}

func NewSteamController(container *container.Container) *SteamController {
	var onboardSteamUserCommand steam_in.OnboardSteamUserCommand
	err := container.Resolve(&onboardSteamUserCommand)

	if err != nil {
		slog.Error("Cannot resolve steam_in.OnboardSteamUserCommand for new SteamController", "err", err)
		panic(err)
	}

	return &SteamController{
		OnboardSteamUserCommand: onboardSteamUserCommand,
		helper:                  NewControllerHelper(),
	}
}

func (c *SteamController) OnboardSteamUser(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "localhost:3000") // TODO: >>> config
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		var steamUserParams steam_entity.SteamUser

		// Decode request using helper
		if err := c.helper.DecodeJSONRequest(w, r, &steamUserParams); err != nil {
			return // Error already handled by helper
		}

		// Validate the steam user parameters
		err := c.OnboardSteamUserCommand.Validate(r.Context(), &steamUserParams)
		if c.helper.HandleError(w, r, err, "error validating steam user") {
			return // Error handled by helper
		}

		// Execute the onboard command
		steamUser, ridToken, err := c.OnboardSteamUserCommand.Exec(r.Context(), &steamUserParams)
		if c.helper.HandleError(w, r, err, "error onboarding steam user") {
			return // Error handled by helper
		}

		// Validate ridToken is not nil
		if ridToken == nil {
			slog.ErrorContext(r.Context(), "error onboarding steam user", "err", "controller: ridToken is nil", "steamUserParams.Steam.ID", steamUserParams.Steam.ID, "steamUserParams.VHash", steamUserParams.VHash)
			c.helper.HandleError(w, r, common.NewAPIError(http.StatusInternalServerError, "INTERNAL_ERROR", "ridToken is nil"), "ridToken validation failed")
			return
		}

		// Set additional headers for resource owner information
		w.Header().Set(ResourceOwnerIDHeaderKey, ridToken.GetID().String())
		w.Header().Set(ResourceOwnerAudTypeHeaderKey, string(ridToken.IntendedAudience))

		// Write successful response
		c.helper.WriteCreated(w, r, steamUser)
	}
}
