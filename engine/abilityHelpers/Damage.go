package abilityHelpers

import "github.com/warebuf/fd3/engine"

//

func DamageAll(state *engine.GameState, dmgAmount int) {
	for i := 0; i < len(state.HP); i++ {
		state.HP[i] -= dmgAmount
	}
}

func DamageOne(state *engine.GameState, targetIdx uint, dmgAmount int) {
	state.HP[targetIdx] -= dmgAmount
}
