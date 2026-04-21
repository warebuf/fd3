package defs

import (
	"github.com/warebuf/fd3/engine/state"
)

// so right now this file contains

type CardDef struct {
	Type int

	OnPlay    []state.Effect
	Triggers  []state.TriggerEffect
	Recurring []state.RecurringEffect
	Passives  []state.Passive

	BaseHP int
}

var CardDefs = []CardDef{
	0: {}, // empty, so DefNo lines up with index
	1: {
		Type:   state.CardTypeUnit,
		BaseHP: 10,
	},
	2: {
		Type: state.CardTypeSpell,
		OnPlay: []state.Effect{
			{
				Type: state.EffectDestroy,
				A:    0, // Destroy all units
			},
		},
	},
}
