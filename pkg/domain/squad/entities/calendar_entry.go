package squad_entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
)

type CalendarEntryType string

const (
	CalendarEntryMatch        CalendarEntryType = "Match"
	CalendarEntryTraining     CalendarEntryType = "Training"
	CalendarEntryAssessment   CalendarEntryType = "Assessment"
	CalendarEntryEvent        CalendarEntryType = "Event"
	CalendarEntryRegistration CalendarEntryType = "Registration"
)

type CalendarEntryCategory string

const (
	SquadCalendar      CalendarEntryCategory = "Squad"
	PlayerCalendar     CalendarEntryCategory = "Player"
	TournamentCalendar CalendarEntryCategory = "Tournament"
	EventCalendar      CalendarEntryCategory = "Event"
)

type CalendarEntryStatus string

const (
	CalendarEntryStatusPending   CalendarEntryStatus = "Pending"
	CalendarEntryStatusConfirmed CalendarEntryStatus = "Confirmed"
	CalendarEntryStatusCancelled CalendarEntryStatus = "Cancelled"
)

type CalendarEntry struct {
	common.BaseEntity
	TournamentID *uuid.UUID            `json:"tournament_id" bson:"tournament_id"`
	SquadIDs     []uuid.UUID           `json:"squad_ids" bson:"squad_ids"`
	PlayerIDs    []uuid.UUID           `json:"player_ids" bson:"player_ids"`
	GameID       common.GameIDKey      `json:"game_id" bson:"game_id"`
	Category     CalendarEntryCategory `json:"category" bson:"category"`
	StartTime    time.Time             `json:"start_time" bson:"start_time"`
	EndTime      time.Time             `json:"end_time" bson:"end_time"`
	Title        string                `json:"title" bson:"title"`
	Description  string                `json:"description" bson:"description"`
	Location     string                `json:"location" bson:"location"`
	Region       string                `json:"region" bson:"region"`
	Passphrase   string                `json:"-" bson:"passphrase"` // TODO: getPassphrase will be separated route
	Type         CalendarEntryType     `json:"type" bson:"type"`
	Status       string                `json:"status" bson:"status"`
}
