package state

const (
	CardTypeStructure = iota
	CardTypeUnit
	CardTypeTrap
	CardTypeSpell
)

const (
	ArchetypeNone = iota
	ArchetypeTree
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

type CardStatus int

const (
	CardStatusDefault   CardStatus = iota // spells, structures
	CardStatusSummoning                   // unit: on field, summon window not yet resolved
	CardStatusUnitActive                   // unit: fully resolved and active
	CardStatusSet                         // trap: face-down on field
	CardStatusTrapActive                  // trap: triggered/active
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
	Archetype   int
	Taunt       bool

	// state info
	Team          int
	Type          int        // structure, unit, trap, spell
	Zone          int        // 0: deck, 1: hand, 2: field, 3: graveyard, 4: chain
	Status        CardStatus
	StageOrder    int // 0: not staged; 1+: reveal order during board setup phase
	AutoSummon    int // 0: not queued; 1+: priority order in play phase autosummon queue
	AbilityTimers []AbilityTimer
}
