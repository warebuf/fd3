package functions

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/actions"
	"github.com/warebuf/fd3/engine/functions_shared"
)

// Heal any unit on team for 50
func A00002(state *engine.GameState, casterIndex int, functionIndex int, logs []actions.Action) []actions.Action {
	teamIdx := state.Team[casterIndex]
	targetIdx, ok := functions_shared.RandomSameTeam(state, teamIdx)
	if !ok {
		return logs
	}
	functions_shared.HealOne(state, targetIdx, 50)

	// append log
	logs = append(logs, actions.Action{
		Timestamp:     state.Timestamp,
		Type:          "function",
		UnitIndex:     casterIndex,
		FunctionIndex: functionIndex,
		FunctionNo:    2,
	})

	// reset charge
	state.FunctionCharge[functionIndex] = 0

	return logs
}

func A00002Valid(state *engine.GameState, casterIdx int) bool {
	if state.Zone[casterIdx] != 1 {
		return false
	}
	teamIdx := state.Team[casterIdx]
	return functions_shared.RandomSameTeamCheck(state, teamIdx)
}
