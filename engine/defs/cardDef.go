package defs

import (
	"github.com/warebuf/fd3/engine/state"
)

type Recurring struct {
	Milliseconds int64
	Effects      []state.Effect
}

type CardDef struct {
	Type        int
	Name        string
	Description string

	Archetype int

	OnPlay    []state.Effect
	Recurring []Recurring

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
		Cost:        100,
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
		Cost:        5000,
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
		Description: "Gain 1 mana every 5 seconds",
		BaseHP:      2,
		Cost:        500,
		Recurring: []Recurring{
			{
				Milliseconds: 5000,
				Effects: []state.Effect{
					{
						Type: state.EffectMana, // fill A, which is team
						B:    1,                // amount
					},
				},
			},
		},
	},
	4: {
		Type:        state.CardTypeSpell,
		Name:        "Downforce",
		Description: "Gain 500 mana",
		BaseHP:      0,
		Cost:        0,
		OnPlay: []state.Effect{
			{
				Type: state.EffectMana,
				B:    500, // amount
			},
		},
	},
	5: {
		Type:        state.CardTypeUnit,
		Name:        "Barkbreaker",
		Description: "Prioritizes and double-damages 'Tree' archetype. deal 1 damage to a random enemy every 20 seconds",
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
}
