package abilityHelpers

import "github.com/warebuf/fd3/engine"

//

func DamageAll(state *engine.GameState, dmgAmount int) {
	for i := 0; i < len(state.HP); i++ {
		if state.HP[i] > 0 && state.Zone[i] == 1 {
			state.HP[i] -= dmgAmount
		}
	}
}

func DamageOne(state *engine.GameState, targetIdx int, dmgAmount int) {
	state.HP[targetIdx] -= dmgAmount
}
