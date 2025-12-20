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
	"github.com/replay-api/replay-api/cmd/rest-api/docs"
	"github.com/replay-api/replay-api/cmd/rest-api/middlewares"
	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
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
	AuthLogin              string = "/auth/login"
	AuthRefresh            string = "/auth/refresh"
	AuthLogout             string = "/auth/logout"
	AuthVerifyEmail         string = "/auth/verify-email"
	AuthVerificationSend    string = "/auth/verification/send"
	AuthVerificationResend  string = "/auth/verification/resend"
	AuthVerificationStatus  string = "/auth/verification/status"
	AuthPasswordReset       string = "/auth/password-reset"
	AuthPasswordResetConfirm string = "/auth/password-reset/confirm"
	AuthPasswordResetValidate string = "/auth/password-reset/validate"
	AuthMFASetup             string = "/auth/mfa/setup"
	AuthMFAVerify            string = "/auth/mfa/verify"
	AuthMFAValidate          string = "/auth/mfa/validate"
	AuthMFADisable           string = "/auth/mfa/disable"
	AuthMFAStatus            string = "/auth/mfa/status"
	AuthMFABackupCodes       string = "/auth/mfa/backup-codes/regenerate"

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
	emailController := controllers.NewEmailController(&container)
	authController := controllers.NewAuthController(&container)
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
	walletQueryController := query_controllers.NewWalletQueryController(container)
	ratingController := query_controllers.NewRatingController(container)

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

	// CORS middleware is applied as a wrapper (not r.Use()) to handle OPTIONS
	// preflight requests even when route path matches but method doesn't
	corsMiddleware := middlewares.NewCORSMiddleware()

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

	// cors

	// search mux
	r.HandleFunc(Search, searchMux.Dispatch).Methods("GET")

	// Health checks (Kubernetes-ready)
	r.HandleFunc(Health, healthController.HealthCheck(ctx)).Methods("GET")
	r.HandleFunc("/health/live", healthController.LivenessCheck(ctx)).Methods("GET")
	r.HandleFunc("/health/ready", healthController.ReadinessCheck(ctx)).Methods("GET")
	r.HandleFunc("/health/detailed", healthController.DetailedHealthCheck(ctx)).Methods("GET")
	r.HandleFunc("/health/component", healthController.ComponentHealth(ctx)).Methods("GET")
	r.HandleFunc("/ready", healthController.ReadinessCheck(ctx)).Methods("GET") // Legacy compatibility

	// Prometheus metrics
	r.Handle("/metrics", healthController.MetricsHandler()).Methods("GET")

	// API Documentation (Swagger UI, ReDoc, OpenAPI spec)
	docsConfig := docs.DefaultSwaggerConfig()
	r.HandleFunc("/api/docs", docs.DocsIndexHandler(docsConfig)).Methods("GET")
	r.HandleFunc("/api/docs/swagger", docs.SwaggerUIHandler(docsConfig)).Methods("GET")
	r.HandleFunc("/api/docs/redoc", docs.RedocHandler(docsConfig)).Methods("GET")
	r.HandleFunc("/api/docs/openapi.yaml", docs.OpenAPISpecHandler()).Methods("GET")

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

	// auth/refresh - Token refresh for session extension
	r.HandleFunc(AuthRefresh, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(AuthRefresh, authController.RefreshToken(ctx)).Methods("POST")

	// auth/logout - Token revocation
	r.HandleFunc(AuthLogout, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(AuthLogout, authController.Logout(ctx)).Methods("POST")

	// Email Verification API
	emailVerificationController := controllers.NewEmailVerificationController(&container)
	r.HandleFunc(AuthVerifyEmail, emailVerificationController.VerifyEmailByToken(ctx)).Methods("GET")
	r.HandleFunc(AuthVerifyEmail, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(AuthVerifyEmail, emailVerificationController.VerifyEmail(ctx)).Methods("POST")
	r.HandleFunc(AuthVerificationSend, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(AuthVerificationSend, emailVerificationController.SendVerificationEmail(ctx)).Methods("POST")
	r.HandleFunc(AuthVerificationResend, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(AuthVerificationResend, emailVerificationController.ResendVerification(ctx)).Methods("POST")
	r.HandleFunc(AuthVerificationStatus, emailVerificationController.GetVerificationStatus(ctx)).Methods("GET")

	// Password Reset API
	passwordResetController := controllers.NewPasswordResetController(&container)
	r.HandleFunc(AuthPasswordReset, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(AuthPasswordReset, passwordResetController.RequestPasswordReset(ctx)).Methods("POST")
	r.HandleFunc(AuthPasswordResetConfirm, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(AuthPasswordResetConfirm, passwordResetController.ConfirmPasswordReset(ctx)).Methods("POST")
	r.HandleFunc(AuthPasswordResetValidate, passwordResetController.ValidateResetToken(ctx)).Methods("GET")

	// MFA API
	mfaController := controllers.NewMFAController(container)
	r.HandleFunc(AuthMFASetup, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(AuthMFASetup, mfaController.SetupMFAHandler(ctx)).Methods("POST")
	r.HandleFunc(AuthMFAVerify, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(AuthMFAVerify, mfaController.VerifyMFAHandler(ctx)).Methods("POST")
	r.HandleFunc(AuthMFAValidate, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(AuthMFAValidate, mfaController.ValidateMFAHandler(ctx)).Methods("POST")
	r.HandleFunc(AuthMFADisable, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(AuthMFADisable, mfaController.DisableMFAHandler(ctx)).Methods("POST")
	r.HandleFunc(AuthMFAStatus, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(AuthMFAStatus, mfaController.GetMFAStatusHandler(ctx)).Methods("GET")
	r.HandleFunc(AuthMFABackupCodes, OptionsHandler).Methods("OPTIONS")
	r.HandleFunc(AuthMFABackupCodes, mfaController.RegenerateBackupCodesHandler(ctx)).Methods("POST")

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
	r.HandleFunc("/games/{game_id}/replays/{id}/events", fileController.GetReplayEvents(ctx)).Methods("GET")
	r.HandleFunc("/games/{game_id}/replays/{id}/scoreboard", fileController.GetReplayScoreboard(ctx)).Methods("GET")
	r.HandleFunc("/games/{game_id}/replays/{id}/timeline", fileController.GetReplayTimeline(ctx)).Methods("GET")

	// Game Events API
	r.HandleFunc(GameEvents, eventController.DefaultSearchHandler).Methods("GET")

	// Match Analytics API (heatmaps, trajectories, positioning)
	matchAnalyticsController := query_controllers.NewMatchAnalyticsController(container)
	r.HandleFunc("/games/{game_id}/matches/{match_id}/trajectory", matchAnalyticsController.GetMatchTrajectoryHandler).Methods("GET")
	r.HandleFunc("/games/{game_id}/matches/{match_id}/rounds/{round_number}/trajectory", matchAnalyticsController.GetRoundTrajectoryHandler).Methods("GET")
	r.HandleFunc("/games/{game_id}/matches/{match_id}/heatmap", matchAnalyticsController.GetMatchHeatmapHandler).Methods("GET")
	r.HandleFunc("/games/{game_id}/matches/{match_id}/rounds/{round_number}/heatmap", matchAnalyticsController.GetRoundHeatmapHandler).Methods("GET")
	r.HandleFunc("/games/{game_id}/matches/{match_id}/positioning-stats", matchAnalyticsController.GetPositioningStatsHandler).Methods("GET")

	// Match History API
	r.HandleFunc("/matches/player/{player_id}", matchController.GetPlayerMatchHistoryHandler).Methods("GET")
	r.HandleFunc("/matches/squad/{squad_id}", matchController.GetSquadMatchHistoryHandler).Methods("GET")

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
	r.HandleFunc("/squads/{id}/stats", squadController.GetSquadStatsHandler(ctx)).Methods("GET")
	r.HandleFunc("/squads/{id}/invitations", squadController.GetSquadInvitationsHandler(ctx)).Methods("GET")
	r.HandleFunc("/squads/{id}/invitations", squadController.InvitePlayerHandler(ctx)).Methods("POST")
	r.HandleFunc("/squads/{id}/join-requests", squadController.RequestJoinHandler(ctx)).Methods("POST")
	r.HandleFunc("/invitations/{invitation_id}/respond", squadController.RespondToInvitationHandler(ctx)).Methods("POST")
	r.HandleFunc("/invitations/{invitation_id}", squadController.CancelInvitationHandler(ctx)).Methods("DELETE")
	r.HandleFunc("/players/{id}/invitations", squadController.GetPlayerInvitationsHandler(ctx)).Methods("GET")

	// Player Profiles API
	r.HandleFunc("/players", playerProfileController.CreatePlayerProfileHandler(ctx)).Methods("POST")
	r.HandleFunc("/players", playerProfileQueryController.DefaultSearchHandler).Methods("GET")
	r.HandleFunc("/players/{id}", playerProfileController.GetPlayerProfileHandler(ctx)).Methods("GET")
	r.HandleFunc("/players/{id}", playerProfileController.UpdatePlayerProfileHandler(ctx)).Methods("PUT")
	r.HandleFunc("/players/{id}", playerProfileController.DeletePlayerProfileHandler(ctx)).Methods("DELETE")
	r.HandleFunc("/players/{id}/stats", playerProfileQueryController.GetPlayerStatsHandler).Methods("GET")

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

	// Rating & Leaderboard API (Glicko-2)
	r.HandleFunc("/players/{id}/rating", ratingController.GetPlayerRatingHandler).Methods("GET")
	r.HandleFunc("/leaderboard", ratingController.GetLeaderboardHandler).Methods("GET")
	r.HandleFunc("/ranks/distribution", ratingController.GetRankDistributionHandler).Methods("GET")

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
	r.HandleFunc("/tournaments/{id}/brackets", tournamentCommandController.GenerateBracketsHandler(ctx)).Methods("POST")
	r.HandleFunc("/tournaments/{id}/schedule", tournamentCommandController.ScheduleMatchesHandler(ctx)).Methods("POST")
	r.HandleFunc("/tournaments/{id}/matches/{match_id}/schedule", tournamentCommandController.RescheduleMatchHandler(ctx)).Methods("PUT")
	r.HandleFunc("/tournaments/{id}/matches/{match_id}/result", tournamentCommandController.ReportMatchResultHandler(ctx)).Methods("POST")
	r.HandleFunc("/players/{player_id}/tournaments", tournamentQueryController.GetPlayerTournamentsHandler).Methods("GET")
	r.HandleFunc("/organizers/{organizer_id}/tournaments", tournamentQueryController.GetOrganizerTournamentsHandler).Methods("GET")

	// Wallet API
	r.HandleFunc("/wallet/balance", walletQueryController.GetWalletBalanceHandler).Methods("GET")
	r.HandleFunc("/wallet/transactions", walletQueryController.GetWalletTransactionsHandler).Methods("GET")

	// Withdrawal API
	withdrawalController := cmd_controllers.NewWithdrawalController(container)
	r.HandleFunc("/withdrawals", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/withdrawals", withdrawalController.CreateWithdrawalHandler(ctx)).Methods("POST")
	r.HandleFunc("/withdrawals", withdrawalController.ListWithdrawalsHandler(ctx)).Methods("GET")
	r.HandleFunc("/withdrawals/{id}", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/withdrawals/{id}", withdrawalController.GetWithdrawalHandler(ctx)).Methods("GET")
	r.HandleFunc("/withdrawals/{id}/cancel", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/withdrawals/{id}/cancel", withdrawalController.CancelWithdrawalHandler(ctx)).Methods("POST")

	// Subscription Management API
	subscriptionController := cmd_controllers.NewSubscriptionController(container)
	r.HandleFunc("/subscriptions/upgrade", subscriptionController.UpgradeSubscriptionHandler()).Methods("POST")
	r.HandleFunc("/subscriptions/downgrade", subscriptionController.DowngradeSubscriptionHandler()).Methods("POST")

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

	// Challenge API (Bug Reports, VAR, Round Restart)
	challengeController := cmd_controllers.NewChallengeController(container)
	r.HandleFunc("/challenges", challengeController.CreateChallengeHandler(ctx)).Methods("POST")
	r.HandleFunc("/challenges", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/challenges/pending", challengeController.GetPendingChallengesHandler(ctx)).Methods("GET")
	r.HandleFunc("/challenges/{id}", challengeController.GetChallengeHandler(ctx)).Methods("GET")
	r.HandleFunc("/challenges/{id}", challengeController.CancelHandler(ctx)).Methods("DELETE")
	r.HandleFunc("/challenges/{id}/evidence", challengeController.AddEvidenceHandler(ctx)).Methods("POST")
	r.HandleFunc("/challenges/{id}/vote", challengeController.VoteHandler(ctx)).Methods("POST")
	r.HandleFunc("/challenges/{id}/resolve", challengeController.ResolveHandler(ctx)).Methods("POST")
	r.HandleFunc("/matches/{match_id}/challenges", challengeController.GetChallengesByMatchHandler(ctx)).Methods("GET")

	// Wrap the router with CORS handler to handle preflight OPTIONS requests
	// This is necessary because mux middleware doesn't run for 405 responses
	return corsMiddleware.Handler(r)
}
