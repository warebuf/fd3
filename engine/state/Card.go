package state

const (
	CardTypeStructure = iota
	CardTypeUnit
	CardTypeTrap
	CardTypeContinuous
	CardTypeEquip
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
	ZonePending
	ZoneContinuous
	ZoneTrap
	ZoneEquip
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
	TriggerAllyTrapResolve TriggerType = iota + 1 // fires when this card's owner has a trap resolve
	TriggerHostWouldBeDestroyed
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
	Target          int // preselected target for this ability slot, -1 if none
	TargetStructure int // preselected enemy structure slot for this ability slot, -1 if none
	TargetFlags     int // uses the target enum for bitmasking

	// condition stuff
	Cons      [3]ConditionType // what condition(s) must be met for this ability to be activated
	Cost      int
	Unlimited bool // this ability ignores Charges

	// autoability stuff
	CanAutoFire bool // whether this ability may be toggled to auto-fire at all
	AutoAbility bool // autofires if true, player-toggleable

	ChainEntry ChainEntry // how using this ability may enter the chain

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
	Type      int  // structure, unit, trap, continuous
	Zone      int  // 0: deck, 1: hand, 2: field, 3: graveyard, 4: pending, 5: continuous, 6: trap
	IsToken   bool // tokens skip the graveyard on destruction

	ChainEntry ChainEntry // how playing this card may enter the chain

	// charge related
	Charges        int // shared charge pool for this card; any ability slot may gain or spend from it
	MaxCharges     int
	PendingCharges int // activations currently in flight on the chain across all Active slots, not yet resolved

	// trap related, only meaningful while this card sits in ZoneTrap
	SetStartAtSimTime int64 // sim time this trap was set, so the client can draw the countdown
	ReadyAtSimTime    int64 // sim time it finishes setting
	Ready             bool  // whether it has finished setting and may be activated

	// autoplay related
	AutoSummon int // 0: not queued; 1+: priority order in play phase autosummon queue

	// ability related
	OnTick  [3]TickAbility
	Actives [3]ActiveAbility
	Taunt   bool

	// on-play targeting related
	OnPlay [3]OnPlayAbility // mirrors CardDef.OnPlay's TargetFlags; Target is preselected before the card is played

	EquipHostIndex int // -1 is no host
}
