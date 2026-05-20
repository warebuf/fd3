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
	ZoneChain
)

const (
	AbilityTimerNormal = iota
	AbilityTimerSummon
)

type AbilityTimer struct {
	AbilityID  int
	AbilityIdx int   // index into card def's ability list; -1 for summon window
	Type       int   // AbilityTimerNormal, AbilityTimerSummon
	StartAtSimTime   int64
	ExecuteAtSimTime int64
}

type Card struct {
	// descriptive info
	DefNo       int // the card's serial number
	Name        string
	Description string
	HP          int
	Cost        int

	// state info
	Team          int
	Type          int // structure, unit, trap, spell
	Zone          int // 0: deck, 1: hand, 2: field, 3: graveyard, 4: chain
	Armed         bool
	AutoSummon    int // 0: not staged; 1+: order in which to autosummon during reveal phase
	AbilityTimers []AbilityTimer
}
