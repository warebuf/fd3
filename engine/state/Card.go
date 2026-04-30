package state

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
	ZoneTemporaryUnaccessible
)

type Card struct {

	// descriptive info
	DefNo       int // the card's serial number
	Name        string
	Description string

	// state info
	CardID int // card ID
	Team   int
	Type   int // structure, unit, trap, spell
	Zone   int // 0: deck, 1: hand, 2: field, 3: graveyard, 4: temporarily unaccessible
	Armed  bool
}
