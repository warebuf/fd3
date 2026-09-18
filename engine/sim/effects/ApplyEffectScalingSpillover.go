package effects

import "github.com/warebuf/fd3/engine/state"

// deals eff.B damage per unit currently on the opponent's field to a random enemy - scaling with
// how developed their board is. If that hit kills its target with damage to spare, the leftover
// carries onto a second random enemy (another unit if one's left, otherwise a structure) via a
// second ApplyEffectAttack call, which already falls back to structures on its own.
func ApplyEffectScalingSpillover(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, eff state.Effect, sourceCardIndex int, executeAtWallTime int64) {
	teamIndex := gs.Cards[sourceCardIndex].Team
	enemyTeam := 1 - teamIndex
	count := len(gs.Field[enemyTeam])
	damage := eff.B * count
	if damage <= 0 {
		return
	}

	primaryEff := state.Effect{Type: state.EffectAttack, A: 0, B: damage}
	_, overkill := ApplyEffectAttack(gs, pq, logList, UpdateID, primaryEff, sourceCardIndex, -1, executeAtWallTime)

	if overkill > 0 {
		spilloverEff := state.Effect{Type: state.EffectAttack, A: 0, B: overkill}
		ApplyEffectAttack(gs, pq, logList, UpdateID, spilloverEff, sourceCardIndex, -1, executeAtWallTime)
	}
}
