package primitives

import (
	"math/rand"

	"github.com/warebuf/fd3/engine/states/gameState"
)

// implementing reservoir sampling, as it is considered the practically fastest method

// includes self, negative if nothing
func RandomSameTeam(state *gameState.GameState, teamIdx int) int {
	chosen := -1
	count := 0

	for i := 0; i < len(state.Permanents); i++ {
		if state.Permanents[i].Team == teamIdx && state.Permanents[i].HP > 0 {
			count++
			if rand.Intn(count) == 0 {
				chosen = i
			}
		}
	}

	return chosen
}
