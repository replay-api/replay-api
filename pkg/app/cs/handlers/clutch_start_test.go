package handlers_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	infocs "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	infocsDomain "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
	evt "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/sendtables"
	dispatch "github.com/markus-wa/godispatch"
	h "github.com/replay-api/replay-api/pkg/app/cs/handlers"
	"github.com/replay-api/replay-api/pkg/app/cs/state"
	common "github.com/replay-api/replay-api/pkg/domain"
	e "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	"github.com/stretchr/testify/mock"
)

type args struct {
	p            infocs.Parser
	matchContext *state.CS2MatchContext
	out          chan *e.GameEvent
}

type testMatrixArgs struct {
	name         string
	args         args
	expectEvents []*e.GameEvent
	eventInfo    evt.Kill
}

// Ensure mockParser implements the required interface
var _ infocs.Parser = (*mockParser)(nil)

type mockParser struct {
	mock.Mock
}

// Cancel implements infocs.Parser.
func (m *mockParser) Cancel() {
	panic("unimplemented")
}

// Close implements infocs.Parser.
func (m *mockParser) Close() error {
	panic("unimplemented")
}

// CurrentFrame implements demoInfoCsParserMock.
func (m *mockParser) CurrentFrame() int {
	panic("unimplemented")
}

// CurrentTime implements demoInfoCsParserMock.
func (m *mockParser) CurrentTime() time.Duration {
	panic("unimplemented")
}

// GameState implements demoInfoCsParserMock.
func (m *mockParser) GameState() infocs.GameState {
	return m.Called().Get(0).(infocs.GameState)
}

// Header implements demoInfoCsParserMock.
func (m *mockParser) Header() infocsDomain.DemoHeader {
	panic("unimplemented")
}

// ParseHeader implements demoInfoCsParserMock.
func (m *mockParser) ParseHeader() (infocsDomain.DemoHeader, error) {
	panic("unimplemented")
}

// ParseNextFrame implements demoInfoCsParserMock.
func (m *mockParser) ParseNextFrame() (moreFrames bool, err error) {
	panic("unimplemented")
}

// ParseToEnd implements demoInfoCsParserMock.
func (m *mockParser) ParseToEnd() (err error) {
	panic("unimplemented")
}

// Progress implements demoInfoCsParserMock.
func (m *mockParser) Progress() float32 {
	panic("unimplemented")
}

// RegisterEventHandler implements demoInfoCsParserMock.
func (m *mockParser) RegisterEventHandler(handler any) dispatch.HandlerIdentifier {
	panic("unimplemented")
}

// RegisterNetMessageHandler implements demoInfoCsParserMock.
func (m *mockParser) RegisterNetMessageHandler(handler any) dispatch.HandlerIdentifier {
	panic("unimplemented")
}

// ServerClasses implements demoInfoCsParserMock.
func (m *mockParser) ServerClasses() sendtables.ServerClasses {
	panic("unimplemented")
}

// TickRate implements demoInfoCsParserMock.
func (m *mockParser) TickRate() float64 {
	panic("unimplemented")
}

// TickTime implements demoInfoCsParserMock.
func (m *mockParser) TickTime() time.Duration {
	panic("unimplemented")
}

// UnregisterEventHandler implements demoInfoCsParserMock.
func (m *mockParser) UnregisterEventHandler(identifier dispatch.HandlerIdentifier) {
	panic("unimplemented")
}

// UnregisterNetMessageHandler implements demoInfoCsParserMock.
func (m *mockParser) UnregisterNetMessageHandler(identifier dispatch.HandlerIdentifier) {
	panic("unimplemented")
}

type mockGameState struct {
	mock.Mock
	infocs.GameState
}
type mockParticipants struct {
	mock.Mock
	infocs.Participants
}
type mockPlayer struct {
	mock.Mock
	infocsDomain.Player
}

func getParserMock(mockGameState demoInfoCsGameStateMock) demoInfoCsParserMock {
	mock := new(mockParser)

	mock.On("GameState").Return(mockGameState)
	mock.On("CurrentTime").Return(time.Duration(0))
	mock.On("Cancel").Return(
		func() {
		},
	)

	return mock
}

func getGameStateMock(totalRoundsPlayed int, mockParticipants *mockParticipants) demoInfoCsGameStateMock {
	mock := new(mockGameState)

	mock.On("TotalRoundsPlayed").Return(totalRoundsPlayed)
	mock.On("Participants").Return(mockParticipants)

	return mock
}

func getParticipantsMock(players []*mockPlayer) *mockParticipants {
	mock := new(mockParticipants)

	mock.On("Playing").Return(players)

	return mock
}

func getPlayerMock(isAlive bool) *mockPlayer {
	mock := new(mockPlayer)

	mock.On("IsAlive").Return(isAlive)

	return mock
}

func TestClutchStart(t *testing.T) {
	t.Parallel()

	tests := []testMatrixArgs{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h.ClutchStart(tc.args.p, tc.args.matchContext, tc.args.out)(
				tc.eventInfo,
			)
			defer close(tc.args.out)

			receivedEvents := []*e.GameEvent{}

			go func() {
				for rec := range tc.args.out {
					receivedEvents = append(receivedEvents, rec)
				}
			}()

			<-time.After(1 * time.Second)

			if len(receivedEvents) != len(tc.expectEvents) {
				t.Errorf("Expected %d events, but got %d on %s", len(tc.expectEvents), len(receivedEvents), tc.name)
			}

			for i, rec := range receivedEvents {
				if rec.Type != tc.expectEvents[i].Type {
					t.Errorf("Expected event type %s, but got %s on %s", tc.expectEvents[i].Type, rec.Type, tc.name)
				}
			}

			for i, rec := range receivedEvents {
				if rec.GameTime != tc.expectEvents[i].GameTime {
					t.Errorf("Expected event time %v, but got %v on %s", tc.expectEvents[i].GameTime, rec.GameTime, tc.name)
				}
			}

			for i, rec := range receivedEvents {
				if rec.ResourceOwner != tc.expectEvents[i].ResourceOwner {
					t.Errorf("Expected event resource owner %s, but got %s on %s", tc.expectEvents[i].ResourceOwner, rec.ResourceOwner, tc.name)
				}
			}

			for i, rec := range receivedEvents {
				if rec.MatchID != tc.expectEvents[i].MatchID {
					t.Errorf("Expected event match id %s, but got %s on %s", tc.expectEvents[i].MatchID, rec.MatchID, tc.name)
				}
			}
		})
	}
}

func getUserContext() context.Context {
	tentantContext := context.WithValue(context.Background(), common.TenantIDKey, "test-tenant")
	clientContext := context.WithValue(tentantContext, common.ClientIDKey, "test-client")
	groupContext := context.WithValue(clientContext, common.GroupIDKey, "test-group")
	userContext := context.WithValue(groupContext, common.UserIDKey, "test-user")

	return userContext
}

func getDefaultTestCaseArgs() testMatrixArgs {
	matchCtx := state.NewCS2MatchContext(getUserContext(), uuid.New())
	participants := getParticipantsMock([]*mockPlayer{getPlayerMock(true)})
	gs := getGameStateMock(1, participants)
	p := getParserMock(gs)

	return testMatrixArgs{
		name: "Test ClutchStart (Default)",
		args: args{
			p:            p,
			matchContext: matchCtx,
			out:          make(chan *e.GameEvent, 10),
		},
		expectEvents: []*e.GameEvent{
			{
				ID:            uuid.New(),
				MatchID:       matchCtx.MatchID,
				Type:          common.Event_ClutchStartID,
				GameTime:      time.Duration(1),
				ResourceOwner: matchCtx.ResourceOwner,
			},
		},
		eventInfo: evt.Kill{},
	}
}

func getNotInClutchTestCaseArgs() testMatrixArgs {
	matchCtx := state.NewCS2MatchContext(getUserContext(), uuid.New())
	participants := getParticipantsMock([]*mockPlayer{getPlayerMock(true), getPlayerMock(true)})
	gs := getGameStateMock(1, participants)
	p := getParserMock(gs)

	return testMatrixArgs{
		name: "Test ClutchStart (Not in clutch situation)",
		args: args{
			p:            p,
			matchContext: matchCtx,
			out:          make(chan *e.GameEvent, 10),
		},
		expectEvents: []*e.GameEvent{},
		eventInfo:    evt.Kill{},
	}
}
