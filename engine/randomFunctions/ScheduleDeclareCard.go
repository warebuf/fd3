package randomFunctions

import (
	"container/heap"

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

	// check if we are in an indefinite pause
	curPausePhase := gs.GetPauseMode()
	if curPausePhase == 0 {
		return
	}

	// check if we are in chain phase or sim phase
	if curPausePhase == 1 { // sim pause, validate whether the card can be played in the current chain

		// TODO:
		// reject any sim phase cards
		// add the card to the chain

	} else if curPausePhase == 2 { // simulating

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

		// can either start a chain or summon a card (for now)
		if gs.Cards[cardIndex].Type == state.CardTypeUnit { // summon a card
			// updating the card's state
			gs.Cards[cardIndex].Zone = state.ZoneField

			// remove card from hand
			gs.Hand[teamIndex] = append(gs.Hand[teamIndex][:handIndex], gs.Hand[teamIndex][handIndex+1:]...)

			// summon
			handleSummonDeclaration(gs, logList, UpdateID, teamIndex, cardIndex, popTime)

		} else if gs.Cards[cardIndex].Type == state.CardTypeSpell { // start a chain

			// change it to change state
			gs.Cards[cardIndex].Zone = state.ZoneChain

			// remove card from hand
			gs.Hand[teamIndex] = append(gs.Hand[teamIndex][:handIndex], gs.Hand[teamIndex][handIndex+1:]...)

			// add it to the chain
			startChain(gs, logList, UpdateID, teamIndex, cardIndex, popTime+constants.LEAD)
		}

	}

}

func handleSummonDeclaration(gs *state.GameState, logList *[]any, UpdateID *int, teamIndex int, cardIndex int, popTime int64) {

	// create summon resolution effect for a unit
	eff := state.Effect{
		Type: state.SummonResolution,
		A:    teamIndex,
		B:    cardIndex, // index of card played
	}

	// create the item
	simItem := &state.SimulationItem{
		StartTime: popTime + constants.LEAD,
		PopTime:   popTime + constants.LEAD + constants.REACTION_WINDOW - constants.LEAD,
		OccurTime: popTime + constants.LEAD + constants.REACTION_WINDOW,

		CardID:  cardIndex,
		Type:    state.SimItemSummon,
		Effects: []state.Effect{eff},
	}
	heap.Push(&gs.SimulationPQ, simItem)

	// create the log
	*logList = append(*logList, logs.ScheduleChainLog{
		UpdateID:  *UpdateID,
		Timestamp: popTime + constants.LEAD,
		Type:      logs.ScheduleChain,
		ChainID:   s.ChainID,
		Step:      se,
	})
	*UpdateID += 1

}

func startChain(gs *state.GameState, logList *[]any, UpdateID *int, teamIndex int, cardIndex int, declarationTime int64) {
	card := gs.Cards[cardIndex]

	// TODO: pause the sim phase

	abilities := defs.CardDefs[card.DefNo].OnPlay
	for i := 0; i < len(abilities); i++ {

		// create the template
		templ := abilities[i]

		// fill out the template
		newEff := FulfillEffectTemplate(templ, teamIndex, cardIndex)

		// push the chain item
		chainItem := &state.ChainItem{
			StartTime: declarationTime,
			PopTime:   declarationTime + constants.REACTION_WINDOW - constants.LEAD,
			OccurTime: declarationTime + constants.REACTION_WINDOW,

			CardID:  cardIndex,
			Type:    state.SimItemSummon,
			Effects: []state.Effect{newEff},
		}
		gs.Chain = append(gs.Chain, *chainItem)

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
