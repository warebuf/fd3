package functions

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/functionHelpers"
)

// counter

// reflect and return mult 1x
func A00004(state *engine.GameState, defIdx int, targetIdx int, dmg int) int {
	functionHelpers.HealOne(state, defIdx, dmg)
	functionHelpers.DamageOne(state, targetIdx, dmg)
	return dmg
}

func A00004CounterValid(state *engine.GameState, funcIdx int) bool {
	return state.FunctionCharge[funcIdx] >= state.FunctionSPD[funcIdx]
}
