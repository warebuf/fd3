package defs

import (
	"github.com/warebuf/fd3/engine/stackPQ"
	"github.com/warebuf/fd3/engine/states"
)

// the signature for every effect
type EffectSignature func(gs *states.GameState, PQ *stackPQ.PriorityQueue, eff Effect)

// EffectRegistry is basically a global variable on the package level
var EffectRegistry = map[int]EffectFunc{
	EffectDamage: func(gs *GameState, source *Permanent, targets []*Permanent, e Effect) {
		for _, t := range targets {
			t.HP -= e.Amount
		}
	},

	EffectHeal: func(gs *GameState, source *Permanent, targets []*Permanent, e Effect) {
		for _, t := range targets {
			t.HP += e.Amount
		}
	},

	EffectDrawCard: func(gs *GameState, source *Permanent, targets []*Permanent, e Effect) {
		DrawCards(gs, source.Team, e.Amount)
	},

	EffectGainMana: func(gs *GameState, source *Permanent, targets []*Permanent, e Effect) {
		gs.Mana[source.Team] += e.Amount
	},

	EffectSilence: func(gs *GameState, source *Permanent, targets []*Permanent, e Effect) {
		for _, t := range targets {
			// remove all triggers this permanent registered
			UnregisterTriggers(gs, t.PID)
			// suppress all its base abilities
			t.SuppressedKinds = AllKinds(gs, t)
		}
	},

	EffectGrantAbility: func(gs *GameState, source *Permanent, targets []*Permanent, e Effect) {
		// e.Amount2 holds the trigger kind to grant
		// the actual trigger definition would need to be looked up or passed differently
		// placeholder for now
	},
}

// ExecuteEffect looks up the effect kind in the registry and runs it
func ExecuteEffect(gs *GameState, source *Permanent, targets []*Permanent, e Effect) {
	fn, ok := EffectRegistry[e.Kind]
	if !ok {
		// unknown effect kind, do nothing
		return
	}
	fn(gs, source, targets, e)
}
