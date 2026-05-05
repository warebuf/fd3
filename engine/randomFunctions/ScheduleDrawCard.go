package randomFunctions

import (
	"container/heap"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func ScheduleDrawCard(pq *state.PriorityQueue, logList *[]any, UpdateID *int, teamIndex int, t int64) {

	eff := state.Effect{
		Type: state.EffectDraw,
		A:    teamIndex,
		B:    1, // number of cards to draw
	}

	se := state.ChainStep{
		Type:       state.ChainStepOther,
		CardID:     -1, // no ID for prematch draw
		StartTime:  t - constants.LEAD,
		PopTime:    t - constants.LEAD, // must pop 200 ms earlier, so we can draw for the user at t
		OccurTime:  t,
		Effects:    []state.Effect{eff},
		Persistent: false,
	}

	s := &state.Chain{
		PopTime:    t - constants.LEAD,
		ChainSteps: []state.ChainStep{se},
	}

	heap.Push(pq, s)
	*logList = append(*logList, logs.ScheduleChainLog{
		UpdateID:  *UpdateID,
		Timestamp: s.PopTime,
		Type:      logs.ScheduleChain,
		ChainID:   s.ChainID,
		Step:      se,
	})
	*UpdateID += 1

}
