package simulate

import (
	"container/heap"
	"time"

	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/simulate/abilities"
)

func AddSummonToPQ(state *engine.GameState, pq *PriorityQueue, unitIndex int, currentTime time.Time, newEffectIndex uint64) bool {

	// check if unit is a valid summon
	if state.HP[unitIndex] <= 0 || state.Zone[unitIndex] != 0 {
		return false
	}

	// set zone to 3 (pending summon)
	state.Zone[unitIndex] = 3

	WINDOW := int(5000) // HARDCODED number

	// create the PQ object
	var event Node
	modelNo := state.ModelNo[unitIndex]
	switch modelNo {
	case 0:
		event = abilities.S00000{
			SummonedIndex:   unitIndex,
			DeclarationTime: currentTime.UnixMilli(),
			Window:          WINDOW,
			Valid:           true,
		}
	}

	// for each summon, create a new stack, add the new PQ object to it
	newStack := &Stack{
		EffectIndex:     newEffectIndex,
		DeclarationTime: currentTime.UnixMilli(),
		WindowTime:      WINDOW, // 5 seconds to react
		ResolutionTime:  currentTime.UnixMilli() + int64(WINDOW),
		Effects:         []Node{event},
		index:           0, // heap.Push assigns a correct index
	}
	heap.Push(pq, newStack)

	// apply any on summon declaration effects

	return true
}
