package entities

import (
	"github.com/google/uuid"
)

type ClutchSituationStatusKey string

const (
	NotInClutchSituation ClutchSituationStatusKey = "not_in_clutch_situation"
	ClutchInitiatedKey   ClutchSituationStatusKey = "clutch_initiated"
	ClutchProgressKey    ClutchSituationStatusKey = "clutch_progress"
	ClutchLostKey        ClutchSituationStatusKey = "clutch_lost"
	ClutchWonKey         ClutchSituationStatusKey = "clutch_won"
)

type CSClutchStats struct {
	RoundNumber     int
	PlayerID        *uuid.UUID
	NetworkPlayerID uint64
	OpponentsStats  []CSPlayerStats
	Status          ClutchSituationStatusKey
}
