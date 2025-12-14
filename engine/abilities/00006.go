package abilities

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/abilityHelpers"
)

// Heal any unit on team for 25

func A00006(state *engine.GameState, casterIdx int) bool {
	teamIdx := state.Team[casterIdx]
	targetIdx, ok := abilityHelpers.RandomSameTeam(state, teamIdx)
	if !ok {
		return false
	}
	//fmt.Println(casterIdx, "Heal", targetIdx)
	abilityHelpers.HealOne(state, targetIdx, 25)
	state.AbilityCharge[casterIdx] = 0
	return true
}

func A00006Valid(state *engine.GameState, casterIdx int) bool {
	if state.Zone[casterIdx] != 1 {
		return false
	}
	teamIdx := state.Team[casterIdx]
	return abilityHelpers.RandomSameTeamCheck(state, teamIdx)
}
