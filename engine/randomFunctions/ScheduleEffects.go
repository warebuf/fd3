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
		PopTime:     t - 200, // must pop 200 ms earlier, so we can draw for the user at t
		ResolveTime: t,
		Effects:     []state.Effect{eff},
	}

	heap.Push(pq, s)

}
