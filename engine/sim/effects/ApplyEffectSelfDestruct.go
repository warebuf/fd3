package effects

import (
	"github.com/warebuf/fd3/engine/sim/destruction"
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffectSelfDestruct(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, sourceCardIndex int, executeAtWallTime int64) {
	if gs.Cards[sourceCardIndex].Type == state.CardTypeUnit {
		destruction.HandleUnitDestruction(gs, pq, logList, UpdateID, sourceCardIndex, executeAtWallTime)
	} else if gs.Cards[sourceCardIndex].Type == state.CardTypeTrap {
		destruction.HandleTrapDestruction(gs, pq, logList, UpdateID, sourceCardIndex, executeAtWallTime)
	} else if gs.Cards[sourceCardIndex].Type == state.CardTypeEquip {
		destruction.HandleEquipDestruction(gs, pq, logList, UpdateID, sourceCardIndex, executeAtWallTime)
	}
}
