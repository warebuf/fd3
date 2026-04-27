package randomFunctions

import (
	"container/heap"

	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

// returns numer of effects generated
func ScheduleDeclareCard(gs *state.GameState, effectIndex uint64, eventLogs *[]logs.Log, logIndex int, teamIndex int, cardIndex int, popTime int64) int {

	// check if card is in hand state
	if gs.Cards[cardIndex].Zone != 1 {
		return 0
	}

	// check if card is in hand arr
	handIndex := -1
	for i := 0; i < len(gs.Hand[teamIndex]); i++ {
		if cardIndex == gs.Hand[teamIndex][i] {
			handIndex = i
		}
	}
	if handIndex < 0 {
		return 0
	}

	// TODO: check any of the card's activation conditions

	// create a log for the activation/declaration
	*eventLogs = append(*eventLogs, logs.Log{
		UpdateID:  logIndex,
		Timestamp: popTime + 200, // clients are allowed to receive logs 200ms early but see them in 200ms
		Type:      logs.DeclareCard,
		A:         teamIndex, // A: team index
		B:         cardIndex, // B: card index
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
		return fulfillEffectTemplate(gs, pq, teamIndex, cardIndex, effectIndex, popTime)

	} else if gs.Cards[cardIndex].Type == state.CardTypeSpell {

		// make the card temporarily unaccessible while resolving
		gs.Cards[cardIndex].Zone = state.ZoneTemporaryUnaccessible

		// remove card from hand
		gs.Hand[teamIndex] = append(gs.Hand[teamIndex][:handIndex], gs.Hand[teamIndex][handIndex+1:]...)

		// TODO: schedule the on-play effect depending on the card, fulfill the template

		// schedule the resolution
		return fulfillEffectTemplate(gs, pq, teamIndex, cardIndex, effectIndex, popTime)
	}

	return 0

}

func fulfillEffectTemplate(gs *state.GameState, teamIndex int, cardIndex int, effectIndex uint64, popTime int64) int {

	stacksCreated := 0
	card := gs.Cards[cardIndex]

	// create summon resolution effect if it's a unit
	if card.Type == state.CardTypeUnit {
		s := &stackPQ.Stack{
			EffectIndex: effectIndex,
			PopTime:     popTime + 200 + 4000 - 200,
			Effects: []state.Effect{
				{
					Type: state.SummonResolution,
					A:    teamIndex,
					B:    cardIndex, // index of card played
				},
			},
		}
		heap.Push(pq, s)
		effectIndex += 1
		stacksCreated += 1
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

			s := &stackPQ.Stack{
				EffectIndex: effectIndex,
				PopTime:     popTime + 200 + 4000 - 200,
				Effects: []state.Effect{
					templ,
				},
			}
			heap.Push(pq, s)
			effectIndex += 1
			stacksCreated += 1
		}

	}

	return stacksCreated

}
