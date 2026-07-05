package effects

import (
	"math"
	"math/rand"

	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffectHeal(gs *state.GameState, logList *[]any, UpdateID *int, eff state.Effect, sourceCardIndex int, executeAtWallTime int64) {
	healerIndex := sourceCardIndex
	teamIndex := gs.Cards[healerIndex].Team
	amount := eff.B

	friendlyField := gs.Field[teamIndex]
	if len(friendlyField) == 0 {
		return
	}

	targetIndex := friendlyField[rand.Intn(len(friendlyField))]

	if amount > math.MaxInt-gs.Cards[targetIndex].HP {
		gs.Cards[targetIndex].HP = math.MaxInt
	} else {
		gs.Cards[targetIndex].HP += amount
	}
	*logList = append(*logList, logs.HealLog{
		UpdateID:        *UpdateID,
		WallTimestamp:   executeAtWallTime,
		Type:            logs.Heal,
		HealerIndex:     healerIndex,
		TargetCardIndex: targetIndex,
		Amount:          amount,
	})
	*UpdateID++
}
