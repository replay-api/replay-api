package entities

import "time"

type CSGameRules struct {
	RoundTime  time.Duration
	FreezeTime time.Duration
	BombTime   time.Duration
	ConVars    map[string]string // https://developer.valvesoftware.com/wiki/List_of_CS:GO_Cvars.
}
