package sim

import (
	"container/heap"
	"time"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/schedule"
	"github.com/warebuf/fd3/engine/state"
)

func ManageSimulationPQ(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int) {
	for {
		if gs.Winner >= 0 || gs.Winner == -2 {
			return
		}
		earliestSimItem := pq.Peek()
		if earliestSimItem == nil {
			return
		}
		if gs.PausePhase != state.PausePhaseSimulating {
			return
		}
		currentSimTime := time.Now().UnixMilli() - gs.SimPauseOffset
		if earliestSimItem.PopAtSimTime > currentSimTime {
			return
		}
		handleLatestSimulationItem(gs, pq, logList, UpdateID)
	}
}

func handleLatestSimulationItem(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int) {

	currSimItem := heap.Pop(pq).(*state.SimulationItem)

	executeAtWallTime := currSimItem.PopAtSimTime + gs.SimPauseOffset + constants.LEAD
	for i := 0; i < len(currSimItem.Effects); i++ {
		ApplyEffect(gs, pq, logList, UpdateID, currSimItem.Effects[i], currSimItem.CardIndex, -1, executeAtWallTime)
	}

	switch currSimItem.Type {
	case state.SimItemOnTick:
		clearTickAbility(gs.Cards, logList, UpdateID, currSimItem.CardIndex, currSimItem.AbilityIndex, executeAtWallTime)
		schedule.ScheduleOnTickAbility(gs, pq, logList, UpdateID, currSimItem.CardIndex, currSimItem.AbilityIndex, currSimItem.PopAtSimTime+constants.LEAD, executeAtWallTime)

	case state.SimItemStructureRegen:
		schedule.ScheduleStructureManaRegen(gs, pq, logList, UpdateID, currSimItem.TeamIndex, currSimItem.PopAtSimTime+constants.LEAD)

	case state.SimItemSystemDraw:
		// the draw effect was already applied above; nothing further to schedule
	}
}

// we require management of ontick abilities on the gameState because its the best way to display in the UI
func clearTickAbility(cards []state.Card, logList *[]any, UpdateID *int, cardIndex int, abilityIndex int, wallTime int64) {
	if !cards[cardIndex].OnTick[abilityIndex].Active {
		return
	}
	cards[cardIndex].OnTick[abilityIndex] = state.TickAbility{}
	*logList = append(*logList, logs.ClearTickAbilityLog{
		UpdateID:      *UpdateID,
		WallTimestamp: wallTime,
		Type:          logs.ClearTickAbility,
		CardIndex:     cardIndex,
		AbilityIndex:  abilityIndex,
	})
	*UpdateID++
}
