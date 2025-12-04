package summon

import (
	"fmt"
	"time"

	"github.com/warebuf/fd3/engine"
)

func BasicSummon(unitModels []engine.Model, prevState *engine.GameState, timestamp time.Time, team int, unitIdx int) *engine.GameState {

	newState := engine.CopyGameState(timestamp, prevState)

	// get model
	//modelNo := prevState.ModelNo[unitIdx]
	//model := unitModels[modelNo]

	// check if valid bench unit, move it from bench to field
	if newState.Team[unitIdx] == team && newState.Zone[unitIdx] == 0 {
		newState.Zone[unitIdx] = 1
	} else {
		fmt.Println("invalid BasicSummon", unitIdx, team, newState.Team[unitIdx], newState.Zone[unitIdx])
		return newState
	}

	// trigger any effects

	return newState
}
