package state

import (

	// infocs "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
	infocs "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
	cs_entity "github.com/replay-api/replay-api/pkg/domain/cs/entities"
)

// config/params

type CSEconomyContext struct {
	State cs_entity.CSEconomyState
}

func NewCSEconomyContext(team []*infocs.Player) *CSEconomyContext {
	return &CSEconomyContext{
		State: DetermineBuyType(team),
		// TotalSpend
		// TotalUtilitySpend
		// TotalRiflesPurchased
	}
}

func DetermineBuyType(team []*infocs.Player) cs_entity.CSEconomyState {
	totalCost := 0.0
	numPrimaries := 0
	numUtility := 0
	numArmor := 0
	numPlayers := len(team)

	for _, player := range team {
		totalCost += float64(player.EquipmentValueCurrent())

		if player.Armor() > 0 {
			numArmor++
		}

		if player.HasHelmet() {
			numArmor++
		}

		for _, equipment := range player.Weapons() {
			if equipment.Class() == infocs.EqClassGrenade {
				numUtility++
			}

			if equipment.Class() == infocs.EqClassRifle || equipment.Class() == infocs.EqClassSMG {
				numPrimaries++
			}
		}
	}

	avgCost := totalCost / float64(numPlayers)

	// Calculate proportional thresholds
	fullBuyPrimaries := numPlayers
	fullBuyUtility := numPlayers
	fullBuyArmor := 2 * numPlayers
	forceBuyPrimaries := (numPlayers + 1) / 2
	halfBuyPrimaries := (numPlayers + 2) / 3
	halfBuyUtility := numPlayers / 2

	// CSEconomyStateFullBuy
	if avgCost > 2500 && numPrimaries >= fullBuyPrimaries && numUtility >= fullBuyUtility && numArmor >= fullBuyArmor {
		return cs_entity.CSEconomyStateFullBuy
	}
	// CSEconomyStateHalfBuy
	if avgCost > 1500 && numPrimaries >= halfBuyPrimaries && numUtility >= halfBuyUtility {
		return cs_entity.CSEconomyStateHalfBuy
	}

	// CSEconomyStateForceBuy
	if avgCost > 1000 && numPrimaries >= forceBuyPrimaries {
		return cs_entity.CSEconomyStateForceBuy
	}

	// CSEconomyStateEco and CSEconomyStateAntiEco
	if avgCost < 500 && numArmor == 0 {
		return cs_entity.CSEconomyStateEco
	} else if avgCost < 500 && numArmor <= numPlayers {
		return cs_entity.CSEconomyStateAntiEco
	}

	// Fallback (Semi-Eco or other categories)
	return cs_entity.CSEconomyStateSave
}

// Calculates and populates the CSEconomyStats for a given CS2MatchContext
// func (r *CS2RoundContext) CalculateEconomyStats(team *infocs.TeamState) *cs_entity.CSEconomyStats {
// stats := &cs_entity.CSEconomyStats{
// 	TotalWeapons: cs_entity.CSItemStats{
// 		Count:  0,
// 		Price:  cs_entity.CSEconomyTotalizer{},
// 		Frags:  cs_entity.CSEconomicReturn{},
// 		Damage: cs_entity.CSEconomyTotalizer{},
// 		Reward: cs_entity.CSEconomyTotalizer{},
// 	},
// }

// players := team.Members()
// playerCosts := make([]int, len(players))
// weaponCosts := make([]int, 0)
// utilityCosts := make([]int, 0)
// damageValues := make([]int, 0)
// rewardValues := make([]int, 0)

// for i, player := range players {
// 	playerCost := player.EquipmentValueCurrent()
// 	playerCosts[i] = playerCost
// 	stats.TotalSpend.Price.Total += playerCost
// 	stats.TotalSpend.Count++

// 	for _, weapon := range player.Weapons() {
// 		weaponCost := weapon.Price()

// 		switch weapon.Class() {
// 		case infocs.EqClassRifle, infocs.EqClassSMG:
// 			stats.TotalRifles.Price.Total += weaponCost
// 			stats.TotalRifles.Count++
// 			weaponCosts = append(weaponCosts, weaponCost)
// 		case infocs.EqClassPistols:
// 			stats.TotalPistols.Price.Total += weaponCost
// 			stats.TotalPistols.Count++
// 			weaponCosts = append(weaponCosts, weaponCost)
// 		case infocs.EqClassGrenade:
// 			stats.TotalUtilities.Price.Total += weaponCost
// 			stats.TotalUtilities.Count++
// 			utilityCosts = append(utilityCosts, weaponCost)
// 			nadeName := weapon.String()
// 			stats.NadeTypeCounts[nadeName]++
// 			switch weapon {
// 			case infocs.EqHE:
// 				stats.TotalHEPurchased.Price.Total += weaponCost
// 				stats.TotalHEPurchased.Count++
// 			case infocs.EqFlash:
// 				stats.TotalFlashPurchased.Price.Total += weaponCost
// 				stats.TotalFlashPurchased.Count++
// 			case infocs.EqSmoke:
// 				stats.TotalSmokePurchased.Price.Total += weaponCost
// 				stats.TotalSmokePurchased.Count++
// 			case infocs.EqDecoy:
// 				stats.TotalDecoyPurchased.Price.Total += weaponCost
// 				stats.TotalDecoyPurchased.Count++
// 			}
// 		}

// 		weaponName := weapon.String()
// 		stats.WeaponTypeCounts[weaponName]++
// 	}

// 	if player.Armor() > 0 {
// 		if player.HasHelmet() {
// 			stats.TotalFullArmorPurchased.Price.Total += 1000
// 			stats.TotalFullArmorPurchased.Count++
// 		} else {
// 			stats.TotalKevlarPurchased.Price.Total += 650
// 			stats.TotalKevlarPurchased.Count++
// 		}
// 		stats.TotalArmors.Price.Total += 1000
// 		stats.TotalArmors.Count++
// 	}

// 	stats.PlayerEquipmentValues[player.UserID] = playerCost
// }

// // Calculate averages and medians after the loop (for efficiency)
// if stats.TotalSpend.Count > 0 {
// 	stats.TotalSpend.Price.Avg = stats.TotalSpend.Price.Total / stats.TotalSpend.Count
// }

// if len(playerCosts) > 0 {
// 	sort.Ints(playerCosts)
// 	stats.TotalSpend.Price.Median = playerCosts[len(playerCosts)/2] // Calculate median
// 	stats.TotalSpend.Price.Min = playerCosts[0]
// 	stats.TotalSpend.Price.Max = playerCosts[len(playerCosts)-1]
// }

// if len(weaponCosts) > 0 {
// 	sort.Ints(weaponCosts)
// 	stats.TotalWeapons.Price.Median = weaponCosts[len(weaponCosts)/2]
// 	stats.TotalWeapons.Price.Min = weaponCosts[0]
// 	stats.TotalWeapons.Price.Max = weaponCosts[len(weaponCosts)-1]
// }

// Add logic to calculate damage and reward values based on round outcome
// ...

// return stats
// }

// func NewCSEconomyContext(team *cs2.Team) *CSEconomyContext {
// 	return cs_entity.&CSEconomyContext{
// 		// RoundNumber: roundNumber,
// 		// Team:  team,
// 		// State: EconomyStateUndefined,
// 	}
// }

// // State Machine Logic
// func (c *CSEconomyContext) UpdateState(event string) {
// 	switch c.State {
// 	case EconomyStateUndefined:
// 		if event == "RoundStart" {
// 			// Analyze team equipment to determine initial state (FullBuy, HalfBuy, Eco, etc.)
// 			// ... (logic to assess equipment and set c.State)
// 		}

// 	case EconomyStateEco:
// 		if event == "RoundEnd" {
// 			if c.Team.Score() > c.RoundNumber/2 {
// 				c.State = EconomyStateHalfBuy
// 			} else {
// 				c.State = EconomyStateEco
// 			}
// 		}

// 	case EconomyStateHalfBuy, EconomyStateFullBuy:
// 		if event == "RoundEnd" {
// 			if c.Team.Score() <= c.RoundNumber/2 { // Lost more than won
// 				c.State = EconomyStateEco // Reset to Eco
// 			} // Otherwise, maintain current state
// 		}

// 		// ... (Add more state transitions as needed)
// 	}
// }

func (c *CSEconomyContext) UpdateState(roundContext CS2RoundContext) {
	// c.PriorState = c.State

	// lossBonus := calculateCS2LossBonus(c.Team.Losses(), c.RoundNumber, c.ConVars)
	// c.Money += lossBonus

	// if c.Money >= minFullBuy {
	//     c.State = EconomyStateFullBuy
	//     return
	// } else if c.Money >= minHalfBuy {
	//     c.State = EconomyStateHalfBuy
	//     return
	// } else if c.PriorState == EconomyStateEco {
	//     c.State = EconomyStateForceBuy
	//     return
	// } else if lossBonus+c.Money >= 3400 {
	//     c.State = EconomyStateAntiEco
	//     return
	// }
	// ... inside UpdateState() function

	// ... (Pistol round check) ...

	// if c.Money >= c.Match.GetMinFullBuy() {
	// 	if c.PriorState == EconomyStatePistolRound && c.Team.GetLosses() == 1 {
	// 		c.State = EconomyStateFullBuySecond
	// 	} else {
	// 		c.State = EconomyStateFullBuy
	// 	}
	// } else if c.Money >= c.Match.GetMinHalfBuy() {
	// 	// ... (similar logic for HalfBuy and HalfBuySecond)
	// } // ... (rest of the logic)

	// c.State = EconomyStateEco
}

// func calculateCS2LossBonus(losses, roundNumber int, convars map[string]string) int {
// 	// Example: Use ConVars to get dynamic loss bonus values (assuming they exist)
// 	lossBonus1 := parseIntOrDefault(convars["mp_consecutive_loss_a_bonus_1"], 1400)
// 	lossBonus2 := parseIntOrDefault(convars["mp_consecutive_loss_a_bonus_2"], 1900)
// 	// ... and so on

// 	// Your loss bonus calculation logic using lossBonus1, lossBonus2, etc.
// 	// ... (adapt the previous logic to use the ConVar values)
// }

// func (c *CSEconomyContext) UpdateStats() {
// 	// Update total money spent
// 	c.Stats.TotalSpend = c.Team.GetTotalMoney() - c.GetMoney()

// 	// Update individual player equipment values
// 	for _, p := range c.Players {
// 		c.Stats.PlayerEquipmentValues[p.Player.UserID] = p.GetEquipmentValue()
// 	}

// 	// Update average equipment value
// 	totalValue := 0
// 	for _, value := range c.Stats.PlayerEquipmentValues {
// 		totalValue += value
// 	}
// 	c.Stats.AvgEquipmentValue = totalValue / len(c.Players)

// 	for _, player := range c.Players {
// 		for _, weapon := range player.Weapons() {
// 			weaponName := weapon.Weapon.Name // ie: awp / operator
// 			c.Stats.WeaponTypeCounts[weaponName]++
// 			switch weapon.Class() { // TROCAR PARA MAP configurável
// 			case cs2.EqClassRifle:
// 				c.Stats.NumRifles++
// 			case cs2.EqClassSMG:
// 				c.Stats.NumSMGs++
// 			case cs2.EqClassPistols:
// 				c.Stats.NumPistols++
// 			case cs2.EqClassHeavy:
// 				c.Stats.NumHeavy++
// 			}
// 		}
// 	}
// }

// func (rc *ReplayContext) UpdateEconomy(roundNumber int, team *cs2.Team, players []*cs2.Player, conVars map[string]string) {
//     ctx, ok := rc.EconomyBreakdown[roundNumber-1] // Round number is 1-indexed, but slice is 0-indexed
//     if !ok {
//         ctx = NewCSEconomyContext(rc.Match, team, roundNumber, players, conVars)
//         rc.EconomyBreakdown = append(rc.EconomyBreakdown, ctx)
//     } else {
//         ctx.SetPlayers(players)
//     }

//     ctx.UpdateState()
//     ctx.UpdateStats()
// }
