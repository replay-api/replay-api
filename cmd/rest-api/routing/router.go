package routing

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/gorilla/mux"
	"github.com/replay-api/replay-api/cmd/rest-api/controllers"
	cmd_controllers "github.com/replay-api/replay-api/cmd/rest-api/controllers/command"
	query_controllers "github.com/replay-api/replay-api/cmd/rest-api/controllers/query"
	websocket_controllers "github.com/replay-api/replay-api/cmd/rest-api/controllers/websocket"
	"github.com/replay-api/replay-api/cmd/rest-api/middlewares"
	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	"github.com/replay-api/replay-api/pkg/infra/metrics"
	websocket "github.com/replay-api/replay-api/pkg/infra/websocket"
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
	OnboardEmail  string = "/onboarding/email"
	AuthLogin     string = "/auth/login"
	AuthGuest     string = "/auth/guest"
	AuthRefresh   string = "/auth/refresh"
	AuthLogout    string = "/auth/logout"

	PlayerProfile string = "/players"

	// IAM
	Group string = "/groups"

	Search string = "/search/{query:.*}"

	// Aliases (plural forms for frontend compatibility)
	Matches            string = "/games/{game_id}/matches"
	MatchesDetail      string = "/games/{game_id}/matches/{match_id}"
	MatchesTrajectory  string = "/games/{game_id}/matches/{match_id}/trajectory"
	MatchesHeatmap     string = "/games/{game_id}/matches/{match_id}/heatmap"
	MatchesPosStats    string = "/games/{game_id}/matches/{match_id}/positioning-stats"
	MatchesEvents      string = "/games/{game_id}/matches/{match_id}/events"
	MatchesScoreboard  string = "/games/{game_id}/matches/{match_id}/scoreboard"
	RoundTrajectory    string = "/games/{game_id}/matches/{match_id}/rounds/{round_number}/trajectory"
	RoundHeatmap       string = "/games/{game_id}/matches/{match_id}/rounds/{round_number}/heatmap"

	// Notifications
	Notifications       string = "/notifications"
	NotificationDetail  string = "/notifications/{notification_id}"
	NotificationRead    string = "/notifications/{notification_id}/read"
	NotificationsReadAll string = "/notifications/read-all"
)

func NewRouter(ctx context.Context, container container.Container) http.Handler {
	// middleware
	resourceContextMiddleware := middlewares.NewResourceContextMiddleware(&container)

	// metadataController := controllers.NewMetadataController(container)
	fileController := cmd_controllers.NewFileController(container)
	healthController := controllers.NewHealthController(container)
	authController := controllers.NewAuthController(&container)
	steamController := controllers.NewSteamController(&container)
	googleController := controllers.NewGoogleController(&container)
	emailController := controllers.NewEmailController(&container)
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
	walletQueryController := query_controllers.NewWalletQueryController(container)

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

	// Global OPTIONS handler - must be registered BEFORE other routes
	// This handles CORS preflight for all routes
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		origin := req.Header.Get("Origin")
		if origin == "" {
			origin = "http://localhost:3030"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Resource-Owner-ID, X-Intended-Audience, X-Request-ID, X-Search, x-search")
		w.Header().Set("Access-Control-Expose-Headers", "X-Resource-Owner-ID, X-Intended-Audience")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.WriteHeader(http.StatusOK)
	}).Methods("OPTIONS")

	r.Use(middlewares.ErrorMiddleware)
	r.Use(mux.CORSMethodMiddleware(r))
	r.Use(resourceContextMiddleware.Handler)

	// Enable rate limiting (100 requests per minute per IP)
	rateLimitMiddleware := middlewares.NewRateLimitMiddleware()
	r.Use(rateLimitMiddleware.Handler)

	// Enable auth middleware for protected routes
	authMiddleware := middlewares.NewAuthMiddleware()
	r.Use(authMiddleware.Handler)

	// Enable request signing for sensitive financial operations
	requestSigningMiddleware := middlewares.NewRequestSigningMiddleware()
	r.Use(requestSigningMiddleware.Handler)

	// Enable CORS for browser access
	corsMiddleware := middlewares.NewCORSMiddleware()
	r.Use(corsMiddleware.Handler)

	// search mux
	r.HandleFunc(Search, searchMux.Dispatch).Methods("GET")

	// health
	r.HandleFunc(Health, healthController.HealthCheck(ctx)).Methods("GET")

	// Prometheus metrics
	r.Handle("/metrics", metrics.Handler()).Methods("GET")

	r.HandleFunc(CI, func(w http.ResponseWriter, r *http.Request) {
		slog.Info("CI route up.")
		http.ServeFile(w, r, "/app/coverage/coverage.html")
	}).Methods("GET")

	// onboarding/steam
	r.HandleFunc(OnboardSteam, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(OnboardSteam, steamController.OnboardSteamUser(ctx)).Methods("POST")

	r.HandleFunc(OnboardGoogle, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(OnboardGoogle, googleController.OnboardGoogleUser(ctx)).Methods("POST")

	// onboarding/email
	r.HandleFunc(OnboardEmail, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(OnboardEmail, emailController.OnboardEmailUser(ctx)).Methods("POST")

	// auth/login
	r.HandleFunc(AuthLogin, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(AuthLogin, emailController.LoginEmailUser(ctx)).Methods("POST")

	// auth/guest - Create guest token for unauthenticated users
	r.HandleFunc(AuthGuest, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(AuthGuest, authController.CreateGuestToken(ctx)).Methods("POST")

	// auth/refresh - Refresh existing token
	r.HandleFunc(AuthRefresh, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(AuthRefresh, authController.RefreshToken(ctx)).Methods("POST")

	// auth/logout - Revoke token and logout
	r.HandleFunc(AuthLogout, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(AuthLogout, authController.Logout(ctx)).Methods("POST")

	// Matches API
	r.HandleFunc(Match, matchController.DefaultSearchHandler).Methods("GET")
	r.HandleFunc(MatchDetail, matchController.GetMatchDetailHandler).Methods("GET")
	// r.HandleFunc("/games/{game_id}/matches/{match_id}/share", metadataController.GetEventsByGameIDAndMatchID(ctx)).Methods("POST")

	// Replay Files Query API (search/list)
	replayFileQueryController := query_controllers.NewReplayFileQueryController(container)
	r.HandleFunc(Replay, replayFileQueryController.ListReplayFilesHandler).Methods("GET")

	// Replay API (upload)
	r.HandleFunc(Replay, fileController.UploadHandler(ctx)).Methods("POST")
	r.HandleFunc(Replay, OptionsHandler).Methods("OPTIONS") // TODO: remover
	r.HandleFunc("/games/{game_id}/replays/{id}", fileController.GetReplayMetadata(ctx)).Methods("GET")
	r.HandleFunc("/games/{game_id}/replays/{id}", fileController.UpdateReplayMetadata(ctx)).Methods("PUT")
	r.HandleFunc("/games/{game_id}/replays/{id}", fileController.DeleteReplayFile(ctx)).Methods("DELETE")
	r.HandleFunc("/games/{game_id}/replays/{id}/download", fileController.DownloadReplayFile(ctx)).Methods("GET")
	r.HandleFunc("/games/{game_id}/replays/{id}/status", fileController.GetReplayProcessingStatus(ctx)).Methods("GET")

	// Game Events API
	r.HandleFunc(GameEvents, eventController.DefaultSearchHandler).Methods("GET")

	// Match Analytics API (heatmaps, trajectories, positioning)
	matchAnalyticsController := query_controllers.NewMatchAnalyticsController(container)
	r.HandleFunc("/games/{game_id}/matches/{match_id}/trajectory", matchAnalyticsController.GetMatchTrajectoryHandler).Methods("GET")
	r.HandleFunc("/games/{game_id}/matches/{match_id}/rounds/{round_number}/trajectory", matchAnalyticsController.GetRoundTrajectoryHandler).Methods("GET")
	r.HandleFunc("/games/{game_id}/matches/{match_id}/heatmap", matchAnalyticsController.GetMatchHeatmapHandler).Methods("GET")
	r.HandleFunc("/games/{game_id}/matches/{match_id}/rounds/{round_number}/heatmap", matchAnalyticsController.GetRoundHeatmapHandler).Methods("GET")
	r.HandleFunc("/games/{game_id}/matches/{match_id}/positioning-stats", matchAnalyticsController.GetPositioningStatsHandler).Methods("GET")

	// Share Token API
	shareTokenController := cmd_controllers.NewShareTokenController(container)
	r.HandleFunc("/share-tokens", shareTokenController.CreateShareToken(ctx)).Methods("POST")
	r.HandleFunc("/share-tokens", shareTokenController.ListShareTokens(ctx)).Methods("GET")
	r.HandleFunc("/share-tokens/{token}", shareTokenController.GetShareToken(ctx)).Methods("GET")
	r.HandleFunc("/share-tokens/{token}", shareTokenController.RevokeShareToken(ctx)).Methods("DELETE")

	// Squad API
	r.HandleFunc("/squads", squadQueryController.DefaultSearchHandler).Methods("GET")
	r.HandleFunc("/squads", squadController.CreateSquadHandler(ctx)).Methods("POST")
	r.HandleFunc("/squads/{id}", squadController.GetSquadHandler(ctx)).Methods("GET")
	r.HandleFunc("/squads/{id}", squadController.UpdateSquadHandler(ctx)).Methods("PUT")
	r.HandleFunc("/squads/{id}", squadController.DeleteSquadHandler(ctx)).Methods("DELETE")
	r.HandleFunc("/squads/{id}/members", squadController.AddMemberHandler(ctx)).Methods("POST")
	r.HandleFunc("/squads/{id}/members/{player_id}", squadController.RemoveMemberHandler(ctx)).Methods("DELETE")
	r.HandleFunc("/squads/{id}/members/{player_id}/role", squadController.UpdateMemberRoleHandler(ctx)).Methods("PUT")

	// Teams API (alias for Squads - frontend compatibility)
	r.HandleFunc("/teams", squadQueryController.DefaultSearchHandler).Methods("GET")
	r.HandleFunc("/teams", squadController.CreateSquadHandler(ctx)).Methods("POST")
	r.HandleFunc("/teams/{id}", squadController.GetSquadHandler(ctx)).Methods("GET")
	r.HandleFunc("/teams/{id}", squadController.UpdateSquadHandler(ctx)).Methods("PUT")
	r.HandleFunc("/teams/{id}", squadController.DeleteSquadHandler(ctx)).Methods("DELETE")
	r.HandleFunc("/teams/{id}/members", squadController.AddMemberHandler(ctx)).Methods("POST")
	r.HandleFunc("/teams/{id}/members/{player_id}", squadController.RemoveMemberHandler(ctx)).Methods("DELETE")
	r.HandleFunc("/teams/{id}/members/{player_id}/role", squadController.UpdateMemberRoleHandler(ctx)).Methods("PUT")

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

	// Match-Making API
	r.HandleFunc("/match-making/queue", matchmakingController.JoinQueueHandler(ctx)).Methods("POST")
	r.HandleFunc("/match-making/queue/{session_id}", matchmakingController.LeaveQueueHandler(ctx)).Methods("DELETE")
	r.HandleFunc("/match-making/session/{session_id}", matchmakingController.GetSessionStatusHandler(ctx)).Methods("GET")
	r.HandleFunc("/match-making/pools/{game_id}", matchmakingController.GetPoolStatsHandler(ctx)).Methods("GET")

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

	// Wallet API
	r.HandleFunc("/wallet/balance", walletQueryController.GetWalletBalanceHandler).Methods("GET")
	r.HandleFunc("/wallet/transactions", walletQueryController.GetWalletTransactionsHandler).Methods("GET")

	// Payment API
	paymentController := cmd_controllers.NewPaymentController(container)
	r.HandleFunc("/payments", paymentController.CreatePaymentIntentHandler(ctx)).Methods("POST")
	r.HandleFunc("/payments", paymentController.GetUserPaymentsHandler(ctx)).Methods("GET")
	r.HandleFunc("/payments/{payment_id}", paymentController.GetPaymentHandler(ctx)).Methods("GET")
	r.HandleFunc("/payments/{payment_id}/confirm", paymentController.ConfirmPaymentHandler(ctx)).Methods("POST")
	r.HandleFunc("/payments/{payment_id}/cancel", paymentController.CancelPaymentHandler(ctx)).Methods("POST")
	r.HandleFunc("/payments/{payment_id}/refund", paymentController.RefundPaymentHandler(ctx)).Methods("POST")

	// Stripe Webhook (no auth required)
	r.HandleFunc("/webhooks/stripe", paymentController.StripeWebhookHandler(ctx)).Methods("POST")

	// Matches API (plural routes - aliases for frontend compatibility)
	r.HandleFunc(Matches, matchController.DefaultSearchHandler).Methods("GET")
	r.HandleFunc(Matches, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(MatchesDetail, matchController.GetMatchDetailHandler).Methods("GET")
	r.HandleFunc(MatchesDetail, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(MatchesTrajectory, matchAnalyticsController.GetMatchTrajectoryHandler).Methods("GET")
	r.HandleFunc(MatchesTrajectory, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(MatchesHeatmap, matchAnalyticsController.GetMatchHeatmapHandler).Methods("GET")
	r.HandleFunc(MatchesHeatmap, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(MatchesPosStats, matchAnalyticsController.GetPositioningStatsHandler).Methods("GET")
	r.HandleFunc(MatchesPosStats, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(RoundTrajectory, matchAnalyticsController.GetRoundTrajectoryHandler).Methods("GET")
	r.HandleFunc(RoundTrajectory, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(RoundHeatmap, matchAnalyticsController.GetRoundHeatmapHandler).Methods("GET")
	r.HandleFunc(RoundHeatmap, OptionsHandler).Methods("OPTIONS")

	// Notifications API (stub - returns empty for now)
	notificationHandler := NewNotificationStubHandler()
	r.HandleFunc(Notifications, notificationHandler.ListNotifications).Methods("GET")
	r.HandleFunc(Notifications, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(NotificationDetail, notificationHandler.GetNotification).Methods("GET")
	r.HandleFunc(NotificationDetail, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(NotificationRead, notificationHandler.MarkAsRead).Methods("PUT", "POST")
	r.HandleFunc(NotificationRead, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(NotificationsReadAll, notificationHandler.MarkAllAsRead).Methods("PUT", "POST")
	r.HandleFunc(NotificationsReadAll, OptionsHandler).Methods("OPTIONS")

	// Add NotFound handler with CORS headers
	r.NotFoundHandler = corsMiddleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 page not found"))
	}))

	return r
}
