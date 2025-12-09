package abilities

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/abilityHelpers"
)

// deal 25 damage to a random enemy

func A00007(state *engine.GameState, casterIdx int) bool {
	teamIdx := state.Team[casterIdx]
	targetIdx, ok := abilityHelpers.RandomEnemy(state, teamIdx)
	if !ok {
		return false
	}
	//fmt.Println(casterIdx, "Atk", targetIdx)
	abilityHelpers.DamageOne(state, targetIdx, 25)
	return true
}

func A00007Valid(state *engine.GameState, casterIdx int) bool {
	if state.Zone[casterIdx] != 1 {
		return false
	}
	teamIdx := state.Team[casterIdx]
	return abilityHelpers.RandomEnemyCheck(state, teamIdx)
}
