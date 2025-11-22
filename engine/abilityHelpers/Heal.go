package abilityHelpers

import "github.com/warebuf/fd3/engine"

//

// does not check if dead
// will heal over maxHP
func HealOne(state *engine.GameState, targetIdx int, healAmount int) bool {
	if state.HP[targetIdx] > 0 {
		state.HP[targetIdx] += healAmount
		return true
	}
	return false
}
