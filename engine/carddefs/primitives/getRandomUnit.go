package primitives

import (
	"math/rand"

	"github.com/warebuf/fd3/engine/states"
)

// implementing reservoir sampling, as it is considered the practically fastest method

// includes self, negative if nothing
func RandomSameTeam(state *states.GameState, teamIdx int) int {
	chosen := -1
	count := 0

	for i := 0; i < len(state.Cards); i++ {
		if state.Cards[i].Zone == 1 && state.Cards[i].Team == teamIdx && state.Cards[i].HP > 0 {
			count++
			if rand.Intn(count) == 0 {
				chosen = i
			}
		}
	}

	return chosen
}
