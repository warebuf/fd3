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

	// check cost condition
	cost := gs.Cards[cardIndex].Cost
	if gs.Mana[teamIndex] < cost {
		return
	}

	// TODO: okay now we have to determine chain or sim stuff
	//
	// if it is chain stuff, run the pause
	//
	// if it is a unit, just summon
	//
	curPausePhase := gs.GetPauseMode()
	if curPausePhase == 0 { // currently in indefinite pause, cannot play a declaration
		return
	}

	if curPausePhase == 1 { // sim pause, validate whether the card can be played in the current chain

		// TODO:

	} else if curPausePhase == 2 { // simulating

		// can either start a chain or summon a card (for now)

		if gs.Cards[cardIndex].Type == state.CardTypeUnit { // summon a card
			// updating the card's state
			gs.Cards[cardIndex].Zone = state.ZoneField

			// remove card from hand
			gs.Hand[teamIndex] = append(gs.Hand[teamIndex][:handIndex], gs.Hand[teamIndex][handIndex+1:]...)

			// summon
			handleSummonDeclaration(gs, logList, UpdateID, teamIndex, cardIndex, popTime)

		} else if gs.Cards[cardIndex].Type == state.CardTypeSpell { // start a chain

			// remove the mana
			gs.Mana[teamIndex] -= cost
			*logList = append(*logList, logs.SpendManaLog{
				UpdateID:  *UpdateID,
				Timestamp: popTime,
				Type:      logs.SpendMana,
				TeamIndex: teamIndex,
				Amount:    cost,
			})
			*UpdateID += 1

			// create a log for the activation/declaration
			*logList = append(*logList, logs.DeclareCardLog{
				UpdateID:  *UpdateID,
				Timestamp: popTime + constants.LEAD,
				Type:      logs.DeclareCard,
				TeamIndex: teamIndex,
				CardIndex: cardIndex,
			})

			// remove card from hand
			gs.Hand[teamIndex] = append(gs.Hand[teamIndex][:handIndex], gs.Hand[teamIndex][handIndex+1:]...)

			// add it to grave
			gs.Cards[cardIndex].Zone = state.ZoneChain

			// add it to the chain
		}

	}

}

func handleSummonDeclaration(gs *state.GameState, logList *[]any, UpdateID *int, teamIndex int, cardIndex int, popTime int64) {

	card := gs.Cards[cardIndex]

	// create summon resolution effect if it's a unit
	eff := state.Effect{
		Type: state.SummonResolution,
		A:    teamIndex,
		B:    cardIndex, // index of card played
	}
	se := state.ChainStep{
		Type:       state.ChainStepSummon,
		CardID:     cardIndex,
		StartTime:  popTime + constants.LEAD,
		PopTime:    popTime + constants.LEAD + constants.REACTION_WINDOW - constants.LEAD,
		OccurTime:  popTime + constants.LEAD + constants.REACTION_WINDOW,
		Effects:    []state.Effect{eff},
		Persistent: false,
	}
	s := &state.Chain{
		PopTime:    se.PopTime,
		ChainSteps: []state.ChainStep{se},
	}
	heap.Push(&gs.SimulationPQ, s)
	*logList = append(*logList, logs.ScheduleChainLog{
		UpdateID:  *UpdateID,
		Timestamp: popTime + constants.LEAD,
		Type:      logs.ScheduleChain,
		ChainID:   s.ChainID,
		Step:      se,
	})
	*UpdateID += 1

}

func handleSummonResolution(gs *state.GameState, logList *[]any, UpdateID *int, teamIndex int, cardIndex int, popTime int64) {

	card := gs.Cards[cardIndex]

	abilities := defs.CardDefs[card.DefNo].OnPlay
	for i := 0; i < len(abilities); i++ {

		// create the template
		templ := abilities[i]

		// fill out the template
		newEff := FulfillEffectTemplate(templ, teamIndex, cardIndex)

		se := state.ChainStep{
			Type:       state.ChainStepSpell,
			CardID:     cardIndex,
			StartTime:  popTime + constants.LEAD,
			PopTime:    popTime + constants.LEAD + constants.REACTION_WINDOW - constants.LEAD,
			OccurTime:  popTime + constants.LEAD + constants.REACTION_WINDOW,
			Effects:    []state.Effect{newEff},
			Persistent: false,
		}
		s := &state.Chain{
			PopTime:    se.PopTime,
			ChainSteps: []state.ChainStep{se},
		}
		heap.Push(&gs.SimulationPQ, s)
		*logList = append(*logList, logs.ScheduleChainLog{
			UpdateID:  *UpdateID,
			Timestamp: popTime + constants.LEAD,
			Type:      logs.ScheduleChain,
			ChainID:   s.ChainID,
			Step:      se,
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

	log.Printf("unknown effect type: %v", templ.Type)
	return templ
}
