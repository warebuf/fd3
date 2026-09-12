package match

import (
	"github.com/warebuf/fd3/engine/logs"
)

// the chain changed, so the window each team waived is gone; drop every skip
func (match *Match) clearChainSkips(wallTime int64) {
	for teamIndex, skip := range match.GameState.SkipChain {
		if !skip {
			continue
		}
		match.GameState.SkipChain[teamIndex] = false
		match.EventLog = append(match.EventLog, logs.UnskipChainLog{
			WallTimestamp: wallTime,
			UpdateID:      match.NextUpdateID,
			Type:          logs.UnskipChain,
			TeamIndex:     teamIndex,
		})
		match.NextUpdateID++
	}
}
