package abilities

import (
	"github.com/warebuf/fd3/engine"
)

// normal summon with summon effect
type S00000 struct {
	// base info
	NodeIndex       uint
	SummonedIndex   int
	DeclarationTime int64
	Window          int
	Negated         bool
	phase           int // phase for simulation

	//
	EffectNegated bool // in this case, if summon not negated, but effect may be
}

func (ability S00000) Declaration(state *engine.GameState) bool {

	// check if valid (alive and in hand)
	if state.HP[ability.SummonedIndex] <= 0 || state.Zone[ability.SummonedIndex] != 0 {
		return false
	}

	// TODO: check for autocounters

	// set zone to 3
	state.Zone[ability.SummonedIndex] = 3

	return true
}

// returns if summon successful and whether summon effect should be created
func (ability S00000) Resolution(state *engine.GameState) (bool, bool) {

	// check if valid (alive, pending summon, not negated)
	if state.HP[ability.SummonedIndex] <= 0 || state.Zone[ability.SummonedIndex] != 3 || ability.Negated {
		return false, false
	}

	// TODO: check for autocounters

	// set zone to 1
	state.Zone[ability.SummonedIndex] = 1

	// TODO: check for effect counters

	// iterate through on field units, deal 5 damage
	if !ability.EffectNegated {
		for i := 0; i < len(state.Zone); i++ {
			if state.HP[i] > 0 && state.Zone[i] == 1 {
				state.HP[i] -= 5
				// TODO: check for attack triggers
			}
		}
	}

	return true, true
}

func (S00000) Code() string { return "S00000" }

// Give 2 HP to a random ally unit
type A00000 struct {

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

func (A00000) Type() string { return "A00000" }

func (ability A00000) Declaration(state *engine.GameState) {

	// check if valid

	// get the random ally target

	// set the ability in pending state

}

func (ability A00000) Resolution(state *engine.GameState) {

	// check if valid, if not return

	// check if target is alive, if not return

	//

}
