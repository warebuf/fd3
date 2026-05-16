package randomFunctions

import (
	"container/heap"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

// Returns true if a chain was started (caller should enter sim pause).
func ScheduleDeclareCard(gs *state.GameState, logList *[]any, UpdateID *int, teamIndex int, cardIndex int, popTime int64) bool {

	// check if card is in hand state
	if gs.Cards[cardIndex].Zone != 1 {
		return false
	}

	// check if card is in hand arr
	handIndex := -1
	for i := 0; i < len(gs.Hand[teamIndex]); i++ {
		if cardIndex == gs.Hand[teamIndex][i] {
			handIndex = i
		}
	}
	if handIndex < 0 {
		return false
	}

	// check cost condition
	cost := gs.Cards[cardIndex].Cost
	if gs.Mana[teamIndex] < cost {
		return false
	}

	// check if we are in an indefinite pause
	if gs.PausePhase == 0 {
		return false
	}

	// check if we are in chain phase or sim phase
	if gs.PausePhase == 1 { // sim pause, validate whether the card can be played in the current chain

		// TODO:
		// reject any sim phase cards
		// add the card to the chain

	} else if gs.PausePhase == 2 { // simulating

		// remove the mana
		gs.Mana[teamIndex] -= cost
		*logList = append(*logList, logs.SpendManaLog{
			UpdateID:  *UpdateID,
			Timestamp: popTime + constants.LEAD,
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

			// change it to chain state
			gs.Cards[cardIndex].Zone = state.ZoneChain

			// remove card from hand
			gs.Hand[teamIndex] = append(gs.Hand[teamIndex][:handIndex], gs.Hand[teamIndex][handIndex+1:]...)

			// add it to the chain
			startChain(gs, logList, UpdateID, teamIndex, cardIndex, popTime+constants.LEAD)
			return true
		}

	}

	return false
}

func handleSummonDeclaration(gs *state.GameState, logList *[]any, UpdateID *int, teamIndex int, cardIndex int, popTime int64) {

	// create summon resolution effect for a unit
	eff := state.Effect{
		Type: state.SummonResolution,
		A:    teamIndex,
		B:    cardIndex, // index of card played
	}

	// convert to sim time
	simTime := popTime - gs.SimPauseOffset

	// create the item
	simItem := &state.SimulationItem{
		StartTime: simTime + constants.LEAD,
		PopTime:   simTime + constants.REACTION_WINDOW,
		OccurTime: simTime + constants.REACTION_WINDOW + constants.LEAD,

		CardID:  cardIndex,
		Type:    state.SimItemSummon,
		Effects: []state.Effect{eff},
	}
	heap.Push(&gs.SimulationPQ, simItem)

	// create the log
	*logList = append(*logList, logs.AddSimItemLog{
		UpdateID:  *UpdateID,
		Timestamp: popTime + constants.LEAD,
		Type:      logs.AddSimItem,

		SeqID:       simItem.SeqID,
		StartTime:   simItem.StartTime,
		OccurTime:   simItem.OccurTime,
		CardID:      simItem.CardID,
		SimItemType: int(state.SimItemSummon),
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

		// convert to chain time (wall time minus accumulated indefinite-pause offset)
		chainTime := declarationTime - gs.FullPauseOffset

		// push the chain item
		chainItem := &state.ChainItem{
			StartTime: chainTime,
			PopTime:   chainTime + constants.REACTION_WINDOW - constants.LEAD,
			OccurTime: chainTime + constants.REACTION_WINDOW,

			CardID:  cardIndex,
			Type:    state.ChainItemSpell,
			Effects: []state.Effect{newEff},
		}
		gs.Chain = append(gs.Chain, *chainItem)

		*logList = append(*logList, logs.AddChainLog{
			UpdateID:  *UpdateID,
			Timestamp: declarationTime,
			Type:      logs.AddChain,
			StartTime: chainItem.StartTime,
			OccurTime: chainItem.OccurTime,
			CardID:    chainItem.CardID,
		})
		*UpdateID += 1
	}
}
