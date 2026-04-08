package defs

import (
	"github.com/warebuf/fd3/engine/state"
)

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
	1: {Type: state.CardTypeUnit},
	2: {Type: state.CardTypeUnit, OnPlay: []state.Effect{}},
}
