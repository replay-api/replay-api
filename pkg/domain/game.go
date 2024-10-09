package common

type GameIDKey string

const (
	CS2_GAME_ID   GameIDKey = "cs2"
	CSGO_GAME_ID  GameIDKey = "csgo"
	VLRNT_GAME_ID GameIDKey = "vlrnt"
)

type EventIDKey string

type TickIDType float64

const (
	Event_MatchStartID           EventIDKey = "MatchStart"
	Event_RoundMVPAnnouncementID EventIDKey = "RoundMVPAnnouncement"
	Event_RoundEndID             EventIDKey = "RoundEndID"
	Event_FragOrScoreID          EventIDKey = "FragOrScoreID"
	Event_WeaponFireID           EventIDKey = "WeaponFireID"
	Event_HitID                  EventIDKey = "HitID"
	Event_GenericGameEventID     EventIDKey = "GenericGameEvent"
	Event_ClutchStartID          EventIDKey = "ClutchStart"
	Event_ClutchProgressID       EventIDKey = "ClutchProgress"
	Event_ClutchEndID            EventIDKey = "ClutchEnd"
	Event_Economy                EventIDKey = "EconomyEvent"
)

type Game struct {
	ID     GameIDKey    `json:"id"`             // ID is the unique identifier of the game.
	Name   string       `json:"name"`           // Name is the name of the game.
	Events []EventIDKey `json:"in_game_events"` // Events is a map of SUPPORTED/IMPLEMENTED in-game events to their corresponding event names.
}

func mapCSEvents() []EventIDKey {
	return []EventIDKey{
		Event_MatchStartID,
		Event_RoundMVPAnnouncementID,
		Event_RoundEndID,
		Event_GenericGameEventID,
		Event_ClutchStartID,
		Event_ClutchProgressID,
		Event_ClutchEndID,
		Event_Economy,
	}
}

// func mapVlrntEvents() map[EventIDKey]string {
// 	events := make(map[EventIDKey]string, 0)

// 	events[Event_GenericGameEventID] = "vlrnt::generic_game_event"
// 	events[Event_MatchStartID] = "vlrnt::new_match"

// 	return events
// }

var (
	CS2 = &Game{
		ID:     CS2_GAME_ID,
		Name:   "Counter-Strike: 2",
		Events: mapCSEvents(),
	}

	CSGO = &Game{
		ID:     CSGO_GAME_ID,
		Name:   "Counter-Strike: Global Offensive",
		Events: mapCSEvents(),
	}

	// VLRNT = &Game{
	// 	ID:     VLRNT_GAME_ID,
	// 	Name:   "Valorant",
	// 	Events: mapVlrntEvents(),
	// }
)
