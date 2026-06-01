package sim

import (
	"container/heap"
	"fmt"
	"time"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/effects"
	"github.com/warebuf/fd3/engine/state"
)

func ManageSimulationPQ(gs *state.GameState, pq *state.SimulationPQ, abilityIDCounter *int, logList *[]any, UpdateID *int) {
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
		handleLatestSimulationItem(gs, pq, abilityIDCounter, logList, UpdateID)
	}
}

func handleLatestSimulationItem(gs *state.GameState, pq *state.SimulationPQ, abilityIDCounter *int, logList *[]any, UpdateID *int) {

	currSimItem := heap.Pop(pq).(*state.SimulationItem)

	executeAtWallTime := currSimItem.PopAtSimTime + gs.SimPauseOffset + constants.LEAD
	for i := 0; i < len(currSimItem.Effects); i++ {
		ApplyEffect(gs, pq, logList, UpdateID, currSimItem.Effects[i], currSimItem.CardID, executeAtWallTime)
	}

	if currSimItem.CardID >= 0 {
		if currSimItem.Type == state.SimItemSummonResolve {
			teamIndex := gs.Cards[currSimItem.CardID].Team
			gs.Cards[currSimItem.CardID].Status = state.CardStatusUnitActive

			*logList = append(*logList, logs.SummonResolutionLog{
				UpdateID:      *UpdateID,
				WallTimestamp: executeAtWallTime,
				Type:          logs.SummonResolution,
				TeamIndex:     teamIndex,
				CardIndex:     currSimItem.CardID,
			})
			*UpdateID += 1

			ScheduleRecurringAbilities(gs, pq, abilityIDCounter, logList, UpdateID, currSimItem.CardID, currSimItem.PopAtSimTime+constants.LEAD)

		} else if currSimItem.Type == state.SimItemRecurring {
			if gs.Cards[currSimItem.CardID].Zone == state.ZoneField {
				for _, t := range gs.Cards[currSimItem.CardID].AbilityTimers {
					if t.AbilityID == currSimItem.AbilityID {
						scheduleRecurringAbility(gs, pq, abilityIDCounter, logList, UpdateID, currSimItem.CardID, t.AbilityIdx, currSimItem.PopAtSimTime+constants.LEAD)
						break
					}
				}
			}
		}

		clearAbilityTimer(gs.Cards, logList, UpdateID, currSimItem.CardID, currSimItem.AbilityID, executeAtWallTime)
	}

	if currSimItem.Type == state.SimItemStructureRegen {
		teamIndex := currSimItem.Effects[0].A
		ScheduleStructureManaRegen(gs, pq, logList, UpdateID, teamIndex, currSimItem.PopAtSimTime+constants.LEAD)
	}
}

func clearAbilityTimer(cards []state.Card, logList *[]any, UpdateID *int, cardID int, abilityID int, wallTime int64) {
	timers := cards[cardID].AbilityTimers
	for i, t := range timers {
		if t.AbilityID == abilityID {
			cards[cardID].AbilityTimers = append(timers[:i], timers[i+1:]...)
			*logList = append(*logList, logs.ClearAbilityTimerLog{
				UpdateID:      *UpdateID,
				WallTimestamp: wallTime,
				Type:          logs.ClearAbilityTimer,
				CardIndex:     cardID,
				AbilityID:     abilityID,
			})
			*UpdateID++
			return
		}
	}
}

func ApplyEffect(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, eff state.Effect, sourceCardID int, executeAtWallTime int64) {
	switch eff.Type {

	case state.EffectDraw:
		effects.ApplyEffectDraw(gs, logList, UpdateID, eff, executeAtWallTime)

	case state.EffectDestroy:
		destroyedIDs := map[int]bool{}
		for a := 0; a < len(gs.Field); a++ {
			for b := 0; b < len(gs.Field[a]); b++ {
				cardIndex := gs.Field[a][b]
				gs.Cards[cardIndex].Zone = state.ZoneGraveyard
				gs.Grave[a] = append(gs.Grave[a], cardIndex)
				destroyedIDs[cardIndex] = true
			}
			gs.Field[a] = []int{}
		}
		pq.RemoveSimItemsByCardIDs(destroyedIDs)
		for cardID := range destroyedIDs {
			gs.Cards[cardID].AbilityTimers = []state.AbilityTimer{}
			gs.Cards[cardID].Status = state.CardStatusDefault
		}
		*logList = append(*logList, logs.DestroyCardLog{
			UpdateID:        *UpdateID,
			WallTimestamp:   executeAtWallTime,
			Type:            logs.DestroyCard,
			TeamIndex:       gs.Cards[eff.C].Team,
			CardIndex:       eff.C,
			TargetAll:       eff.A == 0,
			TargetTeamIndex: -1,
			TargetCardIndex: -1,
		})
		*UpdateID += 1

	case state.EffectMana:
		gs.Mana[eff.A] += eff.B
		*logList = append(*logList, logs.GainManaLog{
			UpdateID:      *UpdateID,
			WallTimestamp: executeAtWallTime,
			Type:          logs.GainMana,
			TeamIndex:     eff.A,
			Amount:        eff.B,
		})
		*UpdateID += 1

	case state.EffectAttack:
		effects.ApplyEffectAttack(gs, pq, logList, UpdateID, eff, sourceCardID, executeAtWallTime)

	default:
		fmt.Println("unhandled effect type:", eff.Type)
	}
}

func FulfillEffectTemplate(templ state.Effect, teamIndex int, cardIndex int) state.Effect {
	switch templ.Type {
	case state.EffectDestroy:
		templ.C = cardIndex
		return templ
	case state.EffectMana:
		templ.A = teamIndex
		return templ
	case state.EffectAttack:
		return templ
	}
	return templ
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
		AbilityID:        -1,
		CardID:           -1,
		Type:             state.SimItemStructureRegen,
		Effects:          []state.Effect{{Type: state.EffectMana, A: teamIndex, B: structDef.ManaRegen}},
	})
}

func ScheduleRecurringAbilities(gs *state.GameState, pq *state.SimulationPQ, abilityIDCounter *int, logList *[]any, UpdateID *int, cardID int, startTime int64) {
	recurring := defs.CardDefs[gs.Cards[cardID].DefNo].Recurring
	for a := 0; a < len(recurring); a++ {
		scheduleRecurringAbility(gs, pq, abilityIDCounter, logList, UpdateID, cardID, a, startTime)
	}
}

func scheduleRecurringAbility(gs *state.GameState, pq *state.SimulationPQ, abilityIDCounter *int, logList *[]any, UpdateID *int, cardID int, abilityIdx int, startAtSimTime int64) {
	teamIndex := gs.Cards[cardID].Team
	entry := defs.CardDefs[gs.Cards[cardID].DefNo].Recurring[abilityIdx]

	effs := make([]state.Effect, len(entry.Effects))
	for b := 0; b < len(entry.Effects); b++ {
		effs[b] = FulfillEffectTemplate(entry.Effects[b], teamIndex, cardID)
	}

	abilityID := *abilityIDCounter
	*abilityIDCounter++

	scheduledTime := startAtSimTime + entry.Milliseconds - constants.LEAD
	item := &state.SimulationItem{
		StartAtSimTime:   startAtSimTime,
		PopAtSimTime:     scheduledTime,
		ExecuteAtSimTime: scheduledTime + constants.LEAD,

		AbilityID: abilityID,
		CardID:    cardID,
		Type:      state.SimItemRecurring,
		Effects:   effs,
	}
	heap.Push(pq, item)

	gs.Cards[cardID].AbilityTimers = append(gs.Cards[cardID].AbilityTimers, state.AbilityTimer{
		AbilityID:        abilityID,
		AbilityIdx:       abilityIdx,
		Type:             state.AbilityTimerNormal,
		StartAtSimTime:   item.StartAtSimTime,
		ExecuteAtSimTime: item.ExecuteAtSimTime,
	})
	*logList = append(*logList, logs.AddAbilityTimerLog{
		UpdateID:         *UpdateID,
		WallTimestamp:    startAtSimTime + gs.SimPauseOffset,
		Type:             logs.AddAbilityTimer,
		CardIndex:        cardID,
		AbilityID:        abilityID,
		AbilityIdx:       abilityIdx,
		TimerType:        state.AbilityTimerNormal,
		StartAtSimTime:   item.StartAtSimTime,
		ExecuteAtSimTime: item.ExecuteAtSimTime,
	})
	*UpdateID++
}
