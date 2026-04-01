package microevents

import (
	"time"

	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/states/gameState"
)

func InitDecks(deckDefs [][]int) *gameState.GameState {
	gs := &gameState.GameState{
		GameStateID: 0,
		Timestamp:   time.Now(),
		Phase:       0,

		Cards: []gameState.Card{},
		Deck:  make([][]int, len(deckDefs)),
		Hand:  make([][]int, len(deckDefs)),
		Field: make([][]int, len(deckDefs)),
		Grave: make([][]int, len(deckDefs)),

		Permanents:     []gameState.Permanent{},
		Mana:           make([]int, len(deckDefs)),
		ActiveTriggers: []gameState.TriggerEffect{},
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

			card := gameState.Card{
				CardID: nextCardID,
				DefNo:  defNo,
				Team:   team,
				Type:   def.Type,
				Zone:   gameState.ZoneDeck,
				Armed:  false,
			}

			gs.Cards = append(gs.Cards, card)
			gs.Deck[team] = append(gs.Deck[team], nextCardID)

			nextCardID++
		}
	}

	return gs
}

func IsValidDeck(deck []int) bool {
	if len(deck) > 30 {
		return false
	}

	cardDefsLength := len(defs.CardDefs)
	counts := make([]int, cardDefsLength)

	for i := 0; i < len(deck); i++ {
		cardNo := deck[i]

		// bounds check
		if cardNo <= 0 || cardNo >= cardDefsLength {
			return false
		}

		// count occurrences
		counts[cardNo]++

		// duplicate rule (example: max 3 copies)
		if counts[cardNo] > 3 {
			return false
		}
	}

	return true
}
