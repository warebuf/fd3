package functions

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/functions_shared"
)

// counter

// reflect and return mult 1x
func A00004(state *engine.GameState, defIdx int, targetIdx int, dmg int) int {
	functions_shared.HealOne(state, defIdx, dmg)
	functions_shared.DamageOne(state, targetIdx, dmg)
	return dmg
}

func A00004CounterValid(state *engine.GameState, funcIdx int) bool {
	return state.FunctionCharge[funcIdx] >= state.FunctionSPD[funcIdx]
}
