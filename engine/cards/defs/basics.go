package defs

import (
	"github.com/warebuf/fd3/engine/stackPQ"
	"github.com/warebuf/fd3/engine/states"
)

const (
	EffectAttack       = iota + 1 // 1
	EffectHeal                    // 2
	EffectDrawCard                // 3
	EffectGainMana                // 4
	EffectSilence                 // 5
	EffectGrantAbility            // 6
)

// the signature for every effect
type EffectSignature func(gs *states.GameState, PQ *stackPQ.PriorityQueue, eff Effect)

// EffectRegistry is basically a global variable on the package level
var EffectRegistry = map[int]EffectSignature{
	EffectAttack: func(gs *states.GameState, pq *stackPQ.PriorityQueue, eff Effect) {
		// placeholder
	},

	EffectHeal: func(gs *states.GameState, pq *stackPQ.PriorityQueue, eff Effect) {
		// placeholder
	},

	EffectDrawCard: func(gs *states.GameState, pq *stackPQ.PriorityQueue, eff Effect) {
		// placeholder
	},

	EffectGainMana: func(gs *states.GameState, pq *stackPQ.PriorityQueue, eff Effect) {
		// placeholder
	},

	EffectSilence: func(gs *states.GameState, pq *stackPQ.PriorityQueue, eff Effect) {
		// placeholder
	},

	EffectGrantAbility: func(gs *states.GameState, pq *stackPQ.PriorityQueue, eff Effect) {
		// placeholder
	},
}

func ExecuteEffect(gs *states.GameState, pq *stackPQ.PriorityQueue, eff Effect) {
	fn, ok := EffectRegistry[eff.Kind]
	if !ok {
		return
	}
	fn(gs, pq, eff)
}
