package effects

import (
	"github.com/warebuf/fd3/engine/sim/destruction"
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffectSacrifice(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, eff state.Effect, allEffects []state.Effect, sourceCardIndex int, executeAtWallTime int64) {
	teamIndex := gs.Cards[sourceCardIndex].Team

	targetIndex := eff.A
	valid := false
	for _, ci := range gs.Field[teamIndex] {
		if ci == targetIndex {
			valid = true
			break
		}
	}
	if !valid {
		return // preselected target is gone or was never set, ability fizzles
	}

	damage := gs.Cards[targetIndex].Cost / 10
	for i := range allEffects {
		if allEffects[i].Type == state.EffectAttack {
			allEffects[i].B = damage
			break
		}
	}

	destruction.HandleUnitDestruction(gs, pq, logList, UpdateID, targetIndex, executeAtWallTime)
}
