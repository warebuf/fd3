package randomFunctions

import (
	"container/heap"

	"github.com/warebuf/fd3/engine/stackPQ"
	"github.com/warebuf/fd3/engine/state"
)

func ScheduleActivateCard(gs *state.GameState, pq *stackPQ.PriorityQueue, effectIndex uint64, teamIndex int, cardIndex int, popTime int64) bool {

	// check if the card is in hand
	found := false
	for i := 0; i < len(gs.Hand[teamIndex]); i++ {
		if cardIndex == gs.Hand[teamIndex][i] {
			found = true
		}
	}
	if !found {
		return false
	}

	// check if the card meets the activation conditions

	// TODO: create a log for the activation

	// schedule the resolution
	eff := state.Effect{
		Type: state.EffectActivate,
		A:    teamIndex,
		B:    cardIndex, // index of card to be summoned
	}

	s := &stackPQ.Stack{
		EffectIndex: effectIndex,
		PopTime:     popTime,
		ResolveTime: popTime + 200,
		Effects:     []state.Effect{eff},
	}

	heap.Push(pq, s)

	return true

}
