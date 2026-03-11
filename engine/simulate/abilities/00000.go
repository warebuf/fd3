package abilities

import (
	"github.com/warebuf/fd3/engine"
)

// normal summon with summon effect
type S00000 struct {

	// base info
	NodeIndex     uint
	SummonedIndex int

	// timing
	DeclarationTime int64
	Window          int

	//
	Negated bool

	// resume mechanism
	phase   int // phase for simulation
	counter int // counter that we use so we can restart at the same position

	// special case
	EffectNegated bool // in this case, if summon not negated, but effect may be
}

func (ability S00000) Declaration(state *engine.GameState) bool {

	// check if valid (alive and in hand)
	if state.HP[ability.SummonedIndex] <= 0 || state.Zone[ability.SummonedIndex] != 0 {
		return false
	}

	// TODO: check for autocounters
	for i := ability.counter; i < len(state.HP); i++ {
		ability.counter += 1
	}
	ability.phase += 1 // set to phase 1
	ability.counter = 0

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

	// TODO: check for summon autocounters
	if ability.phase < 2 {
		for i := ability.counter; i < len(state.HP); i++ {
			ability.counter += 1
		}
	}
	ability.phase += 1 // set to phase 2
	ability.counter = 0

	// set zone to 1
	state.Zone[ability.SummonedIndex] = 1

	// TODO: check for post summon triggers
	if ability.phase < 3 {
		for i := ability.counter; i < len(state.HP); i++ {
			ability.counter += 1
		}
	}
	ability.phase += 1 // set to phase 3
	ability.counter = 0

	// TODO: check for effect counters
	if ability.phase < 4 {
		for i := ability.counter; i < len(state.HP); i++ {
			ability.counter += 1
		}
	}
	ability.phase += 1 // set to phase 4
	ability.counter = 0

	// deal 5 damage to all field units, check for individual counters
	if ability.phase < 5 && !ability.EffectNegated {
		for i := ability.counter; i < len(state.Zone); i++ {

			ability.counter += 1 // increment so we don't check this unit again

			if state.HP[i] > 0 && state.Zone[i] == 1 {

				// TODO: check for attack counter triggers

				state.HP[i] -= 5

			}
		}
	}
	ability.phase += 1 // set to phase 5
	ability.counter = 0

	// TODO: check for post attack triggers
	if ability.phase < 6 {
		for i := ability.counter; i < len(state.HP); i++ {
			ability.counter += 1
		}
	}
	ability.phase += 1 // set to phase 6
	ability.counter = 0

	return true, !ability.EffectNegated
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
	startIndex := state.AbilityStartIdx[ability.CasterIndex]
	abilityCount := state.AbilityCount[ability.CasterIndex]
	for i := startIndex; i < startIndex+abilityCount; i++ {

	}

	// get the random ally target

	// set the ability in pending state

}

func (ability A00000) Resolution(state *engine.GameState) {

	// check if valid, if not return

	// check if target is alive, if not return

	//

}
