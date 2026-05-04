package randomFunctions

import (
	"container/heap"
	"log"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

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
		Timestamp: popTime + constants.LEAD,
		Type:      logs.DeclareCard,
		TeamIndex: teamIndex,
		CardIndex: cardIndex,
	})

	// schedule the effect
	if gs.Cards[cardIndex].Type == state.CardTypeUnit {
		// updating the card's state
		gs.Cards[cardIndex].Zone = state.ZoneTemporaryUnaccessible

		// remove card from hand
		gs.Hand[teamIndex] = append(gs.Hand[teamIndex][:handIndex], gs.Hand[teamIndex][handIndex+1:]...)

		// schedule the resolution
		scheduleDeclareCardEffects(gs, logList, UpdateID, teamIndex, cardIndex, popTime)

	} else if gs.Cards[cardIndex].Type == state.CardTypeSpell {

		// make the card temporarily unaccessible while resolving
		gs.Cards[cardIndex].Zone = state.ZoneTemporaryUnaccessible

		// remove card from hand
		gs.Hand[teamIndex] = append(gs.Hand[teamIndex][:handIndex], gs.Hand[teamIndex][handIndex+1:]...)

		// schedule the resolution
		scheduleDeclareCardEffects(gs, logList, UpdateID, teamIndex, cardIndex, popTime)
	}

}

func scheduleDeclareCardEffects(gs *state.GameState, logList *[]any, UpdateID *int, teamIndex int, cardIndex int, popTime int64) {

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
			PopTime: popTime + constants.LEAD + constants.REACTION_WINDOW - constants.LEAD,
			Effects: []state.Effect{eff},
		}
		s := &state.Stack{
			PopTime:      se.PopTime,
			StackEntries: []state.StackEntry{se},
		}
		heap.Push(&gs.ReactionPQ, s)
		*logList = append(*logList, logs.ScheduleStackLog{
			UpdateID:   *UpdateID,
			Timestamp:  popTime + constants.LEAD,
			Type:       logs.ScheduleStack,
			StackID:    s.StackID,
			StackEntry: se,
		})
		*UpdateID += 1
	}

	// create the entries for a card
	effects := defs.CardDefs[card.DefNo].OnPlay
	for i := 0; i < len(effects); i++ {

		// create the template
		templ := effects[i]

		// fill out the template
		newEff := FulfillEffectTemplate(templ, teamIndex, cardIndex)

		se := state.StackEntry{
			Type:    5, // spell
			CardID:  cardIndex,
			PopTime: popTime + constants.LEAD + constants.REACTION_WINDOW - constants.LEAD, // the effect is being declared 200 ms from now, and applied 4 s after that, band broadcasted 200 ms before that
			Effects: []state.Effect{newEff},
		}
		s := &state.Stack{
			PopTime:      se.PopTime,
			StackEntries: []state.StackEntry{se},
		}
		heap.Push(&gs.ReactionPQ, s)
		*logList = append(*logList, logs.ScheduleStackLog{
			UpdateID:   *UpdateID,
			Timestamp:  popTime + constants.LEAD, // when the effect occurs, process the log
			Type:       logs.ScheduleStack,
			StackID:    s.StackID,
			StackEntry: se,
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
	}

	log.Printf("unknown effect type: %v", templ.Type)
	return templ
}
