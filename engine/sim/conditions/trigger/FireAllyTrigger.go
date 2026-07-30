package trigger

import (
	"github.com/warebuf/fd3/engine/sim/reaction"
	"github.com/warebuf/fd3/engine/state"
)

func FireAllyTrigger(gs *state.GameState, logList *[]any, UpdateID *int, teamIndex int, trigger state.TriggerType, executeAtWallTime int64) {
	cardIndexes, triggerIndexes := CheckAllyTriggers(gs, teamIndex, trigger)

	for i := range cardIndexes {
		reaction.PushTriggerEffectsOnStack(gs, logList, UpdateID, teamIndex, cardIndexes[i], triggerIndexes[i], executeAtWallTime)
	}
}
