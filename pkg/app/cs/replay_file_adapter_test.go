package cs2_test

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	cs2 "github.com/replay-api/replay-api/pkg/app/cs"
	e "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	fps_events "github.com/replay-api/replay-common/pkg/replay/events/game/fps"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

func TestCS2ReplayAdapter_GetEvents(t *testing.T) {
	slog.Error("TestCS2ReplayAdapter_GetEvents")
	filePath := "../../../test/sample_replays/cs2/sound.dem"

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open demo file: %v", err)
	}

	defer file.Close()

	adapter := cs2.NewCS2ReplayAdapter()

	results := make([]*e.GameEvent, 0)
	types := make(map[fps_events.EventIDKey]int)

	clutchEventCountPerPlayerAndEvent := make(map[shared.PlayerIDType]map[fps_events.EventIDKey]int)

	eventsChan := make(chan *e.GameEvent)
	mutex := &sync.Mutex{}

	eventCount := 0
	go func() {
		for ge := range eventsChan {
			// slog.Info("Event: %v", "ge", ge)
			mutex.Lock()
			eventCount++
			results = append(results, ge)
			types[ge.Type]++

			if ge.Type == fps_events.Event_ClutchProgressID || ge.Type == fps_events.Event_ClutchStartID || ge.Type == fps_events.Event_ClutchEndID {
				playerIDs, err := ge.GetPlayerIDs()

				if err != nil {
					t.Logf("cannot get player id due to %v", err)
				}

				if len(playerIDs) != 0 {
					playerID := shared.PlayerIDType(playerIDs[0])
					if clutchEventCountPerPlayerAndEvent[playerID] == nil {
						clutchEventCountPerPlayerAndEvent[playerID] = make(map[fps_events.EventIDKey]int)
					}

					clutchEventCountPerPlayerAndEvent[playerID][ge.Type]++
				}
			}

			mutex.Unlock()
		}
	}()

	ctx := context.WithValue(context.Background(), shared.TenantIDKey, replay_common.TeamPROTenantID)
	ctx = context.WithValue(ctx, shared.ClientIDKey, replay_common.TeamPROAppClientID)
	ctx = context.WithValue(ctx, shared.UserIDKey, uuid.New())

	match := &e.Match{
		ID:            uuid.New(),
		ReplayFileID:  uuid.New(),
		ResourceOwner: shared.GetResourceOwner(ctx),
	}

	err = adapter.Parse(ctx, match.ID, file, eventsChan)

	if err != nil {
		t.Fatalf("GetEvents returned an error: %v", err)
	}

	for k, v := range types {
		slog.InfoContext(ctx, "Event type: %v, count: %v", string(k), v)
	}

	slog.InfoContext(ctx, "Total events: %v", "eventCount", eventCount)

	for playerID, events := range clutchEventCountPerPlayerAndEvent {
		slog.InfoContext(ctx, "Player: %v", "playerID", playerID)

		eventCountMap := make(map[fps_events.EventIDKey]int)
		for k, v := range events {
			eventCountMap[k] = v
			slog.InfoContext(ctx, "Event type: %s, count: %d", string(k), v)
		}

		clutchStartEventCount := eventCountMap[fps_events.Event_ClutchStartID]
		clutchEndEventCount := eventCountMap[fps_events.Event_ClutchEndID]

		if clutchStartEventCount != clutchEndEventCount {
			t.Errorf("Expected %d events of %s, got %d", clutchStartEventCount, fps_events.Event_ClutchEndID, clutchEndEventCount)
		}
	}

	if len(results) == 0 {
		t.Errorf("Expected >= 1 events, got %d", len(results))
	}
}
