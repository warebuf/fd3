package chain

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/conditions"
	"github.com/warebuf/fd3/engine/state"
)

func PushTrapOnChain(gs *state.GameState, logList *[]any, UpdateID *int, cardIndex int, activeIndex int, executeAtWallTime int64) bool {

	if cardIndex < 0 || cardIndex >= len(gs.Cards) {
		return false
	}

	card := gs.Cards[cardIndex]
	teamIndex := card.Team

	if card.Zone != state.ZoneTrap || !card.Ready {
		return false
	}
	if activeIndex < 0 || activeIndex >= len(card.Actives) {
		return false
	}

	if gs.PausePhase == state.PausePhaseFullPause {
		return false
	}
	if gs.Phase != state.PhasePlay {
		return false
	}

	activeDef := defs.CardDefs[card.DefNo].Active[activeIndex]

	if !activeDef.ChainEntry.CanEnter(gs) {
		return false
	}

	if activeDef.TargetFlags != 0 {
		if !conditions.IsValidTarget(gs, activeDef.TargetFlags, cardIndex, card.Actives[activeIndex].Target, teamIndex) {
			return false
		}
	}

	effs := []state.Effect{}
	for _, eff := range activeDef.Effects {
		if eff.Type == 0 {
			continue // unused Effect slot
		}
		effs = append(effs, state.FulfillEffectTemplate(gs, eff, teamIndex, cardIndex, activeIndex))
	}
	if len(effs) == 0 {
		return false // no real ability on this slot
	}

	// mana check last. the cost sits on the card, not the ability slot, because setting was free
	cost := card.Cost
	if gs.Mana[teamIndex] < cost {
		return false
	}

	gs.Mana[teamIndex] -= cost
	*logList = append(*logList, logs.SpendManaLog{
		UpdateID:      *UpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.SpendMana,
		TeamIndex:     teamIndex,
		Amount:        cost,
	})
	*UpdateID++

	*logList = append(*logList, logs.DeclareCardLog{
		UpdateID:      *UpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.DeclareCard,
		TeamIndex:     teamIndex,
		CardIndex:     cardIndex,
	})
	*UpdateID++

	gs.Cards[cardIndex].Ready = false
	gs.Cards[cardIndex].SetStartAtSimTime = 0
	gs.Cards[cardIndex].ReadyAtSimTime = 0

	startAtActiveTime := executeAtWallTime - gs.FullPauseOffset
	chainItem := state.ChainItem{
		StartAtActiveTime:   startAtActiveTime,
		PopAtActiveTime:     startAtActiveTime + constants.CHAIN_WINDOW - constants.LEAD,
		ExecuteAtActiveTime: startAtActiveTime + constants.CHAIN_WINDOW,
		CardIndex:           cardIndex,
		ActiveIndex:         activeIndex,
		TeamIndex:           teamIndex,
		Type:                state.ChainTrap,
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

	gs.Priority = 1 - teamIndex

	return true
}
