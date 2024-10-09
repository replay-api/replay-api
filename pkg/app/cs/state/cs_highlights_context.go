package state

import (
	"time"

	cs2 "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
)

type Highlight struct {
	Type           string                 // "Kill", "Clutch", "Economy", etc.
	Round          int                    // Round number
	Tick           int                    // Game tick when the event occurred
	Time           time.Duration          // Real time since the match started
	Players        []*cs2.Player          // Players involved (killer, Opponent, etc.)
	Team           *cs2.Team              // Team involved
	AdditionalData map[string]interface{} // Additional data specific to highlight type
}

// func GenerateHighlights(events ...Event) {
// 	for _, event := range events {
// 		switch e := event.(type) {
// 		case event.Kill:
// 			if IsTrade(e, rc.Match.PlayerLastKill) {
// 				rc.AddHighlight(&Highlight{
// 					Type:    "Trade",
// 					Round:   rc.Match.RoundNumber,
// 					Tick:    e.Tick,
// 					Players: []*cs2.Player{e.Killer, e.Opponent},
// 				})
// 			}
// 			// ... (logic for other kill-related highlights like clutches)

// 		case event.RoundStart:
// 			// ... (logic for round start highlights)

// 			// ... (add more cases for other event types you want to track)
// 		}
// 	}
// }

// func (rc *ReplayContext) GenerateHighlights(p *cs2.Parser) {
// 	p.RegisterEventHandler(func(e event.Kill) {
// 		if IsTrade(e, rc.Match.PlayerLastKill) {
// 			rc.AddHighlight(&Highlight{
// 				Type:    "Trade",
// 				Round:   rc.Match.RoundNumber,
// 				Tick:    e.Tick,
// 				Players: []*cs2.Player{e.Killer, e.Opponent},
// 				Time:    p.CurrentTime(),
// 			})
// 		}

// 		if IsClutch(e, p.GameState()) {
// 			rc.AddHighlight(&Highlight{
// 				Type:    "Clutch",
// 				Round:   rc.Match.RoundNumber,
// 				Tick:    e.Tick,
// 				Players: []*cs2.Player{e.Killer},
// 				Time:    p.CurrentTime(),
// 				AdditionalData: map[string]interface{}{
// 					"clutchType": fmt.Sprintf("1v%d", len(p.GameState().Participants().OtherTeam(e.Killer.Team))) - 1,
// 				},
// 			})
// 		}

// 		if IsEntryFrag(e, p.GameState()) {
// 			rc.AddHighlight(&Highlight{
// 				Type:    "Entry Frag",
// 				Round:   rc.Match.RoundNumber,
// 				Tick:    e.Tick,
// 				Players: []*cs2.Player{e.Killer},
// 				Time:    p.CurrentTime(),
// 			})
// 		}

// 		// ... Add more kill-related highlight checks here ...
// 	})

// 	p.RegisterEventHandler(func(e event.RoundStart) {
// 		// Logic to handle round start highlights (e.g., economy reset)
// 		rc.AddHighlight(&Highlight{
// 			Type:  "Round Start",
// 			Round: rc.Match.RoundNumber,
// 			Tick:  e.Tick,
// 			Time:  p.CurrentTime(),
// 		})
// 	})

// 	p.RegisterEventHandler(func(e event.BombPlanted) {
// 		rc.AddHighlight(&Highlight{
// 			Type:    "Bomb Planted",
// 			Round:   rc.Match.RoundNumber,
// 			Tick:    e.Tick,
// 			Players: []*cs2.Player{e.Player},
// 			Time:    p.CurrentTime(),
// 		})
// 	})

// 	p.RegisterEventHandler(func(e event.BombDefused) {
// 		rc.AddHighlight(&Highlight{
// 			Type:    "Bomb Defused",
// 			Round:   rc.Match.RoundNumber,
// 			Tick:    e.Tick,
// 			Players: []*cs2.Player{e.Player},
// 			Time:    p.CurrentTime(),
// 		})
// 	})

// 	// ... Add more event handlers for other highlight types ...
// }

// // Helper functions

// func IsTrade(kill event.Kill, lastKills map[uint64]*event.Kill) bool {
// 	// Logic to check if the kill is a trade within the last 5 seconds
// 	// ...
// }

// func IsClutch(kill event.Kill, gs *cs2.GameState) bool {
// 	// Logic to check if the kill is a clutch (1vX situation)
// 	// ...
// }

// func IsEntryFrag(kill event.Kill, gs *cs2.GameState) bool {
// 	// Logic to check if the kill is an entry frag
// 	// ...
// }

// // ... (add other helper functions as needed)
