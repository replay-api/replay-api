package entities

import (
	"strings"
	"testing"
	"time"

	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

func TestGenerateMatchSlug(t *testing.T) {
	playedAt := time.Date(2026, 3, 12, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		gameID     replay_common.GameIDKey
		teamA      string
		teamB      string
		playedAt   time.Time
		mapName    string
		gameNumber int
		seriesType string
		expected   string
	}{
		{
			name:     "standard CS2 match (bo1 default)",
			gameID:   "cs2",
			teamA:    "FaZe",
			teamB:    "NAVI",
			playedAt: playedAt,
			mapName:  "Mirage",
			expected: "cs2:faze-vs-navi:2026-03-12:mirage",
		},
		{
			name:     "teams sorted alphabetically",
			gameID:   "cs2",
			teamA:    "NAVI",
			teamB:    "FaZe",
			playedAt: playedAt,
			mapName:  "Mirage",
			expected: "cs2:faze-vs-navi:2026-03-12:mirage",
		},
		{
			name:     "special characters in team names",
			gameID:   "cs2",
			teamA:    "Cloud9 (C9)",
			teamB:    "G2 Esports!",
			playedAt: playedAt,
			mapName:  "dust2",
			expected: "cs2:cloud9-c9-vs-g2:2026-03-12:dust2",
		},
		{
			name:     "missing team names",
			gameID:   "cs2",
			teamA:    "",
			teamB:    "",
			playedAt: playedAt,
			mapName:  "inferno",
			expected: "cs2:unknown-vs-unknown:2026-03-12:inferno",
		},
		{
			name:     "missing map name",
			gameID:   "cs2",
			teamA:    "FaZe",
			teamB:    "NAVI",
			playedAt: playedAt,
			mapName:  "",
			expected: "cs2:faze-vs-navi:2026-03-12:unknown",
		},
		{
			name:     "zero played_at time",
			gameID:   "cs2",
			teamA:    "FaZe",
			teamB:    "NAVI",
			playedAt: time.Time{},
			mapName:  "Mirage",
			expected: "cs2:faze-vs-navi::mirage",
		},
		{
			name:     "empty game ID",
			gameID:   "",
			teamA:    "FaZe",
			teamB:    "NAVI",
			playedAt: playedAt,
			mapName:  "Mirage",
			expected: "unknown:faze-vs-navi:2026-03-12:mirage",
		},

		// --- Map Canonicalization ---
		{
			name:     "map alias de_mirage → mirage",
			gameID:   "cs2",
			teamA:    "FaZe",
			teamB:    "NAVI",
			playedAt: playedAt,
			mapName:  "de_mirage",
			expected: "cs2:faze-vs-navi:2026-03-12:mirage",
		},
		{
			name:     "map alias dust_ii → dust2",
			gameID:   "cs2",
			teamA:    "FaZe",
			teamB:    "NAVI",
			playedAt: playedAt,
			mapName:  "dust_ii",
			expected: "cs2:faze-vs-navi:2026-03-12:dust2",
		},
		{
			name:     "map alias de_dust2 → dust2",
			gameID:   "cs2",
			teamA:    "FaZe",
			teamB:    "NAVI",
			playedAt: playedAt,
			mapName:  "de_dust2",
			expected: "cs2:faze-vs-navi:2026-03-12:dust2",
		},
		{
			name:     "map alias de_ancient → ancient",
			gameID:   "cs2",
			teamA:    "FaZe",
			teamB:    "NAVI",
			playedAt: playedAt,
			mapName:  "de_ancient",
			expected: "cs2:faze-vs-navi:2026-03-12:ancient",
		},

		// --- Team Canonicalization ---
		{
			name:     "team alias Natus Vincere → navi",
			gameID:   "cs2",
			teamA:    "Natus Vincere",
			teamB:    "FaZe Clan",
			playedAt: playedAt,
			mapName:  "Mirage",
			expected: "cs2:faze-vs-navi:2026-03-12:mirage",
		},
		{
			name:     "team alias Na'Vi → navi",
			gameID:   "cs2",
			teamA:    "Na'Vi",
			teamB:    "Heroic",
			playedAt: playedAt,
			mapName:  "Nuke",
			expected: "cs2:heroic-vs-navi:2026-03-12:nuke",
		},
		{
			name:     "team alias FaZe Clan → faze",
			gameID:   "cs2",
			teamA:    "FaZe Clan",
			teamB:    "MOUZ",
			playedAt: playedAt,
			mapName:  "Vertigo",
			expected: "cs2:faze-vs-mouz:2026-03-12:vertigo",
		},

		// --- Bo3/Bo5 game numbers ---
		{
			name:       "bo3 game 1",
			gameID:     "cs2",
			teamA:      "FaZe",
			teamB:      "NAVI",
			playedAt:   playedAt,
			mapName:    "Mirage",
			gameNumber: 1,
			seriesType: "bo3",
			expected:   "cs2:faze-vs-navi:2026-03-12:mirage:bo3-g1",
		},
		{
			name:       "bo3 game 2 different map",
			gameID:     "cs2",
			teamA:      "FaZe",
			teamB:      "NAVI",
			playedAt:   playedAt,
			mapName:    "Inferno",
			gameNumber: 2,
			seriesType: "bo3",
			expected:   "cs2:faze-vs-navi:2026-03-12:inferno:bo3-g2",
		},
		{
			name:       "bo5 game 5",
			gameID:     "cs2",
			teamA:      "G2",
			teamB:      "Vitality",
			playedAt:   playedAt,
			mapName:    "Ancient",
			gameNumber: 5,
			seriesType: "bo5",
			expected:   "cs2:g2-vs-vitality:2026-03-12:ancient:bo5-g5",
		},
		{
			name:       "game number 0 omits series suffix",
			gameID:     "cs2",
			teamA:      "FaZe",
			teamB:      "NAVI",
			playedAt:   playedAt,
			mapName:    "Mirage",
			gameNumber: 0,
			seriesType: "bo3",
			expected:   "cs2:faze-vs-navi:2026-03-12:mirage",
		},
		{
			name:       "game number with empty series type defaults to bo1",
			gameID:     "cs2",
			teamA:      "FaZe",
			teamB:      "NAVI",
			playedAt:   playedAt,
			mapName:    "Mirage",
			gameNumber: 1,
			seriesType: "",
			expected:   "cs2:faze-vs-navi:2026-03-12:mirage:bo1-g1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateMatchSlug(tt.gameID, tt.teamA, tt.teamB, tt.playedAt, tt.mapName, tt.gameNumber, tt.seriesType)
			if result != tt.expected {
				t.Errorf("GenerateMatchSlug() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGenerateMatchSlug_Deterministic(t *testing.T) {
	playedAt := time.Date(2026, 3, 12, 14, 30, 0, 0, time.UTC)

	// Same match with teams in different order should produce same slug
	slug1 := GenerateMatchSlug("cs2", "FaZe", "NAVI", playedAt, "Mirage", 0, "")
	slug2 := GenerateMatchSlug("cs2", "NAVI", "FaZe", playedAt, "Mirage", 0, "")

	if slug1 != slug2 {
		t.Errorf("Expected deterministic slugs, got %q and %q", slug1, slug2)
	}
}

func TestGenerateMatchSlug_DeterministicWithSeries(t *testing.T) {
	playedAt := time.Date(2026, 3, 12, 14, 30, 0, 0, time.UTC)

	slug1 := GenerateMatchSlug("cs2", "FaZe", "NAVI", playedAt, "Mirage", 2, "bo3")
	slug2 := GenerateMatchSlug("cs2", "NAVI", "FaZe", playedAt, "Mirage", 2, "bo3")

	if slug1 != slug2 {
		t.Errorf("Expected deterministic slugs with series, got %q and %q", slug1, slug2)
	}
}

func TestGenerateMatchSlug_TeamAliasSymmetry(t *testing.T) {
	playedAt := time.Date(2026, 3, 12, 14, 30, 0, 0, time.UTC)

	// Different team name variants should produce the same slug
	slug1 := GenerateMatchSlug("cs2", "Natus Vincere", "FaZe Clan", playedAt, "Mirage", 0, "")
	slug2 := GenerateMatchSlug("cs2", "Na'Vi", "FaZe", playedAt, "Mirage", 0, "")
	slug3 := GenerateMatchSlug("cs2", "NAVI", "FaZe", playedAt, "Mirage", 0, "")

	if slug1 != slug2 || slug2 != slug3 {
		t.Errorf("Expected same slug for team aliases:\n  Natus Vincere = %q\n  Na'Vi        = %q\n  NAVI          = %q", slug1, slug2, slug3)
	}
}

func TestGenerateMatchSlug_MapAliasesConverge(t *testing.T) {
	playedAt := time.Date(2026, 3, 12, 14, 30, 0, 0, time.UTC)

	slug1 := GenerateMatchSlug("cs2", "FaZe", "NAVI", playedAt, "de_mirage", 0, "")
	slug2 := GenerateMatchSlug("cs2", "FaZe", "NAVI", playedAt, "Mirage", 0, "")
	slug3 := GenerateMatchSlug("cs2", "FaZe", "NAVI", playedAt, "mirage", 0, "")

	if slug1 != slug2 || slug2 != slug3 {
		t.Errorf("Expected same slug for map aliases:\n  de_mirage = %q\n  Mirage    = %q\n  mirage    = %q", slug1, slug2, slug3)
	}
}

func TestSlugDateVariants(t *testing.T) {
	playedAt := time.Date(2026, 3, 12, 14, 30, 0, 0, time.UTC)

	variants := SlugDateVariants(playedAt)
	if len(variants) != 3 {
		t.Fatalf("Expected 3 date variants, got %d", len(variants))
	}
	if variants[0] != "2026-03-12" {
		t.Errorf("Primary date expected 2026-03-12, got %s", variants[0])
	}
	if variants[1] != "2026-03-11" {
		t.Errorf("Previous day expected 2026-03-11, got %s", variants[1])
	}
	if variants[2] != "2026-03-13" {
		t.Errorf("Next day expected 2026-03-13, got %s", variants[2])
	}
}

func TestSlugDateVariants_ZeroTime(t *testing.T) {
	variants := SlugDateVariants(time.Time{})
	if len(variants) != 1 {
		t.Fatalf("Expected 1 variant for zero time, got %d", len(variants))
	}
	if variants[0] != "" {
		t.Errorf("Expected empty string for zero time variant, got %q", variants[0])
	}
}

func TestGenerateSlugVariants(t *testing.T) {
	playedAt := time.Date(2026, 3, 12, 14, 30, 0, 0, time.UTC)

	variants := GenerateSlugVariants("cs2", "FaZe", "NAVI", playedAt, "Mirage", 0, "")
	if len(variants) != 3 {
		t.Fatalf("Expected 3 slug variants, got %d: %v", len(variants), variants)
	}
	if !strings.Contains(variants[0], "2026-03-12") {
		t.Errorf("Primary slug should contain primary date, got %q", variants[0])
	}
	if !strings.Contains(variants[1], "2026-03-11") {
		t.Errorf("Second slug should contain previous date, got %q", variants[1])
	}
	if !strings.Contains(variants[2], "2026-03-13") {
		t.Errorf("Third slug should contain next date, got %q", variants[2])
	}
}

func TestGenerateSlugVariants_ZeroTime(t *testing.T) {
	variants := GenerateSlugVariants("cs2", "FaZe", "NAVI", time.Time{}, "Mirage", 0, "")
	if len(variants) != 1 {
		t.Fatalf("Expected 1 variant for zero time, got %d", len(variants))
	}
}

func TestUpgradeSlugWithTeamIDs(t *testing.T) {
	tests := []struct {
		name     string
		slug     string
		teamAID  string
		teamBID  string
		expected string
	}{
		{
			name:     "upgrade name-based slug to UUID-based",
			slug:     "cs2:faze-vs-navi:2026-03-12:mirage",
			teamAID:  "aaaa-1111",
			teamBID:  "bbbb-2222",
			expected: "cs2:aaaa-1111-vs-bbbb-2222:2026-03-12:mirage",
		},
		{
			name:     "team IDs sorted alphabetically",
			slug:     "cs2:faze-vs-navi:2026-03-12:mirage",
			teamAID:  "zzzz-9999",
			teamBID:  "aaaa-1111",
			expected: "cs2:aaaa-1111-vs-zzzz-9999:2026-03-12:mirage",
		},
		{
			name:     "malformed slug returns as-is",
			slug:     "invalid",
			teamAID:  "aaaa",
			teamBID:  "bbbb",
			expected: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UpgradeSlugWithTeamIDs(tt.slug, tt.teamAID, tt.teamBID)
			if result != tt.expected {
				t.Errorf("UpgradeSlugWithTeamIDs() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// --- Canonicalization unit tests ---

func TestCanonicalizeMapName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"de_mirage", "mirage"},
		{"Mirage", "mirage"},
		{"de_dust2", "dust2"},
		{"dust_ii", "dust2"},
		{"de_inferno", "inferno"},
		{"de_ancient", "ancient"},
		{"de_nuke", "nuke"},
		{"de_overpass", "overpass"},
		{"de_vertigo", "vertigo"},
		{"cs_office", "office"},
		{"", ""},
		{"    ", ""},
		{"SomeNewMap", "somenewmap"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := CanonicalizeMapName(tt.input)
			if result != tt.expected {
				t.Errorf("CanonicalizeMapName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCanonicalizeTeamName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Natus Vincere", "navi"},
		{"Na'Vi", "navi"},
		{"NAVI", "navi"},
		{"FaZe Clan", "faze"},
		{"FaZe", "faze"},
		{"G2 Esports", "g2"},
		{"Ninjas in Pyjamas", "nip"},
		{"MOUZ", "mouz"},
		{"mousesports", "mouz"},
		{"Complexity Gaming", "complexity"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := CanonicalizeTeamName(tt.input)
			if result != tt.expected {
				t.Errorf("CanonicalizeTeamName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
