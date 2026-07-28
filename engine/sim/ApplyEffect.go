package sim

import (
	"fmt"

	"github.com/warebuf/fd3/engine/sim/effects"
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffect(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, eff state.Effect, allEffects []state.Effect, sourceCardIndex int, activeIndex int, executeAtWallTime int64) {
	switch eff.Type {

	case state.EffectDraw:
		effects.ApplyEffectDraw(gs, logList, UpdateID, eff, executeAtWallTime)

	case state.EffectDestroy:
		effects.ApplyEffectDestroyAll(gs, pq, logList, UpdateID, eff, executeAtWallTime)

	case state.EffectMana:
		effects.ApplyEffectMana(gs, logList, UpdateID, eff, executeAtWallTime)

	case state.EffectAttack:
		effects.ApplyEffectAttack(gs, pq, logList, UpdateID, eff, sourceCardIndex, activeIndex, executeAtWallTime)

	case state.EffectHeal:
		effects.ApplyEffectHeal(gs, logList, UpdateID, eff, sourceCardIndex, executeAtWallTime)

	case state.EffectCounter:
		effects.ApplyEffectCounterAndDestroy(gs, pq, logList, UpdateID, sourceCardIndex, executeAtWallTime)

	case state.EffectGainCharge:
		effects.ApplyEffectGainCharge(gs, logList, UpdateID, eff, sourceCardIndex, executeAtWallTime)

	case state.EffectSacrifice:
		effects.ApplyEffectSacrifice(gs, pq, logList, UpdateID, eff, allEffects, sourceCardIndex, executeAtWallTime)

	case state.EffectConsume:
		effects.ApplyEffectConsume(gs, pq, logList, UpdateID, eff, sourceCardIndex, executeAtWallTime)

	case state.EffectSelfDestruct:
		effects.ApplyEffectSelfDestruct(gs, pq, logList, UpdateID, sourceCardIndex, executeAtWallTime)

	case state.EffectConvertToTokens:
		effects.ApplyEffectConvertToTokens(gs, pq, logList, UpdateID, eff, sourceCardIndex, executeAtWallTime)

	case state.EffectAddCharges:
		effects.ApplyEffectAddCharges(gs, logList, UpdateID, eff, sourceCardIndex, executeAtWallTime)

	default:
		fmt.Println("unhandled effect type:", eff.Type)
	}
}
