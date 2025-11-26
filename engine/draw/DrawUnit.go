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
	newState.HP = append(newState.HP, model.HP)

	newState.AbilityStartIdx = append(newState.AbilityStartIdx, len(newState.Abilities))
	newState.AbilityCount = append(newState.AbilityCount, len(model.Abilities))
	for i := 0; i < len(model.Abilities); i++ {
		newState.Abilities = append(newState.Abilities, model.Abilities[i].AbilityID)
		newState.AbilitySPD = append(newState.AbilitySPD, model.Abilities[i].CD*1000)
		newState.AbilityCharge = append(newState.AbilityCharge, 0)
	}

	// trigger any effects

	return newState
}
