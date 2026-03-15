package oracle_services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeterministicTeamUUID(t *testing.T) {
	// Same name should produce the same UUID
	id1 := deterministicTeamUUID("NAVI")
	id2 := deterministicTeamUUID("NAVI")
	assert.Equal(t, id1, id2)

	// Different names should produce different UUIDs
	id3 := deterministicTeamUUID("FaZe")
	assert.NotEqual(t, id1, id3)

	// Empty name should still produce a valid UUID
	id4 := deterministicTeamUUID("")
	assert.NotEqual(t, id4.String(), "00000000-0000-0000-0000-000000000000")
}

func TestStreamMonitor_IsDuplicate(t *testing.T) {
	monitor := &StreamMonitor{}

	score1 := &ParsedScore{
		TeamAName:  "navi",
		TeamBName:  "faze",
		TeamAScore: 13,
		TeamBScore: 7,
	}

	// First score should not be duplicate
	assert.False(t, monitor.isDuplicate(score1))

	// After updating, same score should be duplicate
	monitor.updateLastScore(score1)
	assert.True(t, monitor.isDuplicate(score1))

	// Different score should not be duplicate
	score2 := &ParsedScore{
		TeamAName:  "navi",
		TeamBName:  "faze",
		TeamAScore: 14,
		TeamBScore: 7,
	}
	assert.False(t, monitor.isDuplicate(score2))

	// Update with new score
	monitor.updateLastScore(score2)
	assert.True(t, monitor.isDuplicate(score2))
	assert.False(t, monitor.isDuplicate(score1))
}

func TestStreamMonitor_IsDuplicate_DifferentTeams(t *testing.T) {
	monitor := &StreamMonitor{}

	score1 := &ParsedScore{
		TeamAName:  "navi",
		TeamBName:  "faze",
		TeamAScore: 13,
		TeamBScore: 7,
	}
	monitor.updateLastScore(score1)

	// Same scores but different teams should not be duplicate
	score2 := &ParsedScore{
		TeamAName:  "spirit",
		TeamBName:  "vitality",
		TeamAScore: 13,
		TeamBScore: 7,
	}
	assert.False(t, monitor.isDuplicate(score2))
}
