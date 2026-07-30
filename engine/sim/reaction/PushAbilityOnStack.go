package reaction

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func PushAbilityOnStack(gs *state.GameState, logList *[]any, UpdateID *int, cardIndex int, activeIndex int, executeAtWallTime int64) bool {
	card := gs.Cards[cardIndex]
	teamIndex := card.Team
	activeDef := defs.CardDefs[card.DefNo].Active[activeIndex]

	effs := []state.Effect{}
	for i := 0; i < len(activeDef.Effects); i++ {
		eff := activeDef.Effects[i]
		if eff.Type == 0 {
			continue // unused Effect slot
		}
		effs = append(effs, state.FulfillEffectTemplate(gs, eff, teamIndex, cardIndex, activeIndex))
	}
	if len(effs) == 0 {
		return false // no real ability on this slot
	}

	for _, eff := range effs {
		if eff.Type != state.EffectSacrifice {
			continue
		}
		validTarget := false
		for _, fieldCardIndex := range gs.Field[teamIndex] {
			if fieldCardIndex == eff.A {
				validTarget = true
				break
			}
		}
		if !validTarget {
			return false // sacrifice target is no longer on the field
		}
	}

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

	startAtActiveTime := executeAtWallTime - gs.FullPauseOffset
	reactionItem := state.ReactionItem{
		StartAtActiveTime:   startAtActiveTime,
		PopAtActiveTime:     startAtActiveTime + constants.REACTION_WINDOW - constants.LEAD,
		ExecuteAtActiveTime: startAtActiveTime + constants.REACTION_WINDOW,
		CardIndex:           cardIndex,
		ActiveIndex:         activeIndex,
		TeamIndex:           teamIndex,
		Type:                reactionItemType,
		Effects:             effs,
	}
	gs.ReactionStack = append(gs.ReactionStack, reactionItem)
	gs.Cards[cardIndex].PendingCharges++
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
		TeamIndex:           reactionItem.TeamIndex,
		Effects:             reactionItem.Effects,
	})
	*UpdateID++
	return true
}
