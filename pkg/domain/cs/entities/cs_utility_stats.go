package entities

import common "github.com/psavelis/team-pro/replay-api/pkg/domain"

type CSUtilityIDType = int
type CSUtilityStats struct {
	PlayerUtilityStats map[common.PlayerIDType]CSUtilityStats // utilities positions etc
	TeamUtilityStats   map[TeamIDType]CSUtilityStats          // utilities positions etc

	// TotalUtilityDamage      int            // Total damage dealt by utility (if tracking)
	// UtilityDamageBreakdown  map[string]int // Damage by grenade type (if tracking)
	// TotalUtilityFrags       int            // Total Frags by utility (if tracking)
	// UtilityFragsBreakdown   map[string]int // Frags by grenade type (if tracking)
	// TotalUtilityAssists     int            // Total assists by utility (if tracking)
	// UtilityAssistsBreakdown map[string]int // Assists by grenade type (if tracking)

	// // ie: smoke <- he or
	// TotalUtilityDisabled     int            // Total players Disabled by utility (if tracking)
	// UtilityDisabledBreakdown map[string]int // Players Disabled by grenade type (if tracking)g)
	// //sugestao p talvez em vlrnt, revisar
	// TotalUtilityDamaged     int            // Total players damaged by utility (if tracking)
	// UtilityDamagedBreakdown map[string]int // Players damaged by grenade type (if tracking)
}
