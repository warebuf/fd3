package match

import (
	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/match/eventpq"
)

func (match *Match) handleRevealEvent(ev *eventpq.RevealEvent) {

	wallTime := ev.WallTime.UnixMilli() + constants.LEAD

	if len(match.RevealQueue) == 0 {
		match.PendingRevealTime = 0
		PushPhase(match, wallTime-match.GameState.FullPauseOffset+2000+constants.REACTION_WINDOW, 4)
		return
	}

	next := match.RevealQueue[0]
	match.RevealQueue = match.RevealQueue[1:]

	oldLogsSize := len(match.EventLog)
	chainStarted := ScheduleDeclareCard(match.GameState, &match.SimulationPQ, &match.AbilityIDCounter, &match.EventLog, &match.NextUpdateID, next.TeamIdx, next.CardIndex, match.Phase3StartTime)
	if chainStarted {
		scheduleChainResolution(match)
	} else {
		pushRevealEvent(match, wallTime-match.GameState.FullPauseOffset+800)
	}
	match.publish(wallTime, oldLogsSize)

}
