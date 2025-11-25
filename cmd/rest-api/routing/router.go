package routing

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/gorilla/mux"
	"github.com/psavelis/team-pro/replay-api/cmd/rest-api/controllers"
	cmd_controllers "github.com/psavelis/team-pro/replay-api/cmd/rest-api/controllers/command"
	websocket_controllers "github.com/psavelis/team-pro/replay-api/cmd/rest-api/controllers/websocket"
	query_controllers "github.com/psavelis/team-pro/replay-api/cmd/rest-api/controllers/query"
	"github.com/psavelis/team-pro/replay-api/cmd/rest-api/middlewares"
	matchmaking_in "github.com/psavelis/team-pro/replay-api/pkg/domain/matchmaking/ports/in"
	websocket "github.com/psavelis/team-pro/replay-api/pkg/infra/websocket"
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

	PlayerProfile string = "/players"

	// IAM
	Group string = "/groups"

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
	groupController := query_controllers.NewGroupController(&container)
	squadController := cmd_controllers.NewSquadController(container)
	playerProfileQueryController := query_controllers.NewPlayerProfileQueryController(container)
	playerProfileController := cmd_controllers.NewPlayerProfileController(container)
	matchmakingController := cmd_controllers.NewMatchmakingController(container)
	prizePoolQueryController := query_controllers.NewPrizePoolQueryController(container)
	tournamentCommandController := cmd_controllers.NewTournamentCommandController(container)
	tournamentQueryController := query_controllers.NewTournamentQueryController(container)

	// Prize pool matchmaking controllers
	var lobbyCommand matchmaking_in.LobbyCommand
	if err := container.Resolve(&lobbyCommand); err != nil {
		slog.ErrorContext(ctx, "Failed to resolve LobbyCommand", "error", err)
	}
	var wsHub *websocket.WebSocketHub
	if err := container.Resolve(&wsHub); err != nil {
		slog.ErrorContext(ctx, "Failed to resolve WebSocketHub", "error", err)
	}
	lobbyController := cmd_controllers.NewLobbyController(container, lobbyCommand)
	lobbyWebSocketHandler := websocket_controllers.NewLobbyWebSocketHandler(container, wsHub)

	// search controllers
	searchMux := query_controllers.NewSearchMux(&container)

	r := mux.NewRouter()

	r.Use(middlewares.ErrorMiddleware)
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
	r.HandleFunc(Match, matchController.DefaultSearchHandler).Methods("GET")
	r.HandleFunc(MatchDetail, matchController.GetMatchDetailHandler).Methods("GET")
	// r.HandleFunc("/games/{game_id}/matches/{match_id}/share", metadataController.GetEventsByGameIDAndMatchID(ctx)).Methods("POST")

	// Replay API
	r.HandleFunc(Replay, fileController.UploadHandler(ctx)).Methods("POST")
	r.HandleFunc(Replay, OptionsHandler).Methods("OPTIONS") // TODO: remover
	r.HandleFunc("/games/{game_id}/replays/{id}", fileController.GetReplayMetadata(ctx)).Methods("GET")
	r.HandleFunc("/games/{game_id}/replays/{id}", fileController.UpdateReplayMetadata(ctx)).Methods("PUT")
	r.HandleFunc("/games/{game_id}/replays/{id}", fileController.DeleteReplayFile(ctx)).Methods("DELETE")
	r.HandleFunc("/games/{game_id}/replays/{id}/download", fileController.DownloadReplayFile(ctx)).Methods("GET")
	r.HandleFunc("/games/{game_id}/replays/{id}/status", fileController.GetReplayProcessingStatus(ctx)).Methods("GET")

	// Game Events API
	r.HandleFunc(GameEvents, eventController.DefaultSearchHandler).Methods("GET")

	// Share Token API
	shareTokenController := cmd_controllers.NewShareTokenController(container)
	r.HandleFunc("/share-tokens", shareTokenController.CreateShareToken(ctx)).Methods("POST")
	r.HandleFunc("/share-tokens", shareTokenController.ListShareTokens(ctx)).Methods("GET")
	r.HandleFunc("/share-tokens/{token}", shareTokenController.GetShareToken(ctx)).Methods("GET")
	r.HandleFunc("/share-tokens/{token}", shareTokenController.RevokeShareToken(ctx)).Methods("DELETE")

	// Squad API
	r.HandleFunc("/squads", squadController.CreateSquadHandler(ctx)).Methods("POST")
	r.HandleFunc("/squads/{id}", squadController.GetSquadHandler(ctx)).Methods("GET")
	r.HandleFunc("/squads/{id}", squadController.UpdateSquadHandler(ctx)).Methods("PUT")
	r.HandleFunc("/squads/{id}", squadController.DeleteSquadHandler(ctx)).Methods("DELETE")
	r.HandleFunc("/squads/{id}/members", squadController.AddMemberHandler(ctx)).Methods("POST")
	r.HandleFunc("/squads/{id}/members/{player_id}", squadController.RemoveMemberHandler(ctx)).Methods("DELETE")
	r.HandleFunc("/squads/{id}/members/{player_id}/role", squadController.UpdateMemberRoleHandler(ctx)).Methods("PUT")

	// Player Profiles API
	r.HandleFunc("/players", playerProfileController.CreatePlayerProfileHandler(ctx)).Methods("POST")
	r.HandleFunc("/players", playerProfileQueryController.DefaultSearchHandler).Methods("GET")
	r.HandleFunc("/players/{id}", playerProfileController.GetPlayerProfileHandler(ctx)).Methods("GET")
	r.HandleFunc("/players/{id}", playerProfileController.UpdatePlayerProfileHandler(ctx)).Methods("PUT")
	r.HandleFunc("/players/{id}", playerProfileController.DeletePlayerProfileHandler(ctx)).Methods("DELETE")

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

	// IAM API
	r.HandleFunc(Group, groupController.HandleListMemberGroups).Methods("GET")

	// Matchmaking API
	r.HandleFunc("/matchmaking/queue", matchmakingController.JoinQueueHandler(ctx)).Methods("POST")
	r.HandleFunc("/matchmaking/queue/{session_id}", matchmakingController.LeaveQueueHandler(ctx)).Methods("DELETE")
	r.HandleFunc("/matchmaking/session/{session_id}", matchmakingController.GetSessionStatusHandler(ctx)).Methods("GET")
	r.HandleFunc("/matchmaking/pools/{game_id}", matchmakingController.GetPoolStatsHandler(ctx)).Methods("GET")

	// Prize Pool Lobby API
	r.HandleFunc("/api/lobbies", lobbyController.CreateLobbyHandler(ctx)).Methods("POST")
	r.HandleFunc("/api/lobbies/{lobby_id}/join", lobbyController.JoinLobbyHandler(ctx)).Methods("POST")
	r.HandleFunc("/api/lobbies/{lobby_id}/leave", lobbyController.LeaveLobbyHandler(ctx)).Methods("DELETE")
	r.HandleFunc("/api/lobbies/{lobby_id}/ready", lobbyController.SetPlayerReadyHandler(ctx)).Methods("PUT")
	r.HandleFunc("/api/lobbies/{lobby_id}/start", lobbyController.StartMatchHandler(ctx)).Methods("POST")
	r.HandleFunc("/api/lobbies/{lobby_id}", lobbyController.CancelLobbyHandler(ctx)).Methods("DELETE")

	// WebSocket for real-time lobby updates
	r.HandleFunc("/ws/lobby/{lobby_id}", lobbyWebSocketHandler.UpgradeConnection(ctx)).Methods("GET")

	// Prize Pool API
	r.HandleFunc("/prize-pools/{id}", prizePoolQueryController.GetPrizePoolHandler).Methods("GET")
	r.HandleFunc("/prize-pools/{id}/history", prizePoolQueryController.GetPrizePoolHistoryHandler).Methods("GET")
	r.HandleFunc("/matches/{match_id}/prize-pool", prizePoolQueryController.GetPrizePoolByMatchHandler).Methods("GET")
	r.HandleFunc("/prize-pools/pending-distributions", prizePoolQueryController.GetPendingDistributionsHandler).Methods("GET")

	// Tournament API
	r.HandleFunc("/tournaments", tournamentCommandController.CreateTournamentHandler(ctx)).Methods("POST")
	r.HandleFunc("/tournaments", tournamentQueryController.ListTournamentsHandler).Methods("GET")
	r.HandleFunc("/tournaments/upcoming", tournamentQueryController.GetUpcomingTournamentsHandler).Methods("GET")
	r.HandleFunc("/tournaments/{id}", tournamentQueryController.GetTournamentHandler).Methods("GET")
	r.HandleFunc("/tournaments/{id}", tournamentCommandController.UpdateTournamentHandler(ctx)).Methods("PUT")
	r.HandleFunc("/tournaments/{id}", tournamentCommandController.DeleteTournamentHandler(ctx)).Methods("DELETE")
	r.HandleFunc("/tournaments/{id}/register", tournamentCommandController.RegisterPlayerHandler(ctx)).Methods("POST")
	r.HandleFunc("/tournaments/{id}/register", tournamentCommandController.UnregisterPlayerHandler(ctx)).Methods("DELETE")
	r.HandleFunc("/tournaments/{id}/start", tournamentCommandController.StartTournamentHandler(ctx)).Methods("POST")
	r.HandleFunc("/players/{player_id}/tournaments", tournamentQueryController.GetPlayerTournamentsHandler).Methods("GET")
	r.HandleFunc("/organizers/{organizer_id}/tournaments", tournamentQueryController.GetOrganizerTournamentsHandler).Methods("GET")

	return r
}
