package schedule

import (
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/state"
)

func ScheduleAllOnTicks(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, cardIndex int, startAtSimTime int64, wallTimestamp int64) {
	onTick := defs.CardDefs[gs.Cards[cardIndex].DefNo].OnTick
	for a := 0; a < len(onTick); a++ {
		ScheduleOnTickAbility(gs, pq, logList, UpdateID, cardIndex, a, startAtSimTime, wallTimestamp)
	}
}
