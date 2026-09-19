package effects

import "github.com/warebuf/fd3/engine/state"

// attacks a random enemy for eff.B damage, and only if that hit lands drains eff.C mana and heals the source by the HP it actually removed
func ApplyEffectLifeLeech(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, eff state.Effect, sourceCardIndex int, executeAtWallTime int64) {
	if eff.B <= 0 {
		return
	}
	teamIndex := gs.Cards[sourceCardIndex].Team

	attackEff := state.Effect{Type: state.EffectAttack, A: 0, B: eff.B}
	result := ApplyEffectAttack(gs, pq, logList, UpdateID, attackEff, sourceCardIndex, -1, executeAtWallTime)
	if !result.Landed {
		return
	}

	drainEff := state.Effect{Type: state.EffectDrainMana, A: teamIndex, B: eff.C}
	ApplyEffectDrainMana(gs, logList, UpdateID, drainEff, executeAtWallTime)

	healEff := state.Effect{Type: state.EffectHealSelf, B: result.Dealt}
	ApplyEffectHealSelf(gs, logList, UpdateID, healEff, sourceCardIndex, executeAtWallTime)
}
