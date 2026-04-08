package primitives

import (
	"math/rand"

	"github.com/warebuf/fd3/engine/state"
)

func ShuffleDeck(gs *state.GameState, team int) {
	deck := gs.Deck[team]

	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
}
