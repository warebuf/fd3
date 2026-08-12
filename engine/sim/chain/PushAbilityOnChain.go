package chain

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

func PushAbilityOnChain(gs *state.GameState, logList *[]any, UpdateID *int, cardIndex int, activeIndex int, executeAtWallTime int64) bool {
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

	chainItemType := state.ChainUnitAbility
	if card.Type == state.CardTypeContinuous {
		chainItemType = state.ChainContinuousAbility
	}

	startAtActiveTime := executeAtWallTime - gs.FullPauseOffset
	chainItem := state.ChainItem{
		StartAtActiveTime:   startAtActiveTime,
		PopAtActiveTime:     startAtActiveTime + constants.CHAIN_WINDOW - constants.LEAD,
		ExecuteAtActiveTime: startAtActiveTime + constants.CHAIN_WINDOW,
		CardIndex:           cardIndex,
		ActiveIndex:         activeIndex,
		TeamIndex:           teamIndex,
		Type:                chainItemType,
		Effects:             effs,
	}
	gs.Chain = append(gs.Chain, chainItem)
	gs.Cards[cardIndex].PendingCharges++
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

	gs.Priority = 1 - teamIndex

	return true
}
