package summon

import (
	"time"

	"github.com/warebuf/fd3/engine"
)

func SpawnSummon(unitModels []engine.Model, prevState *engine.GameState, team int, modelNo int) *engine.GameState {

	newState := engine.CopyGameState(time.Now(), prevState)

	// get model
	model := unitModels[modelNo]

	// modify the state
	newState.Team = append(newState.Team, team)
	newState.Zone = append(newState.Zone, 1) // 1 is field
	newState.ModelNo = append(newState.ModelNo, modelNo)
	newState.HP = append(newState.HP, model.HP)
	newState.SummonCondition = append(newState.SummonCondition, model.SummonCondition)

	newState.AbilityStartIdx = append(newState.AbilityStartIdx, len(newState.Abilities))
	newState.AbilityCount = append(newState.AbilityCount, len(model.Abilities))
	for i := 0; i < len(model.Abilities); i++ {
		newState.Abilities = append(newState.Abilities, model.Abilities[i].AbilityID)
		newState.AbilitySPD = append(newState.AbilitySPD, model.Abilities[i].CD*1000)
		newState.AbilityCharge = append(newState.AbilityCharge, 0)
	}

	return newState
}
