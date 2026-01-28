package functions

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/functionHelpers"
)

// TODO: add the summon draw and target (?) functions

// summon a unit
func A99999(state *engine.GameState, casterIdx int) int {
	teamIdx := state.Team[casterIdx]
	targetIdx, ok := functionHelpers.RandomEnemy(state, teamIdx)
	if !ok {
		return -1
	}
	functionHelpers.DamageOne(state, targetIdx, 50)
	return targetIdx
}
