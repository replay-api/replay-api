package controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/golobby/container/v3"
	steam_entity "github.com/replay-api/replay-api/pkg/domain/steam/entities"
	steam_in "github.com/replay-api/replay-api/pkg/domain/steam/ports/in"
)

type SteamController struct {
	OnboardSteamUserCommand steam_in.OnboardSteamUserCommand
}

func NewSteamController(container *container.Container) *SteamController {
	var onboardSteamUserCommand steam_in.OnboardSteamUserCommand
	err := container.Resolve(&onboardSteamUserCommand)

	if err != nil {
		slog.Error("Cannot resolve steam_in.OnboardSteamUserCommand for new SteamController", "err", err)
		panic(err)
	}

	return &SteamController{OnboardSteamUserCommand: onboardSteamUserCommand}
}

func (c *SteamController) OnboardSteamUser(apiContext context.Context) http.HandlerFunc {
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
		var steamUserParams steam_entity.SteamUser
		err := decoder.Decode(&steamUserParams)

		// slog.InfoContext(r.Context(), "SteamUser Received =>", "steam_entity.SteamUser", steamUserParams)

		if err != nil {
			slog.ErrorContext(r.Context(), "error decoding steam user from request", "err", err, "request.body", r.Body)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		err = c.OnboardSteamUserCommand.Validate(r.Context(), &steamUserParams)

		if err != nil {
			slog.ErrorContext(r.Context(), "error validating steam user", "err", err, "steamUserParams", steamUserParams)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		steamUser, ridToken, err := c.OnboardSteamUserCommand.Exec(r.Context(), &steamUserParams)

		if err != nil {
			slog.ErrorContext(r.Context(), "error onboarding steam user", "err", err, "steamUserParams.Steam.ID", steamUserParams.Steam.ID, "steamUserParams.VHash", steamUserParams.VHash)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if ridToken == nil {
			slog.ErrorContext(r.Context(), "error onboarding steam user", "err", "controller: ridToken is nil", "steamUserParams.Steam.ID", steamUserParams.Steam.ID, "steamUserParams.VHash", steamUserParams.VHash)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set(ResourceOwnerIDHeaderKey, ridToken.GetID().String())
		w.Header().Set(ResourceOwnerAudTypeHeaderKey, string(ridToken.IntendedAudience))
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(steamUser)
	}
}
