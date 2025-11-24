package routing

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/psavelis/team-pro/replay-api/cmd/rest-api/controllers"
	cmd_controllers "github.com/psavelis/team-pro/replay-api/cmd/rest-api/controllers/command"
	websocket_controllers "github.com/psavelis/team-pro/replay-api/cmd/rest-api/controllers/websocket"
	query_controllers "github.com/psavelis/team-pro/replay-api/cmd/rest-api/controllers/query"
	"github.com/psavelis/team-pro/replay-api/cmd/rest-api/middlewares"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	matchmaking_in "github.com/psavelis/team-pro/replay-api/pkg/domain/matchmaking/ports/in"
	squad_out "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/ports/out"
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
	squadQueryController := query_controllers.NewSquadQueryController(container)
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

	// Resolve SquadReader for admin role checking
	var squadReader squad_out.SquadReader
	if err := container.Resolve(&squadReader); err != nil {
		slog.ErrorContext(ctx, "Failed to resolve SquadReader", "error", err)
	}

	// Squad admin role checker function
	squadAdminRoleChecker := func(ctx context.Context, resourceID, userID uuid.UUID) (bool, error) {
		squads, err := squadReader.Search(ctx, common.NewSearchByID(ctx, resourceID, common.ClientApplicationAudienceIDKey))
		if err != nil || len(squads) == 0 {
			return false, err
		}
		squad := squads[0]
		for _, membership := range squad.Membership {
			if membership.UserID == userID {
				for _, role := range membership.Roles {
					if role == "admin" {
						return true, nil
					}
				}
			}
		}
		return false, nil
	}

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
	r.Handle(Replay, middlewares.RequireAuthentication()(http.HandlerFunc(fileController.UploadHandler(ctx)))).Methods("POST")
	r.HandleFunc(Replay, OptionsHandler).Methods("OPTIONS") // TODO: remover
	r.HandleFunc("/games/{game_id}/replays/{id}", fileController.GetReplayMetadata(ctx)).Methods("GET")
	r.Handle("/games/{game_id}/replays/{id}", middlewares.ResourceOwnershipMiddleware(middlewares.DefaultOwnershipConfig(common.ResourceTypeReplayFile))(http.HandlerFunc(fileController.UpdateReplayMetadata(ctx)))).Methods("PUT")
	r.Handle("/games/{game_id}/replays/{id}", middlewares.ResourceOwnershipMiddleware(middlewares.DefaultOwnershipConfig(common.ResourceTypeReplayFile))(http.HandlerFunc(fileController.DeleteReplayFile(ctx)))).Methods("DELETE")
	r.HandleFunc("/games/{game_id}/replays/{id}/download", fileController.DownloadReplayFile(ctx)).Methods("GET")
	r.HandleFunc("/games/{game_id}/replays/{id}/status", fileController.GetReplayProcessingStatus(ctx)).Methods("GET")

	// Game Events API
	r.HandleFunc(GameEvents, eventController.DefaultSearchHandler).Methods("GET")

	// Share Token API
	shareTokenController := cmd_controllers.NewShareTokenController(container)
	r.Handle("/share-tokens", middlewares.RequireAuthentication()(http.HandlerFunc(shareTokenController.CreateShareToken(ctx)))).Methods("POST")
	r.Handle("/share-tokens", middlewares.RequireAuthentication()(http.HandlerFunc(shareTokenController.ListShareTokens(ctx)))).Methods("GET")
	r.HandleFunc("/share-tokens/{token}", shareTokenController.GetShareToken(ctx)).Methods("GET")
	r.Handle("/share-tokens/{token}", middlewares.RequireAuthentication()(http.HandlerFunc(shareTokenController.RevokeShareToken(ctx)))).Methods("DELETE")

	// Squad API
	r.Handle("/squads", middlewares.RequireAuthentication()(http.HandlerFunc(squadController.CreateSquadHandler(ctx)))).Methods("POST")
	r.HandleFunc("/squads", squadQueryController.DefaultSearchHandler).Methods("GET")
	r.HandleFunc("/squads/{id}", squadController.GetSquadHandler(ctx)).Methods("GET")
	r.Handle("/squads/{id}", middlewares.RequireOwnerOrAdminRole(common.ResourceTypeSquad, squadAdminRoleChecker)(http.HandlerFunc(squadController.UpdateSquadHandler(ctx)))).Methods("PUT")
	r.Handle("/squads/{id}", middlewares.ResourceOwnershipMiddleware(middlewares.DefaultOwnershipConfig(common.ResourceTypeSquad))(http.HandlerFunc(squadController.DeleteSquadHandler(ctx)))).Methods("DELETE")
	r.Handle("/squads/{id}/members", middlewares.RequireOwnerOrAdminRole(common.ResourceTypeSquad, squadAdminRoleChecker)(http.HandlerFunc(squadController.AddMemberHandler(ctx)))).Methods("POST")
	r.Handle("/squads/{id}/members/{player_id}", middlewares.RequireOwnerOrAdminRole(common.ResourceTypeSquad, squadAdminRoleChecker)(http.HandlerFunc(squadController.RemoveMemberHandler(ctx)))).Methods("DELETE")
	r.Handle("/squads/{id}/members/{player_id}/role", middlewares.RequireOwnerOrAdminRole(common.ResourceTypeSquad, squadAdminRoleChecker)(http.HandlerFunc(squadController.UpdateMemberRoleHandler(ctx)))).Methods("PUT")

	// Player Profiles API
	r.Handle("/players", middlewares.RequireAuthentication()(http.HandlerFunc(playerProfileController.CreatePlayerProfileHandler(ctx)))).Methods("POST")
	r.HandleFunc("/players", playerProfileQueryController.DefaultSearchHandler).Methods("GET")
	r.HandleFunc("/players/{id}", playerProfileController.GetPlayerProfileHandler(ctx)).Methods("GET")
	r.Handle("/players/{id}", middlewares.ResourceOwnershipMiddleware(middlewares.DefaultOwnershipConfig(common.ResourceTypePlayerProfile))(http.HandlerFunc(playerProfileController.UpdatePlayerProfileHandler(ctx)))).Methods("PUT")
	r.Handle("/players/{id}", middlewares.ResourceOwnershipMiddleware(middlewares.DefaultOwnershipConfig(common.ResourceTypePlayerProfile))(http.HandlerFunc(playerProfileController.DeletePlayerProfileHandler(ctx)))).Methods("DELETE")

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
	r.Handle(Group, middlewares.RequireAuthentication()(http.HandlerFunc(groupController.HandleListMemberGroups))).Methods("GET")

	// Matchmaking API
	r.Handle("/matchmaking/queue", middlewares.RequireAuthentication()(http.HandlerFunc(matchmakingController.JoinQueueHandler(ctx)))).Methods("POST")
	r.Handle("/matchmaking/queue/{session_id}", middlewares.RequireAuthentication()(http.HandlerFunc(matchmakingController.LeaveQueueHandler(ctx)))).Methods("DELETE")
	r.Handle("/matchmaking/session/{session_id}", middlewares.RequireAuthentication()(http.HandlerFunc(matchmakingController.GetSessionStatusHandler(ctx)))).Methods("GET")
	r.HandleFunc("/matchmaking/pools/{game_id}", matchmakingController.GetPoolStatsHandler(ctx)).Methods("GET")

	// Prize Pool Lobby API
	r.Handle("/api/lobbies", middlewares.RequireAuthentication()(http.HandlerFunc(lobbyController.CreateLobbyHandler(ctx)))).Methods("POST")
	r.Handle("/api/lobbies/{lobby_id}/join", middlewares.RequireAuthentication()(http.HandlerFunc(lobbyController.JoinLobbyHandler(ctx)))).Methods("POST")
	r.Handle("/api/lobbies/{lobby_id}/leave", middlewares.RequireAuthentication()(http.HandlerFunc(lobbyController.LeaveLobbyHandler(ctx)))).Methods("DELETE")
	r.Handle("/api/lobbies/{lobby_id}/ready", middlewares.RequireAuthentication()(http.HandlerFunc(lobbyController.SetPlayerReadyHandler(ctx)))).Methods("PUT")
	r.Handle("/api/lobbies/{lobby_id}/start", middlewares.RequireAuthentication()(http.HandlerFunc(lobbyController.StartMatchHandler(ctx)))).Methods("POST")
	r.Handle("/api/lobbies/{lobby_id}", middlewares.RequireAuthentication()(http.HandlerFunc(lobbyController.CancelLobbyHandler(ctx)))).Methods("DELETE")

	// WebSocket for real-time lobby updates
	r.HandleFunc("/ws/lobby/{lobby_id}", lobbyWebSocketHandler.UpgradeConnection(ctx)).Methods("GET")

	// Prize Pool API
	r.HandleFunc("/prize-pools/{id}", prizePoolQueryController.GetPrizePoolHandler).Methods("GET")
	r.HandleFunc("/prize-pools/{id}/history", prizePoolQueryController.GetPrizePoolHistoryHandler).Methods("GET")
	r.HandleFunc("/matches/{match_id}/prize-pool", prizePoolQueryController.GetPrizePoolByMatchHandler).Methods("GET")
	r.HandleFunc("/prize-pools/pending-distributions", prizePoolQueryController.GetPendingDistributionsHandler).Methods("GET")

	// Tournament API
	r.Handle("/tournaments", middlewares.RequireAuthentication()(http.HandlerFunc(tournamentCommandController.CreateTournamentHandler(ctx)))).Methods("POST")
	r.HandleFunc("/tournaments", tournamentQueryController.ListTournamentsHandler).Methods("GET")
	r.HandleFunc("/tournaments/upcoming", tournamentQueryController.GetUpcomingTournamentsHandler).Methods("GET")
	r.HandleFunc("/tournaments/{id}", tournamentQueryController.GetTournamentHandler).Methods("GET")
	r.Handle("/tournaments/{id}", middlewares.ResourceOwnershipMiddleware(middlewares.DefaultOwnershipConfig(common.ResourceTypeTournament))(http.HandlerFunc(tournamentCommandController.UpdateTournamentHandler(ctx)))).Methods("PUT")
	r.Handle("/tournaments/{id}", middlewares.ResourceOwnershipMiddleware(middlewares.DefaultOwnershipConfig(common.ResourceTypeTournament))(http.HandlerFunc(tournamentCommandController.DeleteTournamentHandler(ctx)))).Methods("DELETE")
	r.Handle("/tournaments/{id}/register", middlewares.RequireAuthentication()(http.HandlerFunc(tournamentCommandController.RegisterPlayerHandler(ctx)))).Methods("POST")
	r.Handle("/tournaments/{id}/register", middlewares.RequireAuthentication()(http.HandlerFunc(tournamentCommandController.UnregisterPlayerHandler(ctx)))).Methods("DELETE")
	r.Handle("/tournaments/{id}/start", middlewares.ResourceOwnershipMiddleware(middlewares.DefaultOwnershipConfig(common.ResourceTypeTournament))(http.HandlerFunc(tournamentCommandController.StartTournamentHandler(ctx)))).Methods("POST")
	r.HandleFunc("/players/{player_id}/tournaments", tournamentQueryController.GetPlayerTournamentsHandler).Methods("GET")
	r.HandleFunc("/organizers/{organizer_id}/tournaments", tournamentQueryController.GetOrganizerTournamentsHandler).Methods("GET")

	return r
}
