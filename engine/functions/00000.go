package functions

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/actions"
	"github.com/warebuf/fd3/engine/functions_shared"
)

// heal 1 random on same team
func A00000(state *engine.GameState, casterIndex int, functionIndex int, logs []actions.Action) []actions.Action {
	teamIdx := state.Team[casterIndex]
	targetIdx, ok := functions_shared.RandomSameTeam(state, teamIdx)
	if !ok {
		return logs
	}
	functions_shared.HealOne(state, targetIdx, 5)
	state.FunctionCharge[functionIndex] = 0

	logs = append(logs, actions.Action{
		Timestamp:     state.Timestamp,
		Type:          "function",
		UnitIndex:     casterIndex,
		FunctionIndex: functionIndex,
		FunctionNo:    0,
	})

	return logs
}

func A00000Valid(state *engine.GameState, casterIdx int) bool {
	if state.Zone[casterIdx] != 1 {
		return false
	}
	teamIdx := state.Team[casterIdx]
	return functions_shared.RandomSameTeamCheck(state, teamIdx)
}
