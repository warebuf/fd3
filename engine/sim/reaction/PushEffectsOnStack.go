package reaction

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

// ActiveIndex is -1: neither OnTick nor OnTrigger abilities are Active-slot-sourced.
func PushEffectsOnStack(gs *state.GameState, logList *[]any, UpdateID *int, cardIndex int, effects []state.Effect, executeAtWallTime int64) bool {
	if len(effects) == 0 {
		return false
	}

	startAtActiveTime := executeAtWallTime - gs.FullPauseOffset
	reactionItem := state.ReactionItem{
		StartAtActiveTime:   startAtActiveTime,
		PopAtActiveTime:     startAtActiveTime + constants.REACTION_WINDOW - constants.LEAD,
		ExecuteAtActiveTime: startAtActiveTime + constants.REACTION_WINDOW,
		CardIndex:           cardIndex,
		ActiveIndex:         -1,
		TeamIndex:           gs.Cards[cardIndex].Team,
		Type:                state.ReactionUnitAbility,
		Effects:             effects,
	}
	gs.ReactionStack = append(gs.ReactionStack, reactionItem)

	*logList = append(*logList, logs.AddStackLog{
		UpdateID:            *UpdateID,
		WallTimestamp:       executeAtWallTime,
		Type:                logs.AddStack,
		StackType:           int(reactionItem.Type),
		StartAtActiveTime:   reactionItem.StartAtActiveTime,
		PopAtActiveTime:     reactionItem.PopAtActiveTime,
		ExecuteAtActiveTime: reactionItem.ExecuteAtActiveTime,
		CardIndex:           reactionItem.CardIndex,
		ActiveIndex:         reactionItem.ActiveIndex,
		Effects:             reactionItem.Effects,
	})
	*UpdateID++
	return true
}
