package randomFunctions

import (
	"container/heap"

	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/stackPQ"
	"github.com/warebuf/fd3/engine/state"
)

func ScheduleDeclareCard(gs *state.GameState, pq *stackPQ.PriorityQueue, effectIndex uint64, eventLogs *[]logs.Log, logIndex int, teamIndex int, cardIndex int, popTime int64) bool {

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

	// create a log for the activation
	*eventLogs = append(*eventLogs, logs.Log{
		UpdateID:  logIndex,
		Timestamp: popTime + 200, // clients are allowed to receive logs 200ms early, but see them in 200ms
		Type:      logs.DrawExec,
		A:         teamIndex, // A: teamIndex
		B:         cardIndex, // B: num of cards to draw
	})

	// schedule the resolution
	eff := state.Effect{
		Type: state.SummonResolution,
		A:    teamIndex,
		B:    cardIndex, // index of card to be summoned
	}

	s := &stackPQ.Stack{
		EffectIndex: effectIndex,
		PopTime:     popTime + 200 + 2000 - 200,
		Effects:     []state.Effect{eff},
	}

	heap.Push(pq, s)

	return true

}
