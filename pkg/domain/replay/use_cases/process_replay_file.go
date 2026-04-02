package use_cases

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	cs_entity "github.com/replay-api/replay-api/pkg/domain/cs/entities"
	e "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_out "github.com/replay-api/replay-api/pkg/domain/replay/ports/out"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	fps_events "github.com/replay-api/replay-common/pkg/replay/events/game/fps"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

const CHUNK_SIZE = 10

// Re-export types from entities for use in this file
type FinalScoreboardPayload = e.FinalScoreboardPayload
type PlayerScoreboardData = e.PlayerScoreboardDataEntry

// ScoreSubmissionCallback defines the callback interface for submitting scores
// from replay processing. This avoids a direct dependency on the scores domain.
type ScoreSubmissionCallback func(ctx context.Context, match *e.Match, replayFileID uuid.UUID) error

type ProcessReplayFileUseCase struct {
	ReplayMetadataReader replay_out.ReplayFileMetadataReader
	ReplayContentReader  replay_out.ReplayFileContentReader
	ReplayMetadataWriter replay_out.ReplayFileMetadataWriter
	ReplayContentWriter  replay_out.ReplayFileContentWriter

	PlayerMetadataWriter replay_out.PlayerMetadataWriter
	MatchMetadataWriter  replay_out.MatchMetadataWriter

	Parser      replay_out.ReplayParser
	EventWriter replay_out.GameEventWriter

	// Optional: callback to submit scores to the scores domain after processing
	OnMatchProcessed ScoreSubmissionCallback
}

func NewProcessReplayFileUseCase(metadataReader replay_out.ReplayFileMetadataReader, contentReader replay_out.ReplayFileContentReader, metadataWriter replay_out.ReplayFileMetadataWriter, contentWriter replay_out.ReplayFileContentWriter, parser replay_out.ReplayParser, eventWriter replay_out.GameEventWriter, playerMetadataWriter replay_out.PlayerMetadataWriter, matchMetadataWriter replay_out.MatchMetadataWriter) *ProcessReplayFileUseCase {
	return &ProcessReplayFileUseCase{
		ReplayMetadataReader: metadataReader,
		ReplayContentReader:  contentReader,
		ReplayMetadataWriter: metadataWriter,
		ReplayContentWriter:  contentWriter,

		PlayerMetadataWriter: playerMetadataWriter,
		MatchMetadataWriter:  matchMetadataWriter,

		Parser:      parser,
		EventWriter: eventWriter,
	}
}

func (usecase *ProcessReplayFileUseCase) Exec(ctx context.Context, replayFileID uuid.UUID) (*e.Match, error) {
	replayFile, err := usecase.ReplayMetadataReader.GetByID(ctx, replayFileID)
	if err != nil {
		slog.ErrorContext(ctx, "error getting replay metadata", "replayFileID", replayFileID, "err", err)
		return nil, err
	}

	// Update Metadata Status
	replayFile.Status = e.ReplayFileStatusProcessing
	replayFile, err = usecase.ReplayMetadataWriter.Update(ctx, replayFile)

	if err != nil {
		slog.ErrorContext(ctx, "error updating uploaded replay metadata", "replayFile", replayFile, "err", err)
		return nil, err
	}

	slog.InfoContext(ctx, "processing replay file", "replayFile", replayFile)

	match := e.NewCS2MatchWithOwner(replayFile.ResourceOwner, replayFile.ID)

	// For reference replays (duplicates uploaded by different users), use the original replay ID
	// to look up blob content, since the blob is stored under the original replay's ID
	contentID := replayFileID
	if replayFile.OriginalReplayID != nil && *replayFile.OriginalReplayID != uuid.Nil {
		contentID = *replayFile.OriginalReplayID
		slog.InfoContext(ctx, "using original replay ID for content retrieval", "replayFileID", replayFileID, "originalReplayID", contentID)
	}

	file, err := usecase.ReplayContentReader.GetByID(ctx, contentID)
	if err != nil {
		slog.ErrorContext(ctx, "error getting replay file content data", "err", err)
		return nil, err
	}
	defer file.Close()

	slog.InfoContext(ctx, "parsing replay file", "Size", replayFile.Size, "replayFileID", replayFileID)

	eventsChan := make(chan *e.GameEvent, 1000)

	entitiesMap := make(map[shared.ResourceType][]interface{})

	// Batch event buffer — flush every CHUNK_SIZE events to avoid unbounded memory
	const eventBatchSize = 500
	var eventBatch []*e.GameEvent
	var totalEventsWritten int
	
	// Store final scoreboard data when captured
	var finalScoreboard *FinalScoreboardPayload

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for event := range eventsChan {
			slog.InfoContext(ctx, "event", "event.Type", event.Type)
			match.EventCount++

			// Accumulate event for batch write
			eventBatch = append(eventBatch, event)

			// Flush batch when full
			if len(eventBatch) >= eventBatchSize {
				if flushErr := usecase.EventWriter.CreateMany(ctx, eventBatch); flushErr != nil {
					slog.ErrorContext(ctx, "error flushing event batch", "err", flushErr, "batchSize", len(eventBatch))
				} else {
					totalEventsWritten += len(eventBatch)
				}
				eventBatch = make([]*e.GameEvent, 0, eventBatchSize)
			}

			// Extract match details from MatchStart event
			if event.Type == fps_events.Event_MatchStartID {
				if matchStats, ok := event.Payload.(*cs_entity.CSMatchStats); ok && matchStats != nil && matchStats.Header != nil {
					match.MapName = matchStats.Header.MapName
					match.Duration = matchStats.Header.Length.Seconds()
					match.ServerName = matchStats.Header.ServerName
					match.Status = e.MatchStatusCompleted
					match.Mode = "competitive" // Default mode
					slog.InfoContext(ctx, "Extracted match details from MatchStart",
						"map_name", match.MapName,
						"duration", match.Duration,
						"server_name", match.ServerName)
				}
			}

			// Extract final scoreboard from MatchEnd event
			if event.Type == fps_events.Event_MatchEndID {
				// Try to cast to the actual struct type first
				if payload, ok := event.Payload.(FinalScoreboardPayload); ok {
					finalScoreboard = &payload
					slog.InfoContext(ctx, "Extracted final scoreboard from MatchEnd event (struct)",
						"ct_score", finalScoreboard.CTScore,
						"t_score", finalScoreboard.TScore,
						"player_count", len(finalScoreboard.Players))
				} else if payloadPtr, ok := event.Payload.(*FinalScoreboardPayload); ok && payloadPtr != nil {
					finalScoreboard = payloadPtr
					slog.InfoContext(ctx, "Extracted final scoreboard from MatchEnd event (struct ptr)",
						"ct_score", finalScoreboard.CTScore,
						"t_score", finalScoreboard.TScore,
						"player_count", len(finalScoreboard.Players))
				} else if payload, ok := event.Payload.(map[string]interface{}); ok {
					// Fallback for map-based payloads (e.g., from JSON deserialization)
					finalScoreboard = extractFinalScoreboard(payload)
					slog.InfoContext(ctx, "Extracted final scoreboard from MatchEnd event (map)",
						"ct_score", finalScoreboard.CTScore,
						"t_score", finalScoreboard.TScore,
						"player_count", len(finalScoreboard.Players))
				} else {
					slog.WarnContext(ctx, "Could not extract final scoreboard from MatchEnd event",
						"payload_type", fmt.Sprintf("%T", event.Payload))
				}
			}

			for k, v := range event.Entities {
				if entitiesMap[k] == nil {
					entitiesMap[k] = make([]interface{}, 0)
				}

				entitiesMap[k] = append(entitiesMap[k], v...)
			}
		}

		// Flush remaining events
		if len(eventBatch) > 0 {
			if flushErr := usecase.EventWriter.CreateMany(ctx, eventBatch); flushErr != nil {
				slog.ErrorContext(ctx, "error flushing final event batch", "err", flushErr, "batchSize", len(eventBatch))
			} else {
				totalEventsWritten += len(eventBatch)
			}
			eventBatch = nil
		}
		slog.InfoContext(ctx, "total events written to DB", "totalEventsWritten", totalEventsWritten)
	}()

	err = usecase.Parser.Parse(ctx, match.ID, file, eventsChan)

	if err != nil {
		slog.ErrorContext(ctx, "error parsing replay events", "err", err)
		return nil, err
	}

	wg.Wait()

	// Build scoreboard from final scoreboard data if available
	if finalScoreboard != nil {
		match.Scoreboard = buildMatchScoreboard(finalScoreboard, replayFile.ResourceOwner)
		match.Teams = buildTeams(finalScoreboard, replayFile.ResourceOwner)
		slog.InfoContext(ctx, "Built match scoreboard",
			"team_count", len(match.Teams),
			"scoreboard_teams", len(match.Scoreboard.TeamScoreboards))
	}

	// Calculate duration from round durations if not set from header
	if match.Duration == 0 && len(match.Scoreboard.TeamScoreboards) > 0 {
		// Estimate duration based on number of rounds (average ~100 seconds per round)
		totalRounds := 0
		for _, ts := range match.Scoreboard.TeamScoreboards {
			totalRounds += ts.TeamScore
		}
		if totalRounds > 0 {
			// Average CS2 round is about 100 seconds including buy time
			match.Duration = float64(totalRounds) * 100.0
			slog.InfoContext(ctx, "Calculated match duration from rounds",
				"total_rounds", totalRounds,
				"estimated_duration", match.Duration)
		}
	}

	// Set PlayedAt if not set (use replay file creation time as approximation)
	if match.PlayedAt.IsZero() && !replayFile.CreatedAt.IsZero() {
		match.PlayedAt = replayFile.CreatedAt
		slog.InfoContext(ctx, "Set PlayedAt from replay file creation time", "played_at", match.PlayedAt)
	}

	// Add the main match to entitiesMap so it gets saved
	if entitiesMap[replay_common.ResourceTypeMatch] == nil {
		entitiesMap[replay_common.ResourceTypeMatch] = make([]interface{}, 0)
	}
	entitiesMap[replay_common.ResourceTypeMatch] = append(entitiesMap[replay_common.ResourceTypeMatch], match)

	playersMap := make(map[string]*e.PlayerMetadata)

	for _, p := range entitiesMap[replay_common.ResourceTypePlayerMetadata] {
		player := p.(*e.PlayerMetadata)
		playersMap[player.NetworkUserID] = player
	}

	for _, p := range playersMap {
		slog.InfoContext(ctx, "PlayerMetadata", "entity", p)
		err = usecase.PlayerMetadataWriter.Create(ctx, *p)

		if err != nil {
			slog.ErrorContext(ctx, "error writing PlayerMetadata entities", "err", err)
			return nil, err
		}
	}

	for resourceKey, entities := range entitiesMap {
		switch resourceKey {
		case replay_common.ResourceTypeMatch:
			for _, rawEntity := range entities {
				slog.InfoContext(ctx, "MatchMetadata", "entity", rawEntity)
				entity := rawEntity.(*e.Match)

				err = usecase.MatchMetadataWriter.Create(ctx, *entity)

				if err != nil {
					slog.ErrorContext(ctx, "error writing MatchMetadata entities", "err", err)
					return nil, err
				}
			}
		}
	}

	// Save the main match directly
	err = usecase.MatchMetadataWriter.Create(ctx, *match)
	if err != nil {
		slog.ErrorContext(ctx, "error saving main match", "err", err)
		return nil, err
	}

	// Events are already written in batches during parsing (streaming writes)

	// Update Metadata Status
	replayFile.Status = e.ReplayFileStatusCompleted
	replayFile, err = usecase.ReplayMetadataWriter.Update(ctx, replayFile)

	if err != nil {
		slog.ErrorContext(ctx, "error updating uploaded replay metadata status to Completed", "replayFile", replayFile, "err", err)
		return nil, err
	}

	slog.InfoContext(ctx, "Replay file processed", "ReplayFileID", replayFileID)

	// Bridge to scores domain: submit match result from replay data
	if usecase.OnMatchProcessed != nil {
		if err := usecase.OnMatchProcessed(ctx, match, replayFileID); err != nil {
			slog.WarnContext(ctx, "failed to submit replay scores (non-fatal)",
				slog.String("match_id", match.ID.String()),
				slog.String("replay_file_id", replayFileID.String()),
				slog.String("error", err.Error()),
			)
			// Non-fatal: match is already saved, score submission can be retried
		} else {
			slog.InfoContext(ctx, "Replay scores submitted to scores domain",
				slog.String("match_id", match.ID.String()),
			)
		}
	}

	return match, nil
}

// extractFinalScoreboard extracts FinalScoreboardPayload from event payload map
func extractFinalScoreboard(payload map[string]interface{}) *FinalScoreboardPayload {
	scoreboard := &FinalScoreboardPayload{}

	if v, ok := payload["ct_score"].(int); ok {
		scoreboard.CTScore = v
	} else if v, ok := payload["ct_score"].(float64); ok {
		scoreboard.CTScore = int(v)
	}

	if v, ok := payload["t_score"].(int); ok {
		scoreboard.TScore = v
	} else if v, ok := payload["t_score"].(float64); ok {
		scoreboard.TScore = int(v)
	}

	if v, ok := payload["ct_team_name"].(string); ok {
		scoreboard.CTTeamName = v
	}

	if v, ok := payload["t_team_name"].(string); ok {
		scoreboard.TTeamName = v
	}

	if v, ok := payload["total_rounds"].(int); ok {
		scoreboard.TotalRounds = v
	} else if v, ok := payload["total_rounds"].(float64); ok {
		scoreboard.TotalRounds = int(v)
	}

	if v, ok := payload["winner_side"].(string); ok {
		scoreboard.WinnerSide = v
	}

	if v, ok := payload["duration"].(float64); ok {
		scoreboard.Duration = v
	}

	if v, ok := payload["map_name"].(string); ok {
		scoreboard.MapName = v
	}

	// Extract players
	if playersData, ok := payload["players"].([]interface{}); ok {
		for _, p := range playersData {
			if playerMap, ok := p.(map[string]interface{}); ok {
				player := extractPlayerData(playerMap)
				scoreboard.Players = append(scoreboard.Players, player)
			}
		}
	}

	return scoreboard
}

// extractPlayerData extracts PlayerScoreboardData from map
func extractPlayerData(data map[string]interface{}) PlayerScoreboardData {
	player := PlayerScoreboardData{}

	if v, ok := data["network_user_id"].(string); ok {
		player.NetworkUserID = v
	}
	if v, ok := data["name"].(string); ok {
		player.Name = v
	}
	if v, ok := data["team"].(string); ok {
		player.Team = v
	}
	if v, ok := data["side"].(string); ok {
		player.Side = v
	}
	if v, ok := data["kills"].(int); ok {
		player.Kills = v
	} else if v, ok := data["kills"].(float64); ok {
		player.Kills = int(v)
	}
	if v, ok := data["deaths"].(int); ok {
		player.Deaths = v
	} else if v, ok := data["deaths"].(float64); ok {
		player.Deaths = int(v)
	}
	if v, ok := data["assists"].(int); ok {
		player.Assists = v
	} else if v, ok := data["assists"].(float64); ok {
		player.Assists = int(v)
	}
	if v, ok := data["kd_ratio"].(float64); ok {
		player.KDRatio = v
	}
	if v, ok := data["headshots"].(int); ok {
		player.Headshots = v
	} else if v, ok := data["headshots"].(float64); ok {
		player.Headshots = int(v)
	}
	if v, ok := data["total_damage"].(int); ok {
		player.TotalDamage = v
	} else if v, ok := data["total_damage"].(float64); ok {
		player.TotalDamage = int(v)
	}
	if v, ok := data["utility_damage"].(int); ok {
		player.UtilityDamage = v
	} else if v, ok := data["utility_damage"].(float64); ok {
		player.UtilityDamage = int(v)
	}
	if v, ok := data["adr"].(float64); ok {
		player.ADR = v
	}
	if v, ok := data["mvp_count"].(int); ok {
		player.MVPCount = v
	} else if v, ok := data["mvp_count"].(float64); ok {
		player.MVPCount = int(v)
	}
	if v, ok := data["score"].(int); ok {
		player.Score = v
	} else if v, ok := data["score"].(float64); ok {
		player.Score = int(v)
	}

	return player
}

// buildMatchScoreboard builds the Match.Scoreboard from final scoreboard data
func buildMatchScoreboard(scoreboard *FinalScoreboardPayload, resourceOwner shared.ResourceOwner) e.Scoreboard {
	ctPlayers := make([]e.PlayerMetadata, 0)
	tPlayers := make([]e.PlayerMetadata, 0)
	ctPlayerStats := make([]e.PlayerStatsEntry, 0)
	tPlayerStats := make([]e.PlayerStatsEntry, 0)

	var matchMVP *e.PlayerMetadata
	var ctMVP *e.PlayerMetadata
	var tMVP *e.PlayerMetadata
	maxMVPs := 0
	maxCTKills := 0
	maxTKills := 0

	for _, p := range scoreboard.Players {
		playerMeta := e.NewPlayerMetadata(p.Name, p.NetworkUserID, replay_common.SteamNetworkIDKey, p.Team, resourceOwner)
		playerID := playerMeta.GetID().String()

		// Calculate headshot percentage
		headshotPct := 0.0
		if p.Kills > 0 {
			headshotPct = float64(p.Headshots) / float64(p.Kills) * 100.0
		}

		// Calculate HLTV 2.0 rating approximation
		// Rating 2.0 = 0.0073*KAST + 0.3591*KPR − 0.5329*DPR + 0.2372*Impact + 0.0032*ADR + 0.1587
		totalRounds := float64(scoreboard.TotalRounds)
		if totalRounds == 0 {
			totalRounds = float64(scoreboard.CTScore + scoreboard.TScore)
		}
		if totalRounds == 0 {
			totalRounds = 1 // Avoid division by zero
		}
		kpr := float64(p.Kills) / totalRounds
		dpr := float64(p.Deaths) / totalRounds
		// Approximate KAST (simplified: survived rounds + assist rounds + trade rounds)
		estimatedKAST := 70.0 + (float64(p.Assists)*2.0) + (float64(p.TradeKills)*3.0)
		if estimatedKAST > 100 {
			estimatedKAST = 100
		}
		// Approximate impact rating
		impactRating := (2.13*kpr + 0.42*float64(p.Assists)/totalRounds - 0.41)
		if impactRating < 0 {
			impactRating = 0
		}
		// Calculate rating 2.0
		rating2 := 0.0073*estimatedKAST + 0.3591*kpr - 0.5329*dpr + 0.2372*impactRating + 0.0032*p.ADR + 0.1587
		if rating2 < 0 {
			rating2 = 0
		}

		stats := e.PlayerStatsEntry{
			PlayerID:       playerID,
			Kills:          p.Kills,
			Deaths:         p.Deaths,
			Assists:        p.Assists,
			KDRatio:        p.KDRatio,
			Headshots:      p.Headshots,
			HeadshotPct:    headshotPct,
			TotalDamage:    p.TotalDamage,
			UtilityDamage:  p.UtilityDamage,
			ADR:            p.ADR,
			MVPCount:       p.MVPCount,
			Score:          p.Score,
			// Advanced esports stats
			KAST:           estimatedKAST,
			ImpactRating:   impactRating,
			OpeningKills:   p.FirstKills,
			OpeningDeaths:  p.FirstDeaths,
			TradeKills:     p.TradeKills,
			FlashAssists:   p.FlashAssists,
			EnemiesFlashed: p.EnemiesFlashed,
			Rating2:        rating2,
		}

		if p.Side == "CT" {
			ctPlayers = append(ctPlayers, *playerMeta)
			ctPlayerStats = append(ctPlayerStats, stats)
			if p.Kills > maxCTKills {
				maxCTKills = p.Kills
				ctMVP = playerMeta
			}
		} else {
			tPlayers = append(tPlayers, *playerMeta)
			tPlayerStats = append(tPlayerStats, stats)
			if p.Kills > maxTKills {
				maxTKills = p.Kills
				tMVP = playerMeta
			}
		}

		if p.MVPCount > maxMVPs {
			maxMVPs = p.MVPCount
			matchMVP = playerMeta
		}
	}

	ctTeam := e.Team{
		Name:      scoreboard.CTTeamName,
		ShortName: "CT",
		Players:   ctPlayers,
	}

	tTeam := e.Team{
		Name:      scoreboard.TTeamName,
		ShortName: "T",
		Players:   tPlayers,
	}

	teamScoreboards := []e.TeamScoreboard{
		{
			Team:        ctTeam,
			Side:        "CT",
			TeamScore:   scoreboard.CTScore,
			TeamMVP:     ctMVP,
			Players:     ctPlayers,
			PlayerStats: ctPlayerStats,
		},
		{
			Team:        tTeam,
			Side:        "T",
			TeamScore:   scoreboard.TScore,
			TeamMVP:     tMVP,
			Players:     tPlayers,
			PlayerStats: tPlayerStats,
		},
	}

	return e.Scoreboard{
		TeamScoreboards: teamScoreboards,
		MatchMVP:        matchMVP,
	}
}

// buildTeams builds the Match.Teams from final scoreboard data
func buildTeams(scoreboard *FinalScoreboardPayload, resourceOwner shared.ResourceOwner) []e.Team {
	ctPlayers := make([]e.PlayerMetadata, 0)
	tPlayers := make([]e.PlayerMetadata, 0)

	for _, p := range scoreboard.Players {
		playerMeta := e.NewPlayerMetadata(p.Name, p.NetworkUserID, replay_common.SteamNetworkIDKey, p.Team, resourceOwner)

		if p.Side == "CT" {
			ctPlayers = append(ctPlayers, *playerMeta)
		} else {
			tPlayers = append(tPlayers, *playerMeta)
		}
	}

	ctEntity := shared.NewUnrestrictedEntity(resourceOwner)
	tEntity := shared.NewUnrestrictedEntity(resourceOwner)

	return []e.Team{
		{
			BaseEntity: ctEntity,
			Name:       scoreboard.CTTeamName,
			ShortName:  "CT",
			Players:    ctPlayers,
		},
		{
			BaseEntity: tEntity,
			Name:       scoreboard.TTeamName,
			ShortName:  "T",
			Players:    tPlayers,
		},
	}
}
