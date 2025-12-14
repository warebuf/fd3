package abilities

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/abilityHelpers"
)

// counter

func A00008(state *engine.GameState, defenderIdx int, attackerIdx int) int {
	teamIdx := state.Team[defenderIdx]
	targetIdx, ok := abilityHelpers.RandomEnemy(state, teamIdx)
	if !ok {
		return -1
	}
	//fmt.Println(casterIdx, "Atk", targetIdx)
	abilityHelpers.DamageOne(state, targetIdx, 25)
	return targetIdx
}

func A00008CounterValid(state *engine.GameState, defenderIdx int, attackerIdx int) bool {

	if state.AbilityCharge[defenderIdx] < state.AbilitySPD[defenderIdx] {
		return false
	}

}
