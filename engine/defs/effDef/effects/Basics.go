package effects

import "github.com/warebuf/fd3/engine/states/gameState"

func ApplyEffectDraw(state *gameState.GameState, eff gameState.Effect) {
	teamIndex := eff.A
	numCards := eff.B

	for i := 0; i < numCards; i++ {
		if len(state.Deck[teamIndex]) == 0 { // TODO: if deck is empty
			return
		}

		cardIdx := state.Deck[teamIndex][0]
		state.Deck[teamIndex] = state.Deck[teamIndex][1:]              // draw top card
		state.Hand[teamIndex] = append(state.Hand[teamIndex], cardIdx) // move to hand
		state.Cards[cardIdx].Zone = gameState.ZoneHand                 // update card state
	}
}
