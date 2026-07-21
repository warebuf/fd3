package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/sim/reaction"
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

	if defs.CardDefs[card.DefNo].RequiresStackTarget && len(gs.ReactionStack) == 0 {
		return
	}

	if gs.PausePhase == state.PausePhaseFullPause {
		return
	}

	if gs.PausePhase == state.PausePhaseSimPause {
		oldLogsSize := len(match.EventLog)
		success := reaction.ScheduleDeclareCard(gs, &match.EventLog, &match.NextUpdateID, teamIndex, cardIndex, ie.WallTime.UnixMilli())
		if success {
			scheduleReactionStackResolution(match)
			match.publish(ie.WallTime.UnixMilli()+constants.LEAD, oldLogsSize)
		}
		return
	}

	// simulating: try to play immediately, no queue fallback
	oldLogsSize := len(match.EventLog)
	success := reaction.ScheduleDeclareCard(
		gs,
		&match.EventLog,
		&match.NextUpdateID,
		teamIndex,
		cardIndex,
		ie.WallTime.UnixMilli(),
	)

	if success {
		match.TransitionPauseState(state.PausePhaseSimPause, ie.WallTime.UnixMilli()+constants.LEAD, -1)
		scheduleReactionStackResolution(match)
		match.publish(ie.WallTime.UnixMilli()+constants.LEAD, oldLogsSize)
	}

}
