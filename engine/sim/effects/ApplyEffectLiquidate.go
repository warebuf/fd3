package effects

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/chain"
	"github.com/warebuf/fd3/engine/sim/conditions/autoability"
	"github.com/warebuf/fd3/engine/sim/destruction"
	"github.com/warebuf/fd3/engine/state"
)

// destroys a friendly field unit and converts its HP, Charges, and attached equips' Cost into mana
// for its controller. A: target card index (from Card.Actives[abilityIndex].Target).
func ApplyEffectLiquidate(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, eff state.Effect, executeAtWallTime int64) {
	targetIndex := eff.A
	if targetIndex < 0 || targetIndex >= len(gs.Cards) {
		return // preselected target was never set, ability fizzles
	}
	card := gs.Cards[targetIndex]
	if card.Zone != state.ZoneField {
		return // target left the field, ability fizzles
	}
	teamIndex := card.Team

	amount := card.Charges
	if card.HP > 0 {
		amount += card.HP
	}
	for _, ci := range gs.Equip[teamIndex] {
		if gs.Cards[ci].EquipHostIndex == targetIndex {
			amount += gs.Cards[ci].Cost
		}
	}

	destruction.HandleUnitDestruction(gs, pq, logList, UpdateID, targetIndex, executeAtWallTime)

	// a Deathless-style equip may have redirected the destruction onto itself instead of the
	// target; nothing to convert if the target itself never actually reached the graveyard
	if gs.Cards[targetIndex].Zone != state.ZoneGraveyard || amount <= 0 {
		return
	}

	gs.Mana[teamIndex] += amount
	*logList = append(*logList, logs.GainManaLog{
		UpdateID:      *UpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.GainMana,
		TeamIndex:     teamIndex,
		Amount:        amount,
	})
	*UpdateID++

	cardIndexes, activeIndexes := autoability.AutoAbilityManaCheck(gs, teamIndex)
	for i := 0; i < len(cardIndexes); i++ {
		chain.PushAbilityOnChain(gs, logList, UpdateID, cardIndexes[i], activeIndexes[i], executeAtWallTime)
	}

	structureIndexes := autoability.AutoAbilityStructureDrawCheck(gs, teamIndex)
	for i := 0; i < len(structureIndexes); i++ {
		chain.PushStructureDrawOnChain(gs, logList, UpdateID, teamIndex, structureIndexes[i], executeAtWallTime)
	}
}
