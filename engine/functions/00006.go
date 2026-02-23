package functions

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/chainTypes"
	"github.com/warebuf/fd3/engine/functions/helpers"
)

// deal 100 to a random unit, reset all CDs
// TODO: have not implemented CD reset, should either be it's own chain or added to the HPChangeDeclaration funciton
func A00006(state *engine.GameState, casterIndex int) *chainTypes.HPChangeDeclaration {

	if !A00006Valid(state, casterIndex) {
		return nil
	}

	return &chainTypes.HPChangeDeclaration{

		// base info
		CasterIndex: casterIndex,
		Amount:      100,

		// type
		Heal:       false,
		Sequential: true, // does not matter in this case

		// target related
		TargetType:    "enemy", // all units in the same team as the caster
		All:           false,
		TargetCount:   1,
		TargetIndexes: []int{}, // only set if custom

		// status
		Cancelled: false, // always set to false at start

	}

}

func A00006Valid(state *engine.GameState, casterIndex int) bool {
	if state.Zone[casterIndex] != 1 {
		return false
	}
	if state.EXP[casterIndex] < 75 {
		return false
	}

	teamIdx := state.Team[casterIndex]
	return helpers.RandomEnemyCheck(state, teamIdx)
}
