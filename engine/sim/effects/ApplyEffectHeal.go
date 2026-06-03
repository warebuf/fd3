package effects

import (
	"math/rand"

	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffectHeal(gs *state.GameState, logList *[]any, UpdateID *int, eff state.Effect, sourceCardID int, executeAtWallTime int64) {
	healerIndex := sourceCardID
	teamIndex := gs.Cards[healerIndex].Team
	amount := eff.B

	friendlyField := gs.Field[teamIndex]
	if len(friendlyField) == 0 {
		return
	}

	targetIdx := friendlyField[rand.Intn(len(friendlyField))]

	gs.Cards[targetIdx].HP += amount
	*logList = append(*logList, logs.HealLog{
		UpdateID:        *UpdateID,
		WallTimestamp:   executeAtWallTime,
		Type:            logs.Heal,
		HealerIndex:     healerIndex,
		TargetCardIndex: targetIdx,
		Amount:          amount,
	})
	*UpdateID++
}
