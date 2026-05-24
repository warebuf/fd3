package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

func (match *Match) handleRevealEvent(ev *eventpq.RevealEvent) {

	wallTime := ev.WallTime.UnixMilli() + constants.LEAD

	if len(match.RevealQueue) == 0 {
		match.PendingRevealTime = 0
		PushPhaseEvent(match, wallTime-match.GameState.FullPauseOffset+5000, state.PhasePlay)
		return
	}

	next := match.RevealQueue[0]
	match.RevealQueue = match.RevealQueue[1:]

	oldLogsSize := len(match.EventLog)
	chainStarted := ScheduleDeclareCard(
		match.GameState,
		&match.SimulationPQ,
		&match.AbilityIDCounter,
		&match.EventLog,
		&match.NextUpdateID,
		next.TeamIdx,
		next.CardIndex,
		ev.WallTime.UnixMilli(),
	)
	if chainStarted {
		// no need enter Simulation Pause (PausePhase 1) here, should already be in it
		scheduleChainResolution(match)
	} else {
		pushRevealEvent(match, wallTime-match.GameState.FullPauseOffset+constants.REACTION_WINDOW) // make a reveal same amount of time as 1 chain
	}
	match.publish(wallTime, oldLogsSize)

}
