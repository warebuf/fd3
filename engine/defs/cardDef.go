package defs

import (
	"math"

	"github.com/warebuf/fd3/engine/state"
)

const (
	TargetSelf     = 1 << iota // the card itself
	TargetOwnField             // friendly field cards (excluding self)
	TargetOppField             // opponent field cards
	TargetOwnHand              // friendly hand cards
	TargetOppHand              // opponent hand cards
	TargetOwnCont              // friendly continuous zone
	TargetOppCont              // opponent continuous zone
)

// which zone(s) the owning card must be in for OnTrigger
const (
	TriggerZoneField = 1 << iota
	TriggerZoneContinuous
	TriggerZoneGraveyard
	TriggerZoneHand
)

type Ability struct {
	Effects      [3]state.Effect
	Cons         [3]state.ConditionType
	Cost         int
	Unlimited    bool              // if true is always available, and ignores Charges/Pending
	Milliseconds int64             // only used by OnTick
	TargetFlags  int               // bitmask of Target flags
	Trigger      state.TriggerType // only used by OnTrigger, condition we do a trigger check
	TriggerZones int               // only used by OnTrigger, bitmask of TriggerZone flags
	CanAutoFire  bool              // only used by Active
	ChainEntry   state.ChainEntry  // used by Active, OnTick and OnTrigger, the slots that push chain items
}

// change from dynamic arrs to fixed size
type CardDef struct {
	Type        int
	Name        string
	Description string

	Archetype int
	Taunt     bool

	ChainEntry state.ChainEntry
	IsToken    bool // if set true, does not go to grave

	OnPlay    [3]Ability
	OnTick    [3]Ability
	Active    [3]Ability
	OnTrigger [3]Ability

	Charges    int // starting charges; shared pool any ability slot on this card may gain or spend from
	MaxCharges int

	BaseHP int
	Cost   int
}

var CardDefs = []CardDef{
	0: {}, // empty, so DefNo lines up with index
	1: {
		Type:        state.CardTypeUnit,
		Name:        "Fallen Soldier",
		Description: "Deal 1 random damage every 30 ticks to an enemy",
		BaseHP:      2,
		Cost:        60,
		ChainEntry:  state.ChainAction,
		OnTick: [3]Ability{{
			ChainEntry:   state.ChainAction,
			Milliseconds: 30000,
			Effects: [3]state.Effect{
				{
					Type: state.EffectAttack,
					A:    0, // targetType, 0: random, 1: user-targetted, 2: prioritize archetype
					B:    1, // amount
				},
			},
		}},
	},
	2: {
		Type:        state.CardTypeSpell,
		Name:        "Year Zero",
		Description: "Destroy all units on the field",
		Cost:        300,
		ChainEntry:  state.ChainDual,
		OnPlay: [3]Ability{{
			Effects: [3]state.Effect{
				{
					Type: state.EffectDestroy,
					A:    0, // Destroy all units
				},
			},
		}},
	},
	3: {
		Type:        state.CardTypeUnit,
		Archetype:   state.ArchetypeTree,
		Name:        "Money Tree",
		Description: "Gain 10 mana every 10 ticks",
		BaseHP:      1,
		Cost:        60,
		ChainEntry:  state.ChainAction,
		OnTick: [3]Ability{{
			ChainEntry:   state.ChainAction,
			Milliseconds: 10000,
			Effects: [3]state.Effect{
				{
					Type: state.EffectMana, // fill A, which is team
					B:    10,               // amount
				},
			},
		}},
	},
	4: {
		Type:        state.CardTypeSpell,
		Name:        "Downforce",
		Description: "Gain 30 mana",
		Cost:        0,
		ChainEntry:  state.ChainDual,
		OnPlay: [3]Ability{{
			Effects: [3]state.Effect{
				{
					Type: state.EffectMana, // fill A, which is team
					B:    30,               // amount
				},
			},
		}},
	},
	5: {
		Type:        state.CardTypeUnit,
		Name:        "Barkbreaker",
		Description: "Prioritizes and double-damages 'Tree' archetype. Deal 1 damage every 20 ticks to a random",
		BaseHP:      2,
		Cost:        100,
		ChainEntry:  state.ChainAction,
		OnTick: [3]Ability{{
			ChainEntry:   state.ChainAction,
			Milliseconds: 20000,
			Effects: [3]state.Effect{
				{
					Type: state.EffectAttack,
					A:    2,                   // targetType, 0: random, 1: user-targetted, 2: prioritize archetype
					B:    1,                   // base damage
					C:    state.ArchetypeTree, // priority archetype
				},
			},
		}},
	},
	6: {
		Type:        state.CardTypeUnit,
		Name:        "Metal Heart",
		Description: "Taunt",
		BaseHP:      6,
		Cost:        80,
		ChainEntry:  state.ChainAction,
		Taunt:       true,
	},
	7: {
		Type:        state.CardTypeUnit,
		Name:        "Tomoe",
		Description: "Heal 1 HP every 30 ticks to a random friendly unit",
		BaseHP:      2,
		Cost:        60,
		ChainEntry:  state.ChainAction,
		OnTick: [3]Ability{{
			ChainEntry:   state.ChainAction,
			Milliseconds: 30000,
			Effects: [3]state.Effect{
				{
					Type: state.EffectHeal,
					B:    1, // amount
				},
			},
		}},
	},
	8: {
		Type:        state.CardTypeSpell,
		Name:        "Counterspell",
		Description: "Negate a summon or ability and destroy the source card.",
		Cost:        30,
		ChainEntry:  state.ChainReaction,
		OnPlay: [3]Ability{{
			Effects: [3]state.Effect{
				{Type: state.EffectCounter},
			},
		}},
	},
	9: {
		Type:        state.CardTypeContinuous,
		Name:        "Rolling in the Deep",
		Description: "Use (50): Draw a card. 3 uses.",
		Cost:        80,
		ChainEntry:  state.ChainDual,
		Charges:     3,
		MaxCharges:  math.MaxInt,
		Active: [3]Ability{{
			Cost:        50,
			Cons:        [3]state.ConditionType{state.ConditionManaAtLeast},
			CanAutoFire: true,
			ChainEntry:  state.ChainDual,
			Effects:     [3]state.Effect{{Type: state.EffectDraw, B: 1}},
		}},
	},
	10: {
		Type:        state.CardTypeUnit,
		Name:        "Archangel",
		Description: "Maximum Charges: 3. Gain 1 charge every 30 ticks. Spend a charge to deal 1 damage to a target.",
		BaseHP:      2,
		Cost:        70,
		ChainEntry:  state.ChainAction,
		MaxCharges:  3,
		OnTick: [3]Ability{{
			ChainEntry:   state.ChainAction,
			Milliseconds: 30000,
			Effects: [3]state.Effect{
				{Type: state.EffectChargeSelf},
			},
		}},
		Active: [3]Ability{{
			TargetFlags: TargetSelf | TargetOwnField | TargetOppField,
			Cons:        [3]state.ConditionType{state.ConditionChargeAtLeastOne},
			CanAutoFire: true,
			ChainEntry:  state.ChainDual,
			Effects: [3]state.Effect{
				{
					Type: state.EffectAttack,
					A:    1, // targetType, 0: random, 1: user-targetted, 2: prioritize archetype, 3: always struct
					B:    1, // amount
				},
			},
		}},
	},
	11: {
		Type:        state.CardTypeUnit,
		Name:        "Forsaken",
		Description: "Use: Sacrifice a friendly unit you choose. Deal damage directly to the enemy structure equal to 1/10th its mana cost.",
		BaseHP:      3,
		Cost:        60,
		ChainEntry:  state.ChainAction,
		Active: [3]Ability{{
			Unlimited:   true,
			TargetFlags: TargetSelf | TargetOwnField,
			ChainEntry:  state.ChainDual,
			Effects: [3]state.Effect{
				{Type: state.EffectSacrifice},
				{Type: state.EffectAttack, A: 3}, // A: 3, always hits enemy structure; B is filled in by the sacrifice at runtime
			},
		}},
	},
	12: {
		Type:        state.CardTypeContinuous,
		Name:        "Charge Reservoir",
		Description: "Gain 1 charge whenever you cast a spell.",
		Cost:        60,
		ChainEntry:  state.ChainDual,
		MaxCharges:  math.MaxInt,
		OnTrigger: [3]Ability{{
			Trigger:      state.TriggerAllySpellResolve,
			TriggerZones: TriggerZoneContinuous,
			ChainEntry:   state.ChainDual,
			Effects:      [3]state.Effect{{Type: state.EffectChargeSelf}},
		}},
	},
	13: {
		Type:        state.CardTypeUnit,
		Name:        "Charge Consumer",
		Description: "Summon: Destroy a friendly card you choose; gain its charges as HP. Use: Deal 1 damage to an enemy, plus those charges, then this unit is destroyed.",
		BaseHP:      1,
		Cost:        60,
		ChainEntry:  state.ChainAction,
		Charges:     1,
		MaxCharges:  1,
		OnPlay: [3]Ability{{
			TargetFlags: TargetOwnField | TargetOwnCont,
			Effects:     [3]state.Effect{{Type: state.EffectConsume, C: 0}}, // C: 0, feeds Active slot 0's Bonus
		}},
		Active: [3]Ability{{
			TargetFlags: TargetOppField,
			Cons:        [3]state.ConditionType{state.ConditionChargeAtLeastOne},
			CanAutoFire: true,
			ChainEntry:  state.ChainDual,
			Effects: [3]state.Effect{
				{Type: state.EffectAttack, A: 1, B: 1}, // A: 1, user-targeted; base 1 damage, plus Bonus
				{Type: state.EffectSelfDestruct},
			},
		}},
	},
	14: {
		Type:        state.CardTypeUnit,
		Name:        "Charge Token",
		Description: "Destroyed after 10 ticks.",
		BaseHP:      1,
		IsToken:     true,
		ChainEntry:  state.ChainAction,
		OnTick: [3]Ability{{
			ChainEntry:   state.ChainAction,
			Milliseconds: 10000,
			Effects:      [3]state.Effect{{Type: state.EffectSelfDestruct}},
		}},
	},
	15: {
		Type:        state.CardTypeSpell,
		Name:        "Temp Tokens",
		Description: "Destroy a friendly Charge Reservoir you choose; for every 2 charges it had, create a token with 1 HP that lasts 10 ticks.",
		Cost:        60,
		ChainEntry:  state.ChainDual,
		OnPlay: [3]Ability{{
			TargetFlags: TargetOwnField | TargetOwnCont,
			Effects:     [3]state.Effect{{Type: state.EffectConvertToTokens, B: 14}}, // B: 14, Charge Token's DefNo
		}},
	},
	16: {
		Type:        state.CardTypeSpell,
		Name:        "Kindling",
		Description: "If none of your cards have 2+ charges, give a card 2 charges.",
		Cost:        50,
		ChainEntry:  state.ChainDual,
		OnPlay: [3]Ability{{
			TargetFlags: TargetOwnField | TargetOppField | TargetOwnCont | TargetOppCont,
			Effects:     [3]state.Effect{{Type: state.EffectChargeTarget, B: 2}},
		}},
	},
	17: {
		Type:        state.CardTypeSpell,
		Name:        "Rummage",
		Description: "Draw a card.",
		Cost:        40,
		ChainEntry:  state.ChainDual,
		OnPlay: [3]Ability{{
			Effects: [3]state.Effect{{Type: state.EffectDraw, B: 1}},
		}},
	},
}
