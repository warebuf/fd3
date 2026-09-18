package chain

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/conditions"
	"github.com/warebuf/fd3/engine/state"
)

func PushCardOnChain(gs *state.GameState, logList *[]any, UpdateID *int, teamIndex int, cardIndex int, wallTimeAtPop int64) bool {

	if cardIndex < 0 || cardIndex >= len(gs.Cards) {
		return false
	}

	if gs.Cards[cardIndex].Zone != state.ZoneHand {
		return false
	}
	handIndex := -1
	for i := 0; i < len(gs.Hand[teamIndex]); i++ {
		if cardIndex == gs.Hand[teamIndex][i] {
			handIndex = i
		}
	}
	if handIndex < 0 {
		return false
	}

	if gs.PausePhase == state.PausePhaseFullPause {
		return false
	}

	if gs.Phase != state.PhasePlay {
		return false
	}

	if !defs.CardDefs[gs.Cards[cardIndex].DefNo].ChainEntry.CanEnter(gs) {
		return false
	}

	for i, ability := range defs.CardDefs[gs.Cards[cardIndex].DefNo].OnPlay {
		if ability.TargetFlags == 0 {
			continue // unused ability slot, or an ability that doesn't target anything
		}
		if !conditions.IsValidTarget(gs, ability.TargetFlags, cardIndex, gs.Cards[cardIndex].OnPlay[i].Target, teamIndex) {
			return false
		}
	}

	if gs.Cards[cardIndex].Type == state.CardTypeEquip {
		equipTargetFlags := defs.CardDefs[gs.Cards[cardIndex].DefNo].EquipTargetFlags
		if !conditions.IsValidTarget(gs, equipTargetFlags, cardIndex, gs.Cards[cardIndex].EquipHostIndex, teamIndex) {
			return false
		}
	}

	var chainType state.ChainType
	switch gs.Cards[cardIndex].Type {
	case state.CardTypeUnit:
		chainType = state.ChainSummon
	case state.CardTypeContinuous:
		chainType = state.ChainContinuousResolution
	case state.CardTypeEquip:
		chainType = state.ChainEquip
	case state.CardTypeTrap:
		return false
	default:
		return false
	}

	// mana check last
	cost := gs.Cards[cardIndex].Cost
	if gs.Mana[teamIndex] < cost {
		return false
	}

	gs.Mana[teamIndex] -= cost
	*logList = append(*logList, logs.SpendManaLog{
		UpdateID:      *UpdateID,
		WallTimestamp: wallTimeAtPop + constants.LEAD,
		Type:          logs.SpendMana,
		TeamIndex:     teamIndex,
		Amount:        cost,
	})
	*UpdateID += 1
	fireOpponentTrigger(gs, logList, UpdateID, teamIndex, state.TriggerEnemyPaidMana, wallTimeAtPop+constants.LEAD)

	*logList = append(*logList, logs.DeclareCardLog{
		UpdateID:      *UpdateID,
		WallTimestamp: wallTimeAtPop + constants.LEAD,
		Type:          logs.DeclareCard,
		TeamIndex:     teamIndex,
		CardIndex:     cardIndex,
	})
	*UpdateID += 1

	gs.Cards[cardIndex].AutoSummon = 0

	gs.Cards[cardIndex].Zone = state.ZonePending
	gs.Hand[teamIndex] = append(gs.Hand[teamIndex][:handIndex], gs.Hand[teamIndex][handIndex+1:]...)

	declarationTime := wallTimeAtPop + constants.LEAD
	card := gs.Cards[cardIndex]

	effs := []state.Effect{}
	for i, ability := range defs.CardDefs[card.DefNo].OnPlay {
		for _, eff := range ability.Effects {
			if eff.Type == 0 {
				continue // unused Effect slot
			}
			effs = append(effs, state.FulfillEffectTemplate(gs, eff, teamIndex, cardIndex, i))
		}
	}

	startAtActiveTime := declarationTime - gs.FullPauseOffset
	chainItem := &state.ChainItem{
		StartAtActiveTime:   startAtActiveTime,
		PopAtActiveTime:     startAtActiveTime + constants.CHAIN_WINDOW - constants.LEAD,
		ExecuteAtActiveTime: startAtActiveTime + constants.CHAIN_WINDOW,
		CardIndex:           cardIndex,
		ActiveIndex:         -1, // OnPlay effects aren't Active-ability-sourced
		TeamIndex:           teamIndex,
		StructureIndex:      -1,
		Type:                chainType,
		Effects:             effs,
	}
	gs.Chain = append(gs.Chain, *chainItem)

	*logList = append(*logList, logs.AddChainLog{
		UpdateID:            *UpdateID,
		WallTimestamp:       declarationTime,
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
	*UpdateID += 1

	return true
}
