package simulate

import (
	"github.com/warebuf/fd3/engine"
)

// TODO: between tiebreakers, not sure who goes first

func sortDue(state *engine.GameState, due []Due) []Due {

	for i := 0; i < len(due); i++ {

		if due[i].kind == "summon" {
			due[i].prio = 0
		} else if due[i].kind == "function" { // heal has prio over damage
			switch due[i].abilityIndex {
			case 0:
				due[i].prio = 1
			case 1:
				due[i].prio = 2
			case 2:
				due[i].prio = 1
			case 3:
				due[i].prio = 2
			case 4: // counter, should not be sorted because cannot be activated by itself
				due[i].prio = 3
			case 5:
				due[i].prio = 2
			case 6:
				due[i].prio = 2
			}
		}

	}

	return due

}
