package abilities

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/abilityHelpers"
)

// heal 1 random on same team

func A00000(state *engine.GameState, casterIdx int) bool {

	teamIdx := state.Team[casterIdx]
	targetIdx, ok := abilityHelpers.RandomSameTeam(state, teamIdx)
	if !ok {
		return false
	}
	abilityHelpers.HealOne(state, targetIdx, 5)

	return true
}
