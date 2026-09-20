package effects

import "github.com/warebuf/fd3/engine/state"

// attacks a random enemy for eff.B damage, and only if that hit lands drains eff.C mana from the opponent
func ApplyEffectLeech(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, eff state.Effect, sourceCardIndex int, executeAtWallTime int64) {
	if eff.B <= 0 {
		return
	}

	attackEff := state.Effect{Type: state.EffectAttack, A: 0, B: eff.B}
	result := ApplyEffectAttack(gs, pq, logList, UpdateID, attackEff, sourceCardIndex, -1, executeAtWallTime)
	if !result.Landed {
		return
	}

	drainEff := state.Effect{Type: state.EffectDrainMana, A: gs.Cards[sourceCardIndex].Team, B: eff.C}
	ApplyEffectDrainMana(gs, logList, UpdateID, drainEff, executeAtWallTime)
}
