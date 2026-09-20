package effects

import "github.com/warebuf/fd3/engine/state"

// spawns eff.C copies of the token CardDef eff.B onto the source card's controller's field
func ApplyEffectSpawnTokens(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, eff state.Effect, sourceCardIndex int, executeAtWallTime int64) {
	teamIndex := gs.Cards[sourceCardIndex].Team
	for i := 0; i < eff.C; i++ {
		spawnToken(gs, pq, logList, UpdateID, eff.B, teamIndex, executeAtWallTime)
	}
}
