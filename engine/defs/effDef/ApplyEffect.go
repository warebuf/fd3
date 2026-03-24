package defs

import (
	"github.com/warebuf/fd3/engine/states/gameState"
	"github.com/warebuf/fd3/engine/states/stackPQ"
)

func ApplyEffect(gs *gameState.GameState, pq *stackPQ.PriorityQueue, eff gameState.Effect) {
	switch eff.Type {
	case gameState.EffectAttack:
		// placeholder
	case gameState.EffectHeal:
		// placeholder
	case gameState.EffectDraw:
		// placeholder
	case gameState.EffectMana:
		// placeholder
	case gameState.EffectSilence:
		// placeholder
	case gameState.EffectGrantAbility:
		// placeholder
	}
}
