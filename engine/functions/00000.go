package functions

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/functionHelpers"
)

// heal 1 random on same team
func A00000(state *engine.GameState, casterIdx int) bool {
	teamIdx := state.Team[casterIdx]
	targetIdx, ok := functionHelpers.RandomSameTeam(state, teamIdx)
	if !ok {
		return false
	}
	functionHelpers.HealOne(state, targetIdx, 5)
	return true
}

func A00000Valid(state *engine.GameState, casterIdx int) bool {
	if state.Zone[casterIdx] != 1 {
		return false
	}
	teamIdx := state.Team[casterIdx]
	return functionHelpers.RandomSameTeamCheck(state, teamIdx)
}
