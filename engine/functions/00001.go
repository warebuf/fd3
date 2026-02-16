package functions

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/functions_shared"
)

// deal 2 damage to all
func A00001(state *engine.GameState, casterIdx int) bool {
	functions_shared.DamageAll(state, 2)
	return true
}

func A00001Valid(state *engine.GameState, casterIdx int) bool {
	if state.Zone[casterIdx] != 1 {
		return false
	}
	return true
}
