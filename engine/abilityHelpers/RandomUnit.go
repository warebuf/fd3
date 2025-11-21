package abilityHelpers

import (
	"math/rand"

	"github.com/warebuf/fd3/engine"
)

// implementing reservoir sampling, as it is considered the practically fastest method

// same team, includes casting unit
func RandomSameTeam(state *engine.GameState, teamIdx uint) (uint, bool) {
	var chosen uint
	count := 0

	for i := 0; i < len(state.Team); i++ {

		if state.Zone[i] == 1 && state.Team[i] == teamIdx {
			count++
			if rand.Intn(count) == 0 {
				chosen = uint(i)
			}
		}

	}

	if count == 0 {
		return 0, false // no units found
	}

	return chosen, true
}

// same team, does not include casting unit
func RandomSameTeamNoCaster(state *engine.GameState, teamIdx uint, casterIdx uint) (uint, bool) {
	var chosen uint
	count := 0

	for i := 0; i < len(state.HP); i++ {

		if casterIdx != uint(i) && state.Zone[i] == 1 && state.HP[i] > 0 && state.Team[i] == teamIdx {
			count++
			if rand.Intn(count) == 0 {
				chosen = uint(i)
			}
		}

	}

	if count == 0 {
		return 0, false // no units found
	}

	return chosen, true
}

func RandomEnemy(state *engine.GameState, casterTeamIdx uint) (uint, bool) {
	var chosen uint
	count := 0

	for i := 0; i < len(state.Team); i++ {

		if state.Zone[i] == 1 && state.Team[i] != casterTeamIdx {
			count++
			if rand.Intn(count) == 0 {
				chosen = uint(i)
			}
		}

	}

	if count == 0 {
		return 0, false // no units found
	}

	return chosen, true
}
