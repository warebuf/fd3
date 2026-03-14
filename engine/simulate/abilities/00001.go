package abilities

import (
	"github.com/warebuf/fd3/engine"
)

// normal summon with declaration effect - if successfully summoned, negate all effects of a card
type U00001S struct {
	// base info
	NodeIndex     uint
	SummonedIndex int

	// timing
	DeclarationTime int64
	Window          int

	// resume mechanism
	phase   int // phase for simulation
	counter int // counter that we use so we can restart at the same position

	// status
	Negated bool

	// special level info
	Effect *U00000A1 // if declaration is success, spawns an object, therefore we reference it here
}

func (ability *U00001S) Declaration(state *engine.GameState) bool {
	return true
}

func (ability *U00001S) Resolution(state *engine.GameState) bool { // returns summon success
	return true
}

func (U00001S) EffectNo() int { return 2 }

// negate all effects of a card
type U00001A1 struct {
}

func (U00001A1) EffectNo() int { return 3 }

// Drain 1 HP from ally units
type U00001A2 struct {
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

func (U00001A2) EffectNo() int { return 4 }

func (ability U00001A2) Declaration(state *engine.GameState) {

}

func (ability U00001A2) Resolution(state *engine.GameState) {

}
