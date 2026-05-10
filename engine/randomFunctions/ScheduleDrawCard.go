package randomFunctions

import (
	"container/heap"

	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func ScheduleDrawCard(pq *state.SimulationPQ, logList *[]any, UpdateID *int, teamIndex int, t int64) {

	// schedule the draw
	item := &state.SimulationItem{
		StartTime: t,
		PopTime:   t,
		OccurTime: t,
		CardID:    -1,
		Effects:   []state.Effect{{Type: state.EffectDraw, A: teamIndex, B: 1}},
	}
	heap.Push(pq, item)

	// add the scheduling log
	*logList = append(*logList, logs.ScheduleChainLog{
		UpdateID:  *UpdateID,
		Timestamp: t,
		Type:      logs.ScheduleChain,
		ChainID:   item.SeqID,
		Step:      *item,
	})
	*UpdateID += 1
}
