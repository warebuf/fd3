package microevents

import (
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/states/gameState"
)

func NewCard(cid int, defNo int, team int) gameState.Card {
	def := defs.CardDefs[defNo] // look up the blueprint
	return gameState.Card{
		CardID: cid,
		DefNo:  defNo,
		Team:   team,
		Type:   def.Type,
		Zone:   gameState.ZoneDeck,
		Armed:  false,
	}
}

func NewDeck(team int, defNos []int) []gameState.Card {
	deck := []gameState.Card{}
	for i, defNo := range defNos {
		deck = append(deck, NewCard(i, defNo, team))
	}
	return deck
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
