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
	common "github.com/replay-api/replay-api/pkg/domain"
	analytics_entities "github.com/replay-api/replay-api/pkg/domain/analytics/entities"
	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	"github.com/replay-api/replay-api/pkg/infra/metrics"
	websocket "github.com/replay-api/replay-api/pkg/infra/websocket"
	"go.mongodb.org/mongo-driver/mongo"
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
	Notifications             string = "/notifications"
	NotificationDetail        string = "/notifications/{notification_id}"
	NotificationRead          string = "/notifications/{notification_id}/read"
	NotificationsReadAll      string = "/notifications/read-all"
	NotificationsUnreadCount  string = "/notifications/unread-count"
	NotificationsSeed         string = "/notifications/seed"

	// Messaging - Match Comments
	MatchComments       string = "/matches/{match_id}/comments"
	MatchCommentDetail  string = "/matches/{match_id}/comments/{comment_id}"
	MatchCommentReplies string = "/matches/{match_id}/comments/{comment_id}/replies"
	MatchCommentReact   string = "/matches/{match_id}/comments/{comment_id}/reactions"

	// Messaging - Direct Messages
	Conversations       string = "/messages/conversations"
	ConversationDetail  string = "/messages/conversations/{user_id}"
	ConversationRead    string = "/messages/conversations/{user_id}/read"
	DirectMessages      string = "/messages/{user_id}"
	DirectMessageDelete string = "/messages/{message_id}/delete"

	// Messaging - Team Messages
	TeamMessages string = "/teams/{team_id}/messages"
	TeamChannels string = "/teams/{team_id}/channels"

	// Messaging WebSocket
	MessagingWS string = "/ws/messaging"

	// Predictions - Markets
	PredictionMarkets       string = "/predictions/markets"
	PredictionMarketDetail  string = "/predictions/markets/{market_id}"
	PredictionMarketLock    string = "/predictions/markets/{market_id}/lock"
	PredictionMarketResolve string = "/predictions/markets/{market_id}/resolve"
	PredictionMarketCancel  string = "/predictions/markets/{market_id}/cancel"
	PredictionMarketBets    string = "/predictions/markets/{market_id}/bets"
	PredictionMarketSummary string = "/predictions/markets/{market_id}/summary"
	PredictionMatchMarkets  string = "/predictions/matches/{match_id}/markets"

	// Predictions - Bets & Leaderboard
	PredictionUserBets   string = "/predictions/bets/me"
	PredictionLeaderboard string = "/predictions/leaderboard"

	// View Analytics
	PlayerViews        string = "/players/{id}/views"
	PlayerViewStats    string = "/players/{id}/views/stats"
	PlayerViewInsights string = "/players/{id}/views/insights"
	TeamViews          string = "/teams/{id}/views"
	TeamViewStats      string = "/teams/{id}/views/stats"
	TeamViewInsights   string = "/teams/{id}/views/insights"
	MatchViews         string = "/games/{game_id}/matches/{match_id}/views"
	MatchViewStats     string = "/games/{game_id}/matches/{match_id}/views/stats"
	ReplayViews        string = "/games/{game_id}/replays/{replay_id}/views"
	ReplayViewStats    string = "/games/{game_id}/replays/{replay_id}/views/stats"
	MyAnalyticsViews   string = "/me/analytics/views"
	MyViewPrivacy      string = "/me/settings/view-privacy"
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
	walletCommandController := cmd_controllers.NewWalletCommandController(container)

	// Subscription and Plan controllers
	subscriptionQueryController := query_controllers.NewSubscriptionQueryController(container)
	planQueryController := query_controllers.NewPlanQueryController(container)

	// Subscription command controllers (upgrade, downgrade)
	subscriptionCommandController := cmd_controllers.NewSubscriptionController(container)

	// Checkout controller (payment → subscription activation)
	checkoutController := cmd_controllers.NewCheckoutController(container)

	// Exchange controller (buy/sell BTC, quotes, rates, orders, fees)
	exchangeController := cmd_controllers.NewExchangeController(container)

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
	notificationWebSocketHandler := websocket_controllers.NewNotificationWebSocketHandler(container, wsHub)

	// Messaging controllers
	messagingCommandController := cmd_controllers.NewMessagingCommandController(container)
	messagingQueryController := query_controllers.NewMessagingQueryController(container)
	messagingWSHandler := websocket_controllers.NewMessagingWebSocketHandler(container, wsHub)

	// Prediction controllers
	predictionCommandController := cmd_controllers.NewPredictionCommandController(container)
	predictionQueryController := query_controllers.NewPredictionQueryController(container)

	// View Analytics controllers
	viewAnalyticsCmdCtrl := cmd_controllers.NewViewAnalyticsCommandController(container)
	viewAnalyticsQueryCtrl := query_controllers.NewViewAnalyticsQueryController(container)

	// search controllers
	searchMux := query_controllers.NewSearchMux(&container)

	// Search schema controller - exposes queryable fields to frontend SDK
	searchSchemaController := query_controllers.NewSearchSchemaController()

	r := mux.NewRouter()

	// Global OPTIONS handler - must be registered BEFORE other routes
	// This handles CORS preflight for all routes
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		origin := req.Header.Get("Origin")
		if origin == "" {
			origin = "http://localhost:3030"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Resource-Owner-ID, X-Intended-Audience, X-Request-ID, X-Search, x-search, X-API-Key")
		w.Header().Set("Access-Control-Expose-Headers", "X-Resource-Owner-ID, X-Intended-Audience")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.WriteHeader(http.StatusOK)
	}).Methods("OPTIONS")

	r.Use(middlewares.ErrorMiddleware)
	r.Use(mux.CORSMethodMiddleware(r))
	r.Use(metrics.Middleware) // Prometheus HTTP request instrumentation
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

	// Search Schema API - Frontend SDK fetches this to discover queryable fields
	r.HandleFunc("/api/search/schema", searchSchemaController.GetSearchSchemaHandler).Methods("GET")
	r.HandleFunc("/api/search/schema/{entity_type}", searchSchemaController.GetEntitySchemaHandler).Methods("GET")

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

	// Matches API (singular form for backwards compatibility)
	r.HandleFunc(Match, matchController.DefaultSearchHandler).Methods("GET")
	r.HandleFunc(MatchDetail, matchController.GetMatchDetailHandler).Methods("GET")
	// Matches API (plural form for frontend SDK compatibility)
	r.HandleFunc(Matches, matchController.DefaultSearchHandler).Methods("GET")
	r.HandleFunc(MatchesDetail, matchController.GetMatchDetailHandler).Methods("GET")
	r.HandleFunc(MatchesScoreboard, matchController.GetMatchScoreboardHandler).Methods("GET")
	r.HandleFunc(MatchesEvents, eventController.GetMatchEventsHandler).Methods("GET")
	// r.HandleFunc("/games/{game_id}/matches/{match_id}/share", metadataController.GetEventsByGameIDAndMatchID(ctx)).Methods("POST")

	// Replay Files Query API (search/list)
	replayFileQueryController := query_controllers.NewReplayFileQueryController(container)
	r.HandleFunc(Replay, replayFileQueryController.ListReplayFilesHandler).Methods("GET")

	// Replay API (upload)
	r.HandleFunc(Replay, fileController.UploadHandler(ctx)).Methods("POST")
	r.HandleFunc(Replay, OptionsHandler).Methods("OPTIONS") // TODO: remover
	r.HandleFunc("/games/{game_id}/replays/{id}", replayFileQueryController.GetReplayFileHandler).Methods("GET")
	r.HandleFunc("/games/{game_id}/replays/{id}", fileController.UpdateReplayMetadata(ctx)).Methods("PUT")
	r.HandleFunc("/games/{game_id}/replays/{id}", fileController.DeleteReplayFile(ctx)).Methods("DELETE")
	r.HandleFunc("/games/{game_id}/replays/{id}/download", fileController.DownloadReplayFile(ctx)).Methods("GET")
	r.HandleFunc("/games/{game_id}/replays/{id}/status", fileController.GetReplayProcessingStatus(ctx)).Methods("GET")
	r.HandleFunc("/games/{game_id}/replays/{id}/events", fileController.GetReplayEvents(ctx)).Methods("GET")
	r.HandleFunc("/games/{game_id}/replays/{id}/scoreboard", fileController.GetReplayScoreboard(ctx)).Methods("GET")
	r.HandleFunc("/games/{game_id}/replays/{id}/timeline", fileController.GetReplayTimeline(ctx)).Methods("GET")

	// Replay Files API (alias for frontend compatibility)
	r.HandleFunc("/games/{game_id}/replay-files/{id}", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/games/{game_id}/replay-files/{id}", fileController.GetReplayMetadata(ctx)).Methods("GET")
	r.HandleFunc("/games/{game_id}/replay-files/{id}", fileController.UpdateReplayMetadata(ctx)).Methods("PUT")
	r.HandleFunc("/games/{game_id}/replay-files/{id}", fileController.DeleteReplayFile(ctx)).Methods("DELETE")
	r.HandleFunc("/games/{game_id}/replay-files/{id}/download", fileController.DownloadReplayFile(ctx)).Methods("GET")
	r.HandleFunc("/games/{game_id}/replays/{id}/status", fileController.GetReplayProcessingStatus(ctx)).Methods("GET")
	r.HandleFunc("/games/{game_id}/replay-files/{id}/events", fileController.GetReplayEvents(ctx)).Methods("GET")
	r.HandleFunc("/games/{game_id}/replay-files/{id}/scoreboard", fileController.GetReplayScoreboard(ctx)).Methods("GET")
	r.HandleFunc("/games/{game_id}/replay-files/{id}/timeline", fileController.GetReplayTimeline(ctx)).Methods("GET")

	// Game Events API
	r.HandleFunc(GameEvents, eventController.DefaultSearchHandler).Methods("GET")
	r.HandleFunc(MatchesEvents, eventController.GetMatchEventsHandler).Methods("GET")
	r.HandleFunc(MatchesEvents, OptionsHandler).Methods("OPTIONS")

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
	r.HandleFunc("/api/lobbies/{lobby_id}", lobbyController.GetLobbyHandler(ctx)).Methods("GET")
	r.HandleFunc("/api/lobbies/{lobby_id}/join", lobbyController.JoinLobbyHandler(ctx)).Methods("POST")
	r.HandleFunc("/api/lobbies/{lobby_id}/leave", lobbyController.LeaveLobbyHandler(ctx)).Methods("DELETE")
	r.HandleFunc("/api/lobbies/{lobby_id}/ready", lobbyController.SetPlayerReadyHandler(ctx)).Methods("PUT")
	r.HandleFunc("/api/lobbies/{lobby_id}/commitments", lobbyController.GetCommitmentSummaryHandler(ctx)).Methods("GET")
	r.HandleFunc("/api/lobbies/{lobby_id}/commitments/confirm", lobbyController.ConfirmReadinessHandler(ctx)).Methods("POST")
	r.HandleFunc("/api/lobbies/{lobby_id}/commitments/decline", lobbyController.DeclineReadinessHandler(ctx)).Methods("POST")
	r.HandleFunc("/api/lobbies/{lobby_id}/connection-info", lobbyController.GetGameConnectionInfoHandler(ctx)).Methods("GET")
	r.HandleFunc("/api/lobbies/{lobby_id}/start", lobbyController.StartMatchHandler(ctx)).Methods("POST")
	r.HandleFunc("/api/lobbies/{lobby_id}/invite", lobbyController.InviteToLobbyHandler(ctx)).Methods("POST")
	r.HandleFunc("/api/lobbies/{lobby_id}", lobbyController.CancelLobbyHandler(ctx)).Methods("DELETE")

	// WebSocket for real-time lobby updates
	r.HandleFunc("/ws/lobby/{lobby_id}", lobbyWebSocketHandler.UpgradeConnection(ctx)).Methods("GET")

	// WebSocket for real-time user notifications
	r.HandleFunc("/ws/notifications", notificationWebSocketHandler.UpgradeConnection(ctx)).Methods("GET")

	// Prize Pool API
	r.HandleFunc("/prize-pools/{id}", prizePoolQueryController.GetPrizePoolHandler).Methods("GET")
	r.HandleFunc("/prize-pools/{id}/history", prizePoolQueryController.GetPrizePoolHistoryHandler).Methods("GET")
	r.HandleFunc("/matches/{match_id}/prize-pool", prizePoolQueryController.GetPrizePoolByMatchHandler).Methods("GET")
	r.HandleFunc("/prize-pools/pending-distributions", prizePoolQueryController.GetPendingDistributionsHandler).Methods("GET")

	// Scores / Match Results API
	matchResultCommandController := cmd_controllers.NewMatchResultCommandController(container)
	matchResultQueryController := query_controllers.NewMatchResultQueryController(container)

	// Match Result Command endpoints (write operations)
	r.HandleFunc("/scores/match-results", matchResultCommandController.SubmitMatchResultHandler(ctx)).Methods("POST")
	r.HandleFunc("/scores/match-results", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/scores/match-results/{id}/verify", matchResultCommandController.VerifyMatchResultHandler(ctx)).Methods("PUT")
	r.HandleFunc("/scores/match-results/{id}/verify", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/scores/match-results/{id}/dispute", matchResultCommandController.DisputeMatchResultHandler(ctx)).Methods("PUT")
	r.HandleFunc("/scores/match-results/{id}/dispute", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/scores/match-results/{id}/conciliate", matchResultCommandController.ConciliateMatchResultHandler(ctx)).Methods("PUT")
	r.HandleFunc("/scores/match-results/{id}/conciliate", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/scores/match-results/{id}/finalize", matchResultCommandController.FinalizeMatchResultHandler(ctx)).Methods("PUT")
	r.HandleFunc("/scores/match-results/{id}/finalize", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/scores/match-results/{id}/cancel", matchResultCommandController.CancelMatchResultHandler(ctx)).Methods("PUT")
	r.HandleFunc("/scores/match-results/{id}/cancel", OptionsHandler).Methods("OPTIONS")

	// Match Result Query endpoints (read operations)
	r.HandleFunc("/scores/match-results", matchResultQueryController.ListMatchResultsHandler).Methods("GET")
	r.HandleFunc("/scores/match-results/{id}", matchResultQueryController.GetMatchResultHandler).Methods("GET")
	r.HandleFunc("/scores/match-results/{id}", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/scores/match-results/by-match/{matchId}", matchResultQueryController.GetMatchResultByMatchHandler).Methods("GET")
	r.HandleFunc("/scores/match-results/by-match/{matchId}", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/scores/tournaments/{tournamentId}/results", matchResultQueryController.GetTournamentResultsHandler).Methods("GET")
	r.HandleFunc("/scores/tournaments/{tournamentId}/results", OptionsHandler).Methods("OPTIONS")

	// ========================
	// Oracle API
	// ========================
	oracleCommandController := cmd_controllers.NewOracleCommandController(container)
	oracleQueryController := query_controllers.NewOracleQueryController(container)
	gameImportController := cmd_controllers.NewGameImportController(container)

	// SECURITY: Oracle write endpoints require API-key or authenticated session.
	// This prevents unauthenticated access to oracle import/command operations.
	apiKeyMiddleware := middlewares.NewAPIKeyMiddleware()

	// Game Import endpoints (discovery → oracle → match result pipeline) — require API key or auth
	r.Handle("/oracle/import", apiKeyMiddleware.RequireAPIKeyOrAuth(gameImportController.ImportDiscoveredMatchHandler(ctx))).Methods("POST")
	r.HandleFunc("/oracle/import", OptionsHandler).Methods("OPTIONS")
	r.Handle("/oracle/import/youtube", apiKeyMiddleware.RequireAPIKeyOrAuth(gameImportController.ImportFromYouTubeVODHandler(ctx))).Methods("POST")
	r.HandleFunc("/oracle/import/youtube", OptionsHandler).Methods("OPTIONS")
	r.Handle("/oracle/results/{id}/bridge", apiKeyMiddleware.RequireAPIKeyOrAuth(gameImportController.BridgeOracleToMatchResultHandler(ctx))).Methods("POST")
	r.HandleFunc("/oracle/results/{id}/bridge", OptionsHandler).Methods("OPTIONS")

	// Oracle Command endpoints (write operations) — require API key or auth
	r.Handle("/oracle/results", apiKeyMiddleware.RequireAPIKeyOrAuth(oracleCommandController.CreateExternalMatchOracleHandler(ctx))).Methods("POST")
	r.HandleFunc("/oracle/results", OptionsHandler).Methods("OPTIONS")
	r.Handle("/oracle/results/{id}/ingest", apiKeyMiddleware.RequireAPIKeyOrAuth(oracleCommandController.IngestExternalScoreHandler(ctx))).Methods("POST")
	r.HandleFunc("/oracle/results/{id}/ingest", OptionsHandler).Methods("OPTIONS")
	r.Handle("/oracle/results/{id}/publish", apiKeyMiddleware.RequireAPIKeyOrAuth(oracleCommandController.PublishToChainHandler(ctx))).Methods("POST")
	r.HandleFunc("/oracle/results/{id}/publish", OptionsHandler).Methods("OPTIONS")
	r.Handle("/oracle/results/{id}/dispute", apiKeyMiddleware.RequireAPIKeyOrAuth(oracleCommandController.DisputeEscalationHandler(ctx))).Methods("POST")
	r.HandleFunc("/oracle/results/{id}/dispute", OptionsHandler).Methods("OPTIONS")
	r.Handle("/oracle/results/trigger-ingestion", apiKeyMiddleware.RequireAPIKeyOrAuth(oracleCommandController.TriggerIngestionHandler(ctx))).Methods("POST")
	r.HandleFunc("/oracle/results/trigger-ingestion", OptionsHandler).Methods("OPTIONS")

	// Oracle Query endpoints (read operations)
	r.HandleFunc("/oracle/results", oracleQueryController.ListOracleResultsHandler).Methods("GET")
	r.HandleFunc("/oracle/results/{id}", oracleQueryController.GetOracleResultHandler).Methods("GET")
	r.HandleFunc("/oracle/results/{id}", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/oracle/results/by-match/{matchId}", oracleQueryController.GetOracleResultByMatchHandler).Methods("GET")
	r.HandleFunc("/oracle/results/by-match/{matchId}", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/oracle/results/{id}/submissions", oracleQueryController.GetSubmissionsHandler).Methods("GET")
	r.HandleFunc("/oracle/results/{id}/submissions", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/oracle/results/{id}/publications", oracleQueryController.GetPublicationStatusHandler).Methods("GET")
	r.HandleFunc("/oracle/results/{id}/publications", OptionsHandler).Methods("OPTIONS")

	// Tournament API
	r.HandleFunc("/tournaments", tournamentCommandController.CreateTournamentHandler(ctx)).Methods("POST")
	r.HandleFunc("/tournaments", tournamentQueryController.ListTournamentsHandler).Methods("GET")
	r.HandleFunc("/tournaments/upcoming", tournamentQueryController.GetUpcomingTournamentsHandler).Methods("GET")
	r.HandleFunc("/tournaments/{id}", tournamentQueryController.GetTournamentHandler).Methods("GET")
	r.HandleFunc("/tournaments/{id}", tournamentCommandController.UpdateTournamentHandler(ctx)).Methods("PUT")
	r.HandleFunc("/tournaments/{id}", tournamentCommandController.DeleteTournamentHandler(ctx)).Methods("DELETE")
	r.HandleFunc("/tournaments/{id}/register", tournamentCommandController.RegisterPlayerHandler(ctx)).Methods("POST")
	r.HandleFunc("/tournaments/{id}/register", tournamentCommandController.UnregisterPlayerHandler(ctx)).Methods("DELETE")
	r.HandleFunc("/tournaments/{id}/registration/open", tournamentCommandController.OpenRegistrationHandler(ctx)).Methods("POST")
	r.HandleFunc("/tournaments/{id}/registration/open", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/tournaments/{id}/registration/close", tournamentCommandController.CloseRegistrationHandler(ctx)).Methods("POST")
	r.HandleFunc("/tournaments/{id}/registration/close", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/tournaments/{id}/start", tournamentCommandController.StartTournamentHandler(ctx)).Methods("POST")
	r.HandleFunc("/tournaments/{id}/complete", tournamentCommandController.CompleteTournamentHandler(ctx)).Methods("POST")
	r.HandleFunc("/tournaments/{id}/complete", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/tournaments/{id}/cancel", tournamentCommandController.CancelTournamentHandler(ctx)).Methods("POST")
	r.HandleFunc("/tournaments/{id}/cancel", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/tournaments/{id}/check-in", tournamentCommandController.CheckInHandler(ctx)).Methods("POST")
	r.HandleFunc("/tournaments/{id}/check-in", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/tournaments/{id}/matches/{match_id}/result", tournamentCommandController.RecordMatchResultHandler(ctx)).Methods("POST")
	r.HandleFunc("/tournaments/{id}/matches/{match_id}/result", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/tournaments/{id}/advance-bracket", tournamentCommandController.AdvanceBracketHandler(ctx)).Methods("POST")
	r.HandleFunc("/tournaments/{id}/advance-bracket", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/players/{player_id}/tournaments", tournamentQueryController.GetPlayerTournamentsHandler).Methods("GET")
	r.HandleFunc("/organizers/{organizer_id}/tournaments", tournamentQueryController.GetOrganizerTournamentsHandler).Methods("GET")

	// Wallet API - Query endpoints
	r.HandleFunc("/wallet/balance", walletQueryController.GetWalletBalanceHandler).Methods("GET")
	r.HandleFunc("/wallet/transactions", walletQueryController.GetWalletTransactionsHandler).Methods("GET")
	r.HandleFunc("/wallet/balance", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/wallet/transactions", OptionsHandler).Methods("OPTIONS")

	// Wallet API - Command endpoints
	r.HandleFunc("/wallet", walletCommandController.CreateWalletHandler(ctx)).Methods("POST")
	r.HandleFunc("/wallet", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/wallet/deposit", walletCommandController.DepositHandler(ctx)).Methods("POST")
	r.HandleFunc("/wallet/deposit", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/wallet/withdraw", walletCommandController.WithdrawHandler(ctx)).Methods("POST")
	r.HandleFunc("/wallet/withdraw", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/wallet/entry-fee", walletCommandController.DeductEntryFeeHandler(ctx)).Methods("POST")
	r.HandleFunc("/wallet/entry-fee", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/wallet/prize", walletCommandController.AddPrizeHandler(ctx)).Methods("POST")
	r.HandleFunc("/wallet/prize", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/wallet/credit", walletCommandController.CreditWalletHandler(ctx)).Methods("POST")
	r.HandleFunc("/wallet/credit", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/wallet/debit", walletCommandController.DebitWalletHandler(ctx)).Methods("POST")
	r.HandleFunc("/wallet/debit", OptionsHandler).Methods("OPTIONS")

	// Team Vault API - Controllers
	vaultCommandController := cmd_controllers.NewVaultCommandController(container)
	vaultQueryController := query_controllers.NewVaultQueryController(container)

	// Team Vault API - Query endpoints
	r.HandleFunc("/squads/{squad_id}/vault", vaultQueryController.GetVaultHandler).Methods("GET")
	r.HandleFunc("/squads/{squad_id}/vault", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/squads/{squad_id}/vault/balance", vaultQueryController.GetVaultBalanceHandler).Methods("GET")
	r.HandleFunc("/squads/{squad_id}/vault/balance", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/squads/{squad_id}/vault/proposals", vaultQueryController.GetVaultProposalsHandler).Methods("GET")
	r.HandleFunc("/squads/{squad_id}/vault/proposals/{proposal_id}", vaultQueryController.GetVaultProposalByIDHandler).Methods("GET")
	r.HandleFunc("/squads/{squad_id}/vault/proposals/{proposal_id}", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/squads/{squad_id}/vault/activity", vaultQueryController.GetVaultActivityHandler).Methods("GET")
	r.HandleFunc("/squads/{squad_id}/vault/activity", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/squads/{squad_id}/vault/analytics", vaultQueryController.GetVaultAnalyticsHandler).Methods("GET")
	r.HandleFunc("/squads/{squad_id}/vault/analytics", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/squads/{squad_id}/vault/inventory", vaultQueryController.GetVaultInventoryHandler).Methods("GET")

	// Team Vault API - Command endpoints
	r.HandleFunc("/squads/{squad_id}/vault", vaultCommandController.CreateVaultHandler(ctx)).Methods("POST")
	r.HandleFunc("/squads/{squad_id}/vault/deposit", vaultCommandController.DepositToVaultHandler(ctx)).Methods("POST")
	r.HandleFunc("/squads/{squad_id}/vault/deposit", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/squads/{squad_id}/vault/proposals", vaultCommandController.ProposeTransactionHandler(ctx)).Methods("POST")
	r.HandleFunc("/squads/{squad_id}/vault/proposals", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/squads/{squad_id}/vault/proposals/{proposal_id}/approve", vaultCommandController.ApproveProposalHandler(ctx)).Methods("POST")
	r.HandleFunc("/squads/{squad_id}/vault/proposals/{proposal_id}/approve", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/squads/{squad_id}/vault/proposals/{proposal_id}/reject", vaultCommandController.RejectProposalHandler(ctx)).Methods("POST")
	r.HandleFunc("/squads/{squad_id}/vault/proposals/{proposal_id}/reject", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/squads/{squad_id}/vault/proposals/{proposal_id}/cancel", vaultCommandController.CancelProposalHandler(ctx)).Methods("POST")
	r.HandleFunc("/squads/{squad_id}/vault/proposals/{proposal_id}/cancel", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/squads/{squad_id}/vault/settings", vaultCommandController.UpdateVaultSettingsHandler(ctx)).Methods("PUT")
	r.HandleFunc("/squads/{squad_id}/vault/settings", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/squads/{squad_id}/vault/inventory", vaultCommandController.DepositItemHandler(ctx)).Methods("POST")
	r.HandleFunc("/squads/{squad_id}/vault/inventory", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/squads/{squad_id}/vault/inventory/transfer", vaultCommandController.ProposeItemTransferHandler(ctx)).Methods("POST")
	r.HandleFunc("/squads/{squad_id}/vault/inventory/transfer", OptionsHandler).Methods("OPTIONS")

	// Exchange API - Buy/Sell BTC, Quotes, Rates, Orders, Fees
	r.HandleFunc("/v1/exchange/buy", exchangeController.PostBuyBitcoin(ctx)).Methods("POST")
	r.HandleFunc("/v1/exchange/buy", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/v1/exchange/sell", exchangeController.PostSellBitcoin(ctx)).Methods("POST")
	r.HandleFunc("/v1/exchange/sell", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/v1/exchange/quote", exchangeController.GetQuote(ctx)).Methods("POST")
	r.HandleFunc("/v1/exchange/quote", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/v1/exchange/rates", exchangeController.GetExchangeRates(ctx)).Methods("GET")
	r.HandleFunc("/v1/exchange/rates", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/v1/exchange/orders", exchangeController.GetOrderHistory(ctx)).Methods("GET")
	r.HandleFunc("/v1/exchange/orders", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/v1/exchange/orders/{id}", exchangeController.GetOrderByID(ctx)).Methods("GET")
	r.HandleFunc("/v1/exchange/orders/{id}", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/v1/exchange/orders/{id}/cancel", exchangeController.PostCancelOrder(ctx)).Methods("POST")
	r.HandleFunc("/v1/exchange/orders/{id}/cancel", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/v1/exchange/fees", exchangeController.GetFeeSchedule(ctx)).Methods("GET")
	r.HandleFunc("/v1/exchange/fees", OptionsHandler).Methods("OPTIONS")

	// Plan API (both /plans and /subscriptions/plans for SDK compatibility)
	// NOTE: These must be registered BEFORE /subscriptions/{subscription_id} to avoid
	// gorilla/mux matching "plans" as a subscription_id parameter
	r.HandleFunc("/plans", planQueryController.ListAvailablePlansHandler).Methods("GET")
	r.HandleFunc("/plans", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/plans/{plan_id}", planQueryController.GetPlanByIDHandler).Methods("GET")
	r.HandleFunc("/plans/{plan_id}", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/subscriptions/plans", planQueryController.ListAvailablePlansHandler).Methods("GET")
	r.HandleFunc("/subscriptions/plans", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/subscriptions/plans/{plan_id}", planQueryController.GetPlanByIDHandler).Methods("GET")
	r.HandleFunc("/subscriptions/plans/{plan_id}", OptionsHandler).Methods("OPTIONS")

	// Subscription API
	r.HandleFunc("/subscriptions/current", subscriptionQueryController.GetCurrentSubscriptionHandler).Methods("GET")
	r.HandleFunc("/subscriptions/current", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/subscriptions/{subscription_id}", subscriptionQueryController.GetSubscriptionByIDHandler).Methods("GET")
	r.HandleFunc("/subscriptions/{subscription_id}", OptionsHandler).Methods("OPTIONS")

	// Subscription Mutations API (upgrade, downgrade)
	r.HandleFunc("/subscriptions/upgrade", subscriptionCommandController.UpgradeSubscriptionHandler()).Methods("POST")
	r.HandleFunc("/subscriptions/upgrade", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/subscriptions/downgrade", subscriptionCommandController.DowngradeSubscriptionHandler()).Methods("POST")
	r.HandleFunc("/subscriptions/downgrade", OptionsHandler).Methods("OPTIONS")

	// Checkout API (payment → subscription activation)
	r.HandleFunc("/checkout", checkoutController.CheckoutHandler(ctx)).Methods("POST")
	r.HandleFunc("/checkout", OptionsHandler).Methods("OPTIONS")

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

	// Notifications API (MongoDB-backed CRUD)
	var mongoClient *mongo.Client
	if err := container.Resolve(&mongoClient); err != nil {
		slog.ErrorContext(ctx, "Failed to resolve mongo.Client for NotificationHandler", "error", err)
	}
	var cfg common.Config
	if err := container.Resolve(&cfg); err != nil {
		slog.ErrorContext(ctx, "Failed to resolve config for NotificationHandler", "error", err)
	}
	notificationHandler := NewNotificationHandler(mongoClient, cfg.MongoDB.DBName)

	// NOTE: Static routes MUST be registered BEFORE parameterized routes
	// to prevent gorilla/mux from matching "read-all", "unread-count", "seed" as {notification_id}
	r.HandleFunc(NotificationsReadAll, notificationHandler.MarkAllAsRead).Methods("PUT", "POST")
	r.HandleFunc(NotificationsReadAll, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(NotificationsUnreadCount, notificationHandler.GetUnreadCount).Methods("GET")
	r.HandleFunc(NotificationsUnreadCount, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(NotificationsSeed, notificationHandler.SeedNotifications).Methods("POST")
	r.HandleFunc(NotificationsSeed, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(Notifications, notificationHandler.ListNotifications).Methods("GET")
	r.HandleFunc(Notifications, notificationHandler.DeleteAllNotifications).Methods("DELETE")
	r.HandleFunc(Notifications, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(NotificationDetail, notificationHandler.GetNotification).Methods("GET")
	r.HandleFunc(NotificationDetail, notificationHandler.DeleteNotification).Methods("DELETE")
	r.HandleFunc(NotificationDetail, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(NotificationRead, notificationHandler.MarkAsRead).Methods("PUT", "POST")
	r.HandleFunc(NotificationRead, OptionsHandler).Methods("OPTIONS")

	// Messaging - Match Comments API
	r.HandleFunc(MatchComments, messagingQueryController.ListMatchCommentsHandler).Methods("GET")
	r.HandleFunc(MatchComments, messagingCommandController.CreateCommentHandler(ctx)).Methods("POST")
	r.HandleFunc(MatchComments, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(MatchCommentDetail, messagingQueryController.GetCommentHandler).Methods("GET")
	r.HandleFunc(MatchCommentDetail, messagingCommandController.EditCommentHandler(ctx)).Methods("PUT")
	r.HandleFunc(MatchCommentDetail, messagingCommandController.DeleteCommentHandler(ctx)).Methods("DELETE")
	r.HandleFunc(MatchCommentDetail, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(MatchCommentReplies, messagingQueryController.GetCommentRepliesHandler).Methods("GET")
	r.HandleFunc(MatchCommentReplies, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(MatchCommentReact, messagingCommandController.ReactToCommentHandler(ctx)).Methods("POST")
	r.HandleFunc(MatchCommentReact, OptionsHandler).Methods("OPTIONS")

	// Messaging - Direct Messages API
	r.HandleFunc(Conversations, messagingQueryController.ListConversationsHandler).Methods("GET")
	r.HandleFunc(Conversations, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(ConversationDetail, messagingQueryController.GetConversationHandler).Methods("GET")
	r.HandleFunc(ConversationDetail, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(ConversationRead, messagingCommandController.MarkConversationReadHandler(ctx)).Methods("PUT", "POST")
	r.HandleFunc(ConversationRead, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(DirectMessages, messagingCommandController.SendDirectMessageHandler(ctx)).Methods("POST")
	r.HandleFunc(DirectMessages, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(DirectMessageDelete, messagingCommandController.DeleteDirectMessageHandler(ctx)).Methods("DELETE")
	r.HandleFunc(DirectMessageDelete, OptionsHandler).Methods("OPTIONS")

	// Messaging - Team Messages API
	r.HandleFunc(TeamMessages, messagingQueryController.ListTeamMessagesHandler).Methods("GET")
	r.HandleFunc(TeamMessages, messagingCommandController.SendTeamMessageHandler(ctx)).Methods("POST")
	r.HandleFunc(TeamMessages, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(TeamChannels, messagingQueryController.ListTeamChannelsHandler).Methods("GET")
	r.HandleFunc(TeamChannels, OptionsHandler).Methods("OPTIONS")

	// Messaging WebSocket
	r.HandleFunc(MessagingWS, messagingWSHandler.UpgradeConnection(ctx)).Methods("GET")

	// View Analytics API - Player Views
	r.HandleFunc(PlayerViews, viewAnalyticsCmdCtrl.RecordViewHandler(analytics_entities.EntityTypePlayer)(ctx)).Methods("POST")
	r.HandleFunc(PlayerViews, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(PlayerViewStats, viewAnalyticsQueryCtrl.GetViewStatisticsHandler(analytics_entities.EntityTypePlayer)(ctx)).Methods("GET")
	r.HandleFunc(PlayerViewStats, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(PlayerViewInsights, viewAnalyticsQueryCtrl.GetViewInsightsHandler(analytics_entities.EntityTypePlayer)(ctx)).Methods("GET")
	r.HandleFunc(PlayerViewInsights, OptionsHandler).Methods("OPTIONS")

	// View Analytics API - Team Views
	r.HandleFunc(TeamViews, viewAnalyticsCmdCtrl.RecordViewHandler(analytics_entities.EntityTypeTeam)(ctx)).Methods("POST")
	r.HandleFunc(TeamViews, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(TeamViewStats, viewAnalyticsQueryCtrl.GetViewStatisticsHandler(analytics_entities.EntityTypeTeam)(ctx)).Methods("GET")
	r.HandleFunc(TeamViewStats, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(TeamViewInsights, viewAnalyticsQueryCtrl.GetViewInsightsHandler(analytics_entities.EntityTypeTeam)(ctx)).Methods("GET")
	r.HandleFunc(TeamViewInsights, OptionsHandler).Methods("OPTIONS")

	// View Analytics API - Match Views
	r.HandleFunc(MatchViews, viewAnalyticsCmdCtrl.RecordViewHandler(analytics_entities.EntityTypeMatch)(ctx)).Methods("POST")
	r.HandleFunc(MatchViews, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(MatchViewStats, viewAnalyticsQueryCtrl.GetViewStatisticsHandler(analytics_entities.EntityTypeMatch)(ctx)).Methods("GET")
	r.HandleFunc(MatchViewStats, OptionsHandler).Methods("OPTIONS")

	// View Analytics API - Replay Views
	r.HandleFunc(ReplayViews, viewAnalyticsCmdCtrl.RecordViewHandler(analytics_entities.EntityTypeReplay)(ctx)).Methods("POST")
	r.HandleFunc(ReplayViews, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(ReplayViewStats, viewAnalyticsQueryCtrl.GetViewStatisticsHandler(analytics_entities.EntityTypeReplay)(ctx)).Methods("GET")
	r.HandleFunc(ReplayViewStats, OptionsHandler).Methods("OPTIONS")

	// View Analytics API - My Analytics & Privacy
	r.HandleFunc(MyAnalyticsViews, viewAnalyticsQueryCtrl.GetMyAnalyticsHandler(ctx)).Methods("GET")
	r.HandleFunc(MyAnalyticsViews, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(MyViewPrivacy, viewAnalyticsQueryCtrl.GetViewPrivacyHandler(ctx)).Methods("GET")
	r.HandleFunc(MyViewPrivacy, viewAnalyticsCmdCtrl.UpdateViewPrivacyHandler(ctx)).Methods("PUT")
	r.HandleFunc(MyViewPrivacy, OptionsHandler).Methods("OPTIONS")

	// Predictions - Markets API
	r.HandleFunc(PredictionMarkets, predictionCommandController.CreateMarketHandler(ctx)).Methods("POST")
	r.HandleFunc(PredictionMarkets, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(PredictionMarketDetail, predictionQueryController.GetMarketHandler).Methods("GET")
	r.HandleFunc(PredictionMarketDetail, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(PredictionMarketLock, predictionCommandController.LockMarketHandler(ctx)).Methods("POST")
	r.HandleFunc(PredictionMarketLock, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(PredictionMarketResolve, predictionCommandController.ResolveMarketHandler(ctx)).Methods("POST")
	r.HandleFunc(PredictionMarketResolve, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(PredictionMarketCancel, predictionCommandController.CancelMarketHandler(ctx)).Methods("POST")
	r.HandleFunc(PredictionMarketCancel, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(PredictionMatchMarkets, predictionQueryController.ListMatchMarketsHandler).Methods("GET")
	r.HandleFunc(PredictionMatchMarkets, OptionsHandler).Methods("OPTIONS")

	// Predictions - Bets API
	r.HandleFunc(PredictionMarketBets, predictionQueryController.GetMarketBetsHandler).Methods("GET")
	r.HandleFunc(PredictionMarketBets, predictionCommandController.PlaceBetHandler(ctx)).Methods("POST")
	r.HandleFunc(PredictionMarketBets, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(PredictionMarketSummary, predictionQueryController.GetUserBetSummaryHandler).Methods("GET")
	r.HandleFunc(PredictionMarketSummary, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(PredictionUserBets, predictionQueryController.GetUserBetsHandler).Methods("GET")
	r.HandleFunc(PredictionUserBets, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(PredictionLeaderboard, predictionQueryController.GetLeaderboardHandler).Methods("GET")
	r.HandleFunc(PredictionLeaderboard, OptionsHandler).Methods("OPTIONS")

	// Add NotFound handler with CORS headers
	r.NotFoundHandler = corsMiddleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 page not found"))
	}))

	return r
}
