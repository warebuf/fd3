package helpers

import (
	"math/rand"

	"github.com/warebuf/fd3/engine"
)

// implementing reservoir sampling, as it is considered the practically fastest method

// includes self, negative if nothing
func RandomSameTeam(state *engine.GameState, teamIdx int) int {
	chosen := -1
	count := 0

	for i := 0; i < len(state.Team); i++ {
		if state.Zone[i] == 1 && state.Team[i] == teamIdx && state.HP[i] > 0 {
			count++
			if rand.Intn(count) == 0 {
				chosen = i
			}
		}
	}

	return chosen
}
