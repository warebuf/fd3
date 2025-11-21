package summon

import (
	"time"

	"github.com/warebuf/fd3/engine"
)

func BasicSummon(prevState *engine.GameState, timestamp time.Time, team uint, modelNo uint) *engine.GameState {

	newState := engine.CopyGameState(timestamp, prevState)

	// check if valid bench unit, move it from bench to field

	// modify the state
	newState.Team = append(newState.Team, team)
	newState.Zone = append(newState.Zone, 1) // 1 is field
	newState.ModelNo = append(newState.ModelNo, modelNo)
	newState.HP = append(newState.HP, 10) // temp: hard code HP
	newState.NextUnitID += 1

	// trigger any effects

	return newState
}
