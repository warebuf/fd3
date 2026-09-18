package chain

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func PushOnTickAbilityOnChain(gs *state.GameState, logList *[]any, UpdateID *int, cardIndex int, abilityIndex int, executeAtWallTime int64) bool {
	card := gs.Cards[cardIndex]
	teamIndex := card.Team
	ability := defs.CardDefs[card.DefNo].OnTick[abilityIndex]

	if !ability.ChainEntry.CanEnter(gs) {
		return false
	}

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
	chainItem := state.ChainItem{
		StartAtActiveTime:   startAtActiveTime,
		PopAtActiveTime:     startAtActiveTime + constants.CHAIN_WINDOW - constants.LEAD,
		ExecuteAtActiveTime: startAtActiveTime + constants.CHAIN_WINDOW,
		CardIndex:           cardIndex,
		ActiveIndex:         -1, // OnTick abilities aren't Active-slot-sourced
		TeamIndex:           teamIndex,
		StructureIndex:      -1,
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
		StructureIndex:      chainItem.StructureIndex,
		Effects:             chainItem.Effects,
	})
	*UpdateID++
	return true
}
