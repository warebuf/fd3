package microevents

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
	newState.EXP = append(newState.EXP, 0)

	newState.FunctionStartIdx = append(newState.FunctionStartIdx, len(newState.Functions))
	newState.FunctionCount = append(newState.FunctionCount, len(model.Functions))
	for i := 0; i < len(model.Functions); i++ {
		newState.Functions = append(newState.Functions, model.Functions[i].FunctionID)
		newState.FunctionSPD = append(newState.FunctionSPD, model.Functions[i].CD*1000)
		newState.FunctionCharge = append(newState.FunctionCharge, 0)

		// each function gets 2 elements for targets
		newState.FunctionTarget = append(newState.FunctionTarget, -1)
		newState.FunctionTarget = append(newState.FunctionTarget, -1)

		// push FunctionType as int
		switch model.Functions[i].Type {
		case "Auto":
			newState.FunctionType = append(newState.FunctionType, 0)
		case "Manual":
			newState.FunctionType = append(newState.FunctionType, 1)
		}
	}

	// trigger any effects

	return newState
}
