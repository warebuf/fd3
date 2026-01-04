package functions

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/functionHelpers"
)

// deal 50 damage to a random unit, gain 25 exp
func A00005(state *engine.GameState, casterIdx int) int {
	teamIdx := state.Team[casterIdx]
	targetIdx, ok := functionHelpers.RandomEnemy(state, teamIdx)
	if !ok {
		return -1
	}
	functionHelpers.DamageOne(state, targetIdx, 50)
	return targetIdx
}

func A00005Success(state *engine.GameState, casterIdx int) {
	state.EXP[casterIdx] += 25
}

func A00005Valid(state *engine.GameState, casterIdx int) bool {
	if state.Zone[casterIdx] != 1 {
		return false
	}
	teamIdx := state.Team[casterIdx]
	return functionHelpers.RandomEnemyCheck(state, teamIdx)
}

// deal 100 to a random unit, reset all CDs
func A00006(state *engine.GameState, casterIdx int) int {
	// deal 100 damage
	teamIdx := state.Team[casterIdx]
	targetIdx, ok := functionHelpers.RandomEnemy(state, teamIdx)
	if !ok {
		return -1
	}

	funcStart := state.FunctionStartIdx[casterIdx]
	funcStart += 1 // increment by 1 because the second func is the manual one
	selectedTarget := state.FunctionTarget[funcStart*2]

	// if no selected target, dead unit, or not in active zone, choose random enemy
	if selectedTarget < 0 || state.HP[selectedTarget] <= 0 || state.Zone[selectedTarget] != 1 {
		selectedTarget = targetIdx
	}

	functionHelpers.DamageOne(state, selectedTarget, 100)

	// reset all CDs
	startIndex := state.FunctionStartIdx[casterIdx]
	count := state.FunctionCount[casterIdx]
	for i := startIndex; i < startIndex+count; i++ {
		state.FunctionCharge[i] = 0
	}

	// spend 75 exp
	state.EXP[casterIdx] -= 75

	return selectedTarget
}

// can reuse A00005 for success

func A00006Valid(state *engine.GameState, casterIdx int) bool {
	if state.Zone[casterIdx] != 1 {
		return false
	}
	if state.EXP[casterIdx] < 75 {
		return false
	}

	teamIdx := state.Team[casterIdx]
	return functionHelpers.RandomEnemyCheck(state, teamIdx)
}
