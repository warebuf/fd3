package effects

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/destruction"
	"github.com/warebuf/fd3/engine/state"
)

// only destroys on field cards
func ApplyEffectDestroyAll(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, eff state.Effect, executeAtWallTime int64) {
	for team := 0; team < len(gs.Field); team++ {
		fieldSnapshot := append([]int{}, gs.Field[team]...)
		for _, cardIndex := range fieldSnapshot {
			destruction.HandleUnitDestruction(gs, pq, logList, UpdateID, cardIndex, executeAtWallTime)
		}
	}

	*logList = append(*logList, logs.DestroyCardLog{
		UpdateID:        *UpdateID,
		WallTimestamp:   executeAtWallTime,
		Type:            logs.DestroyCard,
		TargetAll:       eff.A == 0,
		TargetTeamIndex: -1,
		TargetCardIndex: -1,
	})
	*UpdateID++
}
