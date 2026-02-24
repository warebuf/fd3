package helpers

import (
	"math/rand"

	"github.com/warebuf/fd3/engine"
)

// implementing reservoir sampling, as it is considered the practically fastest method

// same team, includes casting unit
func RandomSameTeam(state *engine.GameState, teamIdx int) (int, bool) {
	var chosen int
	count := 0

	for i := 0; i < len(state.Team); i++ {
		if state.Zone[i] == 1 && state.Team[i] == teamIdx && state.HP[i] > 0 {
			count++
			if rand.Intn(count) == 0 {
				chosen = i
			}
		}
	}

	if count == 0 {
		return 0, false // no units found
	}

	return chosen, true
}

func RandomSameTeamCheck(state *engine.GameState, teamIdx int) bool {
	for i := 0; i < len(state.Team); i++ {
		if state.Zone[i] == 1 && state.Team[i] == teamIdx && state.HP[i] > 0 {
			return true
		}
	}
	return false
}

// same team, does not include casting unit
func RandomSameTeamNoCaster(state *engine.GameState, teamIdx int, casterIdx int) (int, bool) {
	var chosen int
	count := 0

	for i := 0; i < len(state.HP); i++ {

		if casterIdx != i && state.Zone[i] == 1 && state.HP[i] > 0 && state.Team[i] == teamIdx {
			count++
			if rand.Intn(count) == 0 {
				chosen = i
			}
		}

	}

	if count == 0 {
		return 0, false // no units found
	}

	return chosen, true
}

func RandomEnemy(state *engine.GameState, casterTeamIdx int) (int, bool) {
	var chosen int
	count := 0

	for i := 0; i < len(state.Team); i++ {

		if state.Zone[i] == 1 && state.Team[i] != casterTeamIdx && state.HP[i] > 0 {
			count++
			if rand.Intn(count) == 0 {
				chosen = i
			}
		}

	}

	if count == 0 {
		return 0, false // no units found
	}

	return chosen, true
}

func RandomEnemyCheck(state *engine.GameState, casterTeamIdx int) bool {
	for i := 0; i < len(state.Team); i++ {
		if state.Zone[i] == 1 && state.Team[i] != casterTeamIdx && state.HP[i] > 0 {
			return true
		}
	}
	return false
}
