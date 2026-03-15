package entities

import "strings"

// mapAliases maps all known CS2 map name variants to their canonical form.
// This ensures "de_mirage", "Mirage", "DE_MIRAGE" all produce the same slug.
var mapAliases = map[string]string{
	// Mirage
	"de_mirage": "mirage",
	"mirage":    "mirage",
	// Inferno
	"de_inferno": "inferno",
	"inferno":    "inferno",
	// Nuke
	"de_nuke": "nuke",
	"nuke":    "nuke",
	// Overpass
	"de_overpass": "overpass",
	"overpass":    "overpass",
	// Ancient
	"de_ancient": "ancient",
	"ancient":    "ancient",
	// Anubis
	"de_anubis": "anubis",
	"anubis":    "anubis",
	// Vertigo
	"de_vertigo": "vertigo",
	"vertigo":    "vertigo",
	// Train
	"de_train": "train",
	"train":    "train",
	// Dust2 variants
	"de_dust2": "dust2",
	"de_dust":  "dust2",
	"dust2":    "dust2",
	"dust_ii":  "dust2",
	"dust ii":  "dust2",
	"dustii":   "dust2",
	"dust 2":   "dust2",
	// Cache
	"de_cache": "cache",
	"cache":    "cache",
	// Cobblestone
	"de_cobblestone": "cobblestone",
	"cobblestone":    "cobblestone",
	"cobble":         "cobblestone",
	// Tuscan
	"de_tuscan": "tuscan",
	"tuscan":    "tuscan",
	// Season
	"de_season": "season",
	"season":    "season",
	// Office
	"cs_office": "office",
	"office":    "office",
	// Italy
	"cs_italy": "italy",
	"italy":    "italy",
	// Mills
	"de_mills": "mills",
	"mills":    "mills",
	// Thera
	"de_thera": "thera",
	"thera":    "thera",
	// Basalt
	"de_basalt": "basalt",
	"basalt":    "basalt",
	// Edin
	"de_edin": "edin",
	"edin":    "edin",
}

// CanonicalizeMapName normalizes a map name to its canonical form.
// Returns the canonical name if known, otherwise returns the lowercased trimmed input.
// Examples:
//
//	"de_mirage" → "mirage"
//	"Dust_II"  → "dust2"
//	"DE_NUKE"  → "nuke"
//	"unknown_map" → "unknown_map" (returned as-is, lowercased)
func CanonicalizeMapName(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return ""
	}

	if canonical, ok := mapAliases[normalized]; ok {
		return canonical
	}

	// Strip "de_" or "cs_" prefix as last resort
	for _, prefix := range []string{"de_", "cs_"} {
		if strings.HasPrefix(normalized, prefix) {
			stripped := strings.TrimPrefix(normalized, prefix)
			if canonical, ok := mapAliases[stripped]; ok {
				return canonical
			}
			return stripped
		}
	}

	return normalized
}
