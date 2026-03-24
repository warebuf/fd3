package cardDef

import "github.com/warebuf/fd3/engine/states/gameState"

func NewCard(cid int, defNo int, team int) gameState.Card {
	def := CardDefs[defNo] // look up the blueprint
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
