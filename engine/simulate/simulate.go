package simulate

import (
	"fmt"

	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/functions"
)

func Simulate(state *engine.GameState, dueUnits []int, dueAbilities []int) (*engine.GameState, bool) {

	modified := false

	prio := sortFunctionsByPrio(state, dueAbilities)

	for i := 0; i < len(prio); i++ {

		idx := prio[i]

		dueUnit := dueUnits[idx]
		dueAbility := dueAbilities[idx]
		abilityNo := state.Functions[dueAbility]

		// could use an array of functions but will use a switch statement for now
		switch abilityNo {
		case 0:
			occurred := functions.A00000(state, dueUnit)
			if occurred {
				modified = true
				state.FunctionCharge[dueAbility] = 0
			}
		case 1:
			occurred := functions.A00001(state, dueUnit)
			if occurred {
				modified = true
				state.FunctionCharge[dueAbility] = 0
				defenders := []int{}
				attackers := []int{}
				functions := []int{}
				for i := 0; i < len(state.Zone); i++ {
					if state.Zone[i] == 1 {
						defenders = append(defenders, i)
						attackers = append(attackers, dueUnit)
						functions = append(functions, dueAbility)
					}
				}
				ApplyCounterChain(state, defenders, attackers, functions, 2)
			}
		case 2:
			occurred := functions.A00002(state, dueUnit)
			if occurred {
				state.FunctionCharge[dueAbility] = 0
				modified = true
			}
		case 3:
			targetIdx := functions.A00003(state, dueUnit)
			if targetIdx >= 0 {
				modified = true
				state.FunctionCharge[dueAbility] = 0
				defenders := []int{targetIdx}
				attackers := []int{dueUnit}
				functions := []int{dueAbility}
				ApplyCounterChain(state, defenders, attackers, functions, 50)
			}
		case 4:
			fmt.Println("cannot handle counterattack here, there is a bug")
		case 5:
			targetIdx := functions.A00005(state, dueUnit)
			if targetIdx >= 0 {
				modified = true
				state.FunctionCharge[dueAbility] = 0
				defenders := []int{targetIdx}
				attackers := []int{dueUnit}
				functions := []int{dueAbility}
				ApplyCounterChain(state, defenders, attackers, functions, 50)
			}
		case 6:
			targetIdx := functions.A00006(state, dueUnit)
			if targetIdx >= 0 {
				modified = true
				state.FunctionCharge[dueAbility] = 0
				defenders := []int{targetIdx}
				attackers := []int{dueUnit}
				functions := []int{dueAbility}
				ApplyCounterChain(state, defenders, attackers, functions, 100)
			}
		}

	}

	// mark dead as dead
	for i := 0; i < len(state.HP); i++ {
		if state.HP[i] <= 0 {
			state.Zone[i] = 2
		}
	}

	// iterate through targets, if dead, reset target
	for i := 0; i < len(state.FunctionTarget); i++ {
		if state.FunctionTarget[i] >= 0 {
			if state.Zone[state.FunctionTarget[i]] == 2 {
				state.FunctionTarget[i] = -1
			}
		}
	}

	return state, modified
}
