package cs2_test

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	cs2 "github.com/replay-api/replay-api/pkg/app/cs"
	common "github.com/replay-api/replay-api/pkg/domain"
	e "github.com/replay-api/replay-api/pkg/domain/replay/entities"
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
	types := make(map[common.EventIDKey]int)

	clutchEventCountPerPlayerAndEvent := make(map[common.PlayerIDType]map[common.EventIDKey]int)

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

			if ge.Type == common.Event_ClutchProgressID || ge.Type == common.Event_ClutchStartID || ge.Type == common.Event_ClutchEndID {
				playerIDs, err := ge.GetPlayerIDs()

				if err != nil {
					t.Logf("cannot get player id due to %v", err)
				}

				if len(playerIDs) != 0 {
					if clutchEventCountPerPlayerAndEvent[playerIDs[0]] == nil {
						clutchEventCountPerPlayerAndEvent[playerIDs[0]] = make(map[common.EventIDKey]int)
					}

					clutchEventCountPerPlayerAndEvent[playerIDs[0]][ge.Type]++
				}
			}

			mutex.Unlock()
		}
	}()

	ctx := context.WithValue(context.Background(), common.TenantIDKey, common.TeamPROTenantID)
	ctx = context.WithValue(ctx, common.ClientIDKey, common.TeamPROAppClientID)
	ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())

	match := &e.Match{
		ID:            uuid.New(),
		ReplayFileID:  uuid.New(),
		ResourceOwner: common.GetResourceOwner(ctx),
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

		eventCountMap := make(map[common.EventIDKey]int)
		for k, v := range events {
			eventCountMap[k] = v
			slog.InfoContext(ctx, "Event type: %s, count: %d", string(k), v)
		}

		clutchStartEventCount := eventCountMap[common.Event_ClutchStartID]
		clutchEndEventCount := eventCountMap[common.Event_ClutchEndID]

		if clutchStartEventCount != clutchEndEventCount {
			t.Errorf("Expected %d events of %s, got %d", clutchStartEventCount, common.Event_ClutchEndID, clutchEndEventCount)
		}
	}

	if len(results) == 0 {
		t.Errorf("Expected >= 1 events, got %d", len(results))
	}
}
