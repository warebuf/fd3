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

type TickCondition int

const (
	TickIndefinite     TickCondition = iota // always reschedules
	TickUntilChargeMax                      // stops rescheduling once the Active slot it charges hits MaxCharges
)

type AuraModifier struct {
	SourceCardIndex int
	Stat            int
	Amount          int
}

// NOTE: this is weird, we are tracking a lot of things, it is dynamic arr, tickcon is not being used
type TickAbility struct {
	AbilityID        int
	AbilityIndex     int // index into card def's ability list
	Condition        TickCondition
	ActiveIndex      int // which Actives[] slot Condition checks against, -1 if not applicable (Condition == TickIndefinite)
	StartAtSimTime   int64
	ExecuteAtSimTime int64
}

// NOTE: we maintain targets on this object but not Conditions
type ActiveAbility struct {
	Charges     int
	MaxCharges  int
	Target      int // preselected target for this ability slot, -1 if none
	TargetFlags int
	Pending     int  // activations currently in flight on the stack, not yet resolved
	AutoAbility bool // if true, fires itself once Charges/Cost/Cons all hold, instead of waiting for player input; player-toggleable at runtime
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
	Team int
	Type int // structure, unit, trap, spell, continuous
	Zone int // 0: deck, 1: hand, 2: field, 3: graveyard, 4: stack, 5: continuous

	StageOrder    int           // 0: not staged; 1+: reveal order during board setup phase
	AutoSummon    int           // 0: not queued; 1+: priority order in play phase autosummon queue
	OnTick        []TickAbility // NOTE: should make this size 3 prob
	Actives       [3]ActiveAbility
	AuraModifiers []AuraModifier
}
