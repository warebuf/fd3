package sim

import (
	"container/heap"
	"fmt"
	"time"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/conditions"
	"github.com/warebuf/fd3/engine/sim/effects"
	"github.com/warebuf/fd3/engine/state"
)

// not sure where this belongs, maybe not sim package
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
		ApplyEffect(gs, pq, logList, UpdateID, currSimItem.Effects[i], currSimItem.CardIndex, -1, executeAtWallTime)
	}

	// TODO: should change the rest of this function to a clean switch statement
	if currSimItem.CardIndex >= 0 {
		if currSimItem.Type == state.SimItemOnTick {
			if gs.Cards[currSimItem.CardIndex].Zone == state.ZoneField {
				for _, t := range gs.Cards[currSimItem.CardIndex].OnTick {
					if t.AbilityID == currSimItem.AbilityID {
						ScheduleOnTickAbility(gs, pq, abilityIDCounter, logList, UpdateID, currSimItem.CardIndex, t.AbilityIndex, currSimItem.PopAtSimTime+constants.LEAD, executeAtWallTime)
						break
					}
				}
			}
		}

		clearTickAbility(gs.Cards, logList, UpdateID, currSimItem.CardIndex, currSimItem.AbilityID, executeAtWallTime)
	}

	if currSimItem.Type == state.SimItemStructureRegen {
		teamIndex := currSimItem.Effects[0].A
		ScheduleStructureManaRegen(gs, pq, logList, UpdateID, teamIndex, currSimItem.PopAtSimTime+constants.LEAD)
	}
}

// NOTE: i see this requiring work, why should we append objects rather than modifying them
// we require management of ontick abilities on the gameState because its the best way to display in the UI
func clearTickAbility(cards []state.Card, logList *[]any, UpdateID *int, cardIndex int, abilityID int, wallTime int64) {
	onTicks := cards[cardIndex].OnTick
	for i, t := range onTicks {
		if t.AbilityID == abilityID {
			cards[cardIndex].OnTick = append(onTicks[:i], onTicks[i+1:]...)
			*logList = append(*logList, logs.ClearTickAbilityLog{
				UpdateID:      *UpdateID,
				WallTimestamp: wallTime,
				Type:          logs.ClearTickAbility,
				CardIndex:     cardIndex,
				AbilityID:     abilityID,
			})
			*UpdateID++
			return
		}
	}
}

// NOTE: you could probably split this out of this file since it is quite important across the project
func ApplyEffect(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, eff state.Effect, sourceCardIndex int, activeIndex int, executeAtWallTime int64) {
	switch eff.Type {

	case state.EffectDraw:
		effects.ApplyEffectDraw(gs, logList, UpdateID, eff, executeAtWallTime)

	case state.EffectDestroy:
		destroyedIndexes := map[int]bool{}
		for a := 0; a < len(gs.Field); a++ {
			for b := 0; b < len(gs.Field[a]); b++ {
				cardIndex := gs.Field[a][b]
				gs.Cards[cardIndex].Zone = state.ZoneGraveyard
				gs.Grave[a] = append(gs.Grave[a], cardIndex)
				destroyedIndexes[cardIndex] = true
			}
			gs.Field[a] = []int{}
		}
		pq.RemoveSimItemsByCardIndexes(destroyedIndexes)
		for cardIndexes := range destroyedIndexes {
			gs.Cards[cardIndexes].OnTick = []state.TickAbility{}
		}
		// clear any preselected targets pointing at destroyed cards
		for i := range gs.Cards {
			for j := range gs.Cards[i].Actives {
				if destroyedIndexes[gs.Cards[i].Actives[j].Target] {
					gs.Cards[i].Actives[j].Target = -1
				}
			}
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

		cardIndexes, activeIndexes := conditions.AutoAbilityManaCheck(gs, eff.A)
		for i := 0; i < len(cardIndexes); i++ {
			pushAbilityOnStack(gs, logList, UpdateID, cardIndexes[i], activeIndexes[i], executeAtWallTime)
		}

	case state.EffectAttack:
		effects.ApplyEffectAttack(gs, pq, logList, UpdateID, eff, sourceCardIndex, activeIndex, executeAtWallTime)

	case state.EffectHeal:
		effects.ApplyEffectHeal(gs, logList, UpdateID, eff, sourceCardIndex, executeAtWallTime)

	case state.EffectCounter:
		effects.ApplyEffectCounter(gs, logList, UpdateID, sourceCardIndex, executeAtWallTime)

	case state.EffectGainCharge:
		effects.ApplyEffectGainCharge(gs, logList, UpdateID, eff, sourceCardIndex, executeAtWallTime)

		if conditions.AutoAbilityChargeCheck(gs, sourceCardIndex, eff.A) {
			pushAbilityOnStack(gs, logList, UpdateID, sourceCardIndex, eff.A, executeAtWallTime)
		}

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
	case state.EffectDraw:
		templ.A = teamIndex
		return templ
	case state.EffectAttack:
		return templ
	case state.EffectHeal:
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
		CardIndex:        -1,
		Type:             state.SimItemStructureRegen,
		Effects:          []state.Effect{{Type: state.EffectMana, A: teamIndex, B: structDef.ManaRegen}},
	})

}

func ScheduleAllOnTicks(gs *state.GameState, pq *state.SimulationPQ, abilityIDCounter *int, logList *[]any, UpdateID *int, cardIndex int, startAtSimTime int64, wallTimestamp int64) {
	onTick := defs.CardDefs[gs.Cards[cardIndex].DefNo].OnTick
	for a := 0; a < len(onTick); a++ {
		ScheduleOnTickAbility(gs, pq, abilityIDCounter, logList, UpdateID, cardIndex, a, startAtSimTime, wallTimestamp)
	}
}

func ScheduleOnTickAbility(gs *state.GameState, pq *state.SimulationPQ, abilityIDCounter *int, logList *[]any, UpdateID *int, cardIndex int, abilityIndex int, startAtSimTime int64, wallTimestamp int64) {
	card := gs.Cards[cardIndex]
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

	abilityID := *abilityIDCounter
	*abilityIDCounter++

	scheduledTime := startAtSimTime + entry.Milliseconds - constants.LEAD
	item := &state.SimulationItem{
		StartAtSimTime:   startAtSimTime,
		PopAtSimTime:     scheduledTime,
		ExecuteAtSimTime: scheduledTime + constants.LEAD,

		AbilityID: abilityID,
		CardIndex: cardIndex,
		Type:      state.SimItemOnTick,
		Effects:   effs,
	}
	heap.Push(pq, item)

	gs.Cards[cardIndex].OnTick = append(gs.Cards[cardIndex].OnTick, state.TickAbility{
		AbilityID:        abilityID,
		AbilityIndex:     abilityIndex,
		Condition:        condition,
		ActiveIndex:      activeIndex,
		StartAtSimTime:   item.StartAtSimTime,
		ExecuteAtSimTime: item.ExecuteAtSimTime,
	})

	*logList = append(*logList, logs.AddTickAbilityLog{
		UpdateID:         *UpdateID,
		WallTimestamp:    wallTimestamp,
		Type:             logs.AddTickAbility,
		CardIndex:        cardIndex,
		AbilityID:        abilityID,
		AbilityIndex:     abilityIndex,
		Condition:        condition,
		ActiveIndex:      activeIndex,
		StartAtSimTime:   item.StartAtSimTime,
		ExecuteAtSimTime: item.ExecuteAtSimTime,
	})
	*UpdateID++
}
