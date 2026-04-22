package defs

import (
	"github.com/warebuf/fd3/engine/state"
)

// so right now this file contains

type CardDef struct {
	Type        int
	Name        string
	Description string

	OnPlay    []state.Effect
	Triggers  []state.TriggerEffect
	Recurring []state.RecurringEffect
	Passives  []state.Passive

	BaseHP int
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
}
