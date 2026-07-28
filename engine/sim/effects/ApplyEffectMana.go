package effects

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/conditions/autoability"
	"github.com/warebuf/fd3/engine/sim/reaction"
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffectMana(gs *state.GameState, logList *[]any, UpdateID *int, eff state.Effect, executeAtWallTime int64) {
	gs.Mana[eff.A] += eff.B
	*logList = append(*logList, logs.GainManaLog{
		UpdateID:      *UpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.GainMana,
		TeamIndex:     eff.A,
		Amount:        eff.B,
	})
	*UpdateID += 1

	cardIndexes, activeIndexes := autoability.AutoAbilityManaCheck(gs, eff.A)
	for i := 0; i < len(cardIndexes); i++ {
		reaction.PushAbilityOnStack(gs, logList, UpdateID, cardIndexes[i], activeIndexes[i], executeAtWallTime)
	}
}
