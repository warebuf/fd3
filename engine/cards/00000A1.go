package abilities

import (
	"github.com/warebuf/fd3/engine/cards/primitives"
	"github.com/warebuf/fd3/engine/states"
)

// Give 2 HP to a random ally unit
type U00000A1 struct {
	// base info
	NodeIndex   uint
	CasterIndex int

	// timing
	DeclarationTime int
	Window          int

	// phase for simulation
	phase   int
	counter int

	// status
	Negated bool

	// special level info
	AbilityIndex int
	TargetIndex  int
	Amount       int
}

func (ability *U00000A1) Declaration(state *states.GameState) bool {

	// get ability index
	startIndex := state.AbilityStartIdx[ability.CasterIndex]
	abilityCount := state.AbilityCount[ability.CasterIndex]
	effectNo := 2
	for i := startIndex; i < startIndex+abilityCount; i++ {
		if state.Abilities[startIndex].EffectNo == effectNo {
			ability.AbilityIndex = i
			break
		}
	}

	// check if valid (if unit is alive and ability is charged)
	if state.HP[ability.CasterIndex] <= 0 || state.Abilities[ability.AbilityIndex].Charge < state.Abilities[ability.AbilityIndex].SPD {
		return false
	}

	// set the random ally target
	targetIndex := primitives.RandomSameTeam(state, state.Team[ability.CasterIndex])
	if targetIndex < 0 {
		return false
	}
	ability.TargetIndex = targetIndex

	// set the ability in pending state
	state.Abilities[ability.AbilityIndex].Pending = true

	// TODO: check for autotriggers
	for i := ability.counter; i < len(state.HP); i++ {
		ability.counter += 1
	}
	ability.phase += 1 // set to phase 1
	ability.counter = 0

	return true

}

func (ability *U00000A1) Resolution(state *states.GameState) bool {

	// check if valid, if not return
	if state.HP[ability.CasterIndex] <= 0 ||
		state.Abilities[ability.AbilityIndex].Charge < state.Abilities[ability.AbilityIndex].SPD ||
		state.Abilities[ability.AbilityIndex].Pending == false {
		return false
	}

	// check if target is alive, if not reset the charge/pending and return
	if state.HP[ability.TargetIndex] <= 0 {
		state.Abilities[ability.AbilityIndex].Charge = 0
		state.Abilities[ability.AbilityIndex].Pending = false
		return false
	}

	// heal
	state.HP[ability.TargetIndex] += ability.Amount

	// TODO: check for autotriggers
	for i := ability.counter; i < len(state.HP); i++ {
		ability.counter += 1
	}
	ability.phase += 1 // set to phase 2
	ability.counter = 0

	return true
}
