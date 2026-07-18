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
		if handleLatestSimulationItem(gs, pq, logList, UpdateID) {
			return // a reaction was pushed onto the stack; let it resolve before draining more
		}
	}
}

func handleLatestSimulationItem(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int) bool {

	currSimItem := heap.Pop(pq).(*state.SimulationItem)

	executeAtWallTime := currSimItem.PopAtSimTime + gs.SimPauseOffset + constants.LEAD

	pushed := false
	switch currSimItem.Type {
	case state.SimItemOnTick:
		pushed = pushOnTickAbilityOnStack(gs, logList, UpdateID, currSimItem, executeAtWallTime)
		clearTickAbility(gs.Cards, logList, UpdateID, currSimItem.CardIndex, currSimItem.AbilityIndex, executeAtWallTime)
		schedule.ScheduleOnTickAbility(gs, pq, logList, UpdateID, currSimItem.CardIndex, currSimItem.AbilityIndex, currSimItem.PopAtSimTime+constants.LEAD, executeAtWallTime)

	case state.SimItemStructureRegen:
		for i := 0; i < len(currSimItem.Effects); i++ {
			ApplyEffect(gs, pq, logList, UpdateID, currSimItem.Effects[i], currSimItem.Effects, currSimItem.CardIndex, -1, executeAtWallTime)
		}
		schedule.ScheduleStructureManaRegen(gs, pq, logList, UpdateID, currSimItem.TeamIndex, currSimItem.PopAtSimTime+constants.LEAD)

	case state.SimItemSystemDraw:
		for i := 0; i < len(currSimItem.Effects); i++ {
			ApplyEffect(gs, pq, logList, UpdateID, currSimItem.Effects[i], currSimItem.Effects, currSimItem.CardIndex, -1, executeAtWallTime)
		}
	}
	return pushed
}

// ActiveIndex is -1: OnTick abilities aren't Active-slot-sourced.
func pushOnTickAbilityOnStack(gs *state.GameState, logList *[]any, UpdateID *int, currSimItem *state.SimulationItem, executeAtWallTime int64) bool {
	if len(currSimItem.Effects) == 0 {
		return false
	}

	startAtActiveTime := executeAtWallTime - gs.FullPauseOffset
	reactionItem := state.ReactionItem{
		StartAtActiveTime:   startAtActiveTime,
		PopAtActiveTime:     startAtActiveTime + constants.REACTION_WINDOW - constants.LEAD,
		ExecuteAtActiveTime: startAtActiveTime + constants.REACTION_WINDOW,
		CardIndex:           currSimItem.CardIndex,
		ActiveIndex:         -1,
		Type:                state.ReactionUnitAbility,
		Effects:             currSimItem.Effects,
	}
	gs.ReactionStack = append(gs.ReactionStack, reactionItem)

	*logList = append(*logList, logs.AddStackLog{
		UpdateID:            *UpdateID,
		WallTimestamp:       executeAtWallTime,
		Type:                logs.AddStack,
		StackType:           int(reactionItem.Type),
		StartAtActiveTime:   reactionItem.StartAtActiveTime,
		PopAtActiveTime:     reactionItem.PopAtActiveTime,
		ExecuteAtActiveTime: reactionItem.ExecuteAtActiveTime,
		CardIndex:           reactionItem.CardIndex,
		Effects:             reactionItem.Effects,
	})
	*UpdateID++
	return true
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
