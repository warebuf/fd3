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

type TickAbility struct {
	Active           bool // whether this OnTick slot currently has a scheduled tick
	Condition        TickCondition
	StartAtSimTime   int64
	ExecuteAtSimTime int64
}

type OnPlayAbility struct {
	Target      int // preselected target for this card's OnPlay ability, -1 if none
	TargetFlags int // uses the target enum for bitmasking; 0 if this card's OnPlay ability doesn't target anything
}

type ActiveAbility struct {
	// target stuff
	Target      int // preselected target for this ability slot, -1 if none
	TargetFlags int // uses the target enum for bitmasking

	// condition stuff
	Cons      [3]ConditionType // what condition(s) must be met for this ability to be activated
	Cost      int
	Unlimited bool // this ability ignores Charges

	// autoability stuff
	AutoAbility bool // autofires if true, player-toggleable

	// effect stuff
	Bonus int // added to this ability's EffectAttack B when built; accumulated by earlier EffectConsume(s)
}

type Card struct {
	// description related
	DefNo       int // the card's serial number
	Name        string
	Description string

	// state related
	HP        int
	Cost      int
	Archetype int
	Team      int
	Type      int // structure, unit, trap, spell, continuous
	Zone      int // 0: deck, 1: hand, 2: field, 3: graveyard, 4: stack, 5: continuous
	IsToken   bool // tokens skip the graveyard on destruction

	// charge related
	Charges        int // shared charge pool for this card; any ability slot may gain or spend from it
	MaxCharges     int
	PendingCharges int // activations currently in flight on the stack across all Active slots, not yet resolved

	// autoplay related
	StageOrder int // 0: not staged; 1+: reveal order during board setup phase
	AutoSummon int // 0: not queued; 1+: priority order in play phase autosummon queue

	// ability related
	OnTick  [3]TickAbility
	Actives [3]ActiveAbility
	Taunt   bool

	// on-play targeting related
	OnPlay [3]OnPlayAbility // mirrors CardDef.OnPlay's TargetFlags; Target is preselected before the card is played
}
