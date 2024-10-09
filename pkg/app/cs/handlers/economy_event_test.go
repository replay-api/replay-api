package handlers_test

import (
	"testing"

	infocs "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
	cs_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/cs/entities"
)

func createPlayerWithEquipment(equipments ...infocs.EquipmentType) *infocs.Player {
	p := &infocs.Player{
		Inventory: make(map[int]*infocs.Equipment, len(equipments)),
	}

	for i, eq := range equipments {
		p.Inventory[i] = &infocs.Equipment{
			Type: eq,
		}
	}

	return p
}

func TestDetermineBuyType(t *testing.T) {
	testCases := []struct {
		name            string
		teamMembers     []*infocs.Player
		expectedBuyType cs_entity.CSEconomyState
	}{
		{
			name: "FullBuy 5v5 (Heavy Armor)",
			teamMembers: []*infocs.Player{
				createPlayerWithEquipment(infocs.EqAWP, infocs.EqP250, infocs.EqFlash, infocs.EqSmoke, infocs.EqKevlar, infocs.EqHelmet),
				createPlayerWithEquipment(infocs.EqAK47, infocs.EqP250, infocs.EqHE, infocs.EqMolotov, infocs.EqKevlar, infocs.EqHelmet),
				createPlayerWithEquipment(infocs.EqAK47, infocs.EqP250, infocs.EqFlash, infocs.EqDecoy, infocs.EqKevlar, infocs.EqHelmet),
				createPlayerWithEquipment(infocs.EqM4A4, infocs.EqP250, infocs.EqSmoke, infocs.EqIncendiary, infocs.EqKevlar, infocs.EqHelmet),
				createPlayerWithEquipment(infocs.EqAK47, infocs.EqP250, infocs.EqHE, infocs.EqKevlar, infocs.EqHelmet),
			},
			expectedBuyType: cs_entity.CSEconomyStateFullBuy,
		},

		{
			name: "FullBuy 4v4 (Heavy Armor)",
			teamMembers: []*infocs.Player{
				createPlayerWithEquipment(infocs.EqAWP, infocs.EqP250, infocs.EqFlash, infocs.EqSmoke, infocs.EqKevlar, infocs.EqHelmet),
				createPlayerWithEquipment(infocs.EqAK47, infocs.EqP250, infocs.EqHE, infocs.EqMolotov, infocs.EqKevlar, infocs.EqHelmet),
				createPlayerWithEquipment(infocs.EqAK47, infocs.EqP250, infocs.EqFlash, infocs.EqDecoy, infocs.EqKevlar, infocs.EqHelmet),
				createPlayerWithEquipment(infocs.EqM4A4, infocs.EqP250, infocs.EqSmoke, infocs.EqIncendiary, infocs.EqKevlar, infocs.EqHelmet),
			},
			expectedBuyType: cs_entity.CSEconomyStateFullBuy,
		},

		{
			name: "FullBuy 3v3 (Heavy Armor)",
			teamMembers: []*infocs.Player{
				createPlayerWithEquipment(infocs.EqAWP, infocs.EqP250, infocs.EqFlash, infocs.EqSmoke, infocs.EqKevlar, infocs.EqHelmet),
				createPlayerWithEquipment(infocs.EqAK47, infocs.EqP250, infocs.EqHE, infocs.EqMolotov, infocs.EqKevlar, infocs.EqHelmet),
				createPlayerWithEquipment(infocs.EqAK47, infocs.EqP250, infocs.EqFlash, infocs.EqDecoy, infocs.EqKevlar, infocs.EqHelmet),
			},
			expectedBuyType: cs_entity.CSEconomyStateFullBuy,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// TODO: precisa de desacoplamento dos parametros pra poder testar isolado
			// TODO: implementar tests para demais metodos (h.CalculateEconomyStats etc)
			// buyType := h.DetermineBuyType(tc.teamMembers)
			// if buyType != tc.expectedBuyType {
			// 	t.Errorf("Expected %s, but got %s", tc.expectedBuyType, buyType)
			// }
		})
	}
}
