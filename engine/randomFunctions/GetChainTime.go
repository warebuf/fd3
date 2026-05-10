package randomFunctions

import (
	"time"

	"github.com/warebuf/fd3/engine/state"
)

// GetChainTime returns chain time in ms.
// Advances with the wall clock, frozen only during a full pause.
func GetChainTime(gs *state.GameState) int64 {
	if gs.IsFullPaused {
		return gs.FullPauseStart - gs.FullPauseOffset
	}
	return time.Now().UnixMilli() - gs.FullPauseOffset
}
