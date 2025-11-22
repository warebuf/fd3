package draw

import (
	"time"

	"github.com/warebuf/fd3/engine"
)

func DrawUnit(unitModels []engine.Model, prevState *engine.GameState, timestamp time.Time, team int, modelNo int) *engine.GameState {

	newState := engine.CopyGameState(timestamp, prevState)

	// check if valid

	// get model
	model := unitModels[modelNo]

	// modify the state
	newState.Team = append(newState.Team, team)
	newState.Zone = append(newState.Zone, 0) // 0 is bench
	newState.ModelNo = append(newState.ModelNo, modelNo)
	newState.HP = append(newState.HP, model.HP) // temp: hard code HP
	newState.NextUnitID += 1

	// trigger any effects

	return newState
}
