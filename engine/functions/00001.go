package functions

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/actions"
	"github.com/warebuf/fd3/engine/functions_shared"
)

// deal 2 damage to all
func A00001(state *engine.GameState, casterIndex int, functionIndex int, logs []actions.Action) []actions.Action {
	functions_shared.DamageAll(state, 2)
	state.FunctionCharge[functionIndex] = 0

	logs = append(logs, actions.Action{
		Timestamp:     state.Timestamp,
		Type:          "function",
		UnitIndex:     casterIndex,
		FunctionIndex: functionIndex,
		FunctionNo:    1,
	})

	return logs
}

func A00001Valid(state *engine.GameState, casterIdx int) bool {
	if state.Zone[casterIdx] != 1 {
		return false
	}
	return true
}
