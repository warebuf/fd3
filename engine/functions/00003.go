package functions

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/functions_shared"
)

// deal 50 damage to a random enemy

// return target index
func A00003(state *engine.GameState, casterIdx int) int {
	teamIdx := state.Team[casterIdx]
	targetIdx, ok := functions_shared.RandomEnemy(state, teamIdx)
	if !ok {
		return -1
	}
	functions_shared.DamageOne(state, targetIdx, 50)
	return targetIdx
}

func A00003Valid(state *engine.GameState, casterIdx int) bool {
	if state.Zone[casterIdx] != 1 {
		return false
	}
	teamIdx := state.Team[casterIdx]
	return functions_shared.RandomEnemyCheck(state, teamIdx)
}
