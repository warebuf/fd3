package chain

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func PushTriggerEffectsOnChain(gs *state.GameState, logList *[]any, UpdateID *int, teamIndex int, cardIndex int, triggerIndex int, executeAtWallTime int64) {
	ability := defs.CardDefs[gs.Cards[cardIndex].DefNo].OnTrigger[triggerIndex]

	effs := []state.Effect{}
	for _, eff := range ability.Effects {
		if eff.Type == 0 {
			continue // unused Effect slot
		}
		effs = append(effs, state.FulfillEffectTemplate(gs, eff, teamIndex, cardIndex, -1))
	}
	if len(effs) == 0 {
		return
	}

	startAtActiveTime := executeAtWallTime - gs.FullPauseOffset
	chainItem := state.ChainItem{
		StartAtActiveTime:   startAtActiveTime,
		PopAtActiveTime:     startAtActiveTime + constants.CHAIN_WINDOW - constants.LEAD,
		ExecuteAtActiveTime: startAtActiveTime + constants.CHAIN_WINDOW,
		CardIndex:           cardIndex,
		ActiveIndex:         -1, // OnTrigger abilities aren't Active-slot-sourced
		TeamIndex:           teamIndex,
		Type:                state.ChainUnitAbility,
		Effects:             effs,
	}
	gs.Chain = append(gs.Chain, chainItem)

	*logList = append(*logList, logs.AddChainLog{
		UpdateID:            *UpdateID,
		WallTimestamp:       executeAtWallTime,
		Type:                logs.AddChain,
		ChainType:           int(chainItem.Type),
		StartAtActiveTime:   chainItem.StartAtActiveTime,
		PopAtActiveTime:     chainItem.PopAtActiveTime,
		ExecuteAtActiveTime: chainItem.ExecuteAtActiveTime,
		CardIndex:           chainItem.CardIndex,
		ActiveIndex:         chainItem.ActiveIndex,
		TeamIndex:           chainItem.TeamIndex,
		Effects:             chainItem.Effects,
	})
	*UpdateID++
}
