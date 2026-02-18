package simulate

import "github.com/warebuf/fd3/engine"

type Due struct {
	kind         string
	unitIndex    int
	abilityIndex int
	prio         int
}

func getDue(state *engine.GameState) []Due {
	due := []Due{}

	// iterate through units
	for i := 0; i < len(state.FunctionStartIdx); i++ {

		// check for due summons
		if state.Zone[i] == 3 && state.SummonCharge[i] >= 2000 {
			due = append(due, Due{
				kind:         "summon",
				unitIndex:    i,
				abilityIndex: -1,
				prio:         0,
			})
		}

		// iterate through functions
		for j := 0; j < state.FunctionCount[i]; j++ {
			functionIndex := state.FunctionStartIdx[i] + j
			if state.FunctionType[functionIndex] == 0 { // only check auto functions
				isValid := ValidFunction(state, i, state.Functions[functionIndex])
				if state.FunctionCharge[functionIndex] >= state.FunctionSPD[functionIndex] && isValid {
					due = append(due, Due{
						kind:         "function",
						unitIndex:    i,
						abilityIndex: functionIndex,
						prio:         0,
					})
				}
			}
		}

	}

	return due
}
