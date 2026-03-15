package entities

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

var slugCleanRegex = regexp.MustCompile(`[^\p{L}\p{N}\-]`)
var slugMultiDash = regexp.MustCompile(`-{2,}`)

// normalizeSlugPart converts a string to a slug-safe format:
// lowercase, replace non-letter/non-digit/non-hyphen with hyphens, collapse multiple hyphens.
// Preserves unicode letters (CJK, Cyrillic, etc.) — they pass through as valid slug components.
func normalizeSlugPart(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugCleanRegex.ReplaceAllString(s, "-")
	s = slugMultiDash.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// GenerateMatchSlug creates a reconciliation slug for a match.
//
// Format: {game}:{teamA}-vs-{teamB}:{date}:{map}[:{seriesType}-g{gameNumber}]
// Teams are canonicalized and sorted alphabetically for determinism.
// Map names are canonicalized (e.g., "de_mirage" → "mirage").
//
// Examples:
//   - "cs2:faze-vs-navi:2026-03-12:mirage"                (bo1 or unknown series)
//   - "cs2:faze-vs-navi:2026-03-12:mirage:bo3-g1"         (game 1 of a bo3)
//   - "cs2:faze-vs-navi:2026-03-12:inferno:bo3-g2"        (game 2 of a bo3)
//   - "cs2:unknown-vs-unknown:2026-03-12:unknown"          (teams/map unavailable)
//
// If playedAt is zero, DATE is omitted → "cs2:faze-vs-navi::mirage"
// If mapName is empty, "unknown" is used.
// gameNumber and seriesType are only appended when gameNumber > 0.
func GenerateMatchSlug(
	gameID replay_common.GameIDKey,
	teamAName string,
	teamBName string,
	playedAt time.Time,
	mapName string,
	gameNumber int,
	seriesType string,
) string {
	game := normalizeSlugPart(string(gameID))
	if game == "" {
		game = "unknown"
	}

	// Canonicalize team names before slug normalization
	teamACan := CanonicalizeTeamName(teamAName)
	teamBCan := CanonicalizeTeamName(teamBName)

	teamA := normalizeSlugPart(teamACan)
	teamB := normalizeSlugPart(teamBCan)

	if teamA == "" {
		teamA = "unknown"
	}
	if teamB == "" {
		teamB = "unknown"
	}

	// Sort alphabetically for consistent slug generation regardless of team order
	teams := []string{teamA, teamB}
	sort.Strings(teams)

	dateStr := ""
	if !playedAt.IsZero() {
		dateStr = playedAt.UTC().Format("2006-01-02")
	}

	// Canonicalize map name (de_mirage → mirage, dust_ii → dust2)
	canonMap := CanonicalizeMapName(mapName)
	mapSlug := normalizeSlugPart(canonMap)
	if mapSlug == "" {
		mapSlug = "unknown"
	}

	base := fmt.Sprintf("%s:%s-vs-%s:%s:%s", game, teams[0], teams[1], dateStr, mapSlug)

	// Append series/game number only when gameNumber > 0
	if gameNumber > 0 {
		series := strings.ToLower(strings.TrimSpace(seriesType))
		if series == "" {
			series = "bo1"
		}
		base = fmt.Sprintf("%s:%s-g%d", base, series, gameNumber)
	}

	return base
}

// SlugDateVariants returns the primary date slug plus ±1 day variants.
// This handles midnight UTC boundary issues where two sources may record
// the same match on adjacent dates (e.g., 23:50 UTC vs 00:10 UTC).
func SlugDateVariants(playedAt time.Time) []string {
	if playedAt.IsZero() {
		return []string{""}
	}
	primary := playedAt.UTC().Format("2006-01-02")
	prev := playedAt.UTC().AddDate(0, 0, -1).Format("2006-01-02")
	next := playedAt.UTC().AddDate(0, 0, 1).Format("2006-01-02")
	return []string{primary, prev, next}
}

// GenerateSlugVariants generates all slug variants for a match, including date variants.
// Returns primary slug first, then date-shifted variants.
// This is used by the reconciliation service for fuzzy date matching.
func GenerateSlugVariants(
	gameID replay_common.GameIDKey,
	teamAName string,
	teamBName string,
	playedAt time.Time,
	mapName string,
	gameNumber int,
	seriesType string,
) []string {
	primary := GenerateMatchSlug(gameID, teamAName, teamBName, playedAt, mapName, gameNumber, seriesType)
	if playedAt.IsZero() {
		return []string{primary}
	}

	variants := []string{primary}
	dates := SlugDateVariants(playedAt)
	// dates[0] is the primary date — skip it, already in primary slug
	for _, d := range dates[1:] {
		shifted, _ := time.Parse("2006-01-02", d)
		variantSlug := GenerateMatchSlug(gameID, teamAName, teamBName, shifted, mapName, gameNumber, seriesType)
		if variantSlug != primary {
			variants = append(variants, variantSlug)
		}
	}

	return variants
}

// UpgradeSlugWithTeamIDs replaces the name-based team portions of a slug with UUID-based identifiers.
// This is used when teams are later resolved/matched to platform team entities.
//
// Input:  "cs2:faze-vs-navi:2026-03-12:mirage"
// Output: "cs2:{teamAID}-vs-{teamBID}:2026-03-12:mirage"
func UpgradeSlugWithTeamIDs(slug string, teamAID string, teamBID string) string {
	parts := strings.SplitN(slug, ":", 4)
	if len(parts) < 4 {
		return slug
	}

	// Sort team IDs alphabetically for consistency
	ids := []string{teamAID, teamBID}
	sort.Strings(ids)

	return fmt.Sprintf("%s:%s-vs-%s:%s:%s", parts[0], ids[0], ids[1], parts[2], parts[3])
}
