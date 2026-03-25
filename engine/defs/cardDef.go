package defs

import (
	"github.com/warebuf/fd3/engine/states/gameState"
)

type CardDef struct {
	Type int

	OnPlay    []gameState.Effect
	Triggers  []gameState.TriggerEffect
	Recurring []gameState.RecurringEffect
	Passives  []gameState.Passive

	BaseHP int
}

var CardDefs = []CardDef{
	0: {}, // empty, so DefNo lines up with index
	1: {Type: gameState.CardTypeUnit},
	2: {Type: gameState.CardTypeUnit, OnPlay: []gameState.Effect{}},
}
