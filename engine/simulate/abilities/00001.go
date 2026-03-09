package abilities

import (
	"github.com/warebuf/fd3/engine"
)

// normal summon, Negate all effects of a card
type S00001 struct {

	// base info
	SummonedIndex   int
	DeclarationTime int64
	Window          int
	Valid           bool

	// phase for simulation
	phase int
}

func (S00001) Type() (string, int) { return "Summon", 1 }

// Drain 1 HP from ally units
type A00001 struct {

	// base info
	CasterIndex     int
	DeclarationTime int
	Window          int
	Valid           bool

	// ability level info
	Target int
	Amount int

	// phase for simulation
	phase int
}

func (A00001) Type() (string, int) { return "Ability", 1 }

func (ability A00001) Declaration(state *engine.GameState) {

}

func (ability A00001) Resolution(state *engine.GameState) {

}
