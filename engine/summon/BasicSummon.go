package summon

import (
	"time"

	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/actions"
)

func BasicSummon(unitModels []engine.Model, state *engine.GameState, timestamp time.Time, team int, unitIdx int) []actions.Action {

	var logs []actions.Action

	// validation
	if state.Team[unitIdx] != team {
		return logs
	}

	// unit is not on the bench
	if state.Zone[unitIdx] != 0 {
		return logs
	}

	// set the unit to a queued summon
	state.Zone[unitIdx] = 3

	// append log (summon declaration)
	logs = append(logs, actions.Action{
		Timestamp: timestamp,
		Type:      "summon",
		UnitIndex: unitIdx,
	})

	// trigger any effects

	return logs
}
