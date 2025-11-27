package abilities

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/abilityHelpers"
)

// deal 50 damage to a random enemy

func A00002(state *engine.GameState, casterIdx int) bool {

	teamIdx := state.Team[casterIdx]
	targetIdx, ok := abilityHelpers.RandomEnemy(state, teamIdx)
	if !ok {
		return false
	}
	abilityHelpers.DamageOne(state, targetIdx, 50)

	return true
}
