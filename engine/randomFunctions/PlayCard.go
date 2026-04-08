package randomFunctions

import (
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/state"
)

func PlayCard(gs *state.GameState, card *state.Card, pid int) {
	switch card.Type {
	case state.CardTypeUnit, state.CardTypeStructure:
		createPermanent(gs, card, pid)
	case state.CardTypeSpell:
		// resolve effect, send to graveyard
	case state.CardTypeTrap:
		// set in zone, wait for trigger condition
	}
}

func createPermanent(gs *state.GameState, card *state.Card, pid int) {
	card.Zone = state.ZoneField
	def := defs.CardDefs[card.DefNo]
	p := state.Permanent{
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
