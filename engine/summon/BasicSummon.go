package summon

import (
	"fmt"
	"time"

	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/actions"
)

func BasicSummon(unitModels []engine.Model, prevState *engine.GameState, timestamp time.Time, team int, unitIdx int) (*engine.GameState, []actions.Action) {

	newState := engine.CopyGameState(timestamp, prevState)

	var logs []actions.Action

	if newState.Round == 0 && newState.Team[unitIdx] == team { // toggle preround summon
		if newState.Zone[unitIdx] == 0 {
			newState.Zone[unitIdx] = 3
		} else if newState.Zone[unitIdx] == 3 {
			newState.Zone[unitIdx] = 0
		}
	} else if newState.Round > 0 && newState.Team[unitIdx] == team && newState.Zone[unitIdx] == 3 { // check if valid bench unit, move it from bench to field
		newState.Zone[unitIdx] = 1
	} else if newState.Round > 0 && newState.Team[unitIdx] == team && newState.Zone[unitIdx] == 0 { // check if valid bench unit, move it from bench to field
		newState.Zone[unitIdx] = 1
	} else {
		fmt.Println("invalid BasicSummon", unitIdx, team, newState.Team[unitIdx], newState.Zone[unitIdx])
		return newState, logs
	}

	// trigger any effects

	return newState, logs
}
