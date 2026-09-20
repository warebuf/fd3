package sim

import (
	"fmt"

	"github.com/warebuf/fd3/engine/sim/effects"
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffect(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, eff state.Effect, sourceCardIndex int, activeIndex int, executeAtWallTime int64) {
	switch eff.Type {

	case state.EffectDraw:
		effects.ApplyEffectDraw(gs, logList, UpdateID, eff, executeAtWallTime)

	case state.EffectDestroy:
		effects.ApplyEffectDestroyAll(gs, pq, logList, UpdateID, eff, executeAtWallTime)

	case state.EffectMana:
		effects.ApplyEffectMana(gs, logList, UpdateID, eff, executeAtWallTime)

	case state.EffectAttack:
		effects.ApplyEffectAttack(gs, pq, logList, UpdateID, eff, sourceCardIndex, activeIndex, executeAtWallTime)

	case state.EffectCounter:
		effects.ApplyEffectCounterAndDestroy(gs, pq, logList, UpdateID, sourceCardIndex, executeAtWallTime)

	case state.EffectChargeSelf:
		effects.ApplyEffectChargeSelf(gs, logList, UpdateID, eff, sourceCardIndex, executeAtWallTime)

	case state.EffectConsume:
		effects.ApplyEffectConsume(gs, pq, logList, UpdateID, eff, sourceCardIndex, executeAtWallTime)

	case state.EffectSelfDestruct:
		effects.ApplyEffectSelfDestruct(gs, pq, logList, UpdateID, sourceCardIndex, executeAtWallTime)

	case state.EffectConvertToTokens:
		effects.ApplyEffectConvertToTokens(gs, pq, logList, UpdateID, eff, sourceCardIndex, executeAtWallTime)

	case state.EffectChargeTarget:
		effects.ApplyEffectChargeTarget(gs, logList, UpdateID, eff, sourceCardIndex, executeAtWallTime)

	case state.EffectGiveControl:
		effects.ApplyEffectGiveControl(gs, logList, UpdateID, eff, sourceCardIndex, executeAtWallTime)

	case state.EffectHealNonDamaging:
		effects.ApplyEffectHealNonDamaging(gs, logList, UpdateID, eff, executeAtWallTime)

	case state.EffectExtort:
		effects.ApplyEffectExtort(gs, logList, UpdateID, eff, sourceCardIndex, executeAtWallTime)

	case state.EffectDrainMana:
		effects.ApplyEffectDrainMana(gs, logList, UpdateID, eff, executeAtWallTime)

	case state.EffectLiquidate:
		effects.ApplyEffectLiquidate(gs, pq, logList, UpdateID, eff, executeAtWallTime)

	case state.EffectMutualDestruction:
		effects.ApplyEffectMutualDestruction(gs, pq, logList, UpdateID, eff, sourceCardIndex, executeAtWallTime)

	case state.EffectGraveyardBarrage:
		effects.ApplyEffectGraveyardBarrage(gs, pq, logList, UpdateID, eff, sourceCardIndex, executeAtWallTime)

	case state.EffectScalingSpillover:
		effects.ApplyEffectScalingSpillover(gs, pq, logList, UpdateID, eff, sourceCardIndex, executeAtWallTime)

	case state.EffectSpawnTokens:
		effects.ApplyEffectSpawnTokens(gs, pq, logList, UpdateID, eff, sourceCardIndex, executeAtWallTime)

	case state.EffectLeech:
		effects.ApplyEffectLeech(gs, pq, logList, UpdateID, eff, sourceCardIndex, executeAtWallTime)

	default:
		fmt.Println("unhandled effect type:", eff.Type)
	}
}
