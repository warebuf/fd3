package abilities

import (
	"github.com/warebuf/fd3/engine"
)

// the behavior is, create two fresh stacks/nodes
// if the effect node is delayed, only the effect node
// if the summon node is delayed, delay both nodes

// normal summon with summon effect
type U00000S struct {
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
	EffectNodeIndex uint //
}

type U00000SE struct {
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
	SummonNodeIndex uint
}

func (ability *U00000S) Declaration(state *engine.GameState) bool {

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
func (ability *U00000S) Resolution(state *engine.GameState) (bool, bool) { // returns summon success, effect success

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
