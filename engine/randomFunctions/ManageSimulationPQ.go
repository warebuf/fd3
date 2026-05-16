package randomFunctions

import (
	"container/heap"
	"fmt"
	"math/rand"
	"time"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/defs/effects"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

// returns the number of new logs created
func ManageSimulationPQ(gs *state.GameState, logList *[]any, UpdateID *int) {

	for {
		// get the earliest simulation item
		earliestSimItem := gs.SimulationPQ.Peek()
		if earliestSimItem == nil {
			return
		}

		// calculate sim time
		currentSimTime := GetSimTime(gs, time.Now().UnixMilli())

		// if earliestStack is not due
		if earliestSimItem.PopTime > currentSimTime {
			return
		}

		// handle the top effect
		handleLatestSimulationItem(gs, logList, UpdateID)
	}

}

func handleLatestSimulationItem(gs *state.GameState, logList *[]any, UpdateID *int) {

	currSimItem := gs.SimulationPQ.Pop().(*state.SimulationItem)

	// handling all effects of the SimItem
	wallTime := currSimItem.PopTime + gs.SimPauseOffset + constants.LEAD
	for i := 0; i < len(currSimItem.Effects); i++ {

		currentEffect := currSimItem.Effects[i]

		if currentEffect.Type == state.SummonResolution {
			*logList = append(*logList, logs.SummonResolutionLog{
				UpdateID:  *UpdateID,
				Timestamp: wallTime,
				Type:      logs.SummonResolution,
				TeamIndex: currentEffect.A,
				CardIndex: currentEffect.B,
			})
			*UpdateID += 1
			applyOnSummonEffects(gs, logList, UpdateID, currentEffect.A, currentEffect.B, currSimItem.PopTime)
			scheduleRecurringAbilities(gs, logList, UpdateID, currSimItem.CardID, currSimItem.PopTime+constants.LEAD)
			continue
		}

		ApplyEffect(gs, logList, UpdateID, currentEffect, currSimItem.CardID, wallTime)

	}

	// finishing the handling of the SimItem
	if currSimItem.CardID >= 0 {
		if currSimItem.Type == state.SimItemSummon {
			teamIndex := gs.Cards[currSimItem.CardID].Team
			gs.Cards[currSimItem.CardID].Zone = state.ZoneField                   // complete the summon
			gs.Field[teamIndex] = append(gs.Field[teamIndex], currSimItem.CardID) // add card to the field
		} else if currSimItem.Type == state.SimItemRecurring {
			if gs.Cards[currSimItem.CardID].Zone == state.ZoneField {
				scheduleRecurringAbilities(gs, logList, UpdateID, currSimItem.CardID, currSimItem.PopTime+constants.LEAD)
			}
		}
	}

	// add a log
	*logList = append(*logList, logs.RemoveSimItemLog{
		UpdateID:  *UpdateID,
		Timestamp: currSimItem.PopTime + constants.LEAD,
		Type:      logs.RemoveChain,
		CardID:    currSimItem.CardID,
	})
	*UpdateID += 1

}

// ApplyEffect applies a single effect to the game state and appends logs.
// wallTime is the wall-clock ms timestamp to use for all logs emitted by this effect.
// sourceCardID is the card that triggered the effect (used for attack targeting).
func ApplyEffect(gs *state.GameState, logList *[]any, UpdateID *int, eff state.Effect, sourceCardID int, wallTime int64) {
	switch eff.Type {

	case state.EffectDraw:
		effects.ApplyEffectDraw(gs, eff)
		*logList = append(*logList, logs.DrawLog{
			UpdateID:  *UpdateID,
			Timestamp: wallTime,
			Type:      logs.Draw,
			TeamIndex: eff.A,
			NumCards:  1,
		})
		*UpdateID += 1

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
		removedCardIDs := gs.SimulationPQ.RemoveSimItemsByCardIDs(destroyedIDs)
		for _, cardID := range removedCardIDs {
			*logList = append(*logList, logs.RemoveSimItemLog{
				UpdateID:  *UpdateID,
				Timestamp: wallTime,
				Type:      logs.RemoveChain,
				CardID:    cardID,
			})
			*UpdateID += 1
		}
		*logList = append(*logList, logs.DestroyCardLog{
			UpdateID:        *UpdateID,
			Timestamp:       wallTime,
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
			UpdateID:  *UpdateID,
			Timestamp: wallTime,
			Type:      logs.GainMana,
			TeamIndex: eff.A,
			Amount:    eff.B,
		})
		*UpdateID += 1

	case state.EffectAttack:
		unitIndex := sourceCardID
		teamIndex := gs.Cards[unitIndex].Team
		enemyTeam := 1 - teamIndex
		damage := eff.B

		if eff.A == 0 { // random target
			enemyField := gs.Field[enemyTeam]
			if len(enemyField) > 0 {
				targetIdx := enemyField[rand.Intn(len(enemyField))]
				gs.Cards[targetIdx].HP -= damage
				*logList = append(*logList, logs.AttackLog{
					UpdateID:        *UpdateID,
					Timestamp:       wallTime,
					Type:            logs.Attack,
					AttackerIndex:   unitIndex,
					TargetIsUnit:    true,
					TargetTeamIndex: enemyTeam,
					TargetCardIndex: targetIdx,
					Damage:          damage,
				})
				*UpdateID += 1

				if gs.Cards[targetIdx].HP <= 0 {
					for i, ci := range gs.Field[enemyTeam] {
						if ci == targetIdx {
							gs.Field[enemyTeam] = append(gs.Field[enemyTeam][:i], gs.Field[enemyTeam][i+1:]...)
							break
						}
					}
					gs.Cards[targetIdx].Zone = state.ZoneGraveyard
					gs.Grave[enemyTeam] = append(gs.Grave[enemyTeam], targetIdx)

					removedCardIDs := gs.SimulationPQ.RemoveSimItemByCardID(targetIdx)
					for _, cardID := range removedCardIDs {
						*logList = append(*logList, logs.RemoveSimItemLog{
							UpdateID:  *UpdateID,
							Timestamp: wallTime,
							Type:      logs.RemoveChain,
							CardID:    cardID,
						})
						*UpdateID += 1
					}
					*logList = append(*logList, logs.DestroyCardLog{
						UpdateID:        *UpdateID,
						Timestamp:       wallTime,
						Type:            logs.DestroyCard,
						TeamIndex:       teamIndex,
						CardIndex:       targetIdx,
						TargetAll:       false,
						TargetTeamIndex: enemyTeam,
						TargetCardIndex: targetIdx,
					})
					*UpdateID += 1
				}
			} else {
				gs.Structures[enemyTeam].HP -= damage
				*logList = append(*logList, logs.AttackLog{
					UpdateID:        *UpdateID,
					Timestamp:       wallTime,
					Type:            logs.Attack,
					AttackerIndex:   unitIndex,
					TargetIsUnit:    false,
					TargetTeamIndex: enemyTeam,
					TargetCardIndex: -1,
					Damage:          damage,
				})
				*UpdateID += 1
			}
		}

	default:
		fmt.Println("unhandled effect type:", eff.Type)
	}
}

func applyOnSummonEffects(gs *state.GameState, logList *[]any, UpdateID *int, teamIndex int, cardIndex int, currentTime int64) {

	card := gs.Cards[cardIndex]

	abilities := defs.CardDefs[card.DefNo].OnPlay
	for i := 0; i < len(abilities); i++ {

		// create the template
		templ := abilities[i]

		// fill out the template
		newEff := FulfillEffectTemplate(templ, teamIndex, cardIndex)

		simItem := &state.SimulationItem{
			StartTime: currentTime + constants.LEAD,
			PopTime:   currentTime,
			OccurTime: currentTime + constants.LEAD,

			CardID:  cardIndex,
			Type:    state.SimItemSummon,
			Effects: []state.Effect{newEff},
		}
		heap.Push(&gs.SimulationPQ, simItem)

		*logList = append(*logList, logs.AddSimItemLog{
			UpdateID:  *UpdateID,
			Timestamp: currentTime + constants.LEAD,
			Type:      logs.AddSimItem,

			SeqID:       simItem.SeqID,
			StartTime:   simItem.StartTime,
			OccurTime:   simItem.OccurTime,
			CardID:      simItem.CardID,
			SimItemType: int(state.SimItemSummon),
		})
		*UpdateID += 1
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

func scheduleRecurringAbilities(gs *state.GameState, logList *[]any, UpdateID *int, cardID int, startTime int64) {
	teamIndex := gs.Cards[cardID].Team
	recurring := defs.CardDefs[gs.Cards[cardID].DefNo].Recurring

	for a := 0; a < len(recurring); a++ {
		entry := recurring[a]

		// create effects for this entry
		effs := []state.Effect{}
		for b := 0; b < len(entry.Effects); b++ {
			templ := entry.Effects[b]                                 // create the template effect
			newEff := FulfillEffectTemplate(templ, teamIndex, cardID) // completing the template
			effs = append(effs, newEff)
		}

		// create the SimulationItem and add it to the SimulationPQ
		scheduledTime := startTime + entry.Milliseconds - constants.LEAD
		item := &state.SimulationItem{
			StartTime: startTime,
			PopTime:   scheduledTime,
			OccurTime: scheduledTime + constants.LEAD,

			CardID:  cardID,
			Type:    state.SimItemRecurring,
			Effects: effs,
		}
		heap.Push(&gs.SimulationPQ, item)

		// broadcast the new step to all clients
		*logList = append(*logList, logs.AddSimItemLog{
			UpdateID:  *UpdateID,
			Timestamp: startTime,
			Type:      logs.AddSimItem,

			SeqID:       item.SeqID,
			StartTime:   item.StartTime,
			OccurTime:   item.OccurTime,
			CardID:      item.CardID,
			SimItemType: int(state.SimItemRecurring),
		})
		*UpdateID += 1
	}
}
