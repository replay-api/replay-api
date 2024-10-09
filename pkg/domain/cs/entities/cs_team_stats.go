package entities

import "github.com/google/uuid"

type TeamIDType = uuid.UUID
type TeamHashIDType = string
type CSTeamSideIDType = string

const (
	CSTeamSideTID  CSTeamSideIDType = "t"
	CSTeamSideCTID CSTeamSideIDType = "ct"
)

type CSTeamStats struct {
	Efficiency map[CSTeamSideIDType]StatEfficiencyUnit
}
