package effects

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/destruction"
	"github.com/warebuf/fd3/engine/state"
)

// takes control of a field unit the source card's controller originally owned (Cons already gates
// this at activation time via ConditionTargetOwnedBySelf; this is a defensive re-check for state
// that may have changed during the chain window), then forces its Active[0] ability to resolve for
// free. Charges/PendingCharges are pre-compensated here so PQHandlerChain's normal per-resolution
// decrement (which fires unconditionally for any ChainUnitAbility item) nets to zero cost. If that
// ability needs its own target and none was ever preselected (a freshly repossessed unit can't have
// one), the effects that depend on it fizzle or fall back, same as any other unset-target effect.
func ApplyEffectRepossess(gs *state.GameState, logList *[]any, UpdateID *int, eff state.Effect, sourceCardIndex int, executeAtWallTime int64) {
	targetIndex := eff.A
	toTeam := gs.Cards[sourceCardIndex].Team

	if targetIndex < 0 || targetIndex >= len(gs.Cards) {
		return // preselected target was never set, ability fizzles
	}
	card := &gs.Cards[targetIndex]
	fromTeam := card.Team
	if card.Zone != state.ZoneField || fromTeam == toTeam || card.OwnerTeam != toTeam {
		return // target left the field, is already yours, or was never yours to begin with
	}

	moved := false
	for i, ci := range gs.Field[fromTeam] {
		if ci == targetIndex {
			gs.Field[fromTeam] = append(gs.Field[fromTeam][:i], gs.Field[fromTeam][i+1:]...)
			moved = true
			break
		}
	}
	if !moved {
		return
	}

	gs.Field[toTeam] = append(gs.Field[toTeam], targetIndex)
	card.Team = toTeam

	// the old controller's targeting and auto-fire choices don't carry over to the new one
	for i := range card.Actives {
		card.Actives[i].Target = -1
		card.Actives[i].TargetStructure = -1
		card.Actives[i].AutoAbility = false
	}
	destruction.ClearPreselectsTargeting(gs, targetIndex)

	*logList = append(*logList, logs.GiveControlLog{
		UpdateID:      *UpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.GiveControl,
		CardIndex:     targetIndex,
		FromTeamIndex: fromTeam,
		ToTeamIndex:   toTeam,
	})
	*UpdateID++

	activeDef := defs.CardDefs[card.DefNo].Active[0]
	effs := []state.Effect{}
	for _, e := range activeDef.Effects {
		if e.Type == 0 {
			continue
		}
		effs = append(effs, state.FulfillEffectTemplate(gs, e, toTeam, targetIndex, 0))
	}
	if len(effs) == 0 {
		return // this unit has no Active[0] ability to force; control still changed hands above
	}

	card.PendingCharges++
	if !activeDef.Unlimited {
		card.Charges++ // matches the unconditional decrement PQHandlerChain applies on resolution, so this nets to free
	}

	startAtActiveTime := executeAtWallTime - gs.FullPauseOffset
	chainItem := state.ChainItem{
		StartAtActiveTime:   startAtActiveTime,
		PopAtActiveTime:     startAtActiveTime + constants.CHAIN_WINDOW - constants.LEAD,
		ExecuteAtActiveTime: startAtActiveTime + constants.CHAIN_WINDOW,
		CardIndex:           targetIndex,
		ActiveIndex:         0,
		TeamIndex:           toTeam,
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
}
