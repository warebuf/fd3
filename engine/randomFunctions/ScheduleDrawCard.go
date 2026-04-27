package randomFunctions

import (
	"container/heap"

	"github.com/warebuf/fd3/engine/state"
)

func ScheduleDrawCard(pq *state.PriorityQueue, teamIndex int, t int64) {

	eff := state.Effect{
		Type: state.EffectDraw,
		A:    teamIndex,
		B:    1, // number of cards to draw
	}

	se := state.StackEntry{
		Type:    3,       // other
		CardID:  -1,      // no ID for prematch draw
		PopTime: t - 200, // must pop 200 ms earlier, so we can draw for the user at t
		Effects: []state.Effect{eff},
	}

	s := &state.Stack{
		PopTime:      t - 200, // must pop 200 ms earlier, so we can draw for the user at t
		StackEntries: []state.StackEntry{se},
		// note EffectIndex is set inside heap.Push
	}

	heap.Push(pq, s)

}
