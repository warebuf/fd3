package effects

import (
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffectSelfDestruct(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, sourceCardIndex int, executeAtWallTime int64) {
	if gs.Cards[sourceCardIndex].Type == state.CardTypeUnit {
		HandleUnitDestruction(gs, pq, logList, UpdateID, sourceCardIndex, executeAtWallTime)
	} else if gs.Cards[sourceCardIndex].Type == state.CardTypeContinuous {
		HandleContinuousDestruction(gs, pq, logList, UpdateID, sourceCardIndex, executeAtWallTime)
	}
}
