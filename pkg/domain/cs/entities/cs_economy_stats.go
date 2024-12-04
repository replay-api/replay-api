package entities

type CSEconomyState string

const (
	CSEconomyStateUndefined CSEconomyState = "Undefined"

	// Pistol Rounds
	CSEconomyStatePistolRound CSEconomyState = "PistolRound" // First round of each half

	// Full Buys
	CSEconomyStateFullBuy       CSEconomyState = "FullBuy"       // Full equipment and utilities
	CSEconomyStateFullBuySecond CSEconomyState = "FullBuySecond" // After losing pistol, but enough for full buy

	// Partial Buys
	CSEconomyStateHalfBuy        CSEconomyState = "HalfBuy"        // Some rifles/armor, maybe nades
	CSEconomyStateHalfBuySecond  CSEconomyState = "HalfBuySecond"  // After losing pistol, but enough for half buy
	CSEconomyStateHalfBuyUpgrade CSEconomyState = "HalfBuyUpgrade" // Upgrading from a previous half buy

	// Forced Buys
	CSEconomyStateForceBuy CSEconomyState = "ForceBuy" // All-in after losing eco (compare with remaining money)

	// Eco Rounds
	CSEconomyStateEco       CSEconomyState = "Eco"       // Saving for future rounds
	CSEconomyStateEcoSecond CSEconomyState = "EcoSecond" // Second eco in a row
	CSEconomyStateAntiEco   CSEconomyState = "AntiEco"   // Decent equipment despite being on eco

	// Other
	CSEconomyStateSave  CSEconomyState = "Save"  // Not buying much, but not a full eco
	CSEconomyStateMixed CSEconomyState = "Mixed" // Some players full buy, others eco/save
)

type CSTeamEconomyStats struct {
	State     CSEconomyState
	LossBonus StatNumberUnit // review: CSItemStats // rotacionando por round encaixa perfeitamente. (granularity/round)
	CSEconomyStats
}

func NewCSTeamEconomyStats() *CSTeamEconomyStats {
	return &CSTeamEconomyStats{
		State: CSEconomyStateUndefined,
		LossBonus: StatNumberUnit{
			Count: 0,
			Sum:   0,
		},
	}
}

type CSEconomyStats struct {
	Item   CSInventoryStats
	Budget CSTeamBudgetStats
}

type StatFragUnit struct {
	Score        StatNumberUnit
	Assist       StatNumberUnit
	Entry        StatNumberUnit
	Exit         StatNumberUnit
	Trade        StatNumberUnit
	Concede      StatNumberUnit
	Witness      StatNumberUnit
	Rescue       StatNumberUnit
	SelfDropOrWO StatNumberUnit
}

// TODO: dividir para contexto de item? (item/frag/objective??)
type CSPurchasedItemPerformanceStats struct {
	Frags      StatFragUnit
	Objectives map[CSObjectiveKey]ObjectiveUnit
	Hits       map[CSHitBoxType]StatNumberUnit
	Damage     StatNumberUnit
	Reward     StatNumberUnit
	// Losses     CSItemCostStats

}

type CSObjectiveKey int

const (
	All CSObjectiveKey = iota
	Plant
	Detonate
	Defuse
	Wipe
	Timeout

	// alts...
	Disable
	Storm
	Entry
	Count
	Hold
	Pick
	Trade
	Cover
	Fake
	Distract
	Lure
	Rotate
	Mock
	Retake
	Locate
	Boost
	Block
	Gather
	Lurk
	Dispose
	Hijack
	Anticipate
	Chase
	Flee
	Hide
	Eliminate
	Dominate
)

type ObjectiveUnit struct {
	All          StatTimeUnit
	Priorization StatTimeUnit
	Initiation   StatTimeUnit
	Continuation StatTimeUnit
	Interruption StatTimeUnit
	Expiration   StatTimeUnit
	Deprecation  StatTimeUnit
	Completion   StatTimeUnit
}

type CSItemCostStats struct {
	All       StatNumberUnit
	Purchased StatNumberUnit
	Donated   StatNumberUnit
	Saved     StatNumberUnit
	Acquired  StatNumberUnit
}

type CSItemStats struct {
	Values   CSItemCostStats
	Outcomes CSPurchasedItemPerformanceStats
	Losses   CSPurchasedItemPerformanceStats
}

type CSItemUtilityEconomyStats struct {
	All   CSItemStats
	HE    CSItemStats
	Flash CSItemStats
	Smoke CSItemStats
	Decoy CSItemStats
}

type CSItemArmorEconomyStats struct {
	All    CSItemStats
	Helmet CSItemStats // Number of players with full armor
	Kevlar CSItemStats // Number of players with kevlar only
}

type CSItemWeaponEconomyStats struct {
	All     CSItemStats
	Rifles  CSItemStats // Number of rifles purchased in the context (round or match)
	SMGs    CSItemStats // Number of SMGs purchased in the context (round or match)
	Pistols CSItemStats // Number of pistols purchased in the context (round or match)
}

type CSInventoryStats struct {
	All     CSItemStats
	Armor   CSItemArmorEconomyStats   // Total spent on Armor
	Utility CSItemUtilityEconomyStats // Total spent on grenades/utility
	Weapon  CSItemWeaponEconomyStats  // Total spent on Weapons
	Kit     CSItemStats               // Number of defuse kits purchased
}

type CSTeamBudgetStats struct {
	Current  StatNumberUnit
	Earned   StatNumberUnit // when upgrading
	Spend    StatNumberUnit
	Saved    StatNumberUnit // when keeping kev, sub, utility, heavy/rif
	Lost     StatNumberUnit
	Conceded StatNumberUnit
}

// PlayerEquipmentValues               map[PlayerIDType]float64 // Equipment value per player (nao utilizar steamID, utilizar playerID / internal ID da replay msm para anonimizar)
// WeaponTypeCounts                    map[string]float64       // Count of each weapon type purchased (e.g., "AK-47": 2)
// NadeTypeCounts                      map[string]float64       // Count of each grenade type (e.g., "Flashbang": 3)
// MostExpensiveItem                   string               // Name of the most expensive item bought
// MostExpensiveItemCost               float64                  // Cost of the most expensive item bought
// TotalMoney                          float64                  // Total money available to the context (files, matches, rounds, teams, players, maps, etc.))
// TotalMoneySpent                     float64                  // Total money spent by the context (files, matches, rounds, teams, players, maps, etc.))
// TotalMoneyLost                      float64                  // Total money lost by the context (files, matches, rounds, teams, players, maps, etc.))
// TotalMoneyEarned                    float64                  // Total money earned by the context (files, matches, rounds, teams, players, maps, etc.))
// TotalMoneySaved                     float64                  // Total money saved by the context (files, matches, rounds, teams, players, maps, etc.))
// TotalMoneyAvailable                 float64                  // Total money available to the context (files, matches, rounds, teams, players, maps, etc.))
// TotalMoneySpentOnWeapons            float64                  // Total money spent on weapons
// TotalMoneySpentOnUtility            float64                  // Total money spent on utility
// TotalMoneySpentOnArmor              float64                  // Total money spent on armor
// TotalMoneySpentOnKits               float64                  // Total money spent on defuse kits
// TotalMoneySpentOnAmmo               float64                  // Total money spent on ammo
// TotalMoneySpentOnOther              float64                  // Total money spent on other items
// TotalMoneySpentOnMostExpensiveItem  float64                  // Total money spent on the most expensive item
// TotalMoneySpentOnLeastExpensiveItem float64                  // Total money spent on the least expensive item

// // 80/20 metrics
// TotalMoneySpentOnTop20PercentItems   float64 // Total money spent on the top 20% of items
// TotalMoneySpentOnTop20PercentUtility float64 // Total money spent on the top 20% of utility
// TotalMoneySpentOnTop20PercentWeapons float64 // Total money spent on the top 20% of weapons

// // 80/20 counts
// TotalTop20PercentItemsPurchased   float64 // Number of items purchased in the top 20%
// TotalTop20PercentUtilityPurchased float64 // Number of utility items purchased in the top 20%
// TotalTop20PercentWeaponsPurchased float64 // Number of weapons purchased in the top 20%
