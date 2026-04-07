package defs

import (
	"time"

	"github.com/warebuf/fd3/engine/defs/effDef/effects"
	"github.com/warebuf/fd3/engine/states/gameState"
	"github.com/warebuf/fd3/engine/states/stackPQ"
)

func ManageStackPQ(gs *gameState.GameState, pq *stackPQ.PriorityQueue) {

	for {

		// get the top stack
		topStack := pq.Peek()
		if topStack == nil {
			return
		}

		// if topStack is not due
		if topStack.ResolveTime > time.Now().UnixMilli() {
			return
		}

		// handle the top effect

	}

}

func ApplyTopOfStack(gs *gameState.GameState, stack []gameState.Effect) {

	topEffect := stack[len(stack)-1]

	switch topEffect.Type {
	case gameState.EffectDraw:
		effects.ApplyEffectDraw(gs, topEffect)
	case gameState.EffectAttack:
		// placeholder
	case gameState.EffectHeal:
		// placeholder
	case gameState.EffectMana:
		// placeholder
	case gameState.EffectSilence:
		// placeholder
	case gameState.EffectGrantAbility:
		// placeholder
	}
}
