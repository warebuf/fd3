package effects

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffectDraw(gs *state.GameState, logList *[]any, UpdateID *int, eff state.Effect, executeAtWallTime int64) {
	teamIndex := eff.A
	numCards := eff.B

	for i := 0; i < numCards; i++ {
		if len(gs.Deck[teamIndex]) == 0 { // TODO: if deck is empty
			return
		}

		cardIdx := gs.Deck[teamIndex][0]
		gs.Deck[teamIndex] = gs.Deck[teamIndex][1:]              // draw top card
		gs.Hand[teamIndex] = append(gs.Hand[teamIndex], cardIdx) // move to hand
		gs.Cards[cardIdx].Zone = state.ZoneHand                  // update card state
	}

	*logList = append(*logList, logs.DrawLog{
		UpdateID:      *UpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.Draw,
		TeamIndex:     teamIndex,
		NumCards:      numCards,
	})
	*UpdateID += 1
}
