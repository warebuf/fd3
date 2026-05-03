package defs

import (
	"github.com/warebuf/fd3/engine/state"
)

// so right now this file contains

type Recurring struct {
	Milliseconds int64
	Effects      []state.Effect
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
		Description: "",
		BaseHP:      10,
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
			},
		},
	},
}
