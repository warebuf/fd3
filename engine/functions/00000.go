package functions

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/chainTypes"
	"github.com/warebuf/fd3/engine/functions/helpers"
)

// heal 1 random on same team
func A00000(state *engine.GameState, casterIndex int) *chainTypes.HPChangeDeclaration {

	if A00000Valid(state, casterIndex) {
		return nil
	}

	return &chainTypes.HPChangeDeclaration{

		// base info
		CasterIndex: casterIndex,
		Amount:      5,

		// type
		Heal:       true,
		Sequential: true, // does not matter in this case

		// target related
		TargetType:    "team", // all units in the same team as the caster
		All:           false,
		TargetCount:   1,
		TargetIndexes: []int{}, // only set if custom

		// status
		Cancelled: false, // always set to false at start

	}

}

func A00000Valid(state *engine.GameState, casterIndex int) bool {
	if state.Zone[casterIndex] != 1 {
		return false
	}
	teamIdx := state.Team[casterIndex]
	return helpers.RandomSameTeamCheck(state, teamIdx)
}
