package simulate

import (
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/states/gameState"
)

func PlayCard(gs *gameState.GameState, card *gameState.Card, pid int) {
	switch card.Type {
	case gameState.CardTypeUnit, gameState.CardTypeStructure:
		createPermanent(gs, card, pid)
	case gameState.CardTypeSpell:
		// resolve effect, send to graveyard
	case gameState.CardTypeTrap:
		// set in zone, wait for trigger condition
	}
}

func createPermanent(gs *gameState.GameState, card *gameState.Card, pid int) {
	card.Zone = gameState.ZoneField
	def := defs.CardDefs[card.DefNo]
	p := gameState.Permanent{
		PermanentID: pid,
		SourceCard:  card.CardID,
		Type:        card.Type,
		HP:          def.BaseHP,
		Team:        card.Team,
		Master:      false,
	}
	gs.Permanents = append(gs.Permanents, p)
	// register triggers, schedule abilities next
}
