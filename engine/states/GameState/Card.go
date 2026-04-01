package gameState

const (
	CardTypeStructure = iota
	CardTypeUnit
	CardTypeTrap
	CardTypeSpell
)

const (
	ZoneDeck = iota
	ZoneHand
	ZoneField
	ZoneGraveyard
)

type Card struct {
	CardID int // card ID
	DefNo  int // the card's serial number
	Team   int
	Type   int // structure, unit, trap, spell
	Zone   int // 0: deck, 1: hand, 2: field, 3: graveyard
	Armed  bool
}
