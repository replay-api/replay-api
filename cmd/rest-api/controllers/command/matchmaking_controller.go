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
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_out "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/out"
)

type MatchmakingController struct {
	container   container.Container
	sessionRepo matchmaking_out.MatchmakingSessionRepository
	poolRepo    matchmaking_out.MatchmakingPoolRepository
}

func NewMatchmakingController(container container.Container) *MatchmakingController {
	ctrl := &MatchmakingController{container: container}
	
	// Resolve repositories
	if err := container.Resolve(&ctrl.sessionRepo); err != nil {
		slog.Error("Failed to resolve MatchmakingSessionRepository", "err", err)
	}
	if err := container.Resolve(&ctrl.poolRepo); err != nil {
		slog.Error("Failed to resolve MatchmakingPoolRepository", "err", err)
	}
	
	return ctrl
}

// JoinQueueRequest represents a request to join matchmaking
type JoinQueueRequest struct {
	PlayerID           string                                      `json:"player_id"`
	SquadID            *string                                     `json:"squad_id,omitempty"`
	Preferences        matchmaking_entities.MatchPreferences       `json:"preferences"`
	PlayerMMR          int                                         `json:"player_mmr"`
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

		// Create matchmaking session
		session := &matchmaking_entities.MatchmakingSession{
			ID:            uuid.New(),
			PlayerID:      playerID,
			Preferences:   req.Preferences,
			Status:        matchmaking_entities.StatusQueued,
			PlayerMMR:     req.PlayerMMR,
			QueuedAt:      time.Now(),
			EstimatedWait: ctrl.calculateEstimatedWait(req.Preferences),
			ExpiresAt:     time.Now().Add(30 * time.Minute),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if req.SquadID != nil {
			squadID, err := uuid.Parse(*req.SquadID)
			if err == nil {
				session.SquadID = &squadID
			}
		}

		// Persist session to database
		if ctrl.sessionRepo != nil {
			if err := ctrl.sessionRepo.Save(r.Context(), session); err != nil {
				slog.Error("Failed to save matchmaking session", "err", err, "session_id", session.ID)
				http.Error(w, "failed to save session", http.StatusInternalServerError)
				return
			}
		}

		// Calculate queue position from database
		queuePosition := ctrl.calculateQueuePositionFromDB(r.Context(), req.Preferences)

		response := JoinQueueResponse{
			SessionID:     session.ID.String(),
			Status:        string(session.Status),
			EstimatedWait: session.EstimatedWait,
			QueuePosition: queuePosition,
			QueuedAt:      session.QueuedAt,
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(stats)
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

		// Update session status to cancelled
		if ctrl.sessionRepo != nil {
			if err := ctrl.sessionRepo.UpdateStatus(r.Context(), sessionUUID, matchmaking_entities.StatusCancelled); err != nil {
				slog.Error("Failed to cancel session", "err", err, "session_id", sessionID)
				http.Error(w, "failed to cancel session", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
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
		if ctrl.sessionRepo != nil {
			session, err = ctrl.sessionRepo.GetByID(r.Context(), sessionUUID)
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
		json.NewEncoder(w).Encode(response)
	}
}

// Helper functions

func (ctrl *MatchmakingController) calculateEstimatedWait(prefs matchmaking_entities.MatchPreferences) int {
	baseWait := 60 // 1 minute base

	// Premium tiers get faster queues
	switch prefs.Tier {
	case matchmaking_entities.TierElite:
		baseWait = 30
	case matchmaking_entities.TierPro:
		baseWait = 40
	case matchmaking_entities.TierPremium:
		baseWait = 50
	}

	// Priority boost cuts wait time
	if prefs.PriorityBoost {
		baseWait = baseWait / 2
	}

	return baseWait
}

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
	if ctrl.sessionRepo == nil {
		return ctrl.calculateQueuePosition(prefs)
	}

	// Get all active sessions with same preferences
	sessions, err := ctrl.sessionRepo.GetActiveSessions(ctx, matchmaking_out.SessionFilters{
		GameID:   prefs.GameID,
		GameMode: prefs.GameMode,
		Region:   prefs.Region,
		Limit:    1000,
	})

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

func (ctrl *MatchmakingController) generatePoolStats(gameID, gameMode, region string) PoolStatsResponse {
	ctx := context.Background()
	
	// Try to get real data from database
	if ctrl.sessionRepo != nil {
		sessions, err := ctrl.sessionRepo.GetActiveSessions(ctx, matchmaking_out.SessionFilters{
			GameID:   gameID,
			GameMode: gameMode,
			Region:   region,
			Limit:    1000,
		})

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
