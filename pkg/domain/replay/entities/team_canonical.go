package entities

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// teamAliases maps common team name variants to a canonical short name.
// This ensures "Natus Vincere", "NaVi", "Na'Vi" all produce the same slug component.
var teamAliases = map[string]string{
	// NAVI
	"natus vincere": "navi",
	"natusvincere":  "navi",
	"navi":          "navi",
	"navi junior":   "navi-jr",
	// FaZe
	"faze clan": "faze",
	"faze":      "faze",
	// G2
	"g2 esports": "g2",
	"g2":         "g2",
	// Cloud9
	"cloud9":  "c9",
	"cloud 9": "c9",
	"c9":      "c9",
	// Vitality
	"team vitality": "vitality",
	"vitality":      "vitality",
	// Spirit
	"team spirit": "spirit",
	"spirit":      "spirit",
	// Heroic
	"heroic": "heroic",
	// MOUZ
	"mouz":        "mouz",
	"mousesports": "mouz",
	// Complexity
	"complexity gaming": "complexity",
	"complexity":        "complexity",
	"col":               "complexity",
	// Liquid
	"team liquid": "liquid",
	"liquid":      "liquid",
	// NIP
	"ninjas in pyjamas": "nip",
	"ninjas in pajamas": "nip",
	"nip":               "nip",
	// Virtus.pro
	"virtuspro":  "virtuspro",
	"virtus pro": "virtuspro",
	"vp":         "virtuspro",
	// Astralis
	"astralis": "astralis",
	// ENCE
	"ence":         "ence",
	"ence esports": "ence",
	// BIG
	"big":      "big",
	"big clan": "big",
	// Fnatic
	"fnatic": "fnatic",
	// Eternal Fire
	"eternal fire": "eternalfire",
	"eternalfire":  "eternalfire",
	"ef":           "eternalfire",
	// paiN
	"pain":        "pain",
	"pain gaming": "pain",
	// FURIA
	"furia":         "furia",
	"furia esports": "furia",
	// Monte
	"monte": "monte",
	// SAW
	"saw": "saw",
	// GamerLegion
	"gamerlegion":  "gamerlegion",
	"gamer legion": "gamerlegion",
	// Apeks
	"apeks": "apeks",
	// MIBR
	"mibr":           "mibr",
	"made in brazil": "mibr",
	// 9z
	"9z":      "9z",
	"9z team": "9z",
	// Imperial
	"imperial":         "imperial",
	"imperial esports": "imperial",
	// TheMongolz
	"the mongolz": "mongolz",
	"themongolz":  "mongolz",
	"mongolz":     "mongolz",
	// Rare Atom
	"rare atom": "rareatom",
	"rareatom":  "rareatom",
	// Lynn Vision
	"lynn vision": "lynnvision",
	"lynnvision":  "lynnvision",
	// Tyloo
	"tyloo": "tyloo",
}

// teamSuffixes are stripped from team names during canonicalization.
var teamSuffixes = []string{
	" esports",
	" gaming",
	" team",
	" club",
	" org",
	" gg",
	" fe",
}

// nonAlphanumUnicode matches characters that are NOT unicode letters, digits, spaces, or hyphens.
var nonAlphanumUnicode = regexp.MustCompile(`[^\p{L}\p{N}\s\-]`)

// multiSpaceDash collapses runs of whitespace/hyphens.
var multiSpaceDash = regexp.MustCompile(`[\s\-]+`)

// CanonicalizeTeamName normalizes a team name for slug generation.
// Resolution order:
//  1. Unicode NFKD normalization + strip combining marks
//  2. Exact match in alias table
//  3. Suffix stripping + re-check alias table
//  4. Return cleaned name (preserves unicode letters like CJK/Cyrillic)
//
// Examples:
//
//	"Natus Vincere" → "navi"
//	"FaZe Clan"     → "faze"
//	"Team Liquid"   → "liquid"
//	"Unknown Team"  → "unknown" (suffix stripped)
//	"지누스"         → "지누스" (Korean preserved)
func CanonicalizeTeamName(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}

	// Step 1: Unicode NFKD normalization — decomposes characters
	normalized := norm.NFKD.String(raw)

	// Strip combining marks (accents) but keep base characters
	cleaned := strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) { // Mn = Mark, nonspacing (combining marks)
			return -1
		}
		return r
	}, normalized)

	// Lowercase
	cleaned = strings.ToLower(strings.TrimSpace(cleaned))

	// Remove non-alphanumeric (but keep unicode letters, digits, spaces, hyphens)
	cleaned = nonAlphanumUnicode.ReplaceAllString(cleaned, "")
	cleaned = strings.TrimSpace(cleaned)

	// Step 2: Check alias table
	if canonical, ok := teamAliases[cleaned]; ok {
		return canonical
	}

	// Step 3: Try stripping suffixes
	stripped := cleaned
	for _, suffix := range teamSuffixes {
		stripped = strings.TrimSuffix(stripped, suffix)
	}
	stripped = strings.TrimSpace(stripped)

	if stripped != cleaned {
		if canonical, ok := teamAliases[stripped]; ok {
			return canonical
		}
	}

	// Step 4: Return cleaned version — collapse spaces/hyphens to single hyphen
	result := multiSpaceDash.ReplaceAllString(stripped, "-")
	result = strings.Trim(result, "-")

	if result == "" {
		return ""
	}

	return result
}
