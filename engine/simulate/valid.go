package simulate

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/functions"
)

func ValidFunction(state *engine.GameState, casterIndex int, funcNo int) bool {
	switch funcNo {
	case 0:
		return functions.A00000Valid(state, casterIndex)
	case 1:
		return functions.A00001Valid(state, casterIndex)
	case 2:
		return functions.A00002Valid(state, casterIndex)
	case 3:
		return functions.A00003Valid(state, casterIndex)
	case 4: // counter, should never be valid to be initiated by itself
		return false
	case 5:
		return functions.A00005Valid(state, casterIndex)
	case 6:
		return functions.A00006Valid(state, casterIndex)
	}
	return false
}
