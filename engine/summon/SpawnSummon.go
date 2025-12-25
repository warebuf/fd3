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
	newState.EXP = append(newState.EXP, 0)
	newState.SummonCondition = append(newState.SummonCondition, model.SummonCondition)

	newState.FunctionStartIdx = append(newState.FunctionStartIdx, len(newState.Functions))
	newState.FunctionCount = append(newState.FunctionCount, len(model.Functions))
	for i := 0; i < len(model.Functions); i++ {
		newState.Functions = append(newState.Functions, model.Functions[i].FunctionID)
		newState.FunctionSPD = append(newState.FunctionSPD, model.Functions[i].CD*1000)
		newState.FunctionCharge = append(newState.FunctionCharge, 0)

		switch model.Functions[i].Type {
		case "Auto":
			newState.FunctionType = append(newState.FunctionType, 0)
		case "Manual":
			newState.FunctionType = append(newState.FunctionType, 1)
		}
	}

	return newState
}
