package effects

import (
	"math"

	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

// heals the source card itself
func ApplyEffectHealSelf(gs *state.GameState, logList *[]any, UpdateID *int, eff state.Effect, sourceCardIndex int, executeAtWallTime int64) {
	amount := eff.B
	if amount <= 0 {
		return
	}
	card := &gs.Cards[sourceCardIndex]

	if amount > math.MaxInt-card.HP {
		card.HP = math.MaxInt
	} else {
		card.HP += amount
	}
	*logList = append(*logList, logs.HealLog{
		UpdateID:        *UpdateID,
		WallTimestamp:   executeAtWallTime,
		Type:            logs.Heal,
		HealerIndex:     sourceCardIndex,
		TargetCardIndex: sourceCardIndex,
		Amount:          amount,
	})
	*UpdateID++
}
