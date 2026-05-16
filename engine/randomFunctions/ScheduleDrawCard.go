package randomFunctions

import (
	"container/heap"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func ScheduleDrawCard(pq *state.SimulationPQ, logList *[]any, UpdateID *int, teamIndex int, occurTime int64) {

	// schedule the draw
	item := &state.SimulationItem{
		StartTime: occurTime,
		PopTime:   occurTime - constants.LEAD,
		OccurTime: occurTime,
		CardID:    -1,
		Type:      state.SimItemDraw,
		Effects:   []state.Effect{{Type: state.EffectDraw, A: teamIndex, B: 1}},
	}
	heap.Push(pq, item)

	// add the scheduling log
	*logList = append(*logList, logs.AddSimItemLog{
		UpdateID:  *UpdateID,
		Timestamp: occurTime,
		Type:      logs.AddSimItem,

		SeqID:       item.SeqID,
		StartTime:   item.StartTime,
		OccurTime:   item.OccurTime,
		CardID:      item.CardID,
		SimItemType: int(state.SimItemDraw),
	})
	*UpdateID += 1
}
