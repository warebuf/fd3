package abilityHelpers

import "github.com/warebuf/fd3/engine"

//

// does not check if dead
// will heal over maxHP
func HealOne(state *engine.GameState, targetIdx int, healAmount int) {
	state.HP[targetIdx] += healAmount
}
