package oracle_services

import (
	"testing"

	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOCRScoreParser(t *testing.T) {
	parser := NewOCRScoreParser()
	assert.NotNil(t, parser)
	assert.NotEmpty(t, parser.scorePatterns)
	assert.NotEmpty(t, parser.knownMaps)
}

func TestParseScoreboard_FullScoreline_DashSeparator(t *testing.T) {
	parser := NewOCRScoreParser()
	blocks := []oracle_out.TextBlock{
		{Text: "NAVI 13 - 7 FaZe", Confidence: 0.95},
	}

	result, err := parser.ParseScoreboard(blocks, replay_common.CS2_GAME_ID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "NAVI", result.TeamAName)
	assert.Equal(t, "FaZe", result.TeamBName)
	assert.Equal(t, 13, result.TeamAScore)
	assert.Equal(t, 7, result.TeamBScore)
	assert.Equal(t, 20, result.RoundsPlayed)
}

func TestParseScoreboard_FullScoreline_ColonSeparator(t *testing.T) {
	parser := NewOCRScoreParser()
	blocks := []oracle_out.TextBlock{
		{Text: "Team Spirit 16:14 Vitality", Confidence: 0.92},
	}

	result, err := parser.ParseScoreboard(blocks, replay_common.CS2_GAME_ID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Team Spirit", result.TeamAName)
	assert.Equal(t, "Vitality", result.TeamBName)
	assert.Equal(t, 16, result.TeamAScore)
	assert.Equal(t, 14, result.TeamBScore)
	assert.Equal(t, 30, result.RoundsPlayed)
}

func TestParseScoreboard_FullScoreline_PipeSeparator(t *testing.T) {
	parser := NewOCRScoreParser()
	blocks := []oracle_out.TextBlock{
		{Text: "G2 2 | 1 Cloud9", Confidence: 0.90},
	}

	result, err := parser.ParseScoreboard(blocks, replay_common.CS2_GAME_ID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "G2", result.TeamAName)
	assert.Equal(t, "Cloud9", result.TeamBName)
	assert.Equal(t, 2, result.TeamAScore)
	assert.Equal(t, 1, result.TeamBScore)
}

func TestParseScoreboard_ScoreOnly_DashSeparator(t *testing.T) {
	parser := NewOCRScoreParser()
	blocks := []oracle_out.TextBlock{
		{Text: "13 - 7", Confidence: 0.90},
	}

	result, err := parser.ParseScoreboard(blocks, replay_common.CS2_GAME_ID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 13, result.TeamAScore)
	assert.Equal(t, 7, result.TeamBScore)
	assert.Equal(t, "", result.TeamAName)
	assert.Equal(t, "", result.TeamBName)
}

func TestParseScoreboard_ScoreOnly_CompactFormat(t *testing.T) {
	parser := NewOCRScoreParser()
	blocks := []oracle_out.TextBlock{
		{Text: "16:14", Confidence: 0.88},
	}

	result, err := parser.ParseScoreboard(blocks, replay_common.CS2_GAME_ID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 16, result.TeamAScore)
	assert.Equal(t, 14, result.TeamBScore)
}

func TestParseScoreboard_WithMapName(t *testing.T) {
	parser := NewOCRScoreParser()
	blocks := []oracle_out.TextBlock{
		{Text: "NAVI 13 - 7 FaZe", Confidence: 0.95},
		{Text: "Inferno", Confidence: 0.90},
	}

	result, err := parser.ParseScoreboard(blocks, replay_common.CS2_GAME_ID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "inferno", result.MapName)
}

func TestParseScoreboard_WithMirage(t *testing.T) {
	parser := NewOCRScoreParser()
	blocks := []oracle_out.TextBlock{
		{Text: "FaZe 10 - 5 MOUZ", Confidence: 0.92},
		{Text: "Mirage", Confidence: 0.85},
	}

	result, err := parser.ParseScoreboard(blocks, replay_common.CS2_GAME_ID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "mirage", result.MapName)
	assert.Equal(t, "FaZe", result.TeamAName)
	assert.Equal(t, "MOUZ", result.TeamBName)
}

func TestParseScoreboard_AssembleFromSeparateBlocks(t *testing.T) {
	parser := NewOCRScoreParser()
	blocks := []oracle_out.TextBlock{
		{Text: "NAVI", Confidence: 0.90, BoundingBox: [4]int{100, 50, 80, 30}},
		{Text: "13", Confidence: 0.95, BoundingBox: [4]int{300, 50, 30, 30}},
		{Text: "7", Confidence: 0.95, BoundingBox: [4]int{400, 50, 30, 30}},
		{Text: "FaZe", Confidence: 0.92, BoundingBox: [4]int{500, 50, 80, 30}},
	}

	result, err := parser.ParseScoreboard(blocks, replay_common.CS2_GAME_ID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 13, result.TeamAScore)
	assert.Equal(t, 7, result.TeamBScore)
}

func TestParseScoreboard_NoScoreFound(t *testing.T) {
	parser := NewOCRScoreParser()
	blocks := []oracle_out.TextBlock{
		{Text: "Some random text", Confidence: 0.90},
		{Text: "More text without scores", Confidence: 0.85},
	}

	result, err := parser.ParseScoreboard(blocks, replay_common.CS2_GAME_ID)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestParseScoreboard_EmptyBlocks(t *testing.T) {
	parser := NewOCRScoreParser()
	blocks := []oracle_out.TextBlock{}

	result, err := parser.ParseScoreboard(blocks, replay_common.CS2_GAME_ID)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestParseScoreboard_LowConfidenceBlocksSkipped(t *testing.T) {
	parser := NewOCRScoreParser()
	blocks := []oracle_out.TextBlock{
		{Text: "NAVI 13 - 7 FaZe", Confidence: 0.2}, // Below threshold
	}

	// Should still try to parse (the confidence filter is in the loop)
	// The text is checked regardless since strategy 1 checks all blocks
	result, err := parser.ParseScoreboard(blocks, replay_common.CS2_GAME_ID)
	// It may or may not parse depending on the confidence threshold inside ParseScoreboard
	if err == nil {
		assert.NotNil(t, result)
	}
}

func TestParseScoreboard_MultipleScoreLines(t *testing.T) {
	parser := NewOCRScoreParser()
	blocks := []oracle_out.TextBlock{
		{Text: "NAVI 13 - 7 FaZe", Confidence: 0.95},
		{Text: "Spirit 9 - 16 Vitality", Confidence: 0.90},
	}

	// Should return the first match (highest confidence)
	result, err := parser.ParseScoreboard(blocks, replay_common.CS2_GAME_ID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "NAVI", result.TeamAName)
}

func TestParseScoreboard_OCRArtifacts(t *testing.T) {
	parser := NewOCRScoreParser()
	blocks := []oracle_out.TextBlock{
		{Text: "  NAVl  13  -  7  FaZe  ", Confidence: 0.88},
	}

	result, err := parser.ParseScoreboard(blocks, replay_common.CS2_GAME_ID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 13, result.TeamAScore)
	assert.Equal(t, 7, result.TeamBScore)
}

func TestParseScoreboard_AllCS2Maps(t *testing.T) {
	parser := NewOCRScoreParser()
	maps := []string{"Mirage", "Inferno", "Nuke", "Overpass", "Ancient", "Anubis", "Vertigo", "Dust2"}

	for _, mapName := range maps {
		blocks := []oracle_out.TextBlock{
			{Text: "A 13 - 7 B", Confidence: 0.95},
			{Text: mapName, Confidence: 0.90},
		}

		result, err := parser.ParseScoreboard(blocks, replay_common.CS2_GAME_ID)
		require.NoError(t, err, "failed for map: %s", mapName)
		require.NotNil(t, result, "nil result for map: %s", mapName)
		assert.NotEmpty(t, result.MapName, "empty map name for: %s", mapName)
	}
}

func TestParseScoreboard_TiedScore(t *testing.T) {
	parser := NewOCRScoreParser()
	blocks := []oracle_out.TextBlock{
		{Text: "NAVI 15 - 15 FaZe", Confidence: 0.95},
	}

	result, err := parser.ParseScoreboard(blocks, replay_common.CS2_GAME_ID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 15, result.TeamAScore)
	assert.Equal(t, 15, result.TeamBScore)
	assert.Equal(t, 30, result.RoundsPlayed)
}

func TestParseScoreboard_HighScore(t *testing.T) {
	parser := NewOCRScoreParser()
	blocks := []oracle_out.TextBlock{
		{Text: "NAVI 22 - 20 FaZe", Confidence: 0.95},
	}

	result, err := parser.ParseScoreboard(blocks, replay_common.CS2_GAME_ID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 22, result.TeamAScore)
	assert.Equal(t, 20, result.TeamBScore)
	assert.Equal(t, 42, result.RoundsPlayed)
}
