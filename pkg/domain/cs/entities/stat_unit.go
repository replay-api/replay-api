package entities

import "time"

type EfficiencyState = int

const (
	EFStateScore EfficiencyState = iota
	EFStateGoldenScore
	EFStateConcede
	EFStateTie
	EFStateInvalid
)

type StatEfficiencyUnit struct {
	State      EfficiencyState
	Basis      uint32
	Efficiency float64
	Summary    map[EfficiencyState]float64
}

type StatEfficiencyTendency struct {
	Difference float64
	Tendency   map[EfficiencyState]float64
	Forecast   *StatEfficiencyUnit

	// Optional chaining
	PriorEfficiency *StatEfficiencyUnit
	History         []EfficiencyState
}

type StatNumberUnit struct {
	Count uint32

	Sum    float64 // total sum
	Avg    float64
	Median float64
	Min    float64
	Max    float64
}

type StatTimeUnit struct {
	Count uint32

	Sum    time.Duration
	Avg    time.Duration
	Median time.Duration
	Min    time.Duration
	Max    time.Duration
}
