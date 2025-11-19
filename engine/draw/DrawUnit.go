package draw

import (
	"time"

	"github.com/warebuf/fd3/engine"
)

func DrawUnit(prevState *engine.GameState, timestamp time.Time, team uint, modelNo uint) *engine.GameState {

	newState := engine.CopyGameState(timestamp, prevState)

	// check if valid

	// modify the state
	newState.Team = append(newState.Team, team)
	newState.Zone = append(newState.Zone, 0) // 0 is bench
	newState.ModelNo = append(newState.ModelNo, modelNo)
	newState.HP = append(newState.HP, 10) // temp: hard code HP
	newState.TimesAttacked = append(newState.TimesAttacked, 0)
	newState.TimesDefended = append(newState.TimesDefended, 0)
	newState.DamageDealt = append(newState.DamageDealt, 0)
	newState.DamageReceived = append(newState.DamageReceived, 0)
	newState.NextUnitID += 1

	// trigger any effects

	return newState
}
