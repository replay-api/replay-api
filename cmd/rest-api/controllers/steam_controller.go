package controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	steam_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/entities"
	steam_in "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/ports/in"
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
		w.Header().Set("Access-Control-Allow-Origin", "localhost:3000") // TODO: >>> config
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Body == nil {
			slog.ErrorContext(r.Context(), "no request body", "request", r)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		decoder := json.NewDecoder(r.Body)
		var steamUserParams steam_entity.SteamUser
		err := decoder.Decode(&steamUserParams)

		if err != nil {
			slog.ErrorContext(r.Context(), "error decoding steam user from request", "err", err, "request", r)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		err = c.OnboardSteamUserCommand.Validate(r.Context(), steamUserParams.Steam.ID, steamUserParams.VHash)

		if err != nil {
			slog.ErrorContext(r.Context(), "error validating steam user", "err", err, "steamUserParams", steamUserParams)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		steamUser, err := c.OnboardSteamUserCommand.Exec(r.Context(), steamUserParams.Steam.ID, steamUserParams.VHash)

		if err != nil {
			slog.ErrorContext(r.Context(), "error onboarding steam user", "err", err, "steamUserParams.Steam.ID", steamUserParams.Steam.ID, "steamUserParams.VHash", steamUserParams.VHash)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(steamUser)
	}
}