package microevents

import (
	"container/heap"

	"github.com/warebuf/fd3/engine/states/gameState"
	"github.com/warebuf/fd3/engine/states/stackPQ"
)

func ScheduleDrawCard(pq *stackPQ.PriorityQueue, effectIndex uint64, teamIndex int, t int64) {

	eff := gameState.Effect{
		Type: gameState.EffectDraw,
		A:    teamIndex,
		B:    1, // number of cards to draw
	}

	s := &stackPQ.Stack{
		EffectIndex: effectIndex,
		DeclareTime: t,
		ResolveTime: t,
		Effects:     []gameState.Effect{eff},
	}

	heap.Push(pq, s)

}
