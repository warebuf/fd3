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
	TickUntilChargeMax                      // stops rescheduling once the card's Charges hits MaxCharges
)

type ConditionType int

const (
	ConditionManaAtLeast      ConditionType = iota + 1 // checks gs.Mana[team] >= this Ability's own Cost
	ConditionChargeAtLeastOne                          // checks this card's Charges >= 1
)

type TriggerType int

const (
	TriggerAllySpellResolve TriggerType = iota + 1 // fires when this card's owner has a spell resolve
)

type AuraModifier struct {
	SourceCardIndex int
	Stat            int
	Amount          int
}

type TickAbility struct {
	Active           bool // whether this OnTick slot currently has a scheduled tick
	Condition        TickCondition
	StartAtSimTime   int64
	ExecuteAtSimTime int64
}

type ActiveAbility struct {
	Unlimited   bool // if true, this ability ignores Charges/Pending entirely and is always available
	Target      int  // preselected target for this ability slot, -1 if none
	TargetFlags int
	Cons        [3]ConditionType
	Cost        int
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

	Charges    int // shared charge pool for this card; any ability slot may gain or spend from it
	MaxCharges int

	StageOrder    int // 0: not staged; 1+: reveal order during board setup phase
	AutoSummon    int // 0: not queued; 1+: priority order in play phase autosummon queue
	OnTick        [3]TickAbility
	Actives       [3]ActiveAbility
	AuraModifiers []AuraModifier
}
