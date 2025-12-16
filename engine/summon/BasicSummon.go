package summon

import (
	"fmt"
	"time"

	"github.com/warebuf/fd3/engine"
)

func BasicSummon(unitModels []engine.Model, prevState *engine.GameState, timestamp time.Time, team int, unitIdx int) *engine.GameState {

	newState := engine.CopyGameState(timestamp, prevState)

	summonType := newState.SummonCondition[unitIdx]

	if newState.Round == 0 && newState.Team[unitIdx] == team && (summonType == engine.SummonTypeNormal || summonType == engine.SummonTypeStart) { // toggle preround summon
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
		return newState
	}

	// trigger any effects

	return newState
}
