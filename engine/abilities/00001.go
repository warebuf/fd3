package abilities

import (
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/abilityHelpers"
)

// deal 2 damage to all

func A00001(state *engine.GameState) bool {
	abilityHelpers.DamageAll(state, 2)
	return true
}
