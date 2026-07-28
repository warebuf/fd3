package schedule

import (
	"container/heap"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func ScheduleOnTickAbility(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, cardIndex int, abilityIndex int, startAtSimTime int64, wallTimestamp int64) {
	card := gs.Cards[cardIndex]
	if card.Zone == state.ZoneGraveyard {
		return // card is dead, nothing to schedule
	}

	entry := defs.CardDefs[card.DefNo].OnTick[abilityIndex]

	if entry.Milliseconds == 0 {
		return // unused OnTick slot
	}

	// a gain-charge ability stops rescheduling once the card's Charges maxes out
	condition := state.TickIndefinite
	for _, eff := range entry.Effects {
		if eff.Type == state.EffectChargeSelf {
			condition = state.TickUntilChargeMax
			if card.MaxCharges > 0 && card.Charges >= card.MaxCharges {
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
		effs = append(effs, state.FulfillEffectTemplate(gs, e, teamIndex, cardIndex, -1))
	}

	scheduledTime := startAtSimTime + entry.Milliseconds/constants.SIM_SPEED - constants.LEAD
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
		StartAtSimTime:   item.StartAtSimTime,
		ExecuteAtSimTime: item.ExecuteAtSimTime,
	})
	*UpdateID++
}
