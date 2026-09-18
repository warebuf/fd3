package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/sim/chain"
	"github.com/warebuf/fd3/engine/sim/direct"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

func (match *Match) handleInstantPlay(ie *eventpq.InputEvent) {

	gs := match.GameState

	if gs.Phase != state.PhasePlay {
		return
	}

	cardIndex := ie.CardIndex
	teamIndex := ie.TeamIndex
	if cardIndex < 0 || cardIndex >= len(gs.Cards) {
		return
	}

	card := gs.Cards[cardIndex]
	if card.Zone != state.ZoneHand || card.Team != teamIndex {
		return
	}

	// traps are set rather than played, so CardDef.ChainEntry never applies to them
	if card.Type == state.CardTypeTrap {
		match.handleSetTrap(ie, cardIndex, teamIndex)
		return
	}

	if card.Type == state.CardTypeEquip {
		gs.Cards[cardIndex].EquipHostIndex = state.BodyguardRedirect(gs, ie.TargetCardIndex)
	}

	if !defs.CardDefs[card.DefNo].ChainEntry.CanEnter(gs) {
		return
	}

	if gs.PausePhase == state.PausePhaseFullPause {
		return
	}

	// push straight onto the chain, first come first served
	oldLogsSize := len(match.EventLog)
	success := chain.PushCardOnChain(
		gs,
		&match.EventLog,
		&match.NextUpdateID,
		teamIndex,
		cardIndex,
		ie.WallTime.UnixMilli(),
	)

	if success {
		match.clearChainSkips(ie.WallTime.UnixMilli() + constants.LEAD)
		match.TransitionPauseState(state.PausePhaseSimPause, ie.WallTime.UnixMilli()+constants.LEAD, -1)
		scheduleChainResolution(match)
		match.publish(ie.WallTime.UnixMilli()+constants.LEAD, oldLogsSize)
	}

}

// setting bypasses the chain, so there's no window to park it in - refused unless simulating
func (match *Match) handleSetTrap(ie *eventpq.InputEvent, cardIndex int, teamIndex int) {

	gs := match.GameState

	if gs.PausePhase != state.PausePhaseSimulating {
		return
	}

	executeAtWallTime := ie.WallTime.UnixMilli() + constants.LEAD
	oldLogsSize := len(match.EventLog)

	if !direct.SetTrap(gs, &match.SimulationPQ, &match.EventLog, &match.NextUpdateID, teamIndex, cardIndex, executeAtWallTime) {
		return
	}

	scheduleEarliestSimItem(match)
	match.publish(executeAtWallTime, oldLogsSize)

}
