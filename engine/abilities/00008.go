package abilities

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/abilityHelpers"
)

// counter

// reflect and return mult 1x
func A00008(state *engine.GameState, defIdx int, targetIdx int, dmg int) int {
	abilityHelpers.HealOne(state, defIdx, dmg)
	abilityHelpers.DamageOne(state, targetIdx, dmg)
	state.AbilityCharge[defIdx] = 0
	return dmg
}

func A00008CounterValid(state *engine.GameState, defenderIdx int, attackerIdx int) bool {
	return state.AbilityCharge[defenderIdx] >= state.AbilitySPD[defenderIdx]
}
