package summon

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/actions"
)

func SummonResolution(state *engine.GameState, unitIdx int) []actions.Action {

	var logs []actions.Action

	// unit is not queued summoning
	if state.Zone[unitIdx] != 3 {
		return logs
	}

	// set the unit to a queued summon
	state.Zone[unitIdx] = 1

	// append log
	logs = append(logs, actions.Action{
		Timestamp: state.Timestamp,
		Type:      "summon_resolution",
		UnitIndex: unitIdx,
	})

	// trigger any effects

	return logs
}
