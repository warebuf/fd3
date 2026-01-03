package target

import (
	"time"

	"github.com/warebuf/fd3/engine"
)

func SetTargets(prevState *engine.GameState, timestamp time.Time, unitIdx int, funcIdx int, target1 int, target2 int) (*engine.GameState, bool) {

	newState := engine.CopyGameState(timestamp, prevState)

	// check if valid

	// can prevent setting targets by abilities if i want here

	// modify the state
	startIdx := funcIdx * 2
	newState.FunctionTarget[startIdx] = target1
	newState.FunctionTarget[startIdx+1] = target2

	// trigger any effects

	return newState, true
}
