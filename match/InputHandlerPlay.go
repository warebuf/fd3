package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/sim/chain"
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

	if defs.CardDefs[card.DefNo].RequiresChainTarget && len(gs.Chain) == 0 {
		return
	}

	if gs.PausePhase == state.PausePhaseFullPause {
		return
	}

	if gs.PausePhase == state.PausePhaseSimPause {
		// hold the chain item hidden until the window's timer reveals it (see handleChainResolution)
		match.PendingChains[teamIndex] = PendingChain{
			Type:        PendingChainCard,
			CardIndex:   cardIndex,
			ActiveIndex: -1,
		}
		return
	}

	// simulating: try to play immediately, no queue fallback
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
		match.TransitionPauseState(state.PausePhaseSimPause, ie.WallTime.UnixMilli()+constants.LEAD, -1)
		scheduleChainResolution(match)
		match.publish(ie.WallTime.UnixMilli()+constants.LEAD, oldLogsSize)
	}

}
