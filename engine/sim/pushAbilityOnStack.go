package sim

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

// NOTE: im not sure where this belongs, maybe not sim
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

	// NOTE: THIS TYPE SHOULD NOT COME FROM ZONE, IT SHOULD COME FROM CARD TYPE
	stackItemType := state.StackItemAbility
	if card.Zone == state.ZoneContinuous {
		stackItemType = state.StackItemContinuousActivation
	}

	var effs []state.Effect
	for i := 0; i < len(activeDef.Effects); i++ {
		eff := activeDef.Effects[i]
		if eff.Type == 0 {
			continue // unused Effect slot
		}
		effs = append(effs, FulfillEffectTemplate(eff, teamIndex, cardIndex))
	}

	startAtActiveTime := executeAtWallTime - gs.FullPauseOffset
	stackItem := state.StackItem{
		StartAtActiveTime:   startAtActiveTime,
		PopAtActiveTime:     startAtActiveTime + constants.REACTION_WINDOW - constants.LEAD,
		ExecuteAtActiveTime: startAtActiveTime + constants.REACTION_WINDOW,
		CardIndex:           cardIndex,
		ActiveIndex:         activeIndex,
		Type:                stackItemType,
		Effects:             effs,
	}
	gs.Stack = append(gs.Stack, stackItem)
	gs.Cards[cardIndex].Actives[activeIndex].Pending++
	*logList = append(*logList, logs.AddStackLog{
		UpdateID:            *UpdateID,
		WallTimestamp:       executeAtWallTime,
		Type:                logs.AddStack,
		StackType:           stackItem.Type,
		StartAtActiveTime:   stackItem.StartAtActiveTime,
		PopAtActiveTime:     stackItem.PopAtActiveTime,
		ExecuteAtActiveTime: stackItem.ExecuteAtActiveTime,
		CardIndex:           stackItem.CardIndex,
		Effects:             stackItem.Effects,
	})
	*UpdateID++
}
