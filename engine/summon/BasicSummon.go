package summon

import (
	"time"

	"github.com/warebuf/fd3/engine"
)

func BasicSummon(prevState *engine.GameState, timestamp time.Time) *engine.GameState {

	newState := engine.CopyGameState(timestamp, prevState)

	// check if valid bench unit, move it from bench to field

	// modify the state
	newState.Team = append(newState.Team, 0)       // temp: assign to team 0
	newState.Zone = append(newState.Zone, 1)       // 1 is field
	newState.ModelNo = append(newState.ModelNo, 0) // temp: hard code model no
	newState.HP = append(newState.HP, 10)          // temp: hard code HP
	newState.TimesAttacked = append(newState.TimesAttacked, 0)
	newState.TimesDefended = append(newState.TimesDefended, 0)
	newState.DamageDealt = append(newState.DamageDealt, 0)
	newState.DamageReceived = append(newState.DamageReceived, 0)
	newState.NextUnitID += 1

	// trigger any effects

	return newState
}
