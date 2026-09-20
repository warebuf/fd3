package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/sim/abilityslot"
	"github.com/warebuf/fd3/engine/sim/chain"
	"github.com/warebuf/fd3/engine/sim/conditions"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

func (match *Match) handleTrapAbility(ie *eventpq.InputEvent) {

	gs := match.GameState

	if gs.Phase != state.PhasePlay {
		return
	}
	if gs.PausePhase == state.PausePhaseFullPause {
		return
	}

	cardIndex := ie.CardIndex
	teamIndex := ie.TeamIndex
	if cardIndex < 0 || cardIndex >= len(gs.Cards) {
		return
	}

	card := gs.Cards[cardIndex]
	if card.Zone != state.ZoneTrap || card.Team != teamIndex {
		return
	}
	if !card.Ready {
		return
	}

	if defs.CardDefs[card.DefNo].IsAuction {
		if gs.Mana[teamIndex] < card.Cost || gs.StandoffCardIndex != -1 {
			return // can't afford it, or an auction is already open
		}
		match.openStandoffAuction(cardIndex, teamIndex, ie.WallTime.UnixMilli()+constants.LEAD)
		return
	}

	activeIndex := abilityslot.GetActiveSlot(card.DefNo, ie.AbilityIndex)
	if activeIndex < 0 {
		return
	}
	activeDef := defs.CardDefs[card.DefNo].Active[activeIndex]

	if gs.Mana[teamIndex] < card.Cost {
		return
	}

	// check state conditions
	if !conditions.CheckActiveAbilityConditions(gs, cardIndex, activeIndex) {
		return
	}

	// check target conditions
	if activeDef.TargetFlags != 0 && !conditions.IsValidTarget(gs, activeDef.TargetFlags, cardIndex, card.Actives[activeIndex].Target, teamIndex) {
		return
	}

	// check action/reaction/dual condition
	if !activeDef.ChainEntry.CanEnter(gs) {
		return
	}

	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	oldLogsSize := len(match.EventLog)

	if !chain.PushTrapOnChain(gs, &match.EventLog, &match.NextUpdateID, cardIndex, activeIndex, executeAtWallTime) {
		return
	}

	match.clearChainSkips(executeAtWallTime)
	match.TransitionPauseState(state.PausePhaseSimPause, executeAtWallTime, -1)
	scheduleChainResolution(match)
	match.publish(executeAtWallTime, oldLogsSize)

}

func (match *Match) handleUnitAbility(ie *eventpq.InputEvent) {

	gs := match.GameState

	if gs.Phase != state.PhasePlay {
		return
	}
	if gs.PausePhase == state.PausePhaseFullPause {
		return
	}

	cardIndex := ie.CardIndex
	teamIndex := ie.TeamIndex
	if cardIndex < 0 || cardIndex >= len(gs.Cards) {
		return
	}

	card := gs.Cards[cardIndex]
	if card.Zone != state.ZoneField || card.Team != teamIndex {
		return
	}

	// HARDCODE: ALWAYS TARGETS SLOT 0
	const activeIndex = 0
	activeDef := defs.CardDefs[card.DefNo].Active[activeIndex]
	if card.PendingCharges >= card.Charges {
		return
	}

	if gs.Mana[teamIndex] < activeDef.Cost {
		return
	}

	if !conditions.CheckActiveAbilityConditions(gs, cardIndex, activeIndex) {
		return
	}

	if !activeDef.ChainEntry.CanEnter(gs) {
		return
	}

	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	oldLogsSize := len(match.EventLog)

	if !chain.PushAbilityOnChain(gs, &match.EventLog, &match.NextUpdateID, cardIndex, activeIndex, executeAtWallTime) {
		return
	}

	match.clearChainSkips(executeAtWallTime)
	match.TransitionPauseState(state.PausePhaseSimPause, executeAtWallTime, -1)
	scheduleChainResolution(match)
	match.publish(executeAtWallTime, oldLogsSize)

}
