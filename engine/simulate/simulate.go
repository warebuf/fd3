package simulate

import (
	"fmt"

	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/actions"
	"github.com/warebuf/fd3/engine/functions"
	"github.com/warebuf/fd3/engine/summon"
)

func Simulate(state *engine.GameState, due []Due) []actions.Action {

	var logs []actions.Action

	prio := sortDue(state, due)

	// iterate through sorted, applying actions to the state
	for i := 0; i < len(prio); i++ {

		if due[i].kind == "summon" {
			newLogs := summon.SummonResolution(state, prio[i].unitIndex)
			logs = append(logs, newLogs...)
		} else if due[i].kind == "function" {
			abilityNo := state.Functions[due[i].abilityIndex]
			switch abilityNo {
			case 0:
				logs = functions.A00000(state, due[i].unitIndex, due[i].abilityIndex, logs)
			case 1:
				if functions.A00001(state, due[i].abilityIndex) {
					logs = append(logs, actions.Action{
						Timestamp:     state.Timestamp,
						Type:          "function",
						UnitIndex:     due[i].unitIndex,
						FunctionIndex: due[i].abilityIndex,
						FunctionNo:    abilityNo,
					})

					defenders := []int{}
					attackers := []int{}
					functions := []int{}
					for i := 0; i < len(state.Zone); i++ {
						if state.Zone[i] == 1 {
							defenders = append(defenders, i)
							attackers = append(attackers, due[i].unitIndex)
							functions = append(functions, due[i].abilityIndex)
						}
					}
					ApplyCounterChain(state, defenders, attackers, functions, 2)
				}
			case 2:
				occurred := functions.A00002(state, dueUnit)
				if occurred {
					state.FunctionCharge[dueAbility] = 0
				}
			case 3:
				targetIdx := functions.A00003(state, dueUnit)
				if targetIdx >= 0 {
					state.FunctionCharge[dueAbility] = 0
					ApplyCounterChain(state, []int{targetIdx}, []int{dueUnit}, []int{dueAbility}, 50)
				}
			case 4:
				fmt.Println("cannot handle counterattack here, there is a bug")
			case 5:
				targetIdx := functions.A00005(state, dueUnit)
				if targetIdx >= 0 {
					state.FunctionCharge[dueAbility] = 0
					ApplyCounterChain(state, []int{targetIdx}, []int{dueUnit}, []int{dueAbility}, 50)
				}
			case 6:
				targetIdx := functions.A00006(state, dueUnit)
				if targetIdx >= 0 {
					state.FunctionCharge[dueAbility] = 0
					ApplyCounterChain(state, []int{targetIdx}, []int{dueUnit}, []int{dueAbility}, 100)
				}
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

	return logs
}
