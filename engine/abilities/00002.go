package abilities

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/abilityHelpers"
)

// deal 50 damage to a random enemy

func A00002(state *engine.GameState, casterIdx int) int {

	teamIdx := state.Team[casterIdx]
	targetIdx, ok := abilityHelpers.RandomEnemy(state, teamIdx)
	if !ok {
		return -1
	}

	abilityHelpers.DamageOne(state, targetIdx, 50)
	state.AbilityCharge[casterIdx] = 0
	return targetIdx
}

func A00002Valid(state *engine.GameState, casterIdx int) bool {
	if state.Zone[casterIdx] != 1 {
		return false
	}
	teamIdx := state.Team[casterIdx]
	return abilityHelpers.RandomEnemyCheck(state, teamIdx)
}
