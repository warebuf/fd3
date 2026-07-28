package schedule

import (
	"container/heap"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/state"
)

func ScheduleSystemDraw(gs *state.GameState, pq *state.SimulationPQ, teamIndex int, executeAtWallTime int64) {
	executeAtSimTime := executeAtWallTime - gs.SimPauseOffset

	item := &state.SimulationItem{
		StartAtSimTime:   executeAtSimTime,
		PopAtSimTime:     executeAtSimTime - constants.LEAD,
		ExecuteAtSimTime: executeAtSimTime,
		AbilityIndex:     -1,
		CardIndex:        -1,
		TeamIndex:        teamIndex,
		Type:             state.SimItemSystemDraw,
		Effects:          []state.Effect{{Type: state.EffectDraw, A: teamIndex, B: 1}},
	}
	heap.Push(pq, item)
}
