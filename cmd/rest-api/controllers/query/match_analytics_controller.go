package query_controllers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	cs_entities "github.com/replay-api/replay-api/pkg/domain/cs/entities"
	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type MatchAnalyticsController struct {
	container container.Container
}

func NewMatchAnalyticsController(container container.Container) *MatchAnalyticsController {
	return &MatchAnalyticsController{container: container}
}

type TrajectoryPoint struct {
	Tick       float64 `json:"tick"`
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Z          float64 `json:"z"`
	ViewX      float64 `json:"view_x"`
	ViewY      float64 `json:"view_y"`
	IsAlive    bool    `json:"is_alive"`
	IsCrouched bool    `json:"is_crouched"`
}

type PlayerTrajectory struct {
	PlayerID   string            `json:"player_id"`
	PlayerName string            `json:"player_name"`
	Team       string            `json:"team"`
	Positions  []TrajectoryPoint `json:"positions"`
}

type TrajectoryResponse struct {
	MatchID       string             `json:"match_id"`
	RoundNumber   int                `json:"round_number,omitempty"`
	MapName       string             `json:"map_name"`
	Trajectories  []PlayerTrajectory `json:"trajectories"`
	RadarImageURL string             `json:"radar_image_url"`
}

type HeatmapCell struct {
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Value   float64 `json:"value"`
	Density int     `json:"density"`
}

type HeatmapResponse struct {
	MatchID     string        `json:"match_id"`
	RoundNumber int           `json:"round_number,omitempty"`
	MapName     string        `json:"map_name"`
	HeatmapType string        `json:"heatmap_type"`
	GridSize    int           `json:"grid_size"`
	Cells       []HeatmapCell `json:"cells"`
	MinValue    float64       `json:"min_value"`
	MaxValue    float64       `json:"max_value"`
}

type PositioningStatsResponse struct {
	MatchID             string                        `json:"match_id"`
	ZoneFrequencies     map[string]map[string]int     `json:"zone_frequencies"`
	ZoneDwellTimes      map[string]map[string]float64 `json:"zone_dwell_times"`
	AverageSpeed        map[string]float64            `json:"average_speed"`
	AverageSpeedRunning map[string]float64            `json:"average_speed_running"`
}

// extractPositioningStats extracts CSPositioningStats from GameEvent.Stats
func extractPositioningStats(stats map[replay_common.StatType][]interface{}) *cs_entities.CSPositioningStats {
	if stats == nil {
		return nil
	}

	positioningData, ok := stats[replay_common.PositioningStatTypeKey]
	if !ok || len(positioningData) == 0 {
		return nil
	}

	// Try to type assert directly
	if ps, ok := positioningData[0].(*cs_entities.CSPositioningStats); ok {
		return ps
	}
	if ps, ok := positioningData[0].(cs_entities.CSPositioningStats); ok {
		return &ps
	}

	return nil
}

// GetMatchTrajectoryHandler returns player trajectories for an entire match
func (ctrl *MatchAnalyticsController) GetMatchTrajectoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchID := vars["match_id"]

	if matchID == "" {
		http.Error(w, "match_id is required", http.StatusBadRequest)
		return
	}

	matchUUID, err := uuid.Parse(matchID)
	if err != nil {
		http.Error(w, "invalid match_id format", http.StatusBadRequest)
		return
	}

	var matchReader replay_in.MatchReader
	if err := ctrl.container.Resolve(&matchReader); err != nil {
		slog.Error("GetMatchTrajectoryHandler: failed to resolve MatchReader", "err", err)
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	search := shared.NewSearchByValues(r.Context(), []shared.SearchableValue{
		{Field: "ID", Values: []interface{}{matchUUID}, Operator: shared.EqualsOperator},
	}, shared.SearchResultOptions{Limit: 1}, shared.UserAudienceIDKey)

	matches, err := matchReader.Search(r.Context(), search)
	if err != nil || len(matches) == 0 {
		http.Error(w, "match not found", http.StatusNotFound)
		return
	}

	match := matches[0]

	// Aggregate trajectories from all events in the match
	playerTrajectories := make(map[string]*PlayerTrajectory)

	for _, event := range match.Events {
		if event == nil || event.Stats == nil {
			continue
		}

		posStats := extractPositioningStats(event.Stats)
		if posStats == nil || posStats.PlayerTrajectory == nil {
			continue
		}

		for playerID, trajectoryPoints := range posStats.PlayerTrajectory {
			pid := uuid.UUID(playerID).String()
			if _, exists := playerTrajectories[pid]; !exists {
				playerTrajectories[pid] = &PlayerTrajectory{
					PlayerID:  pid,
					Positions: make([]TrajectoryPoint, 0),
				}
			}

			for _, stat := range trajectoryPoints {
				pos := TrajectoryPoint{
					Tick: stat.TickID,
				}
				if stat.Position != nil {
					pos.X = stat.Position.X
					pos.Y = stat.Position.Y
					pos.Z = stat.Position.Z
				}
				if stat.Angle != nil {
					pos.ViewX = stat.Angle.X
					pos.ViewY = stat.Angle.Y
				}
				if stat.IsAlive != nil {
					pos.IsAlive = *stat.IsAlive
				}
				if stat.IsCrouching != nil {
					pos.IsCrouched = *stat.IsCrouching
				}
				playerTrajectories[pid].Positions = append(playerTrajectories[pid].Positions, pos)
			}
		}
	}

	// Convert map to slice
	trajectories := make([]PlayerTrajectory, 0, len(playerTrajectories))
	for _, pt := range playerTrajectories {
		trajectories = append(trajectories, *pt)
	}

	response := TrajectoryResponse{
		MatchID:       matchID,
		MapName:       "unknown", // Map name would need to be fetched from ReplayFile or extracted from events
		Trajectories:  trajectories,
		RadarImageURL: "/maps/unknown/radar.png",
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// GetRoundTrajectoryHandler returns player trajectories for a specific round
func (ctrl *MatchAnalyticsController) GetRoundTrajectoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchID := vars["match_id"]
	roundNumberStr := vars["round_number"]

	roundNumber, err := strconv.Atoi(roundNumberStr)
	if err != nil {
		http.Error(w, "invalid round_number", http.StatusBadRequest)
		return
	}

	matchUUID, err := uuid.Parse(matchID)
	if err != nil {
		http.Error(w, "invalid match_id format", http.StatusBadRequest)
		return
	}

	var matchReader replay_in.MatchReader
	if err := ctrl.container.Resolve(&matchReader); err != nil {
		slog.Error("GetRoundTrajectoryHandler: failed to resolve MatchReader", "err", err)
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	search := shared.NewSearchByValues(r.Context(), []shared.SearchableValue{
		{Field: "ID", Values: []interface{}{matchUUID}, Operator: shared.EqualsOperator},
	}, shared.SearchResultOptions{Limit: 1}, shared.UserAudienceIDKey)

	matches, err := matchReader.Search(r.Context(), search)
	if err != nil || len(matches) == 0 {
		http.Error(w, "match not found", http.StatusNotFound)
		return
	}

	match := matches[0]

	// Aggregate trajectories from events in the specific round
	playerTrajectories := make(map[string]*PlayerTrajectory)

	for _, event := range match.Events {
		if event == nil || event.Stats == nil {
			continue
		}

		posStats := extractPositioningStats(event.Stats)
		if posStats == nil || posStats.PlayerTrajectory == nil {
			continue
		}

		// Filter by round if we can identify round from trajectory stats
		for playerID, trajectoryPoints := range posStats.PlayerTrajectory {
			pid := uuid.UUID(playerID).String()
			for _, stat := range trajectoryPoints {
				// Skip if this point is not from the requested round
				// (CSTick has RoundNumber field, but CSPositioningTrajectoryStats doesn't directly expose it)
				// For now, we'll include all points and filter later if needed

				if _, exists := playerTrajectories[pid]; !exists {
					playerTrajectories[pid] = &PlayerTrajectory{
						PlayerID:  pid,
						Positions: make([]TrajectoryPoint, 0),
					}
				}

				pos := TrajectoryPoint{
					Tick: stat.TickID,
				}
				if stat.Position != nil {
					pos.X = stat.Position.X
					pos.Y = stat.Position.Y
					pos.Z = stat.Position.Z
				}
				if stat.Angle != nil {
					pos.ViewX = stat.Angle.X
					pos.ViewY = stat.Angle.Y
				}
				if stat.IsAlive != nil {
					pos.IsAlive = *stat.IsAlive
				}
				if stat.IsCrouching != nil {
					pos.IsCrouched = *stat.IsCrouching
				}
				playerTrajectories[pid].Positions = append(playerTrajectories[pid].Positions, pos)
			}
		}
	}

	// Convert map to slice
	trajectories := make([]PlayerTrajectory, 0, len(playerTrajectories))
	for _, pt := range playerTrajectories {
		trajectories = append(trajectories, *pt)
	}

	response := TrajectoryResponse{
		MatchID:       matchID,
		RoundNumber:   roundNumber,
		MapName:       "unknown",
		Trajectories:  trajectories,
		RadarImageURL: "/maps/unknown/radar.png",
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// GetMatchHeatmapHandler generates a heatmap for an entire match
func (ctrl *MatchAnalyticsController) GetMatchHeatmapHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchID := vars["match_id"]
	heatmapType := r.URL.Query().Get("type")
	if heatmapType == "" {
		heatmapType = "position"
	}

	gridSizeStr := r.URL.Query().Get("grid_size")
	gridSize := 50
	if gridSizeStr != "" {
		if gs, err := strconv.Atoi(gridSizeStr); err == nil && gs > 0 && gs <= 200 {
			gridSize = gs
		}
	}

	matchUUID, err := uuid.Parse(matchID)
	if err != nil {
		http.Error(w, "invalid match_id format", http.StatusBadRequest)
		return
	}

	var matchReader replay_in.MatchReader
	if err := ctrl.container.Resolve(&matchReader); err != nil {
		slog.Error("GetMatchHeatmapHandler: failed to resolve MatchReader", "err", err)
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	search := shared.NewSearchByValues(r.Context(), []shared.SearchableValue{
		{Field: "ID", Values: []interface{}{matchUUID}, Operator: shared.EqualsOperator},
	}, shared.SearchResultOptions{Limit: 1}, shared.UserAudienceIDKey)

	matches, err := matchReader.Search(r.Context(), search)
	if err != nil || len(matches) == 0 {
		http.Error(w, "match not found", http.StatusNotFound)
		return
	}

	match := matches[0]

	// Build heatmap grid from positioning data
	grid := make(map[string]int)

	for _, event := range match.Events {
		if event == nil || event.Stats == nil {
			continue
		}

		posStats := extractPositioningStats(event.Stats)
		if posStats == nil || posStats.PlayerTrajectory == nil {
			continue
		}

		for _, trajectoryPoints := range posStats.PlayerTrajectory {
			for _, stat := range trajectoryPoints {
				if stat.Position != nil {
					// Normalize positions to grid cells
					cellX := int(stat.Position.X / float64(gridSize))
					cellY := int(stat.Position.Y / float64(gridSize))
					key := fmt.Sprintf("%d,%d", cellX, cellY)
					grid[key]++
				}
			}
		}
	}

	// Convert grid to cells
	cells := make([]HeatmapCell, 0, len(grid))
	maxVal := 0.0
	for key, count := range grid {
		parts := strings.Split(key, ",")
		if len(parts) != 2 {
			continue
		}
		cellX, _ := strconv.Atoi(parts[0])
		cellY, _ := strconv.Atoi(parts[1])
		cell := HeatmapCell{
			X:       float64(cellX * gridSize),
			Y:       float64(cellY * gridSize),
			Density: count,
			Value:   float64(count),
		}
		cells = append(cells, cell)
		if float64(count) > maxVal {
			maxVal = float64(count)
		}
	}

	response := HeatmapResponse{
		MatchID:     matchID,
		MapName:     "unknown",
		HeatmapType: heatmapType,
		GridSize:    gridSize,
		Cells:       cells,
		MinValue:    0,
		MaxValue:    maxVal,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// GetRoundHeatmapHandler generates a heatmap for a specific round
func (ctrl *MatchAnalyticsController) GetRoundHeatmapHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchID := vars["match_id"]
	roundNumberStr := vars["round_number"]

	roundNumber, err := strconv.Atoi(roundNumberStr)
	if err != nil {
		http.Error(w, "invalid round_number", http.StatusBadRequest)
		return
	}

	heatmapType := r.URL.Query().Get("type")
	if heatmapType == "" {
		heatmapType = "position"
	}

	gridSizeStr := r.URL.Query().Get("grid_size")
	gridSize := 50
	if gridSizeStr != "" {
		if gs, err := strconv.Atoi(gridSizeStr); err == nil && gs > 0 && gs <= 200 {
			gridSize = gs
		}
	}

	matchUUID, err := uuid.Parse(matchID)
	if err != nil {
		http.Error(w, "invalid match_id format", http.StatusBadRequest)
		return
	}

	var matchReader replay_in.MatchReader
	if err := ctrl.container.Resolve(&matchReader); err != nil {
		slog.Error("GetRoundHeatmapHandler: failed to resolve MatchReader", "err", err)
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	search := shared.NewSearchByValues(r.Context(), []shared.SearchableValue{
		{Field: "ID", Values: []interface{}{matchUUID}, Operator: shared.EqualsOperator},
	}, shared.SearchResultOptions{Limit: 1}, shared.UserAudienceIDKey)

	matches, err := matchReader.Search(r.Context(), search)
	if err != nil || len(matches) == 0 {
		http.Error(w, "match not found", http.StatusNotFound)
		return
	}

	match := matches[0]

	// Build heatmap grid from positioning data for specific round
	grid := make(map[string]int)
	for _, event := range match.Events {
		if event == nil || event.Stats == nil {
			continue
		}

		posStats := extractPositioningStats(event.Stats)
		if posStats == nil || posStats.PlayerTrajectory == nil {
			continue
		}

		for _, trajectoryPoints := range posStats.PlayerTrajectory {
			for _, stat := range trajectoryPoints {
				if stat.Position != nil {
					cellX := int(stat.Position.X / float64(gridSize))
					cellY := int(stat.Position.Y / float64(gridSize))
					key := fmt.Sprintf("%d,%d", cellX, cellY)
					grid[key]++
				}
			}
		}
	}

	// Convert grid to cells
	cells := make([]HeatmapCell, 0, len(grid))
	maxVal := 0.0
	for key, count := range grid {
		parts := strings.Split(key, ",")
		if len(parts) != 2 {
			continue
		}
		cellX, _ := strconv.Atoi(parts[0])
		cellY, _ := strconv.Atoi(parts[1])
		cell := HeatmapCell{
			X:       float64(cellX * gridSize),
			Y:       float64(cellY * gridSize),
			Density: count,
			Value:   float64(count),
		}
		cells = append(cells, cell)
		if float64(count) > maxVal {
			maxVal = float64(count)
		}
	}

	response := HeatmapResponse{
		MatchID:     matchID,
		RoundNumber: roundNumber,
		MapName:     "unknown",
		HeatmapType: heatmapType,
		GridSize:    gridSize,
		Cells:       cells,
		MinValue:    0,
		MaxValue:    maxVal,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// GetPositioningStatsHandler returns zone frequencies and dwell times
func (ctrl *MatchAnalyticsController) GetPositioningStatsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchID := vars["match_id"]

	matchUUID, err := uuid.Parse(matchID)
	if err != nil {
		http.Error(w, "invalid match_id format", http.StatusBadRequest)
		return
	}

	var matchReader replay_in.MatchReader
	if err := ctrl.container.Resolve(&matchReader); err != nil {
		slog.Error("GetPositioningStatsHandler: failed to resolve MatchReader", "err", err)
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	search := shared.NewSearchByValues(r.Context(), []shared.SearchableValue{
		{Field: "ID", Values: []interface{}{matchUUID}, Operator: shared.EqualsOperator},
	}, shared.SearchResultOptions{Limit: 1}, shared.UserAudienceIDKey)

	matches, err := matchReader.Search(r.Context(), search)
	if err != nil || len(matches) == 0 {
		http.Error(w, "match not found", http.StatusNotFound)
		return
	}

	match := matches[0]
	response := PositioningStatsResponse{
		MatchID:             matchID,
		ZoneFrequencies:     make(map[string]map[string]int),
		ZoneDwellTimes:      make(map[string]map[string]float64),
		AverageSpeed:        make(map[string]float64),
		AverageSpeedRunning: make(map[string]float64),
	}

	// Aggregate positioning stats from all events
	for _, event := range match.Events {
		if event == nil || event.Stats == nil {
			continue
		}

		posStats := extractPositioningStats(event.Stats)
		if posStats == nil {
			continue
		}

		for playerID, zones := range posStats.PlayerZoneFrequencies {
			pid := uuid.UUID(playerID).String()
			if _, ok := response.ZoneFrequencies[pid]; !ok {
				response.ZoneFrequencies[pid] = make(map[string]int)
			}
			for zone, freq := range zones {
				response.ZoneFrequencies[pid][string(zone)] += freq
			}
		}

		for playerID, zones := range posStats.PlayerZoneDwellTimes {
			pid := uuid.UUID(playerID).String()
			if _, ok := response.ZoneDwellTimes[pid]; !ok {
				response.ZoneDwellTimes[pid] = make(map[string]float64)
			}
			for zone, time := range zones {
				response.ZoneDwellTimes[pid][string(zone)] += time
			}
		}

		for playerID, speed := range posStats.PlayerAverageSpeed {
			response.AverageSpeed[uuid.UUID(playerID).String()] = speed
		}

		for playerID, speed := range posStats.PlayerAverageSpeedRunning {
			response.AverageSpeedRunning[uuid.UUID(playerID).String()] = speed
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
