package functions

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/functionHelpers"
)

// Heal any unit on team for 50
func A00002(state *engine.GameState, casterIdx int) bool {
	teamIdx := state.Team[casterIdx]
	targetIdx, ok := functionHelpers.RandomSameTeam(state, teamIdx)
	if !ok {
		return false
	}
	functionHelpers.HealOne(state, targetIdx, 50)
	return true
}

func A00002Valid(state *engine.GameState, casterIdx int) bool {
	if state.Zone[casterIdx] != 1 {
		return false
	}
	teamIdx := state.Team[casterIdx]
	return functionHelpers.RandomSameTeamCheck(state, teamIdx)
}
