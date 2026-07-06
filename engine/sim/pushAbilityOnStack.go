package sim

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func pushAbilityOnStack(gs *state.GameState, logList *[]any, UpdateID *int, cardIndex int, activeIndex int, executeAtWallTime int64) {
	card := gs.Cards[cardIndex]
	teamIndex := card.Team
	activeDef := defs.CardDefs[card.DefNo].Active[activeIndex]

	if activeDef.Cost > 0 {
		gs.Mana[teamIndex] -= activeDef.Cost
		*logList = append(*logList, logs.SpendManaLog{
			UpdateID:      *UpdateID,
			WallTimestamp: executeAtWallTime,
			Type:          logs.SpendMana,
			TeamIndex:     teamIndex,
			Amount:        activeDef.Cost,
		})
		*UpdateID++
	}

	reactionItemType := state.ReactionUnitAbility
	if card.Type == state.CardTypeContinuous {
		reactionItemType = state.ReactionContinuousAbility
	}

	var effs []state.Effect
	for i := 0; i < len(activeDef.Effects); i++ {
		eff := activeDef.Effects[i]
		if eff.Type == 0 {
			continue // unused Effect slot
		}
		effs = append(effs, state.FulfillEffectTemplate(eff, teamIndex, cardIndex))
	}

	startAtActiveTime := executeAtWallTime - gs.FullPauseOffset
	reactionItem := state.ReactionItem{
		StartAtActiveTime:   startAtActiveTime,
		PopAtActiveTime:     startAtActiveTime + constants.REACTION_WINDOW - constants.LEAD,
		ExecuteAtActiveTime: startAtActiveTime + constants.REACTION_WINDOW,
		CardIndex:           cardIndex,
		ActiveIndex:         activeIndex,
		Type:                reactionItemType,
		Effects:             effs,
	}
	gs.ReactionStack = append(gs.ReactionStack, reactionItem)
	gs.Cards[cardIndex].Actives[activeIndex].Pending++
	*logList = append(*logList, logs.AddStackLog{
		UpdateID:            *UpdateID,
		WallTimestamp:       executeAtWallTime,
		Type:                logs.AddStack,
		StackType:           int(reactionItem.Type),
		StartAtActiveTime:   reactionItem.StartAtActiveTime,
		PopAtActiveTime:     reactionItem.PopAtActiveTime,
		ExecuteAtActiveTime: reactionItem.ExecuteAtActiveTime,
		CardIndex:           reactionItem.CardIndex,
		Effects:             reactionItem.Effects,
	})
	*UpdateID++
}
