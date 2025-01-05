package routing

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/gorilla/mux"
	"github.com/psavelis/team-pro/replay-api/cmd/rest-api/controllers"
	cmd_controllers "github.com/psavelis/team-pro/replay-api/cmd/rest-api/controllers/command"
	query_controllers "github.com/psavelis/team-pro/replay-api/cmd/rest-api/controllers/query"
	"github.com/psavelis/team-pro/replay-api/cmd/rest-api/middlewares"
)

const (
	Health string = "/health"
	CI     string = "/coverage"

	Match         string = "/games/{game_id}/match"
	MatchDetail   string = "/games/{game_id}/match/{match_id}"
	MatchEvent    string = "/games/{game_id}/match/{match_id}/events"
	GameEvents    string = "/games/{game_id}/events"
	Replay        string = "/games/{game_id}/replays"
	ReplayDetail  string = "/games/{game_id}/replay/{replay_file_id}"
	Onboard       string = "/onboarding"
	OnboardSteam  string = "/onboarding/steam"
	OnboardGoogle string = "/onboarding/google"

	Search string = "/search/{query:.*}"
)

func NewRouter(ctx context.Context, container container.Container) http.Handler {
	// middleware
	resourceContextMiddleware := middlewares.NewResourceContextMiddleware(&container)

	// metadataController := controllers.NewMetadataController(container)
	fileController := cmd_controllers.NewFileController(container)
	healthController := controllers.NewHealthController(container)
	steamController := controllers.NewSteamController(&container)
	googleController := controllers.NewGoogleController(&container)
	matchController := query_controllers.NewMatchQueryController(container)
	eventController := query_controllers.NewEventQueryController(container)

	// search controllers
	searchMux := query_controllers.NewSearchMux(&container)

	r := mux.NewRouter()
	r.Use(mux.CORSMethodMiddleware(r))
	r.Use(resourceContextMiddleware.Handler)

	// r.Use(middlewares.NewLoggerMiddleware().Handler)
	// r.Use(middlewares.NewRecoveryMiddleware().Handler)
	// r.Use(middlewares.NewResourceContextMiddleware().Handler)
	// r.Use(middlewares.NewAuthMiddleware().Handler)

	// cors

	// search mux
	r.HandleFunc(Search, searchMux.Dispatch).Methods("GET")

	// health
	r.HandleFunc(Health, healthController.HealthCheck(ctx)).Methods("GET")

	r.HandleFunc(CI, func(w http.ResponseWriter, r *http.Request) {
		slog.Info("CI route up.")
		http.ServeFile(w, r, "/app/coverage/coverage.html")
	}).Methods("GET")

	// onboarding/steam
	r.HandleFunc(OnboardSteam, steamController.OnboardSteamUser(ctx)).Methods("POST")

	r.HandleFunc(OnboardGoogle, googleController.OnboardGoogleUser(ctx)).Methods("POST")

	// Matches API
	// r.HandleFunc(MatchEvent, metadataController.GetEventsByGameIDAndMatchID(ctx)).Methods("GET") // DEPRECATED

	// r.HandleFunc("/games/{game_id}/matches/{match_id}/share", metadataController.GetEventsByGameIDAndMatchID(ctx)).Methods("POST")
	// r.HandleFunc("/games/{game_id}/matches", metadataController.GetMatchesByGameID(ctx)).Methods("GET") // ?userID=123&gameID=123&matchID=123

	// Replay API
	r.HandleFunc(Replay, fileController.UploadHandler(ctx)).Methods("POST")
	r.HandleFunc(Replay, OptionsHandler).Methods("OPTIONS") // TODO: remover
	// r.HandleFunc(Replay, metadataController.ReplaySearchHandler(ctx)).Methods("GET")
	r.HandleFunc(Match, matchController.DefaultSearchHandler).Methods("GET")

	// Game Events API
	r.HandleFunc(GameEvents, eventController.DefaultSearchHandler).Methods("GET")

	// r.HandleFunc(ReplayDetail, fileController.ReplayDetailHandler(ctx)).Methods("GET")
	// r.HandleFunc(("/games/{game_id}/replay/{replay_file_id}"), fileController.ProcessReplayFile(ctx)).Methods("PUT")
	// r.HandleFunc(("/games/{game_id}/replay/{replay_file_id}/metadata"), fileController.GetReplayFile(ctx)).Methods("GET")
	// r.HandleFunc(("/games/{game_id}/replay/{replay_file_id}/download"), fileController.DownloadReplayFile(ctx)).Methods("GET")

	// Sharing API
	// r.HandleFunc(("/games/{game_id}/replay/{replay_file_id}/share"), fileController.DownloadReplayFile(ctx)).Methods("POST")
	// r.HandleFunc(("/games/{game_id}/replay/{replay_file_id}/share"), fileController.DownloadReplayFile(ctx)).Methods("GET")
	// r.HandleFunc(("/games/{game_id}/replay/{replay_file_id}/share/{share_token_id}"), fileController.DownloadReplayFile(ctx)).Methods("DELETE")

	// Squad API
	// r.HandleFunc("/games/{game_id}/squad", squadController.GetSquadByGameID(ctx)).Methods("GET")
	// r.HandleFunc("/games/{game_id}/squad", squadController.CreateSquad(ctx)).Methods("POST")
	// r.HandleFunc("/games/{game_id}/squad/{squad_id}", squadController.GetSquadByID(ctx)).Methods("GET")
	// r.HandleFunc("/games/{game_id}/squad/{squad_id}", squadController.UpdateSquad(ctx)).Methods("PUT")
	// r.HandleFunc("/games/{game_id}/squad/{squad_id}", squadController.DeleteSquad(ctx)).Methods("DELETE")

	// User API
	// r.HandleFunc("/games/{game_id}/user", userController.GetUserByGameID(ctx)).Methods("GET")
	// r.HandleFunc("/games/{game_id}/user", userController.CreateUser(ctx)).Methods("POST")
	// r.HandleFunc("/games/{game_id}/user/{user_id}", userController.GetUserByID(ctx)).Methods("GET")
	// r.HandleFunc("/games/{game_id}/user/{user_id}", userController.UpdateUser(ctx)).Methods("PUT")
	// r.HandleFunc("/games/{game_id}/user/{user_id}", userController.DeleteUser(ctx)).Methods("DELETE")

	// Badges API
	// r.HandleFunc("/games/{game_id}/badges", badgeController.GetBadgesByGameID(ctx)).Methods("GET")
	// r.HandleFunc("/games/{game_id}/badge_types", badgeController.GetBadgeTypes(ctx)).Methods("GET")
	// r.HandleFunc("/games/{game_id}/badges/{badge_id}", badgeController.GetBadgeByID(ctx)).Methods("GET")
	// r.HandleFunc("/games/{game_id}/badges/{badge_id}", badgeController.UpdateBadge(ctx)).Methods("PUT")
	// r.HandleFunc("/games/{game_id}/badges/{badge_id}", badgeController.DeleteBadge(ctx)).Methods("DELETE")

	// Stats API
	// r.HandleFunc("/games/{game_id}/stats", statsController.GetStatsByGameID(ctx)).Methods("GET")

	// Leaderboard API
	// r.HandleFunc("/games/{game_id}/leaderboard", leaderboardController.GetLeaderboardByGameID(ctx)).Methods("GET")

	// Game API
	// r.HandleFunc("/games/{game_id}", gameController.GetGameByID(ctx)).Methods("GET")

	return r
}
