package abilities

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/abilityHelpers"
)

// counter

func A00008(state *engine.GameState, targetIdx int, dmg int) bool {
	abilityHelpers.DamageOne(state, targetIdx, dmg)
	return true
}

func A00008CounterValid(state *engine.GameState, defenderIdx int, attackerIdx int) bool {

	if state.AbilityCharge[defenderIdx] < state.AbilitySPD[defenderIdx] {
		return false
	}
	return true
}
