package entities

import common "github.com/replay-api/replay-api/pkg/domain"

type CSStrategyIDType = int

const (
	StrategyUnknown CSStrategyIDType = iota

	// CT-Side Strategies
	StrategyFullSave      //  Full save, no buys
	StrategyForceBuy      //  Force buy with whatever money available
	StrategyEco           //  Partial save, minimal buys
	StrategyDefault       //  Standard setup, defending both sites
	StrategySemiDefault   // One player aggressing, rest on default
	StrategyBDefault      // Default with focus on B site
	StrategyADefault      // Default with focus on A site
	StrategyStackA        // Aggressive defense, most players on A
	StrategyStackB        // Aggressive defense, most players on B
	StrategyLateStackA    // Late round rotation to stack A
	StrategyLateStackB    // Late round rotation to stack B
	StrategyDoubleStackA  // Very aggressive, almost all players on A
	StrategyDoubleStackB  // Very aggressive, almost all players on B
	StrategySplitA        // Two groups, one on A, one elsewhere
	StrategySplitB        // Two groups, one on B, one elsewhere
	StrategyAAggressive   // One player aggressively holding A site
	StrategyBAggressive   // One player aggressively holding B site
	StrategyMidAggressive // One player aggressively holding mid
	StrategyAnchorA       // Passive A site defense (one player holding)
	StrategyAnchorB       // Passive B site defense (one player holding)
	StrategyRotate        // Planned rotation between bombsites
	StrategyAntiEco       // Strategy specifically designed to counter an eco round

	// ... Add more CT-side strategies

	// T-Side Strategies
	StrategyRushA      // All-out rush towards A site
	StrategyRushB      // All-out rush towards B site
	StrategyRushMid    // All-out rush towards mid
	StrategyFastA      // Quick, coordinated A site take
	StrategyFastB      // Quick, coordinated B site take
	StrategyDefaultA   // Standard A site take with post-plant setup
	StrategyDefaultB   // Standard B site take with post-plant setup
	StrategySlowA      // Slower, more methodical A site take
	StrategySlowB      // Slower, more methodical B site take
	StrategyFakeA      // Fake an A attack to lure defenders away
	StrategyFakeB      // Fake a B attack to lure defenders away
	StrategyFakeMid    // Fake a mid push to draw attention away
	StrategyExecuteA   // Late-round A site execute with utility
	StrategyExecuteB   // Late-round B site execute with utility
	StrategyContactA   // Early aggression on A to get info/picks
	StrategyContactB   // Early aggression on B to get info/picks
	StrategyContactMid // Early aggression in mid to gain control
	StrategyRetakeA    // Attempt to retake A site after plant
	StrategyRetakeB    // Attempt to retake B site after plant

	// More Granular Variations (examples):
	StrategyFastALong           // Fast A using long control
	StrategyFastAShort          // Fast A using short/catwalk
	StrategyDefaultAMapSpecific // Default A tailored to a specific map (e.g., Dust2 A Long)
	StrategyDefaultBMapSpecific // Default B tailored to a specific map (e.g., Mirage B Apartments)
	StrategySplitAMapSpecific   // Split A tailored to a specific map (e.g., Inferno Apts + Arch)
	StrategySplitBMapSpecific   // Split B tailored to a specific map (e.g., Overpass B Short + Monster)
	StrategyFakeAMapSpecific    // Fake A tailored to a specific map (e.g., Mirage A Fake)
	StrategyFakeBMapSpecific    // Fake B tailored to a specific map (e.g., Inferno B Fake)
	StrategyExecuteAMapSpecific // Execute A tailored to a specific map (e.g., Train A Site)
	StrategyExecuteBMapSpecific // Execute B tailored to a specific map (e.g., Nuke B Site)

	// Additional Strategic Considerations
	StrategySave               // Saving weapons and utility for later rounds
	StrategyBait               // Sacrificing a player to gain information or create an advantage
	StrategyPick               // Focusing on getting early kills to gain map control
	StrategySlowControl        // Slowly gaining map control with minimal casualties
	StrategyAggro              // High-pressure, aggressive style with lots of duels
	StrategyPassive            // Slower, more defensive style focused on holding angles
	StrategyFake               // Faking an attack to draw out utility or defenders
	StrategyContact            // Aggressive playstyle focused on early engagements
	StrategyExecute            // Coordinated execution of a strategy with utility usage
	StrategyRetake             // Attempting to retake a bombsite after the bomb has been planted
	EconomicStrategyDefaultEco // Economic strategy focused on maximizing utility with minimal buys
	EconomicStrategyForceBuy   // Buying with whatever money is available, regardless of the situation
	EconomicStrategyFullBuy    // Full buy with rifles, armor, and utility
	EconomicStrategyHalfBuy    // Half buy with some rifles and minimal utility
	EconomicStrategyFullSave   // Full save with no buys or investments
	EconomicStrategyMixedBuy   // Mixed buy with a mix of rifles, armor, and utility
)

type CSStrategyStats struct {
	TickID              common.TickIDType
	PlayerStrategyStats map[common.PlayerIDType]CSPlayerStrategyStats
}

type CSPlayerStrategyStats struct {
	Strategy         CSStrategyIDType
	WinRate          float64
	WinRateBreakdown map[CSStrategyIDType]float64

	PlantRate     float64
	DefuseRate    float64
	EntryFragRate float64
	TradeFragRate float64
	ClutchRate    float64
	ClutchWinRate float64
}
