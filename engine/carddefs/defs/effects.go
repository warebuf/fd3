package defs

import (
	"github.com/warebuf/fd3/engine/stackPQ"
	"github.com/warebuf/fd3/engine/states"
)

const (
	EffectAttack = iota + 1
	EffectHeal
	EffectDrawCard
	EffectGainMana
	EffectSilence
	EffectGrantAbility
)

type Effect struct {
	Type       int
	TargetRule int
	Amount     int
}

func applyEff(gs *states.GameState, pq *stackPQ.PriorityQueue, eff Effect) {
	switch eff.Type {
	case EffectAttack:
		// placeholder
	case EffectHeal:
		// placeholder
	case EffectDrawCard:
		// placeholder
	case EffectGainMana:
		// placeholder
	case EffectSilence:
		// placeholder
	case EffectGrantAbility:
		// placeholder
	}
}
