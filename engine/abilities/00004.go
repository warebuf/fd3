package abilities

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/abilityHelpers"
)

// counter

// reflect and return mult 1x
func A00004(state *engine.GameState, defIdx int, targetIdx int, dmg int) int {
	abilityHelpers.HealOne(state, defIdx, dmg)
	abilityHelpers.DamageOne(state, targetIdx, dmg)
	return dmg
}

func A00004CounterValid(state *engine.GameState, abilityIdx int) bool {
	return state.AbilityCharge[abilityIdx] >= state.AbilitySPD[abilityIdx]
}
