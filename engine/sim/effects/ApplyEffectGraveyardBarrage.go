package effects

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

// fires one random-target attack, each dealing eff.B damage, for every card currently in the
// source's own team's graveyard. Reuses ApplyEffectAttack per hit, so taunt/Bodyguard redirect and
// the random-structure fallback (once the enemy field is empty) all apply exactly like a normal
// random attack would; a unit killed by an earlier hit is correctly gone for the next one. Whenever
// a hit actually lands on a structure rather than a unit, the caster also drains mana equal to that
// hit's damage.
func ApplyEffectGraveyardBarrage(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, eff state.Effect, sourceCardIndex int, executeAtWallTime int64) {
	teamIndex := gs.Cards[sourceCardIndex].Team
	hits := len(gs.Grave[teamIndex])
	if hits == 0 || eff.B <= 0 {
		return
	}

	attackEff := state.Effect{Type: state.EffectAttack, A: 0, B: eff.B}
	for i := 0; i < hits; i++ {
		if result := ApplyEffectAttack(gs, pq, logList, UpdateID, attackEff, sourceCardIndex, -1, executeAtWallTime); result.HitStructure {
			gs.Mana[teamIndex] += eff.B
			*logList = append(*logList, logs.GainManaLog{
				UpdateID:      *UpdateID,
				WallTimestamp: executeAtWallTime,
				Type:          logs.GainMana,
				TeamIndex:     teamIndex,
				Amount:        eff.B,
			})
			*UpdateID++
		}
	}
}
