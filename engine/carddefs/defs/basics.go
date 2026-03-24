package defs

import "github.com/warebuf/fd3/engine/states"

type CardDef struct {
	Type     int
	OnPlay   []Effect
	Triggers []Trigger
}

type Trigger struct {
	Condition int
	Effect    Effect
}

var CardDefs = []CardDef{
	0: {}, // empty, so DefNo lines up with index
	1: {Type: states.CardTypeUnit, OnPlay: []Effect{}, Triggers: []Trigger{}},
	2: {Type: states.CardTypeUnit, OnPlay: []Effect{}, Triggers: []Trigger{}},
}

func NewCard(cid int, defNo int, team int) states.Card {
	def := CardDefs[defNo] // look up the blueprint
	return states.Card{
		CID:   cid,
		DefNo: defNo,
		Team:  team,
		Type:  def.Type,
		Zone:  states.ZoneDeck,
		Armed: false,
	}
}

func NewDeck(team int, defNos []int) []states.Card {
	deck := []states.Card{}
	for i, defNo := range defNos {
		deck = append(deck, NewCard(i, defNo, team))
	}
	return deck
}
