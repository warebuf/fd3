package functions

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/chainTypes"
)

// counter

// reflect and return mult 1x
func A00004(state *engine.GameState, casterIndex int, targetIndex int, dmg int) *chainTypes.HPChangeDeclaration {

	if !A00004CounterValid(state, casterIndex) {
		return nil
	}

	return &chainTypes.HPChangeDeclaration{

		// base info
		CasterIndex: casterIndex,
		Amount:      dmg,

		// type
		Heal:       false,
		Sequential: true, // does not matter in this case

		// target related
		TargetType:    "custom", // all units in the same team as the caster
		All:           false,
		TargetCount:   1,
		TargetIndexes: []int{targetIndex}, // only set if custom

		// status
		Cancelled: false, // always set to false at start

	}

}

func A00004CounterValid(state *engine.GameState, funcIdx int) bool {
	return state.FunctionCharge[funcIdx] >= state.FunctionSPD[funcIdx]
}
