package effects

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffectDraw(gs *state.GameState, logList *[]any, UpdateID *int, eff state.Effect, executeAtWallTime int64) {
	teamIndex := eff.A
	numCards := eff.B

	actualDrawn := 0
	for i := 0; i < numCards; i++ {
		if len(gs.Deck[teamIndex]) == 0 {
			break
		}

		cardIndex := gs.Deck[teamIndex][0]
		gs.Deck[teamIndex] = gs.Deck[teamIndex][1:]                // draw top card
		gs.Hand[teamIndex] = append(gs.Hand[teamIndex], cardIndex) // move to hand
		gs.Cards[cardIndex].Zone = state.ZoneHand                  // update card state
		actualDrawn++
	}

	if actualDrawn == 0 {
		return
	}

	*logList = append(*logList, logs.DrawLog{
		UpdateID:      *UpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.Draw,
		TeamIndex:     teamIndex,
		NumCards:      actualDrawn,
	})
	*UpdateID += 1
}
