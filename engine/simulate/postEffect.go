package simulate

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/functions"
)

// maybe post effect shouldnt exist, maybe we should just apply it
func HandlePostEffect(state *engine.GameState, defenderArr []int, attackerArr []int, functionArr []int, attackerIdx int, defenderIdx int, functionIdx int) {

	functionNo := state.Functions[functionIdx]

	switch functionNo {
	case 5:
		functions.A00005Success(state, attackerIdx)
	case 6:
		functions.A00005Success(state, attackerIdx)
	default:
	}

}
