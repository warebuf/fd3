package schedule

import (
	"container/heap"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/state"
)

// the timer rides the SimulationPQ, so it only advances while the game is simulating
func ScheduleTrapReady(gs *state.GameState, pq *state.SimulationPQ, cardIndex int, startAtSimTime int64) {
	card := &gs.Cards[cardIndex]

	scheduledTime := startAtSimTime + constants.TRAP_SET_TIME/constants.SIM_SPEED - constants.LEAD
	item := &state.SimulationItem{
		StartAtSimTime:   startAtSimTime,
		PopAtSimTime:     scheduledTime,
		ExecuteAtSimTime: scheduledTime + constants.LEAD,

		AbilityIndex: -1, // not an ability slot's timer
		CardIndex:    cardIndex,
		TeamIndex:    card.Team,
		Type:         state.SimItemTrapReady,
	}
	heap.Push(pq, item)

	card.SetStartAtSimTime = item.StartAtSimTime
	card.ReadyAtSimTime = item.ExecuteAtSimTime
	card.Ready = false
}
