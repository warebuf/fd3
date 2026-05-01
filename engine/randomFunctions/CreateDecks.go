package randomFunctions

import (
	"time"

	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/state"
)

func CreateDecks(deckDefs [][]int, structureDefNos []int) *state.GameState {
	structures := make([]state.Structure, len(deckDefs))
	for i, defNo := range structureDefNos {
		structures[i] = state.Structure{DefNo: defNo, HP: defs.StructureDefs[defNo].BaseHP}
	}

	gs := &state.GameState{
		UpdateID:  0,
		Timestamp: time.Now(),
		Phase:     0,

		Cards:      []state.Card{},
		Deck:       make([][]int, len(deckDefs)),
		Hand:       make([][]int, len(deckDefs)),
		Field:      make([][]int, len(deckDefs)),
		Grave:      make([][]int, len(deckDefs)),
		Mana:       make([]int, len(deckDefs)),
		Structures: structures,
	}
	for i := range len(deckDefs) {
		gs.Deck[i] = []int{}
		gs.Hand[i] = []int{}
		gs.Field[i] = []int{}
		gs.Grave[i] = []int{}
	}

	nextCardID := 0
	for team := 0; team < len(deckDefs); team++ {
		defNos := deckDefs[team]

		gs.Deck[team] = make([]int, 0, len(defNos))
		gs.Hand[team] = []int{}
		gs.Field[team] = []int{}
		gs.Grave[team] = []int{}

		for _, defNo := range defNos {
			def := defs.CardDefs[defNo]

			card := state.Card{
				//
				DefNo:       defNo,
				Name:        def.Name,
				Description: def.Description,

				//
				CardID: nextCardID,
				Team:   team,
				Type:   def.Type,
				Zone:   state.ZoneDeck,
				Armed:  false,
			}

			gs.Cards = append(gs.Cards, card)
			gs.Deck[team] = append(gs.Deck[team], nextCardID)

			nextCardID++
		}
	}

	return gs
}

func IsValidDeck(deck []int, structNo int) bool {

	if len(deck) > 30 {
		return false
	}

	cardDefsLength := len(defs.CardDefs)
	counts := make([]int, cardDefsLength)

	for i := 0; i < len(deck); i++ {
		cardNo := deck[i]

		// bounds check
		if cardNo <= 0 || cardNo >= cardDefsLength { // rejecting 0, negative, and non-existing numbers
			return false
		}

		// count occurrences
		counts[cardNo]++

		// duplicate rule (example: max 3 copies)
		if counts[cardNo] > 3 {
			return false
		}
	}

	// make sure structNo is available
	structDefsLength := len(defs.StructureDefs)
	if structNo <= 0 || structNo >= structDefsLength {
		return false
	}

	return true
}
