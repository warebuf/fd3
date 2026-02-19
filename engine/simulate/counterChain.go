package simulate

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/functions"
)

func ApplyCounterChain(state *engine.GameState, defenderArr []int, attackerArr []int, functionArr []int, dmg int) {

	currentDmg := dmg

	for j := 0; j < len(defenderArr); j++ {

		currDefender := defenderArr[j]
		currAttacker := attackerArr[j]

		// check all abilities for this unit
		for i := state.FunctionStartIdx[currDefender]; i < state.FunctionStartIdx[currDefender]+state.FunctionCount[currDefender]; i++ {
			switch state.Functions[i] {
			case 4: // reflect and mult 1x
				if functions.A00004CounterValid(state, i) {
					currentDmg = functions.A00004(state, currDefender, currAttacker, currentDmg)
					state.FunctionCharge[i] = 0
					defenderArr = append(defenderArr, currAttacker)
					attackerArr = append(attackerArr, currDefender)
					functionArr = append(functionArr, i)
				}
			default:
			}
		}

	}

}
