package randomFunctions

import "github.com/warebuf/fd3/engine/state"

// GetSimTime returns simulation time in ms.
// Frozen during both a full pause and a sim pause.
func GetSimTime(gs *state.GameState) int64 {
	if gs.IsSimPaused {
		return gs.SimPauseStart - gs.SimPauseOffset
	}
	return GetChainTime(gs) - gs.SimPauseOffset
}

func PauseSim(gs *state.GameState) {
	if gs.IsSimPaused || gs.IsFullPaused {
		return
	}
	gs.IsSimPaused = true
	gs.SimPauseStart = GetChainTime(gs)
}

func UnpauseSim(gs *state.GameState) {
	if !gs.IsSimPaused {
		return
	}
	gs.SimPauseOffset += GetChainTime(gs) - gs.SimPauseStart
	gs.IsSimPaused = false
}
