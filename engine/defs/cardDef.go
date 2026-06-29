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

type Recurring struct {
	Milliseconds int64
	Effects      []state.Effect
}

type AuraEffect struct {
	TargetType int
	Stat       int
	Amount     int
}

type CardDef struct {
	Type        int
	Name        string
	Description string

	Archetype int
	Taunt     bool

	RequiresStackTarget bool

	OnPlay    []state.Effect
	Recurring []Recurring
	Aura      []AuraEffect

	TargetFlags int

	UseEffects []state.Effect
	UseCost    int
	Charges    int
	MaxCharges int

	BaseHP int
	Cost   int
}

var CardDefs = []CardDef{
	0: {}, // empty, so DefNo lines up with index
	1: {
		Type:        state.CardTypeUnit,
		Name:        "Fallen Soldier",
		Description: "Deal 1 random damage every 20 seconds",
		BaseHP:      2,
		Cost:        60,
		Recurring: []Recurring{
			{
				Milliseconds: 20000,
				Effects: []state.Effect{
					{
						Type: state.EffectAttack,
						A:    0, // targetType, 0: random, 1: user-targetted
						B:    1, // amount
					},
				},
			},
		},
	},
	2: {
		Type:        state.CardTypeSpell,
		Name:        "Year Zero",
		Description: "Destroy all summoned units on the field",
		Cost:        300,
		OnPlay: []state.Effect{
			{
				Type: state.EffectDestroy,
				A:    0, // Destroy all units
			},
		},
	},
	3: {
		Type:        state.CardTypeUnit,
		Archetype:   state.ArchetypeTree,
		Name:        "Money Tree",
		Description: "Gain 10 mana every 10 seconds",
		BaseHP:      2,
		Cost:        60,
		Recurring: []Recurring{
			{
				Milliseconds: 10000,
				Effects: []state.Effect{
					{
						Type: state.EffectMana, // fill A, which is team
						B:    10,               // amount
					},
				},
			},
		},
	},
	4: {
		Type:        state.CardTypeSpell,
		Name:        "Downforce",
		Description: "Gain 30 mana",
		BaseHP:      0,
		Cost:        0,
		OnPlay: []state.Effect{
			{
				Type: state.EffectMana,
				B:    30, // amount
			},
		},
	},
	5: {
		Type:        state.CardTypeUnit,
		Name:        "Barkbreaker",
		Description: "Prioritizes and double-damages 'Tree' archetype. Deal 1 damage to a random enemy every 20 seconds",
		BaseHP:      2,
		Cost:        100,
		Recurring: []Recurring{
			{
				Milliseconds: 20000,
				Effects: []state.Effect{
					{
						Type: state.EffectAttack,
						A:    2,                   // targetType, 0: random, 1: user-targetted, 2: prioritize archetype
						B:    1,                   // base damage
						C:    state.ArchetypeTree, // priority archetype
					},
				},
			},
		},
	},
	6: {
		Type:        state.CardTypeUnit,
		Name:        "Metal Heart",
		Description: "Taunt",
		BaseHP:      6,
		Cost:        80,
		Taunt:       true,
	},
	7: {
		Type:        state.CardTypeUnit,
		Name:        "Tomoe",
		Description: "Heal 1 HP to a random friendly unit every 15 seconds",
		BaseHP:      2,
		Cost:        80,
		Recurring: []Recurring{
			{
				Milliseconds: 15000,
				Effects: []state.Effect{
					{
						Type: state.EffectHeal,
						B:    1, // amount
					},
				},
			},
		},
	},
	8: {
		Type:                state.CardTypeSpell,
		Name:                "Counterspell",
		Description:         "Counters a spell",
		Cost:                100,
		RequiresStackTarget: true,
		OnPlay: []state.Effect{
			{Type: state.EffectCounter},
		},
	},
	9: {
		Type:        state.CardTypeContinuous,
		Name:        "Rolling in the Deep",
		Description: "Use (50): Draw a card. 3 uses.",
		Cost:        80,
		Charges:     3,
		MaxCharges:  math.MaxInt,
		UseCost:     50,
		UseEffects:  []state.Effect{{Type: state.EffectDraw, B: 1}},
	},
	10: {
		Type:        state.CardTypeUnit,
		Name:        "Fallen Soldier 2",
		Description: "Maximum Charges: 3. Accumulate 1 charge every 30 seconds. Spend a charge to deal 1 damage to a target.",
		BaseHP:      2,
		Cost:        60,
		Charges:     0,
		MaxCharges:  3,
		TargetFlags: TargetSelf | TargetOwnField | TargetOppField,
		Recurring: []Recurring{
			{
				Milliseconds: 30000,
				Effects: []state.Effect{
					{Type: state.EffectGainCharge},
				},
			},
		},
		UseEffects: []state.Effect{
			{
				Type: state.EffectAttack,
				A:    1, // targetType, 0: random, 1: user-targetted
				B:    1, // amount
			},
		},
	},
}
