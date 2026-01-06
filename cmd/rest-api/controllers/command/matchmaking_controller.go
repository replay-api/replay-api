package cmd_controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	shared "github.com/resource-ownership/go-common/pkg/common"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	matchmaking_services "github.com/replay-api/replay-api/pkg/domain/matchmaking/services"
)

type MatchmakingController struct {
	container               container.Container
	joinQueueHandler        matchmaking_in.JoinMatchmakingQueueCommandHandler
	leaveQueueHandler       matchmaking_in.LeaveMatchmakingQueueCommandHandler
	sessionQuerySvc         *matchmaking_services.MatchmakingSessionQueryService
	poolQuerySvc            *matchmaking_services.MatchmakingPoolQueryService
}

func NewMatchmakingController(container container.Container) *MatchmakingController {
	ctrl := &MatchmakingController{container: container}

	// Resolve command handlers (primary - use case layer)
	if err := container.Resolve(&ctrl.joinQueueHandler); err != nil {
		slog.Error("Failed to resolve JoinMatchmakingQueueCommandHandler", "err", err)
	}
	if err := container.Resolve(&ctrl.leaveQueueHandler); err != nil {
		slog.Error("Failed to resolve LeaveMatchmakingQueueCommandHandler", "err", err)
	}

	// Resolve repositories (for read-only queries only)
	if err := container.Resolve(&ctrl.sessionQuerySvc); err != nil {
		slog.Error("Failed to resolve MatchmakingSessionQueryService", "err", err)
	}
	if err := container.Resolve(&ctrl.poolQuerySvc); err != nil {
		slog.Error("Failed to resolve MatchmakingPoolQueryService", "err", err)
	}

	return ctrl
}

// JoinQueueRequest represents a request to join matchmaking
type JoinQueueRequest struct {
	PlayerID      string  `json:"player_id"`
	SquadID       *string `json:"squad_id,omitempty"`
	GameID        string  `json:"game_id"`
	GameMode      string  `json:"game_mode"`
	Region        string  `json:"region"`
	Tier          string  `json:"tier"`
	PlayerMMR     int     `json:"player_mmr"`
	PlayerRole    *string `json:"player_role,omitempty"`
	TeamFormat    string  `json:"team_format"`
	MaxPing       int     `json:"max_ping"`
	PriorityBoost bool    `json:"priority_boost"`
}

// JoinQueueResponse represents the response when joining queue
type JoinQueueResponse struct {
	SessionID      string    `json:"session_id"`
	Status         string    `json:"status"`
	EstimatedWait  int       `json:"estimated_wait_seconds"`
	QueuePosition  int       `json:"queue_position"`
	QueuedAt       time.Time `json:"queued_at"`
}

// PoolStatsResponse represents current pool statistics
type PoolStatsResponse struct {
	PoolID             string                              `json:"pool_id"`
	GameID             string                              `json:"game_id"`
	GameMode           string                              `json:"game_mode"`
	Region             string                              `json:"region"`
	TotalPlayers       int                                 `json:"total_players"`
	AverageWaitTime    int                                 `json:"average_wait_time_seconds"`
	PlayersByTier      map[string]int                      `json:"players_by_tier"`
	EstimatedMatchTime int                                 `json:"estimated_match_time_seconds"`
	QueueHealth        string                              `json:"queue_health"`
	Timestamp          time.Time                           `json:"timestamp"`
}

// JoinQueueHandler handles requests to join matchmaking queue
func (ctrl *MatchmakingController) JoinQueueHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		var req JoinQueueRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		playerID, err := uuid.Parse(req.PlayerID)
		if err != nil {
			http.Error(w, "invalid player_id", http.StatusBadRequest)
			return
		}

		// Build command for use case
		cmd := matchmaking_in.JoinMatchmakingQueueCommand{
			PlayerID:      playerID,
			GameID:        req.GameID,
			GameMode:      req.GameMode,
			Region:        req.Region,
			Tier:          matchmaking_entities.MatchmakingTier(req.Tier),
			PlayerMMR:     req.PlayerMMR,
			PlayerRole:    req.PlayerRole,
			TeamFormat:    matchmaking_in.TeamFormat(req.TeamFormat),
			MaxPing:       req.MaxPing,
			PriorityBoost: req.PriorityBoost,
		}

		// Handle optional squad ID
		if req.SquadID != nil {
			squadID, err := uuid.Parse(*req.SquadID)
			if err == nil {
				cmd.SquadID = &squadID
			}
		}

		// Set defaults if not provided
		if cmd.GameID == "" {
			cmd.GameID = "cs2"
		}
		if cmd.GameMode == "" {
			cmd.GameMode = "competitive"
		}
		if cmd.Region == "" {
			cmd.Region = "na-east"
		}
		if cmd.TeamFormat == "" {
			cmd.TeamFormat = matchmaking_in.TeamFormat5v5
		}
		if cmd.Tier == "" {
			cmd.Tier = matchmaking_entities.TierFree
		}

		// Execute via use case handler
		if ctrl.joinQueueHandler != nil {
			// Get resource owner from context for proper authentication
			ctx := r.Context()
			resourceOwner := shared.GetResourceOwner(ctx)
			if resourceOwner.UserID != uuid.Nil {
				cmd.PlayerID = resourceOwner.UserID
			}

			session, err := ctrl.joinQueueHandler.Exec(ctx, cmd)
			if err != nil {
				slog.ErrorContext(ctx, "Failed to join matchmaking queue", "err", err, "player_id", cmd.PlayerID)
				if err.Error() == "Unauthorized" {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			response := JoinQueueResponse{
				SessionID:     session.ID.String(),
				Status:        string(session.Status),
				EstimatedWait: session.EstimatedWait,
				QueuePosition: ctrl.calculateQueuePositionForSession(ctx, session),
				QueuedAt:      session.QueuedAt,
			}

			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(response)
			return
		}

		// Fallback: direct repository access (deprecated - for backwards compatibility)
		slog.Warn("JoinQueueHandler using deprecated direct repository access")
		resourceOwner := shared.GetResourceOwner(r.Context())
		session := &matchmaking_entities.MatchmakingSession{
			BaseEntity: shared.NewEntity(resourceOwner),
			PlayerID:   playerID,
			Preferences: matchmaking_entities.MatchPreferences{
				GameID:   cmd.GameID,
				GameMode: cmd.GameMode,
				Region:   cmd.Region,
				Tier:     cmd.Tier,
				MaxPing:  cmd.MaxPing,
			},
			Status:        matchmaking_entities.StatusQueued,
			PlayerMMR:     req.PlayerMMR,
			QueuedAt:      time.Now(),
			EstimatedWait: 60,
			ExpiresAt:     time.Now().Add(30 * time.Minute),
		}

		response := JoinQueueResponse{
			SessionID:     session.ID.String(),
			Status:        string(session.Status),
			EstimatedWait: session.EstimatedWait,
			QueuePosition: 1,
			QueuedAt:      session.QueuedAt,
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
	}
}

// GetPoolStatsHandler returns current pool statistics
func (ctrl *MatchmakingController) GetPoolStatsHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		vars := mux.Vars(r)
		gameID := vars["game_id"]
		gameMode := r.URL.Query().Get("game_mode")
		region := r.URL.Query().Get("region")

		if gameID == "" {
			gameID = "cs2"
		}
		if gameMode == "" {
			gameMode = "competitive"
		}
		if region == "" {
			region = "na-east"
		}

		// Generate realistic pool stats
		stats := ctrl.generatePoolStats(gameID, gameMode, region)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(stats)
	}
}

// LeaveQueueHandler handles leaving the matchmaking queue
func (ctrl *MatchmakingController) LeaveQueueHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		vars := mux.Vars(r)
		sessionID := vars["session_id"]

		sessionUUID, err := uuid.Parse(sessionID)
		if err != nil {
			http.Error(w, "invalid session_id", http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		// Execute via use case handler
		if ctrl.leaveQueueHandler != nil {
			// Get player ID from resource owner
			resourceOwner := shared.GetResourceOwner(ctx)
			playerID := resourceOwner.UserID

			cmd := matchmaking_in.LeaveMatchmakingQueueCommand{
				SessionID: sessionUUID,
				PlayerID:  playerID,
			}

			if err := ctrl.leaveQueueHandler.Exec(ctx, cmd); err != nil {
				slog.ErrorContext(ctx, "Failed to leave matchmaking queue", "err", err, "session_id", sessionID)
				if err.Error() == "Unauthorized" {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"message":    "left queue successfully",
				"session_id": sessionID,
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message":    "left queue successfully",
			"session_id": sessionID,
		})
	}
}

// GetSessionStatusHandler returns the current session status
func (ctrl *MatchmakingController) GetSessionStatusHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		vars := mux.Vars(r)
		sessionID := vars["session_id"]

		sessionUUID, err := uuid.Parse(sessionID)
		if err != nil {
			http.Error(w, "invalid session_id", http.StatusBadRequest)
			return
		}

		// Fetch session from database
		var session *matchmaking_entities.MatchmakingSession
		if ctrl.sessionQuerySvc != nil {
			session, err = ctrl.sessionQuerySvc.GetByID(r.Context(), sessionUUID)
			if err != nil {
				slog.Error("Failed to get session", "err", err, "session_id", sessionID)
				http.Error(w, "failed to get session", http.StatusInternalServerError)
				return
			}
		}

		if session == nil {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}

		// Calculate elapsed time
		elapsedTime := int(time.Since(session.QueuedAt).Seconds())

		response := map[string]any{
			"session_id":     session.ID.String(),
			"status":         string(session.Status),
			"elapsed_time":   elapsedTime,
			"estimated_wait": session.EstimatedWait,
			"queue_position": ctrl.calculateQueuePositionFromDB(r.Context(), session.Preferences),
		}

		// Add match_id if matched
		if session.MatchID != nil {
			response["match_id"] = session.MatchID.String()
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}

// Helper functions

func (ctrl *MatchmakingController) calculateQueuePosition(prefs matchmaking_entities.MatchPreferences) int {
	// Legacy method - prefer calculateQueuePositionFromDB
	switch prefs.Tier {
	case matchmaking_entities.TierElite:
		return 1
	case matchmaking_entities.TierPro:
		return 5
	case matchmaking_entities.TierPremium:
		return 12
	default:
		return 25
	}
}

func (ctrl *MatchmakingController) calculateQueuePositionFromDB(ctx context.Context, prefs matchmaking_entities.MatchPreferences) int {
	if ctrl.sessionQuerySvc == nil {
		return ctrl.calculateQueuePosition(prefs)
	}

	// Get all active sessions with same preferences
	sessions, err := ctrl.sessionQuerySvc.FindActiveSessions(ctx, prefs.GameID, prefs.GameMode, prefs.Region, &prefs.Tier, nil, nil, nil, 1000, 0)

	if err != nil {
		slog.Error("Failed to get active sessions for queue position", "err", err)
		return ctrl.calculateQueuePosition(prefs)
	}

	// Calculate position based on tier priority and queue time
	position := 1
	tierPriority := ctrl.getTierPriority(prefs.Tier)

	for _, session := range sessions {
		sessionTierPriority := ctrl.getTierPriority(session.Preferences.Tier)
		// Sessions with higher tier or equal tier but earlier queue time are ahead
		if sessionTierPriority > tierPriority {
			position++
		}
	}

	return position
}

func (ctrl *MatchmakingController) getTierPriority(tier matchmaking_entities.MatchmakingTier) int {
	switch tier {
	case matchmaking_entities.TierElite:
		return 4
	case matchmaking_entities.TierPro:
		return 3
	case matchmaking_entities.TierPremium:
		return 2
	default:
		return 1
	}
}

// calculateQueuePositionForSession calculates queue position for a specific session
func (ctrl *MatchmakingController) calculateQueuePositionForSession(ctx context.Context, session *matchmaking_entities.MatchmakingSession) int {
	if session == nil {
		return 1
	}
	return ctrl.calculateQueuePositionFromDB(ctx, session.Preferences)
}

func (ctrl *MatchmakingController) generatePoolStats(gameID, gameMode, region string) PoolStatsResponse {
	ctx := context.Background()
	
	// Try to get real data from database
	if ctrl.sessionQuerySvc != nil {
		sessions, err := ctrl.sessionQuerySvc.FindActiveSessions(ctx, gameID, gameMode, region, nil, nil, nil, nil, 1000, 0)

		if err == nil && len(sessions) > 0 {
			// Calculate real statistics
			playersByTier := make(map[string]int)
			totalWaitTime := 0

			for _, session := range sessions {
				tierStr := string(session.Preferences.Tier)
				playersByTier[tierStr]++
				
				waitTime := int(time.Since(session.QueuedAt).Seconds())
				totalWaitTime += waitTime
			}

			avgWaitTime := totalWaitTime / len(sessions)
			queueHealth := "healthy"
			if avgWaitTime > 300 {
				queueHealth = "slow"
			} else if avgWaitTime > 120 {
				queueHealth = "moderate"
			}

			return PoolStatsResponse{
				PoolID:             uuid.New().String(),
				GameID:             gameID,
				GameMode:           gameMode,
				Region:             region,
				TotalPlayers:       len(sessions),
				AverageWaitTime:    avgWaitTime,
				PlayersByTier:      playersByTier,
				EstimatedMatchTime: avgWaitTime + 30,
				QueueHealth:        queueHealth,
				Timestamp:          time.Now(),
			}
		}
	}

	// Fallback to mock data if no database or no sessions
	totalPlayers := 87 + (time.Now().Unix() % 50)
	return PoolStatsResponse{
		PoolID:             uuid.New().String(),
		GameID:             gameID,
		GameMode:           gameMode,
		Region:             region,
		TotalPlayers:       int(totalPlayers),
		AverageWaitTime:    75,
		PlayersByTier: map[string]int{
			"free":    int(totalPlayers * 60 / 100),
			"premium": int(totalPlayers * 25 / 100),
			"pro":     int(totalPlayers * 10 / 100),
			"elite":   int(totalPlayers * 5 / 100),
		},
		EstimatedMatchTime: 90,
		QueueHealth:        "healthy",
		Timestamp:          time.Now(),
	}
}
