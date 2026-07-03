package state

const (
	CardTypeStructure = iota
	CardTypeUnit
	CardTypeTrap
	CardTypeSpell
	CardTypeContinuous
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
	ZoneStack
	ZoneContinuous
)

const (
	AbilityTimerNormal = iota
)

type CardStatus int

const (
	CardStatusDefault          CardStatus = iota // spells, structures, continuous
	CardStatusSummoning                          // unit: on field, summon window not yet resolved
	CardStatusUnitActive                         // unit: fully resolved and active
	CardStatusSet                                // trap: face-down on field
	CardStatusTrapActive                         // trap: triggered/active
	CardStatusContinuousActive                   // continuous spell: in continuous zone
)

type AuraModifier struct {
	SourceCardID int
	Stat         int
	Amount       int
}

type AbilityTimer struct {
	AbilityID        int
	AbilityIdx       int // index into card def's ability list
	Type             int // AbilityTimerNormal, AbilityTimerSummon
	StartAtSimTime   int64
	ExecuteAtSimTime int64
}

// ActiveState is the runtime state for one Active ability slot on a card
// (a card can have up to 3, matching defs.CardDef.Active's [3]Ability).
type ActiveState struct {
	Charges     int
	MaxCharges  int
	Target      int // preselected target for this ability slot, -1 if none
	TargetFlags int
	Pending     int // activations currently in flight on the stack, not yet resolved
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
	Type          int // structure, unit, trap, spell, continuous
	Zone          int // 0: deck, 1: hand, 2: field, 3: graveyard, 4: stack, 5: continuous
	Status        CardStatus
	StageOrder    int // 0: not staged; 1+: reveal order during board setup phase
	AutoSummon    int // 0: not queued; 1+: priority order in play phase autosummon queue
	AbilityTimers []AbilityTimer
	Actives       [3]ActiveState
	AuraModifiers []AuraModifier
}
