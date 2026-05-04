package randomFunctions

import (
	"container/heap"
	"fmt"
	"time"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/defs/effects"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

// returns the number of new logs created
func ManageStackPQ(gs *state.GameState, logList *[]any, UpdateID *int) {

	for {
		// get the earliest stack
		earliestStack := gs.ReactionPQ.Peek()
		if earliestStack == nil {
			return
		}

		// if earliestStack is not due
		if earliestStack.PopTime > time.Now().UnixMilli() {
			return
		}

		// handle the top effect
		SimulateStack(gs, logList, UpdateID)
	}

}

// handling it as a yugioh chain, could be a mtg stack though
func SimulateStack(gs *state.GameState, logList *[]any, UpdateID *int) {

	stackHeader := gs.ReactionPQ.Peek()

	for len(stackHeader.StackEntries) > 0 {
		topEntry := stackHeader.StackEntries[len(stackHeader.StackEntries)-1]

		// handling all effects of an entry
		for i := 0; i < len(topEntry.Effects); i++ {

			currentEffect := topEntry.Effects[i]

			// apply effect
			switch currentEffect.Type {

			case state.EffectDraw:
				effects.ApplyEffectDraw(gs, currentEffect)

				// create log
				*logList = append(*logList, logs.DrawExecEntry{
					UpdateID:  *UpdateID,
					Timestamp: topEntry.PopTime + constants.LEAD,
					Type:      logs.DrawExec,
					TeamIndex: currentEffect.A,
					NumCards:  1,
				})
				*UpdateID += 1

			case state.SummonResolution:

				// create a summon resolution log
				*logList = append(*logList, logs.SummonUnitEntry{
					UpdateID:  *UpdateID,
					Timestamp: topEntry.PopTime + constants.LEAD,
					Type:      logs.SummonUnit,
					TeamIndex: currentEffect.A,
					CardIndex: currentEffect.B,
				})
				*UpdateID += 1

				// schedule recurring abilities
				recurring := defs.CardDefs[gs.Cards[topEntry.CardID].DefNo].Recurring
				for a := 0; a < len(recurring); a++ {

					entry := recurring[a]

					// create effects for this entry
					effs := []state.Effect{}
					for b := 0; b < len(entry.Effects); b++ {
						templ := entry.Effects[b]                                                // create the template effect
						newEff := FulfillEffectTemplate(templ, currentEffect.A, currentEffect.B) // completing the template
						effs = append(effs, newEff)
					}

					// create the entry/stack
					scheduledTime := topEntry.PopTime + entry.Milliseconds + constants.LEAD
					se := state.StackEntry{
						Type:            6, // recurring ability
						CardID:          topEntry.CardID,
						PopTime:         scheduledTime,
						Effects:         effs,
						RecurringPeriod: entry.Milliseconds,
					}
					s := &state.Stack{
						PopTime:      scheduledTime,
						StackEntries: []state.StackEntry{se},
					}
					heap.Push(&gs.ReactionPQ, s)

					// broadcast the new entry to all clients
					*logList = append(*logList, logs.ScheduleStackLog{
						UpdateID:   *UpdateID,
						Timestamp:  topEntry.PopTime + constants.LEAD, // when the effect occurs, process the log
						Type:       logs.ScheduleStack,
						StackID:    s.StackID,
						StackEntry: se,
					})
					*UpdateID += 1

				}

			case state.EffectDestroy:

				// put all field units into the grave, may have to handle the ordering differently tbh (in order of summon or last move/action)
				for a := 0; a < len(gs.Field); a++ {
					for b := 0; b < len(gs.Field[a]); b++ {
						cardIndex := gs.Field[a][b]
						gs.Cards[cardIndex].Zone = state.ZoneGraveyard
						gs.Grave[a] = append(gs.Grave[a], cardIndex)
					}
					gs.Field[a] = []int{}
				}

				// create log
				*logList = append(*logList, logs.DestroyCardEntry{
					UpdateID:        *UpdateID,
					Timestamp:       topEntry.PopTime + constants.LEAD,
					Type:            logs.DestroyCard,
					TeamIndex:       gs.Cards[currentEffect.C].Team,
					CardIndex:       currentEffect.C,
					TargetAll:       currentEffect.A == 0,
					TargetTeamIndex: -1,
					TargetCardIndex: -1,
				})
				*UpdateID += 1

			case state.EffectMana:
				// increase mana for team by the amount
				team := currentEffect.A
				amount := currentEffect.B
				gs.Mana[team] += amount

				// create log
				*logList = append(*logList, logs.GainManaEntry{
					UpdateID:  *UpdateID,
					Timestamp: topEntry.PopTime + constants.LEAD,
					Type:      logs.GainMana,
					TeamIndex: team,
					Amount:    amount,
				})
				*UpdateID += 1

			default:
				fmt.Println("this effect isnt handled")
			}

		}

		// finishing the handling of an entry
		if topEntry.CardID >= 0 {
			if topEntry.Type == 4 { // summon
				teamIndex := gs.Cards[topEntry.CardID].Team
				gs.Cards[topEntry.CardID].Zone = state.ZoneField                   // complete the summon
				gs.Field[teamIndex] = append(gs.Field[teamIndex], topEntry.CardID) // add card to the field
			} else if topEntry.Type == 5 { // spell
				gs.Cards[topEntry.CardID].Zone = state.ZoneGraveyard
			} else if topEntry.Type == 6 { // recurring ability, schedule the next trigger
				if gs.Cards[topEntry.CardID].Zone == state.ZoneField {

					// schedule recurring abilities
					recurring := defs.CardDefs[gs.Cards[topEntry.CardID].DefNo].Recurring
					for a := 0; a < len(recurring); a++ {

						entry := recurring[a]

						// create effects for this entry
						effs := []state.Effect{}
						for b := 0; b < len(entry.Effects); b++ {
							templ := entry.Effects[b]                                                               // create the template effect
							newEff := FulfillEffectTemplate(templ, gs.Cards[topEntry.CardID].Team, topEntry.CardID) // completing the template
							effs = append(effs, newEff)
						}

						// create the entry/stack
						scheduledTime := topEntry.PopTime + entry.Milliseconds + constants.LEAD
						se := state.StackEntry{
							Type:            6, // recurring ability
							CardID:          topEntry.CardID,
							PopTime:         scheduledTime,
							Effects:         effs,
							RecurringPeriod: entry.Milliseconds,
						}
						s := &state.Stack{
							PopTime:      scheduledTime,
							StackEntries: []state.StackEntry{se},
						}
						heap.Push(&gs.ReactionPQ, s)

						// broadcast the new entry to all clients
						*logList = append(*logList, logs.ScheduleStackLog{
							UpdateID:   *UpdateID,
							Timestamp:  topEntry.PopTime + constants.LEAD, // when the effect occurs, process the log
							Type:       logs.ScheduleStack,
							StackID:    s.StackID,
							StackEntry: se,
						})
						*UpdateID += 1

					}

				}
			}
		}

		// removing the top entry from the stack
		stackHeader.StackEntries = stackHeader.StackEntries[:len(stackHeader.StackEntries)-1]

	}

	heap.Pop(&gs.ReactionPQ)
	*logList = append(*logList, logs.RemoveStackLog{
		UpdateID:  *UpdateID,
		Timestamp: stackHeader.PopTime + constants.LEAD,
		Type:      logs.RemoveStack,
		StackID:   stackHeader.StackID,
	})
	*UpdateID += 1

}
