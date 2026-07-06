package sim

import (
	"container/heap"
	"time"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

// not sure where this belongs, maybe not sim package
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
		ScheduleOnTickAbility(gs, pq, logList, UpdateID, currSimItem.CardIndex, currSimItem.AbilityIndex, currSimItem.PopAtSimTime+constants.LEAD, executeAtWallTime)

	case state.SimItemStructureRegen:
		ScheduleStructureManaRegen(gs, pq, logList, UpdateID, currSimItem.TeamIndex, currSimItem.PopAtSimTime+constants.LEAD)

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

func ScheduleStructureManaRegen(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, teamIndex int, startAtSimTime int64) {

	structDef := defs.StructureDefs[gs.Structures[teamIndex].DefNo]
	if structDef.ManaRegen == 0 || structDef.ManaRegenInterval == 0 {
		return
	}
	scheduledTime := startAtSimTime + structDef.ManaRegenInterval - constants.LEAD
	heap.Push(pq, &state.SimulationItem{
		StartAtSimTime:   startAtSimTime,
		PopAtSimTime:     scheduledTime,
		ExecuteAtSimTime: scheduledTime + constants.LEAD,
		AbilityIndex:     -1,
		CardIndex:        -1,
		TeamIndex:        teamIndex,
		Type:             state.SimItemStructureRegen,
		Effects:          []state.Effect{{Type: state.EffectMana, A: teamIndex, B: structDef.ManaRegen}},
	})

}

func ScheduleAllOnTicks(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, cardIndex int, startAtSimTime int64, wallTimestamp int64) {
	onTick := defs.CardDefs[gs.Cards[cardIndex].DefNo].OnTick
	for a := 0; a < len(onTick); a++ {
		ScheduleOnTickAbility(gs, pq, logList, UpdateID, cardIndex, a, startAtSimTime, wallTimestamp)
	}
}

func ScheduleOnTickAbility(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, cardIndex int, abilityIndex int, startAtSimTime int64, wallTimestamp int64) {
	card := gs.Cards[cardIndex]
	if card.Zone == state.ZoneGraveyard {
		return // card is dead, nothing to schedule
	}

	entry := defs.CardDefs[card.DefNo].OnTick[abilityIndex]

	if entry.Milliseconds == 0 {
		return // unused OnTick slot
	}

	// a gain-charge ability stops rescheduling once its targeted Active slot maxes out
	condition := state.TickIndefinite
	activeIndex := -1
	for _, eff := range entry.Effects {
		if eff.Type == state.EffectGainCharge {
			condition = state.TickUntilChargeMax
			activeIndex = eff.A
			active := card.Actives[activeIndex]
			if active.MaxCharges > 0 && active.Charges >= active.MaxCharges {
				return
			}
			break
		}
	}

	teamIndex := card.Team

	var effs []state.Effect
	for _, e := range entry.Effects {
		if e.Type == 0 {
			continue // unused Effect slot
		}
		effs = append(effs, FulfillEffectTemplate(e, teamIndex, cardIndex))
	}

	scheduledTime := startAtSimTime + entry.Milliseconds - constants.LEAD
	item := &state.SimulationItem{
		StartAtSimTime:   startAtSimTime,
		PopAtSimTime:     scheduledTime,
		ExecuteAtSimTime: scheduledTime + constants.LEAD,

		AbilityIndex: abilityIndex,
		CardIndex:    cardIndex,
		TeamIndex:    teamIndex,
		Type:         state.SimItemOnTick,
		Effects:      effs,
	}
	heap.Push(pq, item)

	gs.Cards[cardIndex].OnTick[abilityIndex] = state.TickAbility{
		Active:           true,
		Condition:        condition,
		ActiveIndex:      activeIndex,
		StartAtSimTime:   item.StartAtSimTime,
		ExecuteAtSimTime: item.ExecuteAtSimTime,
	}

	*logList = append(*logList, logs.AddTickAbilityLog{
		UpdateID:         *UpdateID,
		WallTimestamp:    wallTimestamp,
		Type:             logs.AddTickAbility,
		CardIndex:        cardIndex,
		AbilityIndex:     abilityIndex,
		Condition:        condition,
		ActiveIndex:      activeIndex,
		StartAtSimTime:   item.StartAtSimTime,
		ExecuteAtSimTime: item.ExecuteAtSimTime,
	})
	*UpdateID++
}
