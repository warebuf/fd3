package defs

import (
	"github.com/warebuf/fd3/engine/state"
)

type Recurring struct {
	Milliseconds int64
	Effects      []state.Effect
	Persistent   bool
}

type CardDef struct {
	Type        int
	Name        string
	Description string

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
		BaseHP:      3,
		Cost:        1,
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
				Persistent: false,
			},
		},
	},
	2: {
		Type:        state.CardTypeSpell,
		Name:        "Year Zero",
		Description: "Destroy all summoned units on the field",
		OnPlay: []state.Effect{
			{
				Type: state.EffectDestroy,
				A:    0, // Destroy all units
			},
		},
	},
	3: {
		Type:        state.CardTypeUnit,
		Name:        "Matter of Time",
		Description: "Gain 1 mana every 20 seconds",
		BaseHP:      5,
		Cost:        0,
		Recurring: []Recurring{
			{
				Milliseconds: 20000,
				Effects: []state.Effect{
					{
						Type: state.EffectMana, // fill A, which is team
						B:    1,                // amount
					},
				},
				Persistent: false,
			},
		},
	},
}
