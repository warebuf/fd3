package randomFunctions

import (
	"container/heap"

	"github.com/warebuf/fd3/engine/stackPQ"
	"github.com/warebuf/fd3/engine/state"
)

func ScheduleDrawCard(pq *stackPQ.PriorityQueue, effectIndex uint64, teamIndex int, t int64) {

	eff := state.Effect{
		Type: state.EffectDraw,
		A:    teamIndex,
		B:    1, // number of cards to draw
	}

	s := &stackPQ.Stack{
		EffectIndex: effectIndex,
		DeclareTime: t,
		ResolveTime: t - 200, // note that everything in the stackPQ must be scheduled 200 ms early
		Effects:     []state.Effect{eff},
	}

	heap.Push(pq, s)

}
