package functions

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/chainTypes"
)

// deal 2 damage to all
func A00001(state *engine.GameState, casterIndex int) *chainTypes.HPChangeDeclaration {

	if !A00001Valid(state, casterIndex) {
		return nil
	}

	return &chainTypes.HPChangeDeclaration{

		// base info
		CasterIndex: casterIndex,
		Amount:      2,

		// type
		Heal:       false,
		Sequential: false, // concurrent

		// target related
		TargetType:    "all",   // all units
		All:           true,    // apply damage to all
		TargetCount:   0,       // doesn't matter what this is
		TargetIndexes: []int{}, // only set if custom

		// status
		Cancelled: false, // always set to false at start

	}

}

func A00001Valid(state *engine.GameState, casterIndex int) bool {
	if state.Zone[casterIndex] != 1 {
		return false
	}
	return true
}
