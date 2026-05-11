package randomFunctions

import "github.com/warebuf/fd3/engine/state"

func GetSimTime(gs *state.GameState, wallTime int64) int64 {

	if gs.PauseStack[0] != 0 { // if we are in a full pause, the sim time is paused, so get the full pause start time and subtract the SimPauseOffset
		return gs.PauseStack[0] - gs.SimPauseOffset
	} else if gs.PauseStack[1] != 0 { // if we are in an animation pause, ditto
		return gs.PauseStack[1] - gs.SimPauseOffset
	} else if gs.PauseStack[2] != 0 { // if we are in an sim pause, ditto
		return gs.PauseStack[2] - gs.SimPauseOffset
	}

	return wallTime - gs.SimPauseOffset
}

func GetActiveTime(gs *state.GameState, wallTime int64) int64 {

	if gs.PauseStack[0] != 0 { // if we are in a full pause, the active time is paused, so get the full pause start time and subtract the SimPauseOffset
		return gs.PauseStack[0] - gs.FullPauseOffset
	} else if gs.PauseStack[1] != 0 { // if we are an animation pause, the active time is paused, ditto
		return gs.PauseStack[1] - gs.FullPauseOffset
	}

	return wallTime - gs.FullPauseOffset

}
