package primitives

import (
	"math/rand"

	"github.com/warebuf/fd3/engine/states/gameState"
)

func ShuffleDeck(gs *gameState.GameState, team int) {
	deck := gs.Deck[team]

	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
}
