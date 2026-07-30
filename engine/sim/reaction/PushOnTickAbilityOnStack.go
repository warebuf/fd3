package reaction

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func PushOnTickAbilityOnStack(gs *state.GameState, logList *[]any, UpdateID *int, cardIndex int, abilityIndex int, executeAtWallTime int64) bool {
	card := gs.Cards[cardIndex]
	teamIndex := card.Team
	ability := defs.CardDefs[card.DefNo].OnTick[abilityIndex]

	effs := []state.Effect{}
	for _, eff := range ability.Effects {
		if eff.Type == 0 {
			continue // unused Effect slot
		}
		effs = append(effs, state.FulfillEffectTemplate(gs, eff, teamIndex, cardIndex, -1))
	}
	if len(effs) == 0 {
		return false
	}

	startAtActiveTime := executeAtWallTime - gs.FullPauseOffset
	reactionItem := state.ReactionItem{
		StartAtActiveTime:   startAtActiveTime,
		PopAtActiveTime:     startAtActiveTime + constants.REACTION_WINDOW - constants.LEAD,
		ExecuteAtActiveTime: startAtActiveTime + constants.REACTION_WINDOW,
		CardIndex:           cardIndex,
		ActiveIndex:         -1, // OnTick abilities aren't Active-slot-sourced
		TeamIndex:           teamIndex,
		Type:                state.ReactionUnitAbility,
		Effects:             effs,
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
