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

	chainHeader := gs.ReactionPQ.Peek()

	for len(chainHeader.ChainSteps) > 0 {
		topStep := chainHeader.ChainSteps[len(chainHeader.ChainSteps)-1]

		// handling all effects of a step
		for i := 0; i < len(topStep.Effects); i++ {

			currentEffect := topStep.Effects[i]

			// apply effect
			switch currentEffect.Type {

			case state.EffectDraw:
				effects.ApplyEffectDraw(gs, currentEffect)

				// create log
				*logList = append(*logList, logs.DrawLog{
					UpdateID:  *UpdateID,
					Timestamp: topStep.PopTime + constants.LEAD,
					Type:      logs.Draw,
					TeamIndex: currentEffect.A,
					NumCards:  1,
				})
				*UpdateID += 1

			case state.SummonResolution:

				// create a summon resolution log
				*logList = append(*logList, logs.SummonLog{
					UpdateID:  *UpdateID,
					Timestamp: topStep.PopTime + constants.LEAD,
					Type:      logs.Summon,
					TeamIndex: currentEffect.A,
					CardIndex: currentEffect.B,
				})
				*UpdateID += 1

				// schedule recurring abilities
				scheduleRecurringAbilities(gs, logList, UpdateID, topStep.CardID, topStep.PopTime+constants.LEAD)

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
				*logList = append(*logList, logs.DestroyCardLog{
					UpdateID:        *UpdateID,
					Timestamp:       topStep.PopTime + constants.LEAD,
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
				*logList = append(*logList, logs.GainManaLog{
					UpdateID:  *UpdateID,
					Timestamp: topStep.PopTime + constants.LEAD,
					Type:      logs.GainMana,
					TeamIndex: team,
					Amount:    amount,
				})
				*UpdateID += 1

			default:
				fmt.Println("this effect isnt handled")
			}

		}

		// finishing the handling of a step
		if topStep.CardID >= 0 {
			if topStep.Type == state.ChainStepSummon {
				teamIndex := gs.Cards[topStep.CardID].Team
				gs.Cards[topStep.CardID].Zone = state.ZoneField                   // complete the summon
				gs.Field[teamIndex] = append(gs.Field[teamIndex], topStep.CardID) // add card to the field
			} else if topStep.Type == state.ChainStepSpell {
				gs.Cards[topStep.CardID].Zone = state.ZoneGraveyard
			} else if topStep.Type == state.ChainStepRecurring {
				if gs.Cards[topStep.CardID].Zone == state.ZoneField {

					// schedule recurring abilities
					scheduleRecurringAbilities(gs, logList, UpdateID, topStep.CardID, topStep.PopTime+constants.LEAD)

				}
			}
		}

		// removing the top step from the chain
		chainHeader.ChainSteps = chainHeader.ChainSteps[:len(chainHeader.ChainSteps)-1]

	}

	heap.Pop(&gs.ReactionPQ)
	*logList = append(*logList, logs.RemoveChainLog{
		UpdateID:  *UpdateID,
		Timestamp: chainHeader.PopTime + constants.LEAD,
		Type:      logs.RemoveChain,
		ChainID:   chainHeader.ChainID,
	})
	*UpdateID += 1

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

		// create the step/chain
		scheduledTime := startTime + entry.Milliseconds - constants.LEAD
		se := state.ChainStep{
			Type:            state.ChainStepRecurring,
			CardID:          cardID,
			StartTime:       startTime,
			PopTime:         scheduledTime,
			OccurTime:       scheduledTime + constants.LEAD,
			Effects:         effs,
			RecurringPeriod: entry.Milliseconds,
		}
		s := &state.Chain{
			PopTime:    scheduledTime,
			ChainSteps: []state.ChainStep{se},
		}
		heap.Push(&gs.ReactionPQ, s)

		// broadcast the new step to all clients
		*logList = append(*logList, logs.ScheduleChainLog{
			UpdateID:  *UpdateID,
			Timestamp: startTime, // when the effect occurs, process the log
			Type:      logs.ScheduleChain,
			ChainID:   s.ChainID,
			Step:      se,
		})
		*UpdateID += 1
	}
}
