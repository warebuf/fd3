package randomFunctions

import (
	"container/heap"

	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

// returns numer of effects generated
func ScheduleDeclareCard(gs *state.GameState, logList *[]any, UpdateID *int, teamIndex int, cardIndex int, popTime int64) {

	// check if card is in hand state
	if gs.Cards[cardIndex].Zone != 1 {
		return
	}

	// check if card is in hand arr
	handIndex := -1
	for i := 0; i < len(gs.Hand[teamIndex]); i++ {
		if cardIndex == gs.Hand[teamIndex][i] {
			handIndex = i
		}
	}
	if handIndex < 0 {
		return
	}

	// TODO: check any of the card's activation conditions

	// create a log for the activation/declaration
	*logList = append(*logList, logs.DeclareCardEntry{
		UpdateID:  *UpdateID,
		Timestamp: popTime + 200,
		Type:      logs.DeclareCard,
		TeamIndex: teamIndex,
		CardIndex: cardIndex,
	})

	// schedule the effect
	if gs.Cards[cardIndex].Type == state.CardTypeUnit {
		// updating the card's state
		gs.Cards[cardIndex].Zone = state.ZoneField
		gs.Cards[cardIndex].Pending = true

		// remove card from hand
		gs.Hand[teamIndex] = append(gs.Hand[teamIndex][:handIndex], gs.Hand[teamIndex][handIndex+1:]...)

		// add card to the field
		gs.Field[teamIndex] = append(gs.Field[teamIndex], cardIndex)

		// schedule the resolution
		fulfillEffectTemplate(gs, logList, UpdateID, teamIndex, cardIndex, popTime)

	} else if gs.Cards[cardIndex].Type == state.CardTypeSpell {

		// make the card temporarily unaccessible while resolving
		gs.Cards[cardIndex].Zone = state.ZoneTemporaryUnaccessible

		// remove card from hand
		gs.Hand[teamIndex] = append(gs.Hand[teamIndex][:handIndex], gs.Hand[teamIndex][handIndex+1:]...)

		// schedule the resolution
		fulfillEffectTemplate(gs, logList, UpdateID, teamIndex, cardIndex, popTime)
	}

}

func fulfillEffectTemplate(gs *state.GameState, logList *[]any, UpdateID *int, teamIndex int, cardIndex int, popTime int64) {

	card := gs.Cards[cardIndex]

	// create summon resolution effect if it's a unit
	if card.Type == state.CardTypeUnit {
		eff := state.Effect{
			Type: state.SummonResolution,
			A:    teamIndex,
			B:    cardIndex, // index of card played
		}
		se := state.StackEntry{
			Type:    4, // summon
			CardID:  cardIndex,
			PopTime: popTime + 200 + 4000 - 200,
			Effects: []state.Effect{eff},
		}
		s := &state.Stack{
			PopTime:      popTime + 200 + 4000 - 200,
			StackEntries: []state.StackEntry{se},
		}
		heap.Push(&gs.ReactionPQ, s)
		*logList = append(*logList, logs.ScheduleStackLog{
			UpdateID:   *UpdateID,
			Timestamp:  s.PopTime,
			Type:       logs.ScheduleStack,
			StackID:    s.StackID,
			StackEntry: se,
		})
		*UpdateID += 1
	}

	// create any on play effects
	effects := defs.CardDefs[card.DefNo].OnPlay
	for i := 0; i < len(effects); i++ {

		// create the template
		templ := effects[i]

		// fill the rest of the template out
		switch templ.Type {
		case state.EffectDestroy:

			templ.C = cardIndex // card played

			se := state.StackEntry{
				Type:    1, // card eff
				CardID:  cardIndex,
				PopTime: popTime + 200 + 4000 - 200, // the effect is being declared 200 ms from now, and applied 4 s after that, band broadcasted 200 ms before that 
				Effects: []state.Effect{templ},
			}
			s := &state.Stack{
				PopTime:      se.PopTime,
				StackEntries: []state.StackEntry{se},
			}
			heap.Push(&gs.ReactionPQ, s)
			*logList = append(*logList, logs.ScheduleStackLog{
				UpdateID:   *UpdateID,
				Timestamp:  popTime + 200, // when the effect occurs, process the log
				Type:       logs.ScheduleStack,
				StackID:    s.StackID,
				StackEntry: se,
			})
			*UpdateID += 1
		}

	}

}
