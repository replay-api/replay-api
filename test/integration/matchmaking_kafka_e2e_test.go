//go:build integration || e2e
// +build integration e2e

package integration_test

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	kafkaclient "github.com/replay-api/replay-api/pkg/infra/kafka"
)

// getKafkaTestURI returns Kafka connection URI for testing
func getKafkaTestURI() string {
	if uri := os.Getenv("KAFKA_BOOTSTRAP_SERVERS"); uri != "" {
		return uri
	}
	return "localhost:9092" // Default for local testing with Kind
}

// TestKafkaClient tests the Kafka client connectivity and basic operations
func TestKafkaClient_BasicConnectivity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Kafka integration test in short mode")
	}

	config := &kafkaclient.Config{
		BootstrapServers: getKafkaTestURI(),
		SecurityProtocol: "PLAINTEXT",
		Region:           "test",
	}

	client, err := kafkaclient.NewClient(config)
	if err != nil {
		t.Skipf("Skipping Kafka test - unable to create client: %v", err)
	}
	defer client.Close()

	// Test connection by attempting to list topics
	brokers := client.Brokers()
	assert.NotEmpty(t, brokers, "Should have at least one broker")
	t.Logf("Connected to Kafka brokers: %v", brokers)
}

// TestMatchmaking_QueueEvents_Lifecycle tests the complete queue event lifecycle
func TestMatchmaking_QueueEvents_Lifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Kafka integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	config := &kafkaclient.Config{
		BootstrapServers: getKafkaTestURI(),
		SecurityProtocol: "PLAINTEXT",
		Region:           "test",
	}

	client, err := kafkaclient.NewClient(config)
	if err != nil {
		t.Skipf("Skipping Kafka test - unable to create client: %v", err)
	}
	defer client.Close()

	publisher := kafkaclient.NewEventPublisher(client)
	playerID := uuid.New()

	// Create consumer for verification
	receivedEvents := make(chan *kafkaclient.QueueEvent, 10)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		consumeQueueEvents(ctx, t, getKafkaTestURI(), playerID.String(), receivedEvents)
	}()

	// Allow consumer to start
	time.Sleep(2 * time.Second)

	// Test 1: Publish QUEUE_JOINED event
	t.Run("PublishQueueJoined", func(t *testing.T) {
		event := &kafkaclient.QueueEvent{
			PlayerID:  playerID,
			GameType:  "cs2",
			Region:    "us-east",
			MMR:       1500,
			EventType: kafkaclient.EventTypeQueueJoined,
			Metadata: map[string]string{
				"tier":      "premium",
				"squad_id":  uuid.New().String(),
				"game_mode": "competitive",
			},
		}

		err := publisher.PublishQueueEvent(ctx, event)
		require.NoError(t, err, "Should publish queue joined event")
		t.Log("Published QUEUE_JOINED event")
	})

	// Test 2: Publish SEARCHING event
	t.Run("PublishSearching", func(t *testing.T) {
		event := &kafkaclient.QueueEvent{
			PlayerID:  playerID,
			GameType:  "cs2",
			Region:    "us-east",
			MMR:       1500,
			EventType: kafkaclient.EventTypeSearching,
			Metadata: map[string]string{
				"queue_time_seconds": "30",
				"mmr_range":          "200",
			},
		}

		err := publisher.PublishQueueEvent(ctx, event)
		require.NoError(t, err, "Should publish searching event")
		t.Log("Published SEARCHING event")
	})

	// Test 3: Publish QUEUE_LEFT event
	t.Run("PublishQueueLeft", func(t *testing.T) {
		event := &kafkaclient.QueueEvent{
			PlayerID:  playerID,
			GameType:  "cs2",
			Region:    "us-east",
			MMR:       1500,
			EventType: kafkaclient.EventTypeQueueLeft,
			Metadata: map[string]string{
				"reason": "user_cancelled",
			},
		}

		err := publisher.PublishQueueEvent(ctx, event)
		require.NoError(t, err, "Should publish queue left event")
		t.Log("Published QUEUE_LEFT event")
	})

	// Allow time for events to be consumed
	time.Sleep(3 * time.Second)
	cancel()
	wg.Wait()

	close(receivedEvents)
	eventsReceived := 0
	for range receivedEvents {
		eventsReceived++
	}
	t.Logf("Total events received: %d", eventsReceived)
}

// TestMatchmaking_LobbyEvents_Lifecycle tests lobby event publishing
func TestMatchmaking_LobbyEvents_Lifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Kafka integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	config := &kafkaclient.Config{
		BootstrapServers: getKafkaTestURI(),
		SecurityProtocol: "PLAINTEXT",
		Region:           "test",
	}

	client, err := kafkaclient.NewClient(config)
	if err != nil {
		t.Skipf("Skipping Kafka test - unable to create client: %v", err)
	}
	defer client.Close()

	publisher := kafkaclient.NewEventPublisher(client)
	lobbyID := uuid.New()
	playerIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()}

	// Test 1: Publish LOBBY_CREATED event
	t.Run("PublishLobbyCreated", func(t *testing.T) {
		event := &kafkaclient.LobbyEvent{
			LobbyID:   lobbyID,
			EventType: kafkaclient.EventTypeLobbyCreated,
			PlayerIDs: playerIDs,
			GameType:  "cs2",
			Region:    "us-east",
			AvgMMR:    1650,
			Metadata: map[string]string{
				"game_mode": "competitive",
				"tier":      "premium",
			},
		}

		err := publisher.PublishLobbyEvent(ctx, event)
		require.NoError(t, err, "Should publish lobby created event")
		t.Log("Published LOBBY_CREATED event")
	})

	// Test 2: Publish PLAYER_JOINED events for each player
	t.Run("PublishPlayerJoined", func(t *testing.T) {
		for i, playerID := range playerIDs {
			event := &kafkaclient.LobbyEvent{
				LobbyID:   lobbyID,
				EventType: kafkaclient.EventTypePlayerJoined,
				PlayerIDs: []uuid.UUID{playerID},
				GameType:  "cs2",
				Region:    "us-east",
				AvgMMR:    1650,
				Metadata: map[string]string{
					"player_mmr":      "1650",
					"player_position": string(rune('0' + i)),
				},
			}

			err := publisher.PublishLobbyEvent(ctx, event)
			require.NoError(t, err, "Should publish player joined event")
		}
		t.Logf("Published PLAYER_JOINED events for %d players", len(playerIDs))
	})

	// Test 3: Publish READY_STATUS_CHANGED events
	t.Run("PublishReadyStatusChanged", func(t *testing.T) {
		for _, playerID := range playerIDs {
			event := &kafkaclient.LobbyEvent{
				LobbyID:   lobbyID,
				EventType: kafkaclient.EventTypeReadyStatusChanged,
				PlayerIDs: []uuid.UUID{playerID},
				GameType:  "cs2",
				Region:    "us-east",
				Metadata: map[string]string{
					"ready": "true",
				},
			}

			err := publisher.PublishLobbyEvent(ctx, event)
			require.NoError(t, err, "Should publish ready status changed event")
		}
		t.Logf("Published READY_STATUS_CHANGED events for %d players", len(playerIDs))
	})

	// Test 4: Publish LOBBY_READY event
	t.Run("PublishLobbyReady", func(t *testing.T) {
		event := &kafkaclient.LobbyEvent{
			LobbyID:   lobbyID,
			EventType: kafkaclient.EventTypeLobbyReady,
			PlayerIDs: playerIDs,
			GameType:  "cs2",
			Region:    "us-east",
			AvgMMR:    1650,
			Metadata: map[string]string{
				"all_players_ready": "true",
				"lobby_size":        "5",
			},
		}

		err := publisher.PublishLobbyEvent(ctx, event)
		require.NoError(t, err, "Should publish lobby ready event")
		t.Log("Published LOBBY_READY event")
	})

	t.Log("Lobby lifecycle events published successfully")
}

// TestMatchmaking_PrizePoolEvents tests prize pool event publishing
func TestMatchmaking_PrizePoolEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Kafka integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	config := &kafkaclient.Config{
		BootstrapServers: getKafkaTestURI(),
		SecurityProtocol: "PLAINTEXT",
		Region:           "test",
	}

	client, err := kafkaclient.NewClient(config)
	if err != nil {
		t.Skipf("Skipping Kafka test - unable to create client: %v", err)
	}
	defer client.Close()

	publisher := kafkaclient.NewEventPublisher(client)
	poolID := uuid.New()
	lobbyID := uuid.New()

	// Test 1: Initial prize pool creation
	t.Run("PublishPrizePoolCreated", func(t *testing.T) {
		event := &kafkaclient.PrizePoolEvent{
			PoolID:      poolID,
			LobbyID:     lobbyID,
			EventType:   kafkaclient.EventTypePrizePoolUpdated,
			TotalAmount: 500, // $5.00 platform contribution
			Currency:    "usd",
			Metadata: map[string]string{
				"source":     "platform_contribution",
				"match_tier": "premium",
			},
		}

		err := publisher.PublishPrizePoolEvent(ctx, event)
		require.NoError(t, err, "Should publish prize pool created event")
		t.Log("Published initial PRIZE_POOL_UPDATED event")
	})

	// Test 2: Player contribution to prize pool
	t.Run("PublishPlayerContribution", func(t *testing.T) {
		contributorID := uuid.New()
		event := &kafkaclient.PrizePoolEvent{
			PoolID:          poolID,
			LobbyID:         lobbyID,
			EventType:       kafkaclient.EventTypePrizePoolUpdated,
			TotalAmount:     1500, // $15.00 total
			Currency:        "usd",
			ContributorID:   &contributorID,
			ContributionAmt: 1000, // $10.00 contribution
			Metadata: map[string]string{
				"source":         "player_entry_fee",
				"player_tier":    "premium",
				"previous_total": "500",
			},
		}

		err := publisher.PublishPrizePoolEvent(ctx, event)
		require.NoError(t, err, "Should publish player contribution event")
		t.Log("Published player contribution PRIZE_POOL_UPDATED event")
	})

	// Test 3: Multiple player contributions (5v5 match = 10 players)
	t.Run("PublishMultipleContributions", func(t *testing.T) {
		totalAmount := int64(1500) // Starting from previous
		for i := 0; i < 9; i++ {   // 9 more players
			contributorID := uuid.New()
			totalAmount += 100 // $1.00 each

			event := &kafkaclient.PrizePoolEvent{
				PoolID:          poolID,
				LobbyID:         lobbyID,
				EventType:       kafkaclient.EventTypePrizePoolUpdated,
				TotalAmount:     totalAmount,
				Currency:        "usd",
				ContributorID:   &contributorID,
				ContributionAmt: 100,
				Metadata: map[string]string{
					"source":      "player_entry_fee",
					"player_tier": "free",
				},
			}

			err := publisher.PublishPrizePoolEvent(ctx, event)
			require.NoError(t, err, "Should publish contribution event")
		}
		t.Logf("Published 9 additional contributions, final total: $%.2f", float64(totalAmount)/100)
	})

	t.Log("Prize pool events published successfully")
}

// TestMatchmaking_MatchEvents_Lifecycle tests match event publishing
func TestMatchmaking_MatchEvents_Lifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Kafka integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	config := &kafkaclient.Config{
		BootstrapServers: getKafkaTestURI(),
		SecurityProtocol: "PLAINTEXT",
		Region:           "test",
	}

	client, err := kafkaclient.NewClient(config)
	if err != nil {
		t.Skipf("Skipping Kafka test - unable to create client: %v", err)
	}
	defer client.Close()

	publisher := kafkaclient.NewEventPublisher(client)
	matchID := uuid.New()
	lobbyID := uuid.New()
	team1ID := uuid.New()
	team2ID := uuid.New()

	// Create player IDs for both teams
	team1Players := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()}
	team2Players := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()}
	allPlayers := append(team1Players, team2Players...)

	// Test 1: Publish MATCH_CREATED event
	t.Run("PublishMatchCreated", func(t *testing.T) {
		event := &kafkaclient.MatchEvent{
			MatchID:   matchID,
			LobbyID:   lobbyID,
			EventType: kafkaclient.EventTypeMatchCreated,
			GameType:  "cs2",
			Region:    "us-east",
			PlayerIDs: allPlayers,
			Teams: []kafkaclient.TeamInfo{
				{
					TeamID:    team1ID,
					Name:      "Team Alpha",
					PlayerIDs: team1Players,
					Side:      "CT",
				},
				{
					TeamID:    team2ID,
					Name:      "Team Bravo",
					PlayerIDs: team2Players,
					Side:      "T",
				},
			},
			Metadata: map[string]string{
				"game_mode": "competitive",
				"tier":      "premium",
				"map":       "de_dust2",
			},
		}

		err := publisher.PublishMatchCreated(ctx, event)
		require.NoError(t, err, "Should publish match created event")
		t.Log("Published MATCH_CREATED event")
	})

	// Test 2: Publish MATCH_STARTED event
	t.Run("PublishMatchStarted", func(t *testing.T) {
		event := &kafkaclient.MatchEvent{
			MatchID:   matchID,
			LobbyID:   lobbyID,
			EventType: kafkaclient.EventTypeMatchStarted,
			GameType:  "cs2",
			Region:    "us-east",
			PlayerIDs: allPlayers,
			Metadata: map[string]string{
				"server_ip":   "192.168.1.100",
				"server_port": "27015",
			},
		}

		err := publisher.PublishMatchResult(ctx, event)
		require.NoError(t, err, "Should publish match started event")
		t.Log("Published MATCH_STARTED event")
	})

	// Test 3: Publish MATCH_COMPLETED event with results
	t.Run("PublishMatchCompleted", func(t *testing.T) {
		// Generate player stats
		var playerStats []kafkaclient.PlayerMatchStat
		for i, playerID := range team1Players {
			playerStats = append(playerStats, kafkaclient.PlayerMatchStat{
				PlayerID:  playerID,
				Kills:     15 + i*2,
				Deaths:    10 - i,
				Assists:   5 + i,
				Score:     80 + i*10,
				MMRChange: 25,
			})
		}
		for i, playerID := range team2Players {
			playerStats = append(playerStats, kafkaclient.PlayerMatchStat{
				PlayerID:  playerID,
				Kills:     8 + i,
				Deaths:    15 + i,
				Assists:   3 + i,
				Score:     40 + i*5,
				MMRChange: -15,
			})
		}

		event := &kafkaclient.MatchEvent{
			MatchID:   matchID,
			LobbyID:   lobbyID,
			EventType: kafkaclient.EventTypeMatchCompleted,
			GameType:  "cs2",
			Region:    "us-east",
			PlayerIDs: allPlayers,
			Teams: []kafkaclient.TeamInfo{
				{TeamID: team1ID, Name: "Team Alpha", PlayerIDs: team1Players},
				{TeamID: team2ID, Name: "Team Bravo", PlayerIDs: team2Players},
			},
			Result: &kafkaclient.MatchResult{
				WinnerTeamID: &team1ID,
				IsDraw:       false,
				Scores: map[string]int{
					team1ID.String(): 16,
					team2ID.String(): 9,
				},
				Duration:    2700, // 45 minutes
				PlayerStats: playerStats,
				CompletedAt: time.Now().UnixMilli(),
			},
			Metadata: map[string]string{
				"total_rounds": "25",
				"overtime":     "false",
			},
		}

		err := publisher.PublishMatchResult(ctx, event)
		require.NoError(t, err, "Should publish match completed event")
		t.Log("Published MATCH_COMPLETED event with full results")
	})

	t.Log("Match lifecycle events published successfully")
}

// TestMatchmaking_WebSocketBroadcast tests WebSocket broadcast event publishing
func TestMatchmaking_WebSocketBroadcast(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Kafka integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	config := &kafkaclient.Config{
		BootstrapServers: getKafkaTestURI(),
		SecurityProtocol: "PLAINTEXT",
		Region:           "test",
	}

	client, err := kafkaclient.NewClient(config)
	if err != nil {
		t.Skipf("Skipping Kafka test - unable to create client: %v", err)
	}
	defer client.Close()

	publisher := kafkaclient.NewEventPublisher(client)
	lobbyID := uuid.New()
	targetUserIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	// Test 1: Broadcast to specific lobby
	t.Run("BroadcastToLobby", func(t *testing.T) {
		event := &kafkaclient.WebSocketBroadcastEvent{
			Type:    "LOBBY_UPDATE",
			LobbyID: &lobbyID,
			Payload: map[string]interface{}{
				"lobby_id":       lobbyID.String(),
				"status":         "ready",
				"players_ready":  5,
				"players_total":  5,
				"countdown":      30,
				"estimated_wait": 15,
			},
		}

		err := publisher.PublishWebSocketBroadcast(ctx, event)
		require.NoError(t, err, "Should publish WebSocket broadcast event")
		t.Log("Published WebSocket broadcast to lobby")
	})

	// Test 2: Targeted broadcast to specific users
	t.Run("BroadcastToSpecificUsers", func(t *testing.T) {
		event := &kafkaclient.WebSocketBroadcastEvent{
			Type:      "MATCH_FOUND",
			TargetIDs: targetUserIDs,
			Payload: map[string]interface{}{
				"match_id":        uuid.New().String(),
				"server_ip":       "192.168.1.100",
				"accept_deadline": time.Now().Add(30 * time.Second).Unix(),
			},
		}

		err := publisher.PublishWebSocketBroadcast(ctx, event)
		require.NoError(t, err, "Should publish targeted WebSocket broadcast")
		t.Log("Published targeted WebSocket broadcast")
	})

	// Test 3: Global broadcast
	t.Run("GlobalBroadcast", func(t *testing.T) {
		event := &kafkaclient.WebSocketBroadcastEvent{
			Type: "SYSTEM_ANNOUNCEMENT",
			Payload: map[string]interface{}{
				"message":    "Server maintenance in 30 minutes",
				"severity":   "warning",
				"expires_at": time.Now().Add(30 * time.Minute).Unix(),
			},
		}

		err := publisher.PublishWebSocketBroadcast(ctx, event)
		require.NoError(t, err, "Should publish global WebSocket broadcast")
		t.Log("Published global WebSocket broadcast")
	})

	t.Log("WebSocket broadcast events published successfully")
}

// TestMatchmaking_PlayerStatus tests player status compacted topic
func TestMatchmaking_PlayerStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Kafka integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	config := &kafkaclient.Config{
		BootstrapServers: getKafkaTestURI(),
		SecurityProtocol: "PLAINTEXT",
		Region:           "test",
	}

	client, err := kafkaclient.NewClient(config)
	if err != nil {
		t.Skipf("Skipping Kafka test - unable to create client: %v", err)
	}
	defer client.Close()

	publisher := kafkaclient.NewEventPublisher(client)
	playerID := uuid.New()

	// Test player status transitions
	statuses := []struct {
		status   string
		metadata map[string]string
	}{
		{
			status: "online",
			metadata: map[string]string{
				"game_id":   "cs2",
				"region":    "us-east",
				"client_ip": "192.168.1.50",
			},
		},
		{
			status: "in_queue",
			metadata: map[string]string{
				"game_mode": "competitive",
				"tier":      "premium",
				"queue_id":  uuid.New().String(),
			},
		},
		{
			status: "in_lobby",
			metadata: map[string]string{
				"lobby_id":  uuid.New().String(),
				"team_side": "CT",
			},
		},
		{
			status: "in_match",
			metadata: map[string]string{
				"match_id":  uuid.New().String(),
				"server_ip": "192.168.1.100:27015",
			},
		},
		{
			status: "offline",
			metadata: map[string]string{
				"reason":        "disconnect",
				"last_activity": time.Now().Format(time.RFC3339),
			},
		},
	}

	for _, s := range statuses {
		t.Run("PublishStatus_"+s.status, func(t *testing.T) {
			err := publisher.PublishPlayerStatus(ctx, playerID, s.status, s.metadata)
			require.NoError(t, err, "Should publish player status: %s", s.status)
			t.Logf("Published player status: %s", s.status)
		})
	}

	t.Log("Player status events published successfully")
}

// TestMatchmaking_DLQ tests dead letter queue publishing
func TestMatchmaking_DLQ(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Kafka integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	config := &kafkaclient.Config{
		BootstrapServers: getKafkaTestURI(),
		SecurityProtocol: "PLAINTEXT",
		Region:           "test",
	}

	client, err := kafkaclient.NewClient(config)
	if err != nil {
		t.Skipf("Skipping Kafka test - unable to create client: %v", err)
	}
	defer client.Close()

	publisher := kafkaclient.NewEventPublisher(client)

	// Test DLQ publishing for failed message
	t.Run("PublishToDLQ", func(t *testing.T) {
		failedEvent := map[string]interface{}{
			"player_id":  uuid.New().String(),
			"event_type": "QUEUE_JOINED",
			"game_type":  "cs2",
		}

		testErr := assert.AnError // Use testify's error for testing

		err := publisher.PublishToDLQ(ctx, kafkaclient.TopicQueueEvents, "test-key", failedEvent, testErr)
		require.NoError(t, err, "Should publish to DLQ")
		t.Log("Published failed event to DLQ")
	})

	t.Log("DLQ events published successfully")
}

// TestMatchmaking_BatchPublishing tests batch message publishing
func TestMatchmaking_BatchPublishing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Kafka integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	config := &kafkaclient.Config{
		BootstrapServers: getKafkaTestURI(),
		SecurityProtocol: "PLAINTEXT",
		Region:           "test",
	}

	client, err := kafkaclient.NewClient(config)
	if err != nil {
		t.Skipf("Skipping Kafka test - unable to create client: %v", err)
	}
	defer client.Close()

	// Test batch publishing of queue events
	t.Run("BatchPublishQueueEvents", func(t *testing.T) {
		numEvents := 100
		msgs := make([]*kafkaclient.Message, numEvents)

		for i := 0; i < numEvents; i++ {
			event := &kafkaclient.QueueEvent{
				EventID:   uuid.New(),
				PlayerID:  uuid.New(),
				GameType:  "cs2",
				Region:    "us-east",
				MMR:       1200 + i*10,
				EventType: kafkaclient.EventTypeQueueJoined,
				QueueTime: time.Now().UnixMilli(),
			}

			msgs[i] = &kafkaclient.Message{
				Key:       event.PlayerID.String(),
				Value:     event,
				Timestamp: time.Now(),
				Headers: map[string]string{
					"event_type": event.EventType,
				},
			}
		}

		err := client.PublishBatch(ctx, kafkaclient.TopicQueueEvents, msgs)
		require.NoError(t, err, "Should batch publish queue events")
		t.Logf("Batch published %d queue events", numEvents)
	})

	t.Log("Batch publishing completed successfully")
}

// TestMatchmaking_FullLifecycle_Integration tests a complete matchmaking flow
func TestMatchmaking_FullLifecycle_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Kafka integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	config := &kafkaclient.Config{
		BootstrapServers: getKafkaTestURI(),
		SecurityProtocol: "PLAINTEXT",
		Region:           "test",
	}

	client, err := kafkaclient.NewClient(config)
	if err != nil {
		t.Skipf("Skipping Kafka test - unable to create client: %v", err)
	}
	defer client.Close()

	publisher := kafkaclient.NewEventPublisher(client)

	// Create 10 players for 5v5 match
	players := make([]uuid.UUID, 10)
	for i := range players {
		players[i] = uuid.New()
	}

	lobbyID := uuid.New()
	matchID := uuid.New()
	poolID := uuid.New()
	team1ID := uuid.New()
	team2ID := uuid.New()

	// Phase 1: Players join queue
	t.Log("Phase 1: Players joining queue...")
	for _, playerID := range players {
		// Set player status to in_queue
		err := publisher.PublishPlayerStatus(ctx, playerID, "in_queue", map[string]string{
			"game_mode": "competitive",
			"tier":      "premium",
		})
		require.NoError(t, err)

		// Publish queue joined event
		err = publisher.PublishQueueEvent(ctx, &kafkaclient.QueueEvent{
			PlayerID:  playerID,
			GameType:  "cs2",
			Region:    "us-east",
			MMR:       1500 + int(time.Now().UnixNano()%200),
			EventType: kafkaclient.EventTypeQueueJoined,
		})
		require.NoError(t, err)
	}
	t.Logf("10 players joined queue")

	// Phase 2: Matchmaking finds a match, creates lobby
	t.Log("Phase 2: Creating lobby...")
	err = publisher.PublishLobbyEvent(ctx, &kafkaclient.LobbyEvent{
		LobbyID:   lobbyID,
		EventType: kafkaclient.EventTypeLobbyCreated,
		PlayerIDs: players,
		GameType:  "cs2",
		Region:    "us-east",
		AvgMMR:    1550,
	})
	require.NoError(t, err)

	// Update player statuses to in_lobby
	for _, playerID := range players {
		err := publisher.PublishPlayerStatus(ctx, playerID, "in_lobby", map[string]string{
			"lobby_id": lobbyID.String(),
		})
		require.NoError(t, err)
	}
	t.Log("Lobby created with 10 players")

	// Phase 3: Initialize prize pool
	t.Log("Phase 3: Setting up prize pool...")
	totalPrizePool := int64(500) // Platform contribution
	err = publisher.PublishPrizePoolEvent(ctx, &kafkaclient.PrizePoolEvent{
		PoolID:      poolID,
		LobbyID:     lobbyID,
		EventType:   kafkaclient.EventTypePrizePoolUpdated,
		TotalAmount: totalPrizePool,
		Currency:    "usd",
	})
	require.NoError(t, err)

	// Add player contributions
	for _, playerID := range players {
		totalPrizePool += 100
		err = publisher.PublishPrizePoolEvent(ctx, &kafkaclient.PrizePoolEvent{
			PoolID:          poolID,
			LobbyID:         lobbyID,
			EventType:       kafkaclient.EventTypePrizePoolUpdated,
			TotalAmount:     totalPrizePool,
			Currency:        "usd",
			ContributorID:   &playerID,
			ContributionAmt: 100,
		})
		require.NoError(t, err)
	}
	t.Logf("Prize pool setup: $%.2f", float64(totalPrizePool)/100)

	// Phase 4: Players ready up
	t.Log("Phase 4: Players readying up...")
	for _, playerID := range players {
		err = publisher.PublishLobbyEvent(ctx, &kafkaclient.LobbyEvent{
			LobbyID:   lobbyID,
			EventType: kafkaclient.EventTypeReadyStatusChanged,
			PlayerIDs: []uuid.UUID{playerID},
			GameType:  "cs2",
			Region:    "us-east",
			Metadata:  map[string]string{"ready": "true"},
		})
		require.NoError(t, err)
	}

	// Lobby ready
	err = publisher.PublishLobbyEvent(ctx, &kafkaclient.LobbyEvent{
		LobbyID:   lobbyID,
		EventType: kafkaclient.EventTypeLobbyReady,
		PlayerIDs: players,
		GameType:  "cs2",
		Region:    "us-east",
		AvgMMR:    1550,
	})
	require.NoError(t, err)
	t.Log("All players ready, lobby ready")

	// Phase 5: Match created
	t.Log("Phase 5: Creating match...")
	team1Players := players[:5]
	team2Players := players[5:]

	err = publisher.PublishMatchCreated(ctx, &kafkaclient.MatchEvent{
		MatchID:   matchID,
		LobbyID:   lobbyID,
		EventType: kafkaclient.EventTypeMatchCreated,
		GameType:  "cs2",
		Region:    "us-east",
		PlayerIDs: players,
		Teams: []kafkaclient.TeamInfo{
			{TeamID: team1ID, Name: "Team Alpha", PlayerIDs: team1Players, Side: "CT"},
			{TeamID: team2ID, Name: "Team Bravo", PlayerIDs: team2Players, Side: "T"},
		},
	})
	require.NoError(t, err)

	// Update player statuses to in_match
	for _, playerID := range players {
		err := publisher.PublishPlayerStatus(ctx, playerID, "in_match", map[string]string{
			"match_id": matchID.String(),
		})
		require.NoError(t, err)
	}
	t.Log("Match created")

	// Phase 6: Match started
	err = publisher.PublishMatchResult(ctx, &kafkaclient.MatchEvent{
		MatchID:   matchID,
		LobbyID:   lobbyID,
		EventType: kafkaclient.EventTypeMatchStarted,
		GameType:  "cs2",
		Region:    "us-east",
		PlayerIDs: players,
	})
	require.NoError(t, err)

	// Broadcast match started to WebSocket clients
	err = publisher.PublishWebSocketBroadcast(ctx, &kafkaclient.WebSocketBroadcastEvent{
		Type:      "MATCH_STARTED",
		LobbyID:   &lobbyID,
		TargetIDs: players,
		Payload: map[string]interface{}{
			"match_id":   matchID.String(),
			"server_ip":  "192.168.1.100:27015",
			"map":        "de_dust2",
			"game_mode":  "competitive",
			"prize_pool": totalPrizePool,
		},
	})
	require.NoError(t, err)
	t.Log("Match started")

	// Phase 7: Match completed
	t.Log("Phase 7: Match completed...")
	var playerStats []kafkaclient.PlayerMatchStat
	for i, playerID := range team1Players {
		playerStats = append(playerStats, kafkaclient.PlayerMatchStat{
			PlayerID:  playerID,
			Kills:     18 + i,
			Deaths:    12 - i,
			Assists:   6 + i,
			Score:     90 + i*5,
			MMRChange: 25,
		})
	}
	for i, playerID := range team2Players {
		playerStats = append(playerStats, kafkaclient.PlayerMatchStat{
			PlayerID:  playerID,
			Kills:     10 + i,
			Deaths:    16 + i,
			Assists:   4 + i,
			Score:     50 + i*3,
			MMRChange: -15,
		})
	}

	err = publisher.PublishMatchResult(ctx, &kafkaclient.MatchEvent{
		MatchID:   matchID,
		LobbyID:   lobbyID,
		EventType: kafkaclient.EventTypeMatchCompleted,
		GameType:  "cs2",
		Region:    "us-east",
		PlayerIDs: players,
		Teams: []kafkaclient.TeamInfo{
			{TeamID: team1ID, Name: "Team Alpha", PlayerIDs: team1Players},
			{TeamID: team2ID, Name: "Team Bravo", PlayerIDs: team2Players},
		},
		Result: &kafkaclient.MatchResult{
			WinnerTeamID: &team1ID,
			IsDraw:       false,
			Scores: map[string]int{
				team1ID.String(): 16,
				team2ID.String(): 12,
			},
			Duration:    2400, // 40 minutes
			PlayerStats: playerStats,
			CompletedAt: time.Now().UnixMilli(),
		},
	})
	require.NoError(t, err)

	// Update player statuses back to online
	for _, playerID := range players {
		err := publisher.PublishPlayerStatus(ctx, playerID, "online", map[string]string{
			"last_match": matchID.String(),
		})
		require.NoError(t, err)
	}

	// Broadcast match results
	err = publisher.PublishWebSocketBroadcast(ctx, &kafkaclient.WebSocketBroadcastEvent{
		Type:      "MATCH_COMPLETED",
		TargetIDs: players,
		Payload: map[string]interface{}{
			"match_id":     matchID.String(),
			"winner_team":  "Team Alpha",
			"final_score":  "16-12",
			"duration":     2400,
			"prize_pool":   totalPrizePool,
			"winner_share": totalPrizePool * 70 / 100, // 70% to winners
		},
	})
	require.NoError(t, err)

	t.Log("Match completed - Full lifecycle test passed!")
	t.Logf("Summary: 10 players, Prize Pool: $%.2f, Duration: 40 minutes, Score: 16-12", float64(totalPrizePool)/100)
}

// consumeQueueEvents is a helper to consume queue events for testing
func consumeQueueEvents(ctx context.Context, t *testing.T, brokers string, playerKey string, events chan<- *kafkaclient.QueueEvent) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokers},
		Topic:    kafkaclient.TopicQueueEvents,
		GroupID:  "test-consumer-" + uuid.New().String(),
		MinBytes: 1,
		MaxBytes: 10e6,
		MaxWait:  1 * time.Second,
	})
	defer reader.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				continue
			}

			if string(msg.Key) == playerKey {
				var event kafkaclient.QueueEvent
				if err := json.Unmarshal(msg.Value, &event); err != nil {
					t.Logf("Failed to unmarshal event: %v", err)
					continue
				}
				select {
				case events <- &event:
				default:
				}
			}
		}
	}
}

// BenchmarkKafkaPublish benchmarks Kafka publish performance
func BenchmarkKafkaPublish(b *testing.B) {
	config := &kafkaclient.Config{
		BootstrapServers: getKafkaTestURI(),
		SecurityProtocol: "PLAINTEXT",
		Region:           "test",
	}

	client, err := kafkaclient.NewClient(config)
	if err != nil {
		b.Skipf("Skipping benchmark - unable to create client: %v", err)
	}
	defer client.Close()

	publisher := kafkaclient.NewEventPublisher(client)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		event := &kafkaclient.QueueEvent{
			PlayerID:  uuid.New(),
			GameType:  "cs2",
			Region:    "us-east",
			MMR:       1500,
			EventType: kafkaclient.EventTypeQueueJoined,
		}
		_ = publisher.PublishQueueEvent(ctx, event)
	}
}
